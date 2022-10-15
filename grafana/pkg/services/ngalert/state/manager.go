package state

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/ngalert/eval"
	"github.com/grafana/grafana/pkg/services/ngalert/image"
	"github.com/grafana/grafana/pkg/services/ngalert/metrics"
	ngModels "github.com/grafana/grafana/pkg/services/ngalert/models"

	"github.com/grafana/grafana/pkg/services/screenshot"
)

var ResendDelay = 30 * time.Second

// AlertInstanceManager defines the interface for querying the current alert instances.
type AlertInstanceManager interface {
	GetAll(orgID int64) []*State
	GetStatesForRuleUID(orgID int64, alertRuleUID string) []*State
}

type Manager struct {
	log     log.Logger
	metrics *metrics.State

	clock       clock.Clock
	cache       *cache
	quit        chan struct{}
	ResendDelay time.Duration

	ruleStore     RuleReader
	instanceStore InstanceStore
	imageService  image.ImageService
	historian     Historian
	externalURL   *url.URL
}

func NewManager(logger log.Logger, metrics *metrics.State, externalURL *url.URL,
	ruleStore RuleReader, instanceStore InstanceStore, imageService image.ImageService, clock clock.Clock, historian Historian) *Manager {
	manager := &Manager{
		cache:         newCache(),
		quit:          make(chan struct{}),
		ResendDelay:   ResendDelay, // TODO: make this configurable
		log:           logger,
		metrics:       metrics,
		ruleStore:     ruleStore,
		instanceStore: instanceStore,
		imageService:  imageService,
		historian:     historian,
		clock:         clock,
		externalURL:   externalURL,
	}
	go manager.recordMetrics()
	return manager
}

func (st *Manager) Close() {
	st.quit <- struct{}{}
}

func (st *Manager) Warm(ctx context.Context) {
	startTime := time.Now()
	st.log.Info("Warming state cache for startup")

	orgIds, err := st.instanceStore.FetchOrgIds(ctx)
	if err != nil {
		st.log.Error("unable to fetch orgIds", "err", err.Error())
	}

	statesCount := 0
	states := make(map[int64]map[string]*ruleStates, len(orgIds))
	for _, orgId := range orgIds {
		// Get Rules
		ruleCmd := ngModels.ListAlertRulesQuery{
			OrgID: orgId,
		}
		if err := st.ruleStore.ListAlertRules(ctx, &ruleCmd); err != nil {
			st.log.Error("unable to fetch previous state", "msg", err.Error())
		}

		ruleByUID := make(map[string]*ngModels.AlertRule, len(ruleCmd.Result))
		for _, rule := range ruleCmd.Result {
			ruleByUID[rule.UID] = rule
		}

		orgStates := make(map[string]*ruleStates, len(ruleByUID))
		states[orgId] = orgStates

		// Get Instances
		cmd := ngModels.ListAlertInstancesQuery{
			RuleOrgID: orgId,
		}
		if err := st.instanceStore.ListAlertInstances(ctx, &cmd); err != nil {
			st.log.Error("unable to fetch previous state", "msg", err.Error())
		}

		for _, entry := range cmd.Result {
			ruleForEntry, ok := ruleByUID[entry.RuleUID]
			if !ok {
				// TODO Should we delete the orphaned state from the db?
				continue
			}

			rulesStates, ok := orgStates[entry.RuleUID]
			if !ok {
				rulesStates = &ruleStates{states: make(map[string]*State)}
				orgStates[entry.RuleUID] = rulesStates
			}

			lbs := map[string]string(entry.Labels)
			cacheID, err := entry.Labels.StringKey()
			if err != nil {
				st.log.Error("error getting cacheId for entry", "msg", err.Error())
			}
			rulesStates.states[cacheID] = &State{
				AlertRuleUID:         entry.RuleUID,
				OrgID:                entry.RuleOrgID,
				CacheID:              cacheID,
				Labels:               lbs,
				State:                translateInstanceState(entry.CurrentState),
				StateReason:          entry.CurrentReason,
				LastEvaluationString: "",
				StartsAt:             entry.CurrentStateSince,
				EndsAt:               entry.CurrentStateEnd,
				LastEvaluationTime:   entry.LastEvalTime,
				Annotations:          ruleForEntry.Annotations,
			}
			statesCount++
		}
	}
	st.cache.setAllStates(states)
	st.log.Info("State cache has been initialized", "loaded_states", statesCount, "duration", time.Since(startTime))
}

func (st *Manager) Get(orgID int64, alertRuleUID, stateId string) *State {
	return st.cache.get(orgID, alertRuleUID, stateId)
}

// ResetStateByRuleUID deletes all entries in the state manager that match the given rule UID.
func (st *Manager) ResetStateByRuleUID(ctx context.Context, ruleKey ngModels.AlertRuleKey) []*State {
	logger := st.log.New(ruleKey.LogContext()...)
	logger.Debug("resetting state of the rule")
	states := st.cache.removeByRuleUID(ruleKey.OrgID, ruleKey.UID)
	if len(states) > 0 {
		err := st.instanceStore.DeleteAlertInstancesByRule(ctx, ruleKey)
		if err != nil {
			logger.Error("failed to delete states that belong to a rule from database", ruleKey.LogContext()...)
		}
	}
	logger.Info("rules state was reset", "deleted_states", len(states))
	return states
}

// ProcessEvalResults updates the current states that belong to a rule with the evaluation results.
// if extraLabels is not empty, those labels will be added to every state. The extraLabels take precedence over rule labels and result labels
func (st *Manager) ProcessEvalResults(ctx context.Context, evaluatedAt time.Time, alertRule *ngModels.AlertRule, results eval.Results, extraLabels data.Labels) []*State {
	logger := st.log.New(alertRule.GetKey().LogContext()...)
	logger.Debug("state manager processing evaluation results", "resultCount", len(results))
	var states []*State
	processedResults := make(map[string]*State, len(results))
	for _, result := range results {
		s := st.setNextState(ctx, alertRule, result, extraLabels)
		states = append(states, s)
		processedResults[s.CacheID] = s
	}
	resolvedStates := st.staleResultsHandler(ctx, evaluatedAt, alertRule, processedResults)
	if len(states) > 0 {
		logger.Debug("saving new states to the database", "count", len(states))
		_, _ = st.saveAlertStates(ctx, states...)
	}
	return append(states, resolvedStates...)
}

// Maybe take a screenshot. Do it if:
// 1. The alert state is transitioning into the "Alerting" state from something else.
// 2. The alert state has just transitioned to the resolved state.
// 3. The state is alerting and there is no screenshot annotation on the alert state.
func (st *Manager) maybeTakeScreenshot(
	ctx context.Context,
	alertRule *ngModels.AlertRule,
	state *State,
	oldState eval.State,
) error {
	shouldScreenshot := state.Resolved ||
		state.State == eval.Alerting && oldState != eval.Alerting ||
		state.State == eval.Alerting && state.Image == nil
	if !shouldScreenshot {
		return nil
	}

	img, err := st.imageService.NewImage(ctx, alertRule)
	if err != nil &&
		errors.Is(err, screenshot.ErrScreenshotsUnavailable) ||
		errors.Is(err, image.ErrNoDashboard) ||
		errors.Is(err, image.ErrNoPanel) {
		// It's not an error if screenshots are disabled, or our rule isn't allowed to generate screenshots.
		return nil
	} else if err != nil {
		return err
	}
	state.Image = img
	return nil
}

// Set the current state based on evaluation results
func (st *Manager) setNextState(ctx context.Context, alertRule *ngModels.AlertRule, result eval.Result, extraLabels data.Labels) *State {
	currentState := st.cache.getOrCreate(ctx, st.log, alertRule, result, extraLabels, st.externalURL)

	currentState.LastEvaluationTime = result.EvaluatedAt
	currentState.EvaluationDuration = result.EvaluationDuration
	currentState.Results = append(currentState.Results, Evaluation{
		EvaluationTime:  result.EvaluatedAt,
		EvaluationState: result.State,
		Values:          NewEvaluationValues(result.Values),
		Condition:       alertRule.Condition,
	})
	currentState.LastEvaluationString = result.EvaluationString
	currentState.TrimResults(alertRule)
	oldState := currentState.State
	oldReason := currentState.StateReason

	st.log.Debug("setting alert state", "uid", alertRule.UID)
	switch result.State {
	case eval.Normal:
		currentState.resultNormal(alertRule, result)
	case eval.Alerting:
		currentState.resultAlerting(alertRule, result)
	case eval.Error:
		currentState.resultError(alertRule, result)
	case eval.NoData:
		currentState.resultNoData(alertRule, result)
	case eval.Pending: // we do not emit results with this state
	}

	// Set reason iff: result is different than state, reason is not Alerting or Normal
	currentState.StateReason = ""

	if currentState.State != result.State &&
		result.State != eval.Normal &&
		result.State != eval.Alerting {
		currentState.StateReason = result.State.String()
	}

	// Set Resolved property so the scheduler knows to send a postable alert
	// to Alertmanager.
	currentState.Resolved = oldState == eval.Alerting && currentState.State == eval.Normal

	err := st.maybeTakeScreenshot(ctx, alertRule, currentState, oldState)
	if err != nil {
		st.log.Warn("failed to generate a screenshot for an alert instance",
			"alert_rule", alertRule.UID,
			"dashboard", alertRule.DashboardUID,
			"panel", alertRule.PanelID,
			"err", err)
	}

	st.cache.set(currentState)

	shouldUpdateAnnotation := oldState != currentState.State || oldReason != currentState.StateReason
	if shouldUpdateAnnotation {
		go st.historian.RecordState(ctx, alertRule, currentState.Labels, result.EvaluatedAt, InstanceStateAndReason{State: currentState.State, Reason: currentState.StateReason}, InstanceStateAndReason{State: oldState, Reason: oldReason})
	}
	return currentState
}

func (st *Manager) GetAll(orgID int64) []*State {
	return st.cache.getAll(orgID)
}

func (st *Manager) GetStatesForRuleUID(orgID int64, alertRuleUID string) []*State {
	return st.cache.getStatesForRuleUID(orgID, alertRuleUID)
}

func (st *Manager) recordMetrics() {
	// TODO: parameterize?
	// Setting to a reasonable default scrape interval for Prometheus.
	dur := time.Duration(15) * time.Second
	ticker := st.clock.Ticker(dur)
	for {
		select {
		case <-ticker.C:
			st.log.Debug("recording state cache metrics", "now", st.clock.Now())
			st.cache.recordMetrics(st.metrics)
		case <-st.quit:
			st.log.Debug("stopping state cache metrics recording", "now", st.clock.Now())
			ticker.Stop()
			return
		}
	}
}

func (st *Manager) Put(states []*State) {
	for _, s := range states {
		st.cache.set(s)
	}
}

// TODO: Is the `State` type necessary? Should it embed the instance?
func (st *Manager) saveAlertStates(ctx context.Context, states ...*State) (saved, failed int) {
	st.log.Debug("saving alert states", "count", len(states))
	instances := make([]ngModels.AlertInstance, 0, len(states))

	type debugInfo struct {
		OrgID  int64
		Uid    string
		State  string
		Labels string
	}
	debug := make([]debugInfo, 0)

	for _, s := range states {
		labels := ngModels.InstanceLabels(s.Labels)
		_, hash, err := labels.StringAndHash()
		if err != nil {
			debug = append(debug, debugInfo{s.OrgID, s.AlertRuleUID, s.State.String(), s.Labels.String()})
			st.log.Error("failed to save alert instance with invalid labels", "orgID", s.OrgID, "ruleUID", s.AlertRuleUID, "err", err)
			continue
		}
		fields := ngModels.AlertInstance{
			AlertInstanceKey: ngModels.AlertInstanceKey{
				RuleOrgID:  s.OrgID,
				RuleUID:    s.AlertRuleUID,
				LabelsHash: hash,
			},
			Labels:            ngModels.InstanceLabels(s.Labels),
			CurrentState:      ngModels.InstanceStateType(s.State.String()),
			CurrentReason:     s.StateReason,
			LastEvalTime:      s.LastEvaluationTime,
			CurrentStateSince: s.StartsAt,
			CurrentStateEnd:   s.EndsAt,
		}
		instances = append(instances, fields)
	}

	if err := st.instanceStore.SaveAlertInstances(ctx, instances...); err != nil {
		for _, inst := range instances {
			debug = append(debug, debugInfo{inst.RuleOrgID, inst.RuleUID, string(inst.CurrentState), data.Labels(inst.Labels).String()})
		}
		st.log.Error("failed to save alert states", "states", debug, "err", err)
		return 0, len(debug)
	}

	return len(instances), len(debug)
}

// TODO: why wouldn't you allow other types like NoData or Error?
func translateInstanceState(state ngModels.InstanceStateType) eval.State {
	switch {
	case state == ngModels.InstanceStateFiring:
		return eval.Alerting
	case state == ngModels.InstanceStateNormal:
		return eval.Normal
	default:
		return eval.Error
	}
}

// This struct provides grouping of state with reason, and string formatting.
type InstanceStateAndReason struct {
	State  eval.State
	Reason string
}

func (i InstanceStateAndReason) String() string {
	s := fmt.Sprintf("%v", i.State)
	if len(i.Reason) > 0 {
		s += fmt.Sprintf(" (%v)", i.Reason)
	}
	return s
}

func (st *Manager) staleResultsHandler(ctx context.Context, evaluatedAt time.Time, alertRule *ngModels.AlertRule, states map[string]*State) []*State {
	var resolvedStates []*State
	allStates := st.GetStatesForRuleUID(alertRule.OrgID, alertRule.UID)
	toDelete := make([]ngModels.AlertInstanceKey, 0)

	for _, s := range allStates {
		// Is the cached state in our recently processed results? If not, is it stale?
		if _, ok := states[s.CacheID]; !ok && stateIsStale(evaluatedAt, s.LastEvaluationTime, alertRule.IntervalSeconds) {
			st.log.Debug("removing stale state entry", "orgID", s.OrgID, "alertRuleUID", s.AlertRuleUID, "cacheID", s.CacheID)
			st.cache.deleteEntry(s.OrgID, s.AlertRuleUID, s.CacheID)
			ilbs := ngModels.InstanceLabels(s.Labels)
			_, labelsHash, err := ilbs.StringAndHash()
			if err != nil {
				st.log.Error("unable to get labelsHash", "err", err.Error(), "orgID", s.OrgID, "alertRuleUID", s.AlertRuleUID)
			}

			toDelete = append(toDelete, ngModels.AlertInstanceKey{RuleOrgID: s.OrgID, RuleUID: s.AlertRuleUID, LabelsHash: labelsHash})

			if s.State == eval.Alerting {
				previousState := InstanceStateAndReason{State: s.State, Reason: s.StateReason}
				s.State = eval.Normal
				s.StateReason = ngModels.StateReasonMissingSeries
				s.EndsAt = evaluatedAt
				s.Resolved = true
				st.historian.RecordState(ctx, alertRule, s.Labels, evaluatedAt,
					InstanceStateAndReason{State: eval.Normal, Reason: s.StateReason},
					previousState,
				)
				resolvedStates = append(resolvedStates, s)
			}
		}
	}

	if err := st.instanceStore.DeleteAlertInstances(ctx, toDelete...); err != nil {
		st.log.Error("unable to delete stale instances from database", "err", err.Error(),
			"orgID", alertRule.OrgID, "alertRuleUID", alertRule.UID, "count", len(toDelete))
	}
	return resolvedStates
}

func stateIsStale(evaluatedAt time.Time, lastEval time.Time, intervalSeconds int64) bool {
	return !lastEval.Add(2 * time.Duration(intervalSeconds) * time.Second).After(evaluatedAt)
}

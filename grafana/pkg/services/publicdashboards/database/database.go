package database

import (
	"context"
	"encoding/json"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/dashboards"
	"github.com/grafana/grafana/pkg/services/publicdashboards"
	"github.com/grafana/grafana/pkg/services/publicdashboards/internal/tokens"
	. "github.com/grafana/grafana/pkg/services/publicdashboards/models"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/services/sqlstore/db"
	"github.com/grafana/grafana/pkg/util"
)

// Define the storage implementation. We're generating the mock implementation
// automatically
type PublicDashboardStoreImpl struct {
	sqlStore db.DB
	log      log.Logger
}

var LogPrefix = "publicdashboards.store"

// Gives us a compile time error if our database does not adhere to contract of
// the interface
var _ publicdashboards.Store = (*PublicDashboardStoreImpl)(nil)

// Factory used by wire to dependency injection
func ProvideStore(sqlStore db.DB) *PublicDashboardStoreImpl {
	return &PublicDashboardStoreImpl{
		sqlStore: sqlStore,
		log:      log.New(LogPrefix),
	}
}

// Gets list of public dashboards by orgId
func (d *PublicDashboardStoreImpl) ListPublicDashboards(ctx context.Context, orgId int64) ([]PublicDashboardListResponse, error) {
	resp := make([]PublicDashboardListResponse, 0)

	err := d.sqlStore.WithTransactionalDbSession(ctx, func(sess *sqlstore.DBSession) error {
		sess.Table("dashboard_public").
			Join("LEFT", "dashboard", "dashboard.uid = dashboard_public.dashboard_uid AND dashboard.org_id = dashboard_public.org_id").
			Cols("dashboard_public.uid", "dashboard_public.access_token", "dashboard_public.dashboard_uid", "dashboard_public.is_enabled", "dashboard.title").
			Where("dashboard_public.org_id = ?", orgId).
			OrderBy("is_enabled DESC, dashboard.title ASC")

		err := sess.Find(&resp)
		return err
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (d *PublicDashboardStoreImpl) GetDashboard(ctx context.Context, dashboardUid string) (*models.Dashboard, error) {
	dashboard := &models.Dashboard{Uid: dashboardUid}
	err := d.sqlStore.WithTransactionalDbSession(ctx, func(sess *sqlstore.DBSession) error {
		has, err := sess.Get(dashboard)
		if err != nil {
			return err
		}
		if !has {
			return ErrPublicDashboardNotFound
		}
		return nil
	})

	return dashboard, err
}

// Retrieves public dashboard. This method makes 2 queries to the db so that in the
// even that the underlying dashboard is missing it will return a 404.
func (d *PublicDashboardStoreImpl) GetPublicDashboard(ctx context.Context, accessToken string) (*PublicDashboard, *models.Dashboard, error) {
	if accessToken == "" {
		return nil, nil, ErrPublicDashboardIdentifierNotSet
	}

	// get public dashboard
	pdRes := &PublicDashboard{AccessToken: accessToken}
	err := d.sqlStore.WithTransactionalDbSession(ctx, func(sess *sqlstore.DBSession) error {
		has, err := sess.Get(pdRes)
		if err != nil {
			return err
		}
		if !has {
			return ErrPublicDashboardNotFound
		}
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// find dashboard
	dashRes, err := d.GetDashboard(ctx, pdRes.DashboardUid)

	if err != nil {
		return nil, nil, err
	}

	return pdRes, dashRes, err
}

// Generates a new unique uid to retrieve a public dashboard
func (d *PublicDashboardStoreImpl) GenerateNewPublicDashboardUid(ctx context.Context) (string, error) {
	var uid string

	err := d.sqlStore.WithDbSession(ctx, func(sess *sqlstore.DBSession) error {
		for i := 0; i < 3; i++ {
			uid = util.GenerateShortUID()

			exists, err := sess.Get(&PublicDashboard{Uid: uid})
			if err != nil {
				return err
			}

			if !exists {
				return nil
			}
		}

		return ErrPublicDashboardFailedGenerateUniqueUid
	})

	if err != nil {
		return "", err
	}

	return uid, nil
}

// Generates a new unique access token for a new public dashboard
func (d *PublicDashboardStoreImpl) GenerateNewPublicDashboardAccessToken(ctx context.Context) (string, error) {
	var accessToken string

	err := d.sqlStore.WithDbSession(ctx, func(sess *sqlstore.DBSession) error {
		for i := 0; i < 3; i++ {
			var err error
			accessToken, err = tokens.GenerateAccessToken()
			if err != nil {
				continue
			}

			exists, err := sess.Get(&PublicDashboard{AccessToken: accessToken})
			if err != nil {
				return err
			}

			if !exists {
				return nil
			}
		}

		return ErrPublicDashboardFailedGenerateAccessToken
	})

	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// Retrieves public dashboard configuration by Uid
func (d *PublicDashboardStoreImpl) GetPublicDashboardByUid(ctx context.Context, uid string) (*PublicDashboard, error) {
	if uid == "" {
		return nil, nil
	}

	var found bool
	pdRes := &PublicDashboard{Uid: uid}
	err := d.sqlStore.WithTransactionalDbSession(ctx, func(sess *sqlstore.DBSession) error {
		var err error
		found, err = sess.Get(pdRes)
		return err
	})

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, nil
	}

	return pdRes, err
}

// Retrieves public dashboard configuration
func (d *PublicDashboardStoreImpl) GetPublicDashboardConfig(ctx context.Context, orgId int64, dashboardUid string) (*PublicDashboard, error) {
	if dashboardUid == "" {
		return nil, dashboards.ErrDashboardIdentifierNotSet
	}

	pdRes := &PublicDashboard{OrgId: orgId, DashboardUid: dashboardUid}
	err := d.sqlStore.WithTransactionalDbSession(ctx, func(sess *sqlstore.DBSession) error {
		// publicDashboard
		_, err := sess.Get(pdRes)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return pdRes, err
}

// Persists public dashboard configuration
func (d *PublicDashboardStoreImpl) SavePublicDashboardConfig(ctx context.Context, cmd SavePublicDashboardConfigCommand) error {
	if cmd.PublicDashboard.DashboardUid == "" {
		return dashboards.ErrDashboardIdentifierNotSet
	}

	err := d.sqlStore.WithTransactionalDbSession(ctx, func(sess *sqlstore.DBSession) error {
		_, err := sess.UseBool("is_enabled").Insert(&cmd.PublicDashboard)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// Updates existing public dashboard configuration
func (d *PublicDashboardStoreImpl) UpdatePublicDashboardConfig(ctx context.Context, cmd SavePublicDashboardConfigCommand) error {
	err := d.sqlStore.WithTransactionalDbSession(ctx, func(sess *sqlstore.DBSession) error {
		timeSettingsJSON, err := json.Marshal(cmd.PublicDashboard.TimeSettings)
		if err != nil {
			return err
		}

		_, err = sess.Exec("UPDATE dashboard_public SET is_enabled = ?, time_settings = ?, updated_by = ?, updated_at = ? WHERE uid = ?",
			cmd.PublicDashboard.IsEnabled,
			string(timeSettingsJSON),
			cmd.PublicDashboard.UpdatedBy,
			cmd.PublicDashboard.UpdatedAt.UTC().Format("2006-01-02 15:04:05"),
			cmd.PublicDashboard.Uid)

		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// Responds true if public dashboard for a dashboard exists and isEnabled
func (d *PublicDashboardStoreImpl) PublicDashboardEnabled(ctx context.Context, dashboardUid string) (bool, error) {
	hasPublicDashboard := false
	err := d.sqlStore.WithDbSession(ctx, func(dbSession *sqlstore.DBSession) error {
		sql := "SELECT COUNT(*) FROM dashboard_public WHERE dashboard_uid=? AND is_enabled=true"

		result, err := dbSession.SQL(sql, dashboardUid).Count()
		if err != nil {
			return err
		}

		hasPublicDashboard = result > 0

		return err
	})

	return hasPublicDashboard, err
}

// Responds true if accessToken exists and isEnabled. May be renamed in the
// future
func (d *PublicDashboardStoreImpl) AccessTokenExists(ctx context.Context, accessToken string) (bool, error) {
	hasPublicDashboard := false
	err := d.sqlStore.WithDbSession(ctx, func(dbSession *sqlstore.DBSession) error {
		sql := "SELECT COUNT(*) FROM dashboard_public WHERE access_token=? AND is_enabled=true"

		result, err := dbSession.SQL(sql, accessToken).Count()
		if err != nil {
			return err
		}

		hasPublicDashboard = result > 0

		return err
	})

	return hasPublicDashboard, err
}

// Responds with OrgId from if exists and isEnabled.
func (d *PublicDashboardStoreImpl) GetPublicDashboardOrgId(ctx context.Context, accessToken string) (int64, error) {
	var orgId int64
	err := d.sqlStore.WithDbSession(ctx, func(dbSession *sqlstore.DBSession) error {
		sql := "SELECT org_id FROM dashboard_public WHERE access_token=? AND is_enabled=true"

		_, err := dbSession.SQL(sql, accessToken).Get(&orgId)
		if err != nil {
			return err
		}

		return err
	})

	return orgId, err
}

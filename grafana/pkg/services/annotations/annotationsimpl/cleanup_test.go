package annotationsimpl

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/annotations"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnotationCleanUp(t *testing.T) {
	fakeSQL := sqlstore.InitTestDB(t)

	t.Cleanup(func() {
		err := fakeSQL.WithDbSession(context.Background(), func(session *sqlstore.DBSession) error {
			_, err := session.Exec("DELETE FROM annotation")
			return err
		})
		assert.NoError(t, err)
	})

	createTestAnnotations(t, fakeSQL, 21, 6)
	assertAnnotationCount(t, fakeSQL, "", 21)
	assertAnnotationTagCount(t, fakeSQL, 42)

	tests := []struct {
		name                     string
		cfg                      *setting.Cfg
		alertAnnotationCount     int64
		dashboardAnnotationCount int64
		APIAnnotationCount       int64
		affectedAnnotations      int64
	}{
		{
			name: "default settings should not delete any annotations",
			cfg: &setting.Cfg{
				AlertingAnnotationCleanupSetting:   settingsFn(0, 0),
				DashboardAnnotationCleanupSettings: settingsFn(0, 0),
				APIAnnotationCleanupSettings:       settingsFn(0, 0),
			},
			alertAnnotationCount:     7,
			dashboardAnnotationCount: 7,
			APIAnnotationCount:       7,
			affectedAnnotations:      0,
		},
		{
			name: "should remove annotations created before cut off point",
			cfg: &setting.Cfg{
				AlertingAnnotationCleanupSetting:   settingsFn(time.Hour*48, 0),
				DashboardAnnotationCleanupSettings: settingsFn(time.Hour*48, 0),
				APIAnnotationCleanupSettings:       settingsFn(time.Hour*48, 0),
			},
			alertAnnotationCount:     5,
			dashboardAnnotationCount: 5,
			APIAnnotationCount:       5,
			affectedAnnotations:      6,
		},
		{
			name: "should only keep three annotations",
			cfg: &setting.Cfg{
				AlertingAnnotationCleanupSetting:   settingsFn(0, 3),
				DashboardAnnotationCleanupSettings: settingsFn(0, 3),
				APIAnnotationCleanupSettings:       settingsFn(0, 3),
			},
			alertAnnotationCount:     3,
			dashboardAnnotationCount: 3,
			APIAnnotationCount:       3,
			affectedAnnotations:      6,
		},
		{
			name: "running the max count delete again should not remove any annotations",
			cfg: &setting.Cfg{
				AlertingAnnotationCleanupSetting:   settingsFn(0, 3),
				DashboardAnnotationCleanupSettings: settingsFn(0, 3),
				APIAnnotationCleanupSettings:       settingsFn(0, 3),
			},
			alertAnnotationCount:     3,
			dashboardAnnotationCount: 3,
			APIAnnotationCount:       3,
			affectedAnnotations:      0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := setting.NewCfg()
			cfg.AnnotationCleanupJobBatchSize = 1
			cleaner := ProvideCleanupService(fakeSQL, cfg)
			affectedAnnotations, affectedAnnotationTags, err := cleaner.Run(context.Background(), test.cfg)
			require.NoError(t, err)

			assert.Equal(t, test.affectedAnnotations, affectedAnnotations)
			assert.Equal(t, test.affectedAnnotations*2, affectedAnnotationTags)

			assertAnnotationCount(t, fakeSQL, alertAnnotationType, test.alertAnnotationCount)
			assertAnnotationCount(t, fakeSQL, dashboardAnnotationType, test.dashboardAnnotationCount)
			assertAnnotationCount(t, fakeSQL, apiAnnotationType, test.APIAnnotationCount)

			// we create two records in annotation_tag for each sample annotation
			expectedAnnotationTagCount := (test.alertAnnotationCount +
				test.dashboardAnnotationCount +
				test.APIAnnotationCount) * 2
			assertAnnotationTagCount(t, fakeSQL, expectedAnnotationTagCount)
		})
	}
}

func TestOldAnnotationsAreDeletedFirst(t *testing.T) {
	fakeSQL := sqlstore.InitTestDB(t)

	t.Cleanup(func() {
		err := fakeSQL.WithDbSession(context.Background(), func(session *sqlstore.DBSession) error {
			_, err := session.Exec("DELETE FROM annotation")
			return err
		})
		assert.NoError(t, err)
	})

	// create some test annotations
	a := annotations.Item{
		DashboardId: 1,
		OrgId:       1,
		UserId:      1,
		PanelId:     1,
		AlertId:     10,
		Text:        "",
		Created:     time.Now().AddDate(-10, 0, -10).UnixNano() / int64(time.Millisecond),
	}

	err := fakeSQL.WithDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		_, err := sess.Insert(a)
		require.NoError(t, err, "cannot insert annotation")
		_, err = sess.Insert(a)
		require.NoError(t, err, "cannot insert annotation")

		a.AlertId = 20
		_, err = sess.Insert(a)
		require.NoError(t, err, "cannot insert annotation")

		// run the clean up task to keep one annotation.
		cfg := setting.NewCfg()
		cfg.AnnotationCleanupJobBatchSize = 1
		cleaner := &xormRepositoryImpl{cfg: cfg, log: log.New("test-logger"), db: fakeSQL}
		_, err = cleaner.CleanAnnotations(context.Background(), setting.AnnotationCleanupSettings{MaxCount: 1}, alertAnnotationType)
		require.NoError(t, err)

		// assert that the last annotations were kept
		countNew, err := sess.Where("alert_id = 20").Count(&annotations.Item{})
		require.NoError(t, err)
		require.Equal(t, int64(1), countNew, "the last annotations should be kept")

		countOld, err := sess.Where("alert_id = 10").Count(&annotations.Item{})
		require.NoError(t, err)
		require.Equal(t, int64(0), countOld, "the two first annotations should have been deleted")

		return nil
	})
	require.NoError(t, err)
}

func assertAnnotationCount(t *testing.T, fakeSQL *sqlstore.SQLStore, sql string, expectedCount int64) {
	t.Helper()

	err := fakeSQL.WithDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		count, err := sess.Where(sql).Count(&annotations.Item{})
		require.NoError(t, err)
		require.Equal(t, expectedCount, count)
		return nil
	})
	require.NoError(t, err)
}

func assertAnnotationTagCount(t *testing.T, fakeSQL *sqlstore.SQLStore, expectedCount int64) {
	t.Helper()

	err := fakeSQL.WithDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		count, err := sess.SQL("select count(*) from annotation_tag").Count()
		require.NoError(t, err)
		require.Equal(t, expectedCount, count)
		return nil
	})
	require.NoError(t, err)
}

func createTestAnnotations(t *testing.T, store *sqlstore.SQLStore, expectedCount int, oldAnnotations int) {
	t.Helper()

	cutoffDate := time.Now()

	for i := 0; i < expectedCount; i++ {
		a := &annotations.Item{
			DashboardId: 1,
			OrgId:       1,
			UserId:      1,
			PanelId:     1,
			Text:        "",
		}

		// mark every third as an API annotation
		// that does not belong to a dashboard
		if i%3 == 1 {
			a.DashboardId = 0
		}

		// mark every third annotation as an alert annotation
		if i%3 == 0 {
			a.AlertId = 10
			a.DashboardId = 2
		}

		// create epoch as int annotations.go line 40
		a.Created = cutoffDate.UnixNano() / int64(time.Millisecond)

		// set a really old date for the first six annotations
		if i < oldAnnotations {
			a.Created = cutoffDate.AddDate(-10, 0, -10).UnixNano() / int64(time.Millisecond)
		}

		err := store.WithDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
			_, err := sess.Insert(a)
			require.NoError(t, err, "should be able to save annotation", err)

			// mimick the SQL annotation Save logic by writing records to the annotation_tag table
			// we need to ensure they get deleted when we clean up annotations
			for tagID := range []int{1, 2} {
				_, err = sess.Exec("INSERT INTO annotation_tag (annotation_id, tag_id) VALUES(?,?)", a.Id, tagID)
				require.NoError(t, err, "should be able to save annotation tag ID", err)
			}
			return err
		})
		require.NoError(t, err)
	}
}

func settingsFn(maxAge time.Duration, maxCount int64) setting.AnnotationCleanupSettings {
	return setting.AnnotationCleanupSettings{MaxAge: maxAge, MaxCount: maxCount}
}

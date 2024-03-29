package tagimpl

import (
	"context"

	"github.com/grafana/grafana/pkg/services/sqlstore/db"
	"github.com/grafana/grafana/pkg/services/tag"
	"github.com/grafana/grafana/pkg/setting"
)

type Service struct {
	store store
}

func ProvideService(db db.DB, cfg *setting.Cfg) *Service {
	if cfg.IsFeatureToggleEnabled("newDBLibrary") {
		return &Service{
			store: &sqlxStore{
				sess: db.GetSqlxSession(),
			},
		}
	}
	return &Service{
		store: &sqlStore{
			db: db,
		},
	}
}

func (s *Service) EnsureTagsExist(ctx context.Context, tags []*tag.Tag) ([]*tag.Tag, error) {
	return s.store.EnsureTagsExist(ctx, tags)
}

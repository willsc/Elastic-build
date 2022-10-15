package kind

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/store/kind/dummy"
)

func TestKindRegistry(t *testing.T) {
	registry := NewKindRegistry()
	err := registry.Register(dummy.GetObjectKindInfo("test"), dummy.GetObjectSummaryBuilder("test"))
	require.NoError(t, err)

	ids := []string{}
	for _, k := range registry.GetKinds() {
		ids = append(ids, k.ID)
	}
	require.Equal(t, []string{
		"dashboard",
		"dummy",
		"geojson",
		"kind1",
		"kind2",
		"kind3",
		"playlist",
		"png",
		"test",
	}, ids)

	// Check playlist exists
	info, err := registry.GetInfo(models.StandardKindPlaylist)
	require.NoError(t, err)
	require.Equal(t, "Playlist", info.Name)
	require.False(t, info.IsRaw)

	// Check that we registered a test item
	info, err = registry.GetInfo("test")
	require.NoError(t, err)
	require.Equal(t, "test", info.Name)
	require.True(t, info.IsRaw)
}

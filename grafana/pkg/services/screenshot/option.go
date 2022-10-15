package screenshot

import (
	"time"

	"github.com/grafana/grafana/pkg/models"
)

var (
	DefaultTheme   = models.ThemeDark
	DefaultTimeout = 15 * time.Second
	DefaultHeight  = 500
	DefaultWidth   = 1000
)

// ScreenshotOptions are the options for taking a screenshot.
type ScreenshotOptions struct {
	DashboardUID string
	PanelID      int64
	Width        int
	Height       int
	Theme        models.Theme
	Timeout      time.Duration
}

// SetDefaults sets default values for missing or invalid options.
func (s ScreenshotOptions) SetDefaults() ScreenshotOptions {
	if s.Width <= 0 {
		s.Width = DefaultWidth
	}
	if s.Height <= 0 {
		s.Height = DefaultHeight
	}
	switch s.Theme {
	case models.ThemeDark, models.ThemeLight:
	default:
		s.Theme = DefaultTheme
	}
	if s.Timeout <= 0 {
		s.Timeout = DefaultTimeout
	}
	return s
}

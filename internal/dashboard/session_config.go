package dashboard

import (
	"path/filepath"

	"github.com/pinchtab/pinchtab/internal/authn"
	"github.com/pinchtab/pinchtab/internal/config"
)

const dashboardSessionStateFile = "dashboard-auth-sessions.json"

func SessionManagerConfig(runtime *config.RuntimeConfig) authn.SessionConfig {
	if runtime == nil {
		return authn.SessionConfig{}
	}
	return authn.SessionConfig{
		IdleTimeout:                   runtime.Sessions.Dashboard.IdleTimeout,
		MaxLifetime:                   runtime.Sessions.Dashboard.MaxLifetime,
		ElevationWindow:               runtime.Sessions.Dashboard.ElevationWindow,
		Persist:                       runtime.Sessions.Dashboard.Persist,
		PersistPath:                   filepath.Join(runtime.StateDir, dashboardSessionStateFile),
		PersistElevationAcrossRestart: runtime.Sessions.Dashboard.PersistElevationAcrossRestart,
	}
}

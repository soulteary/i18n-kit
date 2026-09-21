package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultMiddlewareConfig(t *testing.T) {
	config := DefaultMiddlewareConfig()

	assert.Equal(t, DefaultDetector, config.Detector)
	assert.Nil(t, config.Bundle)
	assert.False(t, config.SetCookie)
	assert.Equal(t, "lang", config.CookieName)
	assert.Equal(t, 86400*365, config.CookieMaxAge)
	assert.Equal(t, "/", config.CookiePath)
	assert.False(t, config.CookieSecure)
	assert.False(t, config.DisableCookieHTTPOnly)
	assert.Equal(t, "Lax", config.CookieSameSite)
}

func TestResolveMiddlewareConfig_NoArgsGivesDefaults(t *testing.T) {
	assert.Equal(t, DefaultMiddlewareConfig(), ResolveMiddlewareConfig())
}

func TestResolveMiddlewareConfig_ZeroConfigKeepsDefaults(t *testing.T) {
	cfg := ResolveMiddlewareConfig(MiddlewareConfig{})

	assert.Equal(t, DefaultDetector, cfg.Detector)
	assert.Nil(t, cfg.Bundle)
	assert.Equal(t, "lang", cfg.CookieName)
	assert.Equal(t, 86400*365, cfg.CookieMaxAge)
	assert.Equal(t, "/", cfg.CookiePath)
	assert.Equal(t, "Lax", cfg.CookieSameSite)
	assert.False(t, cfg.DisableCookieHTTPOnly)
	assert.False(t, cfg.SetCookie)
	assert.False(t, cfg.CookieSecure)
}

func TestResolveMiddlewareConfig_IgnoresExtraConfigs(t *testing.T) {
	cfg := ResolveMiddlewareConfig(
		MiddlewareConfig{CookieName: "first"},
		MiddlewareConfig{CookieName: "second", CookiePath: "/ignored"},
	)

	assert.Equal(t, "first", cfg.CookieName)
	assert.Equal(t, "/", cfg.CookiePath)
}

// HttpOnly is on unless it is switched off, and nothing else a caller sets can
// switch it off. It used to be inferred: mergeConfig took the caller's
// CookieHTTPOnly only once CookieName or CookieSameSite was also set, so naming
// the cookie and nothing else silently cleared the flag. That guess is gone
// along with the field it worked around.
func TestResolveMiddlewareConfig_HTTPOnlyRule(t *testing.T) {
	tests := []struct {
		name string
		user MiddlewareConfig
		want bool
	}{
		{"untouched is HttpOnly", MiddlewareConfig{}, true},
		{"an unrelated field does not clear it", MiddlewareConfig{CookieMaxAge: 60}, true},
		{"naming the cookie does not clear it", MiddlewareConfig{CookieName: "site_lang"}, true},
		{"setting SameSite does not clear it", MiddlewareConfig{CookieSameSite: "Strict"}, true},
		{"setting several cookie fields does not clear it", MiddlewareConfig{CookieName: "site_lang", CookieSameSite: "Strict", CookiePath: "/app"}, true},
		{"only asking clears it", MiddlewareConfig{DisableCookieHTTPOnly: true}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, !ResolveMiddlewareConfig(tt.user).DisableCookieHTTPOnly)
		})
	}
}

// StdMiddleware must go through ResolveMiddlewareConfig rather than keeping its
// own copy of the defaults: that equivalence is the whole point of exporting it.

func TestResolveMiddlewareConfig_UserValuesOverrideDefaults(t *testing.T) {
	detector := NewDetector(DetectorConfig{Priority: []string{"header"}})
	bundle := NewBundle(LangEN)
	cfg := ResolveMiddlewareConfig(MiddlewareConfig{
		Detector:       detector,
		Bundle:         bundle,
		SetCookie:      true,
		CookieName:     "site_lang",
		CookieMaxAge:   60,
		CookiePath:     "/app",
		CookieSecure:   true,
		CookieSameSite: "Strict",
	})

	assert.Same(t, detector, cfg.Detector)
	assert.Same(t, bundle, cfg.Bundle)
	assert.True(t, cfg.SetCookie)
	assert.Equal(t, "site_lang", cfg.CookieName)
	assert.Equal(t, 60, cfg.CookieMaxAge)
	assert.Equal(t, "/app", cfg.CookiePath)
	assert.True(t, cfg.CookieSecure)
	assert.False(t, cfg.DisableCookieHTTPOnly)
	assert.Equal(t, "Strict", cfg.CookieSameSite)
}

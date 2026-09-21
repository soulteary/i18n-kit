package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ResolveCookieSameSite is the one place CookieSameSite is interpreted. Every
// middleware reads it, so the rules are tested here rather than once per
// framework -- which is what let "strict" mean Strict on Fiber and Lax on
// net/http.
func TestResolveCookieSameSite(t *testing.T) {
	tests := []struct {
		value string
		want  CookieSameSiteMode
	}{
		{"Lax", SameSiteLax},
		{"Strict", SameSiteStrict},
		{"None", SameSiteNone},
		{"disabled", SameSiteDisabled},

		// Case-insensitive: a config read from YAML or an env var rarely
		// arrives in the canonical spelling.
		{"lax", SameSiteLax},
		{"strict", SameSiteStrict},
		{"none", SameSiteNone},
		{"DISABLED", SameSiteDisabled},
		{"StRiCt", SameSiteStrict},

		{"  Strict  ", SameSiteStrict},
		{"\tNone\n", SameSiteNone},

		// Unset and unrecognised both mean the default rather than reaching a
		// browser as a malformed attribute.
		{"", SameSiteLax},
		{"   ", SameSiteLax},
		{"bogus", SameSiteLax},
		{"SameSite=Strict", SameSiteLax},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			assert.Equal(t, tt.want, ResolveCookieSameSite(tt.value))
		})
	}
}

// Resolving is idempotent: a mode's own string value resolves back to itself,
// which is what lets an adapter hand the canonical spelling to a framework that
// takes the attribute as a string.
func TestResolveCookieSameSite_IsIdempotent(t *testing.T) {
	for _, mode := range []CookieSameSiteMode{SameSiteLax, SameSiteStrict, SameSiteNone, SameSiteDisabled} {
		assert.Equal(t, mode, ResolveCookieSameSite(string(mode)))
	}
}

func TestCookieSameSiteMode_RequiresSecure(t *testing.T) {
	assert.True(t, SameSiteNone.RequiresSecure())
	assert.False(t, SameSiteLax.RequiresSecure())
	assert.False(t, SameSiteStrict.RequiresSecure())
	assert.False(t, SameSiteDisabled.RequiresSecure())
}

// StdMiddleware honours RequiresSecure even when CookieSecure says otherwise:
// the alternative is emitting a cookie the browser throws away.

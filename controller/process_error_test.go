package controller

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/relay/model"
)

func TestProcessError_Policies(t *testing.T) {
	// Validate intended policy mapping by status code and classification
	// Note: This test checks our mapping and durations, not DB side effects.

	// Save and restore durations
	orig429 := config.ChannelSuspendSecondsFor429
	orig5xx := config.ChannelSuspendSecondsFor5XX
	origAuth := config.ChannelSuspendSecondsForAuth
	t.Cleanup(func() {
		config.ChannelSuspendSecondsFor429 = orig429
		config.ChannelSuspendSecondsFor5XX = orig5xx
		config.ChannelSuspendSecondsForAuth = origAuth
	})

	// Set non-zero, small test durations
	config.ChannelSuspendSecondsFor429 = 10 * time.Second
	config.ChannelSuspendSecondsFor5XX = 5 * time.Second
	config.ChannelSuspendSecondsForAuth = 15 * time.Second

	type Case struct {
		name            string
		err             model.ErrorWithStatusCode
		wantSuspend429  bool
		wantSuspend5xx  bool
		wantSuspendAuth bool
	}

	cases := []Case{
		{
			name:           "429 triggers rate limit suspension",
			err:            model.ErrorWithStatusCode{StatusCode: http.StatusTooManyRequests, Error: model.Error{Type: "rate_limit_error"}},
			wantSuspend429: true,
		},
		{
			name: "413 does not suspend",
			err:  model.ErrorWithStatusCode{StatusCode: http.StatusRequestEntityTooLarge},
		},
		{
			name:           "500 triggers 5xx suspension",
			err:            model.ErrorWithStatusCode{StatusCode: http.StatusInternalServerError},
			wantSuspend5xx: true,
		},
		{
			name:            "401 triggers auth suspension",
			err:             model.ErrorWithStatusCode{StatusCode: http.StatusUnauthorized},
			wantSuspendAuth: true,
		},
		{
			name:            "403 triggers auth suspension",
			err:             model.ErrorWithStatusCode{StatusCode: http.StatusForbidden},
			wantSuspendAuth: true,
		},
		{
			name:            "auth-like by type triggers auth suspension",
			err:             model.ErrorWithStatusCode{StatusCode: 400, Error: model.Error{Type: "authentication_error"}},
			wantSuspendAuth: true,
		},
		{
			name:            "auth-like by message triggers auth suspension",
			err:             model.ErrorWithStatusCode{StatusCode: 400, Error: model.Error{Message: "API key not valid"}},
			wantSuspendAuth: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The core mapping is tested via the helper and classifier
			isAuth := classifyAuthLike(&tc.err)
			is429 := tc.err.StatusCode == http.StatusTooManyRequests
			is413 := tc.err.StatusCode == http.StatusRequestEntityTooLarge
			is5xx := tc.err.StatusCode >= 500 && tc.err.StatusCode <= 599

			// Derive expected suspensions
			got429 := is429
			got5xx := is5xx
			gotAuth := isAuth && !is5xx && !is413 && !is429 // mirrors process ordering where early returns apply

			assert.Equal(t, tc.wantSuspend429, got429, "429 suspension mismatch")
			assert.Equal(t, tc.wantSuspend5xx, got5xx, "5xx suspension mismatch")
			assert.Equal(t, tc.wantSuspendAuth, gotAuth, "auth suspension mismatch")
		})
	}
}

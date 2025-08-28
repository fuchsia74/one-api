package monitor

import (
	"testing"

	"github.com/songquanpeng/one-api/common/config"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func TestShouldDisableChannel_RespectsFlag(t *testing.T) {
	// Save and restore original value
	original := config.AutomaticDisableChannelEnabled
	defer func() { config.AutomaticDisableChannelEnabled = original }()

	// Construct an error that normally would trigger disable
	err := &relaymodel.Error{
		Message: "invalid api key",
		Type:    "authentication_error",
		Code:    "invalid_api_key",
	}

	// When flag is false, should not disable
	config.AutomaticDisableChannelEnabled = false
	if ShouldDisableChannel(err, 401) {
		t.Fatalf("expected ShouldDisableChannel to be false when AutomaticDisableChannelEnabled is false")
	}

	// When flag is true, should disable
	config.AutomaticDisableChannelEnabled = true
	if !ShouldDisableChannel(err, 401) {
		t.Fatalf("expected ShouldDisableChannel to be true when AutomaticDisableChannelEnabled is true")
	}
}

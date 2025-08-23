package pricing

import (
	"io"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

// localMockAdaptor implements adaptor.Adaptor for tests
type localMockAdaptor struct {
	pricing map[string]adaptor.ModelConfig
}

func (m *localMockAdaptor) GetDefaultModelPricing() map[string]adaptor.ModelConfig { return m.pricing }
func (m *localMockAdaptor) GetModelRatio(modelName string) float64 {
	if p, ok := m.pricing[modelName]; ok {
		return p.Ratio
	}
	return 2.5 * 0.000001
}
func (m *localMockAdaptor) GetCompletionRatio(modelName string) float64 {
	if p, ok := m.pricing[modelName]; ok {
		return p.CompletionRatio
	}
	return 1.0
}
func (m *localMockAdaptor) Init(meta *meta.Meta)                          {}
func (m *localMockAdaptor) GetRequestURL(meta *meta.Meta) (string, error) { return "", nil }
func (m *localMockAdaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	return nil
}
func (m *localMockAdaptor) ConvertRequest(c *gin.Context, relayMode int, request *relaymodel.GeneralOpenAIRequest) (any, error) {
	return nil, nil
}
func (m *localMockAdaptor) ConvertImageRequest(c *gin.Context, request *relaymodel.ImageRequest) (any, error) {
	return nil, nil
}
func (m *localMockAdaptor) ConvertClaudeRequest(c *gin.Context, request *relaymodel.ClaudeRequest) (any, error) {
	return nil, nil
}
func (m *localMockAdaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return nil, nil
}
func (m *localMockAdaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (*relaymodel.Usage, *relaymodel.ErrorWithStatusCode) {
	return nil, nil
}
func (m *localMockAdaptor) GetModelList() []string { return nil }
func (m *localMockAdaptor) GetChannelName() string { return "mock" }

func TestResolveEffectivePricing_BaseNoTiers(t *testing.T) {
	a := &localMockAdaptor{pricing: map[string]adaptor.ModelConfig{
		"m": {Ratio: 1.0, CompletionRatio: 2.0},
	}}

	eff := ResolveEffectivePricing("m", 10, a)
	if eff.InputRatio != 1.0 {
		t.Fatalf("expected input ratio 1.0, got %v", eff.InputRatio)
	}
	if eff.OutputRatio != 2.0 {
		t.Fatalf("expected output ratio 2.0, got %v", eff.OutputRatio)
	}
	if eff.AppliedTierThreshold != 0 {
		t.Fatalf("expected base tier threshold 0, got %v", eff.AppliedTierThreshold)
	}
}

func TestResolveEffectivePricing_TierSelection(t *testing.T) {
	a := &localMockAdaptor{pricing: map[string]adaptor.ModelConfig{
		"m": {
			Ratio:            1.0,
			CompletionRatio:  2.0,
			CachedInputRatio: 0.4,
			Tiers: []adaptor.ModelRatioTier{
				{InputTokenThreshold: 1000, Ratio: 0.5, CompletionRatio: 3.0},
				{InputTokenThreshold: 5000, Ratio: 0.2},
			},
		},
	}}

	// Select first tier (>=1000)
	eff := ResolveEffectivePricing("m", 1500, a)
	if eff.InputRatio != 0.5 || eff.OutputRatio != 1.5 {
		t.Fatalf("unexpected pricing: in=%v out=%v", eff.InputRatio, eff.OutputRatio)
	}
	if eff.AppliedTierThreshold != 1000 {
		t.Fatalf("expected threshold 1000, got %v", eff.AppliedTierThreshold)
	}

	// Select second tier (>=5000)
	eff = ResolveEffectivePricing("m", 6000, a)
	if eff.InputRatio != 0.2 {
		t.Fatalf("expected input ratio 0.2, got %v", eff.InputRatio)
	}
	// Expect inherited completion ratio 3.0 from first tier since second does not set it
	if abs(eff.OutputRatio-0.6) > 1e-8 {
		t.Fatalf("expected output ratio 0.6 (0.2*3.0), got %v", eff.OutputRatio)
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func TestResolveEffectivePricing_CachedNegativeFree(t *testing.T) {
	a := &localMockAdaptor{pricing: map[string]adaptor.ModelConfig{
		"m": {
			Ratio:            1.0,
			CompletionRatio:  2.0,
			CachedInputRatio: -1, // free cached input
		},
	}}

	eff := ResolveEffectivePricing("m", 10, a)
	if eff.CachedInputRatio >= 0 {
		t.Fatalf("expected negative cached input (free), got %v", eff.CachedInputRatio)
	}
	// No cached output pricing; nothing to assert here
}

// New: ensure output price remains input*completion regardless of cached input and across tier transitions
func TestResolveEffectivePricing_TierTransitionCacheIndependence(t *testing.T) {
	a := &localMockAdaptor{pricing: map[string]adaptor.ModelConfig{
		"tm": {
			Ratio:            1.0,
			CompletionRatio:  2.0,
			CachedInputRatio: 0.5,
			Tiers: []adaptor.ModelRatioTier{
				{InputTokenThreshold: 1000, Ratio: 0.8},
				{InputTokenThreshold: 5000, Ratio: 0.6},
			},
		},
	}}

	eff := ResolveEffectivePricing("tm", 6000, a) // tier2
	if eff.InputRatio != 0.6 {
		t.Fatalf("expected tier2 input ratio 0.6, got %v", eff.InputRatio)
	}
	if abs(eff.OutputRatio-1.2) > 1e-8 { // 0.6 * 2.0
		t.Fatalf("expected output ratio 1.2, got %v", eff.OutputRatio)
	}

	// Cached input ratio should not affect output ratio
	if abs(eff.CachedInputRatio-0.5) > 1e-9 {
		t.Fatalf("expected cached input 0.5, got %v", eff.CachedInputRatio)
	}
}

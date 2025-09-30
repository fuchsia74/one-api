package zhipu

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func float64PtrZhipu(v float64) *float64 {
	return &v
}

func newZhipuContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	c.Request = req
	return c
}

func TestConvertRequestClampsParametersV4(t *testing.T) {
	adaptor := &Adaptor{}
	req := &model.GeneralOpenAIRequest{
		Model:       "glm-4",
		TopP:        float64PtrZhipu(2.0),
		Temperature: float64PtrZhipu(-0.5),
	}

	c := newZhipuContext()

	convertedAny, err := adaptor.ConvertRequest(c, relaymode.ChatCompletions, req)
	if err != nil {
		t.Fatalf("ConvertRequest returned error: %v", err)
	}

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	if !ok {
		t.Fatalf("expected v4 conversion to return GeneralOpenAIRequest, got %T", convertedAny)
	}

	if converted.TopP == nil || *converted.TopP != 1 {
		t.Fatalf("expected TopP to be clamped to 1, got %v", converted.TopP)
	}

	if converted.Temperature == nil || *converted.Temperature != 0 {
		t.Fatalf("expected Temperature to be clamped to 0, got %v", converted.Temperature)
	}
}

func TestConvertRequestClampsParametersV3(t *testing.T) {
	adaptor := &Adaptor{}
	req := &model.GeneralOpenAIRequest{
		Model:       "chatglm-3",
		TopP:        float64PtrZhipu(-0.3),
		Temperature: float64PtrZhipu(1.5),
	}

	c := newZhipuContext()

	convertedAny, err := adaptor.ConvertRequest(c, relaymode.ChatCompletions, req)
	if err != nil {
		t.Fatalf("ConvertRequest returned error: %v", err)
	}

	converted, ok := convertedAny.(*Request)
	if !ok {
		t.Fatalf("expected v3 conversion to return *Request, got %T", convertedAny)
	}

	if converted.TopP == nil || *converted.TopP != 0 {
		t.Fatalf("expected TopP to be clamped to 0, got %v", converted.TopP)
	}

	if converted.Temperature == nil || *converted.Temperature != 1 {
		t.Fatalf("expected Temperature to be clamped to 1, got %v", converted.Temperature)
	}
}

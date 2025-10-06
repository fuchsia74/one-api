package aws

import (
	"testing"

	qwen "github.com/songquanpeng/one-api/relay/adaptor/aws/qwen"
)

func TestGetAdaptorReturnsQwenAdaptor(t *testing.T) {
	adaptor := GetAdaptor("qwen3-32b")
	if adaptor == nil {
		t.Fatalf("expected non-nil adaptor for qwen model")
	}

	if _, ok := adaptor.(*qwen.Adaptor); !ok {
		t.Fatalf("expected adaptor type *qwen.Adaptor, got %T", adaptor)
	}
}

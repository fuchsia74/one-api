package controller

import (
	"strings"

	"github.com/gin-gonic/gin"

	openaipayload "github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/apitype"
	"github.com/songquanpeng/one-api/relay/channeltype"
	metalib "github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

// applyThinkingQueryToChatRequest inspects the thinking query parameter and applies
// the corresponding reasoning defaults to a chat completion request when the
// downstream provider supports extended reasoning.
func applyThinkingQueryToChatRequest(c *gin.Context, request *relaymodel.GeneralOpenAIRequest, meta *metalib.Meta) {
	if request == nil || !isThinkingQueryTruthy(c) {
		return
	}

	modelName := resolveModelName(meta, request.Model)
	if !supportsThinkingInjection(meta, modelName) {
		return
	}

	ensureReasoningEffort(c, request, modelName)
	ensureIncludeReasoning(meta, request)
}

// applyThinkingQueryToResponseRequest applies reasoning defaults to Response API
// requests when thinking is enabled via query parameters.
func applyThinkingQueryToResponseRequest(c *gin.Context, request *openaipayload.ResponseAPIRequest, meta *metalib.Meta) {
	if request == nil || !isThinkingQueryTruthy(c) {
		return
	}

	modelName := resolveModelName(meta, request.Model)
	if !supportsThinkingInjection(meta, modelName) {
		return
	}

	ensureResponseReasoning(c, request, modelName)
}

// isThinkingQueryTruthy reports whether the thinking query parameter requests
// auto-enabling reasoning features for the current request context.
func isThinkingQueryTruthy(c *gin.Context) bool {
	if c == nil {
		return false
	}

	return isTruthy(c.Query("thinking"))
}

// isTruthy normalizes a string and returns true when it matches a known truthy token.
func isTruthy(val string) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// resolveModelName determines the effective model name, preferring the mapped
// model stored on meta over the local fallback when available.
func resolveModelName(meta *metalib.Meta, fallback string) string {
	if meta != nil && strings.TrimSpace(meta.ActualModelName) != "" {
		return meta.ActualModelName
	}
	return fallback
}

// supportsThinkingInjection returns true when the channel and model support
// automatic reasoning parameter injection.
func supportsThinkingInjection(meta *metalib.Meta, modelName string) bool {
	if strings.TrimSpace(modelName) == "" {
		return false
	}

	if meta != nil {
		switch meta.APIType {
		case apitype.Anthropic, apitype.AwsClaude:
			return false
		}
	}

	return isReasoningCapableModel(modelName)
}

// ensureReasoningEffort populates reasoning_effort on the chat request when it
// has not been provided by the caller.
func ensureReasoningEffort(c *gin.Context, request *relaymodel.GeneralOpenAIRequest, modelName string) {
	if request.ReasoningEffort != nil && strings.TrimSpace(*request.ReasoningEffort) != "" {
		return
	}

	desired := normalizeReasoningEffort(modelName, c.Query("reasoning_effort"))
	if desired == "" {
		desired = defaultReasoningEffort(modelName)
	}
	if desired == "" {
		return
	}

	request.ReasoningEffort = stringPtr(desired)
}

// ensureIncludeReasoning guarantees OpenRouter requests opt into reasoning payloads.
func ensureIncludeReasoning(meta *metalib.Meta, request *relaymodel.GeneralOpenAIRequest) {
	if meta == nil || meta.ChannelType != channeltype.OpenRouter {
		return
	}
	if request.IncludeReasoning != nil {
		return
	}
	include := true
	request.IncludeReasoning = &include
}

// ensureResponseReasoning ensures Response API requests include a reasoning effort configuration.
func ensureResponseReasoning(c *gin.Context, request *openaipayload.ResponseAPIRequest, modelName string) {
	var existing string
	if request.Reasoning != nil && request.Reasoning.Effort != nil {
		existing = strings.TrimSpace(*request.Reasoning.Effort)
	}
	if existing != "" {
		return
	}

	desired := normalizeReasoningEffort(modelName, c.Query("reasoning_effort"))
	if desired == "" {
		desired = defaultReasoningEffort(modelName)
	}
	if desired == "" {
		return
	}

	if request.Reasoning == nil {
		request.Reasoning = &relaymodel.OpenAIResponseReasoning{}
	}
	request.Reasoning.Effort = stringPtr(desired)
}

// isReasoningCapableModel identifies models that accept reasoning configuration payloads.
func isReasoningCapableModel(modelName string) bool {
	name := strings.ToLower(strings.TrimSpace(modelName))
	if name == "" {
		return false
	}

	switch {
	case strings.HasPrefix(name, "o1"),
		strings.HasPrefix(name, "o3"),
		strings.HasPrefix(name, "o4"),
		strings.HasPrefix(name, "o-"):
		return true
	case strings.HasPrefix(name, "gpt-5") && !strings.HasPrefix(name, "gpt-5-chat"):
		return true
	case strings.Contains(name, "deep-research"):
		return true
	case strings.HasPrefix(name, "grok"):
		return true
	case strings.Contains(name, "deepseek-r1"):
		return true
	case strings.Contains(name, "reasoner"):
		return true
	default:
		return false
	}
}

// defaultReasoningEffort returns the preferred reasoning effort for a model when none is specified.
func defaultReasoningEffort(modelName string) string {
	name := strings.ToLower(strings.TrimSpace(modelName))
	if name == "" {
		return ""
	}
	if strings.Contains(name, "deep-research") {
		return "medium"
	}
	return "high"
}

// normalizeReasoningEffort sanitizes a requested reasoning effort value for a model.
func normalizeReasoningEffort(modelName, effort string) string {
	normalized := strings.ToLower(strings.TrimSpace(effort))
	if normalized == "" {
		return ""
	}
	if !isReasoningEffortAllowed(modelName, normalized) {
		return ""
	}
	return normalized
}

// isReasoningEffortAllowed reports whether the supplied effort is permitted for the model.
func isReasoningEffortAllowed(modelName, effort string) bool {
	if effort == "" {
		return false
	}
	switch effort {
	case "low", "medium", "high":
	default:
		return false
	}

	name := strings.ToLower(strings.TrimSpace(modelName))
	if strings.Contains(name, "deep-research") {
		return effort == "medium"
	}
	return true
}

// stringPtr returns a pointer to a copy of the provided string value.
func stringPtr(v string) *string {
	value := v
	return &value
}

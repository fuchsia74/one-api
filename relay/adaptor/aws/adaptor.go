package aws

import (
	"context"
	"io"
	"net/http"

	"github.com/Laisky/errors/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/adaptor"
	anthropicAdaptor "github.com/songquanpeng/one-api/relay/adaptor/anthropic"
	"github.com/songquanpeng/one-api/relay/adaptor/aws/utils"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

var _ adaptor.Adaptor = new(Adaptor)

type Adaptor struct {
	awsAdapter utils.AwsAdapter
	Config     aws.Config
	Meta       *meta.Meta
	AwsClient  *bedrockruntime.Client
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.Meta = meta
	defaultConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(meta.Config.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			meta.Config.AK, meta.Config.SK, "")))
	if err != nil {
		return
	}
	a.Config = defaultConfig
	a.AwsClient = bedrockruntime.NewFromConfig(defaultConfig)
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	adaptor := GetAdaptor(request.Model)
	if adaptor == nil {
		return nil, errors.New("adaptor not found")
	}

	a.awsAdapter = adaptor
	return adaptor.ConvertRequest(c, relayMode, request)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if a.awsAdapter == nil {
		return nil, utils.WrapErr(errors.New("awsAdapter is nil"))
	}
	return a.awsAdapter.DoResponse(c, a.AwsClient, meta)
}

func (a *Adaptor) GetModelList() (models []string) {
	for model := range adaptors {
		models = append(models, model)
	}
	return
}

func (a *Adaptor) GetChannelName() string {
	return "aws"
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	return "", nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	return nil
}

func (a *Adaptor) ConvertImageRequest(_ *gin.Context, request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	// AWS Bedrock supports Claude Messages natively. Do not convert payload.
	// Just set context for billing/routing and mark direct pass-through.
	sub := GetAdaptor(request.Model)
	if sub == nil {
		return nil, errors.New("adaptor not found for model: " + request.Model)
	}
	a.awsAdapter = sub
	c.Set(ctxkey.ClaudeMessagesNative, true)
	c.Set(ctxkey.ClaudeDirectPassthrough, true)
	c.Set(ctxkey.OriginalClaudeRequest, request)
	c.Set(ctxkey.RequestModel, request.Model)
	// Also parse into anthropic.Request for AWS SDK payload building
	if parsed, perr := anthropicAdaptor.ConvertClaudeRequest(c, *request); perr == nil {
		c.Set(ctxkey.ConvertedRequest, parsed)
	} else {
		return nil, perr
	}
	// Return the original request object; controller will forward original body
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return nil, nil
}

// Pricing methods - AWS adapter manages its own model pricing
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelConfig {
	// Direct map definition - much easier to maintain and edit
	// Pricing from https://aws.amazon.com/bedrock/pricing/
	return map[string]adaptor.ModelConfig{
		// Claude Models on AWS Bedrock
		"claude-instant-1.2":         {Ratio: 0.8 * ratio.MilliTokensUsd, CompletionRatio: 3.125}, // $0.8/$2.5 per 1M tokens
		"claude-2.0":                 {Ratio: 8 * ratio.MilliTokensUsd, CompletionRatio: 3.125},   // $8/$25 per 1M tokens
		"claude-2.1":                 {Ratio: 8 * ratio.MilliTokensUsd, CompletionRatio: 3.125},   // $8/$25 per 1M tokens
		"claude-3-haiku-20240307":    {Ratio: 0.25 * ratio.MilliTokensUsd, CompletionRatio: 5},    // $0.25/$1.25 per 1M tokens
		"claude-3-sonnet-20240229":   {Ratio: 3 * ratio.MilliTokensUsd, CompletionRatio: 5},       // $3/$15 per 1M tokens
		"claude-3-opus-20240229":     {Ratio: 15 * ratio.MilliTokensUsd, CompletionRatio: 5},      // $15/$75 per 1M tokens
		"claude-opus-4-20250514":     {Ratio: 15 * ratio.MilliTokensUsd, CompletionRatio: 5},      // $15/$75 per 1M tokens
		"claude-opus-4-1-20250805":   {Ratio: 15 * ratio.MilliTokensUsd, CompletionRatio: 5},      // $15/$75 per 1M tokens
		"claude-3-5-sonnet-20240620": {Ratio: 3 * ratio.MilliTokensUsd, CompletionRatio: 5},       // $3/$15 per 1M tokens
		"claude-3-5-sonnet-20241022": {Ratio: 3 * ratio.MilliTokensUsd, CompletionRatio: 5},       // $3/$15 per 1M tokens
		"claude-3-5-sonnet-latest":   {Ratio: 3 * ratio.MilliTokensUsd, CompletionRatio: 5},       // $3/$15 per 1M tokens
		"claude-3-5-haiku-20241022":  {Ratio: 1 * ratio.MilliTokensUsd, CompletionRatio: 5},       // $1/$5 per 1M tokens
		"claude-3-7-sonnet-latest":   {Ratio: 3 * ratio.MilliTokensUsd, CompletionRatio: 5},       // $3/$15 per 1M tokens
		"claude-3-7-sonnet-20250219": {Ratio: 3 * ratio.MilliTokensUsd, CompletionRatio: 5},       // $3/$15 per 1M tokens
		"claude-sonnet-4-20250514":   {Ratio: 3 * ratio.MilliTokensUsd, CompletionRatio: 5},       // $3/$15 per 1M tokens

		// Llama Models on AWS Bedrock
		// Note: Pricing may need to be updated later; also this model is significantly faster on AWS GPUs.
		// Llama 4 models - Pricing given per 1K tokens; normalize to $/1M then to $/token via ratio.MilliTokensUsd
		"llama4-maverick-17b-1m": {Ratio: 0.24 * ratio.MilliTokensUsd, CompletionRatio: 4.04}, // $0.24/$0.97 per 1M → $/token
		"llama4-scout-17b-3.5m":  {Ratio: 0.17 * ratio.MilliTokensUsd, CompletionRatio: 3.88}, // $0.17/$0.66 per 1M → $/token

		// Llama 3.3 models
		"llama3-3-70b-128k": {Ratio: 0.72 * ratio.MilliTokensUsd, CompletionRatio: 1}, // $0.72/$0.72 per 1M → $/token

		// Llama 3.2 models
		"llama3-2-1b-131k":         {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},  // $0.1/$0.1 per 1M → $/token
		"llama3-2-3b-131k":         {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 1}, // $0.15/$0.15 per 1M → $/token
		"llama3-2-11b-vision-131k": {Ratio: 0.16 * ratio.MilliTokensUsd, CompletionRatio: 1}, // $0.16/$0.16 per 1M → $/token
		"llama3-2-90b-128k":        {Ratio: 0.72 * ratio.MilliTokensUsd, CompletionRatio: 1}, // $0.72/$0.72 per 1M → $/token

		// Llama 3.1 models
		"llama3-1-8b-128k":  {Ratio: 0.22 * ratio.MilliTokensUsd, CompletionRatio: 1}, // $0.22/$0.22 per 1M → $/token
		"llama3-1-70b-128k": {Ratio: 0.72 * ratio.MilliTokensUsd, CompletionRatio: 1}, // $0.72/$0.72 per 1M → $/token

		// Llama 3 models
		"llama3-8b-8192":  {Ratio: 0.3 * ratio.MilliTokensUsd, CompletionRatio: 2},     // $0.3/$0.6 per 1M → $/token
		"llama3-70b-8192": {Ratio: 2.65 * ratio.MilliTokensUsd, CompletionRatio: 1.32}, // $2.65/$3.5 per 1M → $/token
		// Amazon Nova Models (if supported)
		"amazon-nova-micro":   {Ratio: 0.035 * ratio.MilliTokensUsd, CompletionRatio: 4.28}, // $0.035/$0.15 per 1M tokens
		"amazon-nova-lite":    {Ratio: 0.06 * ratio.MilliTokensUsd, CompletionRatio: 4.17},  // $0.06/$0.25 per 1M tokens
		"amazon-nova-pro":     {Ratio: 0.8 * ratio.MilliTokensUsd, CompletionRatio: 4},      // $0.8/$3.2 per 1M tokens
		"amazon-nova-premier": {Ratio: 2.4 * ratio.MilliTokensUsd, CompletionRatio: 4.17},   // $2.4/$10 per 1M tokens

		// Titan Models (if supported)
		"amazon-titan-text-lite":    {Ratio: 0.3 * ratio.MilliTokensUsd, CompletionRatio: 1.33}, // $0.3/$0.4 per 1M tokens
		"amazon-titan-text-express": {Ratio: 0.8 * ratio.MilliTokensUsd, CompletionRatio: 2},    // $0.8/$1.6 per 1M tokens
		"amazon-titan-embed-text":   {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},    // $0.1 per 1M tokens

		// Cohere Models (if supported)
		"cohere-command-text":       {Ratio: 1.5 * ratio.MilliTokensUsd, CompletionRatio: 1.33}, // $1.5/$2 per 1M tokens
		"cohere-command-light-text": {Ratio: 0.3 * ratio.MilliTokensUsd, CompletionRatio: 2},    // $0.3/$0.6 per 1M tokens

		// AI21 Models (if supported)
		"ai21-j2-mid":    {Ratio: 12.5 * ratio.MilliTokensUsd, CompletionRatio: 1}, // $12.5 per 1M tokens
		"ai21-j2-ultra":  {Ratio: 18.8 * ratio.MilliTokensUsd, CompletionRatio: 1}, // $18.8 per 1M tokens
		"ai21-jamba-1.5": {Ratio: 2 * ratio.MilliTokensUsd, CompletionRatio: 4},    // $2/$8 per 1M tokens

		// Mistral Models (if supported)
		"mistral-7b-instruct":   {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 1.33}, // $0.15/$0.2 per 1M tokens
		"mistral-8x7b-instruct": {Ratio: 0.45 * ratio.MilliTokensUsd, CompletionRatio: 1.56}, // $0.45/$0.7 per 1M tokens
		"mistral-large":         {Ratio: 4 * ratio.MilliTokensUsd, CompletionRatio: 3},       // $4/$12 per 1M tokens
	}
}

func (a *Adaptor) GetModelRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.Ratio
	}
	// Default AWS pricing (Claude-like)
	return 3 * 0.000001 // Default USD pricing
}

func (a *Adaptor) GetCompletionRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.CompletionRatio
	}
	// Default completion ratio for AWS
	return 5.0
}

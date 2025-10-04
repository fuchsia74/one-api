package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"

	"github.com/songquanpeng/one-api/common/ctxkey"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/adaptor/aws/utils"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

var _ utils.AwsAdapter = new(Adaptor)

// Adaptor implements the AWS Bedrock adapter for Qwen3 Coder language models.
//
// This struct provides the core functionality for integrating with AWS Bedrock's
// Qwen3 Coder family models, implementing the AwsAdapter interface to ensure consistent
// behavior across all AWS Bedrock integrations in the One API system.
//
// The adapter handles the complete request-response lifecycle:
//   - Converting OpenAI-compatible requests to AWS Bedrock Qwen format
//   - Processing responses from AWS Bedrock back to OpenAI-compatible format
//   - Managing both streaming and non-streaming response modes
//   - Handling Qwen's code-focused features including multi-language programming support
//   - Managing error conditions and usage tracking
//   - Supporting Qwen's advanced tool calling capabilities for code execution
type Adaptor struct {
}

// ConvertRequest transforms an OpenAI-compatible request into AWS Bedrock Qwen format.
//
// This method performs the critical translation between the One API's unified request format
// and the specific format expected by AWS Bedrock's Qwen3 Coder models. It handles:
//
//   - Message format conversion from OpenAI to Qwen structure
//   - Parameter mapping and validation for Qwen's code-focused features
//   - Tool definitions for code execution and automation
//   - Stop sequence processing for proper generation control
//   - Context storage for downstream processing including multi-language code support
//   - Integration with Qwen's programming capabilities and technical accuracy
//
// Parameters:
//   - c: Gin context for the HTTP request, used for storing converted data
//   - relayMode: Processing mode indicator (embeddings not supported for Qwen)
//   - request: OpenAI-compatible request to be converted to Qwen format
//
// Returns:
//   - any: Converted request in AWS Bedrock Qwen format, ready for API submission
//   - error: Error if the request is invalid or conversion fails
//
// The method stores both the original model name and converted request in the context
// for use by downstream handlers and response processors, enabling proper code-focused
// handling with Qwen's multi-language programming and tool calling features.
func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	var convertedReq any

	switch relayMode {
	case relaymode.Embeddings:
		return nil, errors.New("Qwen models do not support embeddings")
	default:
		convertedReq = ConvertRequest(*request)
	}

	c.Set(ctxkey.RequestModel, request.Model)
	c.Set(ctxkey.ConvertedRequest, convertedReq)
	c.Set("relayMode", relayMode)
	return convertedReq, nil
}

// DoResponse processes the response from AWS Bedrock Qwen and converts it back to OpenAI format.
//
// This method handles the complete response processing pipeline, supporting both streaming
// and non-streaming modes. It coordinates with specialized handlers to:
//
//   - Process AWS Bedrock Qwen responses in their native format
//   - Convert responses back to OpenAI-compatible structure
//   - Handle Qwen's code-focused conversation features and multi-language programming responses
//   - Track token usage for billing and quota management
//   - Manage Qwen's tool calling results and code execution outputs
//   - Handle errors and edge cases appropriately with technical reliability
//
// The method automatically detects the response mode (streaming vs non-streaming) based
// on metadata and delegates to the appropriate specialized handler that understands
// Qwen3 Coder's code generation format and programming capabilities.
//
// Parameters:
//   - c: Gin context containing request data and used for response writing
//   - awsCli: AWS Bedrock Runtime client for making API calls to AWS services
//   - meta: Request metadata including streaming flags and model information
//
// Returns:
//   - usage: Token usage statistics for billing and monitoring purposes with code generation tracking
//   - err: Error with HTTP status code if processing fails, nil on success
//
// Error conditions include network failures, AWS API errors, response parsing issues,
// tool calling errors, code execution failures, and context cancellation scenarios.
func (a *Adaptor) DoResponse(c *gin.Context, awsCli *bedrockruntime.Client, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	relayModeValue, exists := c.Get("relayMode")
	if exists {
		if relayModeInt, ok := relayModeValue.(int); ok {
			switch relayModeInt {
			case relaymode.Embeddings:
				return nil, utils.WrapErr(errors.New("Qwen models do not support embeddings"))
			}
		}
	}

	if meta.IsStream {
		err, usage = StreamHandler(c, awsCli)
	} else {
		err, usage = Handler(c, awsCli, meta.ActualModelName)
	}
	return
}

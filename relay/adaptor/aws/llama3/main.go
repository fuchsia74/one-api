// Package aws provides the AWS adaptor for the relay service.
package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/random"

	"github.com/Laisky/errors/v2"
	gmw "github.com/Laisky/gin-middlewares/v6"
	"github.com/Laisky/zap"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"

	// "github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor/aws/utils"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

// Support for Llama 3, 3.1, 3.2, 3.3, and 4.0 instruction models
// https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html
var AwsModelIDMap = map[string]string{
	// Llama 3 models
	"llama3-8b-8192":  "meta.llama3-8b-instruct-v1:0",
	"llama3-70b-8192": "meta.llama3-70b-instruct-v1:0",

	// Llama 3.1 models
	"llama3-1-8b-128k":  "meta.llama3-1-8b-instruct-v1:0",
	"llama3-1-70b-128k": "meta.llama3-1-70b-instruct-v1:0",

	// Llama 3.2 models
	"llama3-2-90b-128k":        "meta.llama3-2-90b-instruct-v1:0",
	"llama3-2-3b-131k":         "meta.llama3-2-3b-instruct-v1:0",
	"llama3-2-1b-131k":         "meta.llama3-2-1b-instruct-v1:0",
	"llama3-2-11b-vision-131k": "meta.llama3-2-11b-instruct-v1:0",

	// Llama 3.3 models
	"llama3-3-70b-128k": "meta.llama3-3-70b-instruct-v1:0",

	// Llama 4 models
	"llama4-scout-17b-3.5m":  "meta.llama4-scout-17b-instruct-v1:0",
	"llama4-maverick-17b-1m": "meta.llama4-maverick-17b-instruct-v1:0",
}

func awsModelID(requestModel string) (string, error) {
	if awsModelID, ok := AwsModelIDMap[requestModel]; ok {
		return awsModelID, nil
	}

	return "", errors.Errorf("model %s not found", requestModel)
}

// promptTemplate for legacy Llama models (3, 3.1, 3.2, 3.3)
const legacyPromptTemplate = `<|begin_of_text|>{{range .Messages}}<|start_header_id|>{{.Role}}<|end_header_id|>{{.StringContent}}<|eot_id|>{{end}}<|start_header_id|>assistant<|end_header_id|>`

// promptTemplate for Llama 4 models
const llama4PromptTemplate = `<|begin_of_text|>{{range .Messages}}<|start_header_id|>{{.Role}}<|end_header_id|>{{.StringContent}}<|eot|>{{end}}<|start_header_id|>assistant<|end_header_id|>`

var legacyPromptTpl = template.Must(template.New("llama-legacy-chat").Parse(legacyPromptTemplate))
var llama4PromptTpl = template.Must(template.New("llama4-chat").Parse(llama4PromptTemplate))

// isLlama4Model checks if the given model is a Llama 4 model
func isLlama4Model(model string) bool {
	return model == "llama4-scout-17b-3.5m" || model == "llama4-maverick-17b-1m"
}

func RenderPrompt(messages []relaymodel.Message, model string) string {
	var buf bytes.Buffer
	var err error

	if isLlama4Model(model) {
		err = llama4PromptTpl.Execute(&buf, struct{ Messages []relaymodel.Message }{messages})
	} else {
		err = legacyPromptTpl.Execute(&buf, struct{ Messages []relaymodel.Message }{messages})
	}

	if err != nil {
		// rendering prompt failed
	}
	return buf.String()
}

func ConvertRequest(textRequest relaymodel.GeneralOpenAIRequest) *Request {
	llamaRequest := Request{
		MaxGenLen:   textRequest.MaxTokens,
		Temperature: textRequest.Temperature,
		TopP:        textRequest.TopP,
	}
	if llamaRequest.MaxGenLen == 0 {
		llamaRequest.MaxGenLen = config.DefaultMaxToken
	}
	prompt := RenderPrompt(textRequest.Messages, textRequest.Model)
	llamaRequest.Prompt = prompt
	return &llamaRequest
}

func Handler(c *gin.Context, awsCli *bedrockruntime.Client, modelName string) (*relaymodel.ErrorWithStatusCode, *relaymodel.Usage) {
	awsModelID, err := awsModelID(c.GetString(ctxkey.RequestModel))
	if err != nil {
		return utils.WrapErr(errors.Wrap(err, "awsModelID")), nil
	}
	awsModelID = utils.ConvertModelID2CrossRegionProfile(awsModelID, awsCli.Options().Region)
	awsReq := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(awsModelID),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	}

	llamaReq, ok := c.Get(ctxkey.ConvertedRequest)
	if !ok {
		return utils.WrapErr(errors.New("request not found")), nil
	}

	awsReq.Body, err = json.Marshal(llamaReq)
	if err != nil {
		return utils.WrapErr(errors.Wrap(err, "marshal request")), nil
	}

	awsResp, err := awsCli.InvokeModel(gmw.Ctx(c), awsReq)
	if err != nil {
		return utils.WrapErr(errors.Wrap(err, "InvokeModel")), nil
	}

	var llamaResponse Response
	err = json.Unmarshal(awsResp.Body, &llamaResponse)
	if err != nil {
		return utils.WrapErr(errors.Wrap(err, "unmarshal response")), nil
	}

	openaiResp := ResponseLlama2OpenAI(&llamaResponse)
	openaiResp.Model = modelName
	usage := relaymodel.Usage{
		PromptTokens:     llamaResponse.PromptTokenCount,
		CompletionTokens: llamaResponse.GenerationTokenCount,
		TotalTokens:      llamaResponse.PromptTokenCount + llamaResponse.GenerationTokenCount,
	}
	openaiResp.Usage = usage

	c.JSON(http.StatusOK, openaiResp)
	return nil, &usage
}

func ResponseLlama2OpenAI(llamaResponse *Response) *openai.TextResponse {
	var responseText string
	if len(llamaResponse.Generation) > 0 {
		responseText = llamaResponse.Generation
	}
	choice := openai.TextResponseChoice{
		Index: 0,
		Message: relaymodel.Message{
			Role:    "assistant",
			Content: responseText,
			Name:    nil,
		},
		FinishReason: llamaResponse.StopReason,
	}
	fullTextResponse := openai.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", random.GetUUID()),
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: []openai.TextResponseChoice{choice},
	}
	return &fullTextResponse
}

func StreamHandler(c *gin.Context, awsCli *bedrockruntime.Client) (*relaymodel.ErrorWithStatusCode, *relaymodel.Usage) {
	lg := gmw.GetLogger(c)
	createdTime := helper.GetTimestamp()
	awsModelID, err := awsModelID(c.GetString(ctxkey.RequestModel))
	if err != nil {
		return utils.WrapErr(errors.Wrap(err, "awsModelID")), nil
	}
	awsModelID = utils.ConvertModelID2CrossRegionProfile(awsModelID, awsCli.Options().Region)
	awsReq := &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(awsModelID),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	}

	llamaReq, ok := c.Get(ctxkey.ConvertedRequest)
	if !ok {
		return utils.WrapErr(errors.New("request not found")), nil
	}

	awsReq.Body, err = json.Marshal(llamaReq)
	if err != nil {
		return utils.WrapErr(errors.Wrap(err, "marshal request")), nil
	}

	awsResp, err := awsCli.InvokeModelWithResponseStream(gmw.Ctx(c), awsReq)
	if err != nil {
	}
	stream := awsResp.GetStream()
	defer stream.Close()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	var usage relaymodel.Usage
	c.Stream(func(w io.Writer) bool {
		event, ok := <-stream.Events()
		if !ok {
			c.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}

		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			var llamaResp StreamResponse
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&llamaResp)
			if err != nil {
				lg.Error("error unmarshalling stream response", zap.Error(err))
				return false
			}

			if llamaResp.PromptTokenCount > 0 {
				usage.PromptTokens = llamaResp.PromptTokenCount
			}
			if llamaResp.StopReason == "stop" {
				usage.CompletionTokens = llamaResp.GenerationTokenCount
				usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
			}
			response := StreamResponseLlama2OpenAI(&llamaResp)
			response.Id = fmt.Sprintf("chatcmpl-%s", random.GetUUID())
			response.Model = c.GetString(ctxkey.RequestModel)
			response.Created = createdTime
			jsonStr, err := json.Marshal(response)
			if err != nil {
				lg.Error("error marshalling stream response", zap.Error(err))
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonStr)})
			return true
		case *types.UnknownUnionMember:
			fmt.Println("unknown tag:", v.Tag)
			return false
		default:
			fmt.Println("union is nil or unknown type")
			return false
		}
	})

	return nil, &usage
}

func StreamResponseLlama2OpenAI(llamaResponse *StreamResponse) *openai.ChatCompletionsStreamResponse {
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = llamaResponse.Generation
	choice.Delta.Role = "assistant"
	finishReason := llamaResponse.StopReason
	if finishReason != "null" {
		choice.FinishReason = &finishReason
	}
	var openaiResponse openai.ChatCompletionsStreamResponse
	openaiResponse.Object = "chat.completion.chunk"
	openaiResponse.Choices = []openai.ChatCompletionsStreamResponseChoice{choice}
	return &openaiResponse
}

package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	gmw "github.com/Laisky/gin-middlewares/v6"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/relay/model"
)

type GeneralErrorResponse struct {
	Error    model.Error `json:"error"`
	Message  string      `json:"message"`
	Msg      string      `json:"msg"`
	Err      string      `json:"err"`
	ErrorMsg string      `json:"error_msg"`
	Header   struct {
		Message string `json:"message"`
	} `json:"header"`
	Response struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	} `json:"response"`
}

func (e GeneralErrorResponse) ToMessage() string {
	if e.Error.Message != "" {
		return e.Error.Message
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Msg != "" {
		return e.Msg
	}
	if e.Err != "" {
		return e.Err
	}
	if e.ErrorMsg != "" {
		return e.ErrorMsg
	}
	if e.Header.Message != "" {
		return e.Header.Message
	}
	if e.Response.Error.Message != "" {
		return e.Response.Error.Message
	}
	return ""
}

// RelayErrorHandler parses upstream error responses into our unified error model.
// For request-scoped logging, prefer RelayErrorHandlerWithContext.
func RelayErrorHandler(resp *http.Response) (ErrorWithStatusCode *model.ErrorWithStatusCode) {
	if resp == nil {
		return &model.ErrorWithStatusCode{
			StatusCode: 500,
			Error: model.Error{
				Message: "resp is nil",
				Type:    "upstream_error",
				Code:    "bad_response",
			},
		}
	}
	ErrorWithStatusCode = &model.ErrorWithStatusCode{
		StatusCode: resp.StatusCode,
		Error: model.Error{
			Message: "",
			Type:    "upstream_error",
			Code:    "bad_response_status_code",
			Param:   strconv.Itoa(resp.StatusCode),
		},
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	// Intentionally avoid global logger here; use context variant where possible.
	err = resp.Body.Close()
	if err != nil {
		return
	}
	var errResponse GeneralErrorResponse
	err = json.Unmarshal(responseBody, &errResponse)
	if err != nil {
		errResponse.Error.Message = string(responseBody)
		return
	}

	if errResponse.Error.Message != "" {
		// OpenAI format error, so we override the default one
		ErrorWithStatusCode.Error = errResponse.Error
	} else {
		ErrorWithStatusCode.Error.Message = errResponse.ToMessage()
	}

	if ErrorWithStatusCode.Error.Message == "" {
		ErrorWithStatusCode.Error.Message = fmt.Sprintf("bad response status code %d", resp.StatusCode)
	}
	return
}

// RelayErrorHandlerWithContext is a context-aware variant that logs using the request-scoped logger.
func RelayErrorHandlerWithContext(c *gin.Context, resp *http.Response) *model.ErrorWithStatusCode {
	if resp == nil {
		return RelayErrorHandler(resp)
	}
	// Read and restore response body for downstream use
	responseBody, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if config.DebugEnabled {
		gmw.GetLogger(c).Info("error happened",
			zap.Int("status_code", resp.StatusCode),
			zap.ByteString("response", responseBody),
		)
	}
	// Reconstruct a new ReadCloser for any further reads (not commonly needed here)
	resp.Body = io.NopCloser(bytes.NewReader(responseBody))
	return RelayErrorHandler(resp)
}

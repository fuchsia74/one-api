package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Laisky/errors/v2"
	gmw "github.com/Laisky/gin-middlewares/v6"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/network"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/model"
)

func GetRequestCost(c *gin.Context) {
	reqId := c.Param("request_id")
	if reqId == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "request_id should not be empty",
		})
		return
	}

	docu, err := model.GetCostByRequestId(reqId)
	if err != nil {
		helper.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, docu)
}

func GetAllTokens(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	// Get page size from query parameter, default to config value
	size, _ := strconv.Atoi(c.Query("size"))
	if size <= 0 {
		size = config.DefaultItemsPerPage
	}
	if size > config.MaxItemsPerPage {
		size = config.MaxItemsPerPage
	}

	order := c.Query("order")
	sortBy := c.Query("sort")
	sortOrder := c.Query("order")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	tokens, err := model.GetAllUserTokens(userId, p*size, size, order, sortBy, sortOrder)

	if err != nil {
		helper.RespondError(c, err)
		return
	}

	// Get total count for pagination
	totalCount, err := model.GetUserTokenCount(userId)
	if err != nil {
		helper.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    tokens,
		"total":   totalCount,
	})
	return
}

func SearchTokens(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	keyword := c.Query("keyword")
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	size, _ := strconv.Atoi(c.Query("size"))
	if size <= 0 {
		size = config.DefaultItemsPerPage
	}
	if size > config.MaxItemsPerPage {
		size = config.MaxItemsPerPage
	}
	sortBy := c.Query("sort")
	sortOrder := c.Query("order")
	if sortOrder == "" {
		sortOrder = "desc"
	}
	tokens, total, err := model.SearchUserTokens(userId, keyword, p*size, size, sortBy, sortOrder)
	if err != nil {
		helper.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    tokens,
		"total":   total,
	})
	return
}

func GetToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	userId := c.GetInt(ctxkey.Id)
	if err != nil {
		helper.RespondError(c, err)
		return
	}
	token, err := model.GetTokenByIds(id, userId)
	if err != nil {
		helper.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    token,
	})
	return
}

func GetTokenStatus(c *gin.Context) {
	tokenId := c.GetInt(ctxkey.TokenId)
	userId := c.GetInt(ctxkey.Id)
	token, err := model.GetTokenByIds(tokenId, userId)
	if err != nil {
		helper.RespondError(c, err)
		return
	}
	expiredAt := token.ExpiredTime
	if expiredAt == -1 {
		expiredAt = 0
	}
	c.JSON(http.StatusOK, gin.H{
		"object":          "credit_summary",
		"total_granted":   token.RemainQuota,
		"total_used":      0, // not supported currently
		"total_available": token.RemainQuota,
		"expires_at":      expiredAt * 1000,
	})
}

func validateToken(_ *gin.Context, token *model.Token) error {
	if len(token.Name) > 30 {
		return errors.Errorf("Token name is too long")
	}

	if token.Subnet != nil && *token.Subnet != "" {
		err := network.IsValidSubnets(*token.Subnet)
		if err != nil {
			return errors.Wrap(err, "invalid network segment")
		}
	}

	return nil
}

func AddToken(c *gin.Context) {
	token := new(model.Token)
	err := c.ShouldBindJSON(token)
	if err != nil {
		helper.RespondError(c, err)
		return
	}

	// Disallow empty name on create
	if strings.TrimSpace(token.Name) == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Token name is required",
		})
		return
	}

	err = validateToken(c, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("invalid token: %s", err.Error()),
		})
		return
	}

	cleanToken := model.Token{
		UserId:         c.GetInt(ctxkey.Id),
		Name:           token.Name,
		Key:            random.GenerateKey(),
		CreatedTime:    helper.GetTimestamp(),
		AccessedTime:   helper.GetTimestamp(),
		ExpiredTime:    token.ExpiredTime,
		RemainQuota:    token.RemainQuota,
		UnlimitedQuota: token.UnlimitedQuota,
		Models:         token.Models,
		Subnet:         token.Subnet,
	}
	err = cleanToken.Insert()
	if err != nil {
		helper.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    cleanToken,
	})
	return
}

func DeleteToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.GetInt(ctxkey.Id)
	err := model.DeleteTokenById(id, userId)
	if err != nil {
		helper.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

type consumeTokenRequest struct {
	// AddUsedQuota add or subtract used quota from another source
	AddUsedQuota uint `json:"add_used_quota" gorm:"-"`
	// AddReason is the reason for adding or subtracting used quota
	AddReason string `json:"add_reason" gorm:"-"`
}

// ConsumeToken consume token from another source,
// let one-api to gather billing from different sources.
func ConsumeToken(c *gin.Context) {
	// Capture request start time to compute elapsed latency for the synthetic billing log.
	startTime := time.Now()
	ctx := gmw.Ctx(c)
	userID := c.GetInt(ctxkey.Id)
	tokenID := c.GetInt(ctxkey.TokenId)

	// Parse and validate request
	tokenPatch := new(consumeTokenRequest)
	if err := c.ShouldBindJSON(tokenPatch); err != nil {
		helper.RespondError(c, err)
		return
	}

	if tokenPatch.AddUsedQuota == 0 {
		helper.RespondError(c, errors.New("add_used_quota must be greater than 0"))
		return
	}

	if tokenPatch.AddReason == "" {
		helper.RespondError(c, errors.New("add_reason cannot be empty"))
		return
	}

	// Get token
	cleanToken, err := model.GetTokenByIds(tokenID, userID)
	if err != nil {
		helper.RespondError(c, err)
		return
	}

	// Validate token status
	if err := validateTokenForConsumption(cleanToken); err != nil {
		helper.RespondError(c, err)
		return
	}

	// Check quota availability
	if !cleanToken.UnlimitedQuota && cleanToken.RemainQuota < int64(tokenPatch.AddUsedQuota) {
		helper.RespondError(c, errors.New("Insufficient remaining quota"))
		return
	}

	// Transaction: decrease quota and record log
	if err = model.DecreaseTokenQuota(cleanToken.Id, int64(tokenPatch.AddUsedQuota)); err != nil {
		helper.RespondError(c, err)
		return
	}

	// Record consumption log (pass IDs explicitly)
	traceID := ""
	if tid, err := gmw.TraceID(c); err == nil {
		traceID = tid.String()
	}
	model.RecordConsumeLog(ctx, &model.Log{
		UserId:      userID,
		ModelName:   tokenPatch.AddReason,
		TokenName:   cleanToken.Name,
		Quota:       int(tokenPatch.AddUsedQuota),
		Content:     fmt.Sprintf("External (%s) consumed %s", tokenPatch.AddReason, common.LogQuota(int64(tokenPatch.AddUsedQuota))),
		RequestId:   c.GetString(helper.RequestIdKey),
		TraceId:     traceID,
		ElapsedTime: helper.CalcElapsedTime(startTime), // milliseconds
	})

	// Update token data
	cleanToken.RemainQuota -= int64(tokenPatch.AddUsedQuota)
	if err = cleanToken.Update(); err != nil {
		helper.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    cleanToken,
	})
}

func validateTokenForConsumption(token *model.Token) error {
	// Check if token is enabled
	if token.Status != model.TokenStatusEnabled {
		return errors.Errorf("API Key is not enabled")
	}

	// Check if token is expired
	if token.ExpiredTime != -1 && token.ExpiredTime <= helper.GetTimestamp() {
		return errors.Errorf("The token has expired and cannot be used. Please modify the expiration time of the token, or set it to never expire")
	}

	// Check if token is exhausted
	if !token.UnlimitedQuota && token.RemainQuota <= 0 {
		return errors.Errorf("The available quota of the token has been used up. Please add more quota or set it to unlimited")
	}

	return nil
}

func UpdateToken(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	statusOnly := c.Query("status_only")
	tokenPatch := new(model.Token)
	err := c.ShouldBindJSON(tokenPatch)
	if err != nil {
		helper.RespondError(c, err)
		return
	}

	// Disallow empty name when not status_only
	if statusOnly == "" && strings.TrimSpace(tokenPatch.Name) == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Token name cannot be empty",
		})
		return
	}

	token := new(model.Token)
	if err = copier.Copy(token, tokenPatch); err != nil {
		helper.RespondError(c, err)
		return
	}

	err = validateToken(c, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("invalid token: %s", err.Error()),
		})
		return
	}

	cleanToken, err := model.GetTokenByIds(token.Id, userId)
	if err != nil {
		helper.RespondError(c, err)
		return
	}

	switch token.Status {
	case model.TokenStatusEnabled:
		if cleanToken.Status == model.TokenStatusExpired &&
			cleanToken.ExpiredTime <= helper.GetTimestamp() && cleanToken.ExpiredTime != -1 &&
			token.ExpiredTime != -1 && token.ExpiredTime < helper.GetTimestamp() {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "The token has expired and cannot be enabled. Please modify the expiration time of the token, or set it to never expire.",
			})
			return
		}
		if cleanToken.Status == model.TokenStatusExhausted &&
			cleanToken.RemainQuota <= 0 && !cleanToken.UnlimitedQuota &&
			token.RemainQuota <= 0 && !token.UnlimitedQuota {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "The available quota of the token has been used up and cannot be enabled. Please modify the remaining quota of the token, or set it to unlimited quota",
			})
			return
		}
	case model.TokenStatusExhausted:
		if token.RemainQuota > 0 || token.UnlimitedQuota {
			token.Status = model.TokenStatusEnabled
		}
	case model.TokenStatusExpired:
		if token.ExpiredTime == -1 || token.ExpiredTime > helper.GetTimestamp() {
			token.Status = model.TokenStatusEnabled
		}
	}

	if statusOnly != "" {
		cleanToken.Status = token.Status
	} else {
		// If you add more fields, please also update token.Update()
		cleanToken.Name = token.Name
		cleanToken.ExpiredTime = token.ExpiredTime
		cleanToken.UnlimitedQuota = token.UnlimitedQuota
		cleanToken.Models = token.Models
		cleanToken.Subnet = token.Subnet
		cleanToken.RemainQuota = token.RemainQuota
		cleanToken.Status = token.Status
	}

	err = cleanToken.Update()
	if err != nil {
		helper.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    cleanToken,
	})
	return
}

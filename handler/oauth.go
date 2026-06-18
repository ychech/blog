// package handler 提供 HTTP 请求处理函数。
//
// 本文件实现 OAuth2 第三方登录接口。
package handler

import (
	"blog/service"
	"blog/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// OAuthHandler OAuth 登录处理器。
type OAuthHandler struct{}

// NewOAuthHandler 创建 OAuth 处理器。
func NewOAuthHandler() *OAuthHandler {
	return &OAuthHandler{}
}

// GitHubLogin 跳转 GitHub 授权页。
// @Summary GitHub 登录跳转
// @Tags 认证
// @Produce json
// @Success 302 {string} string "重定向到 GitHub"
// @Failure 400 {object} utils.Response
// @Router /auth/oauth/github [get]
func (h *OAuthHandler) GitHubLogin(c *gin.Context) {
	state := "random-state" // 生产环境应使用随机值并校验
	url, err := service.GetGitHubAuthURL(state)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}
	c.Redirect(http.StatusFound, url)
}

// GitHubCallback GitHub 授权回调。
// @Summary GitHub 登录回调
// @Tags 认证
// @Produce json
// @Param code query string true "授权码"
// @Param state query string false "状态值"
// @Success 200 {object} utils.Response{data=model.LoginResponse}
// @Failure 400 {object} utils.Response
// @Router /auth/oauth/github/callback [get]
func (h *OAuthHandler) GitHubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		utils.BadRequest(c, "缺少授权码")
		return
	}

	resp, err := service.HandleGitHubCallback(code)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, resp)
}

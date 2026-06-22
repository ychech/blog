package service

import (
	"blog/config"
	"blog/database"
	"blog/model"
	"blog/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var githubOAuthConfig *oauth2.Config

// InitGitHubOAuth 初始化 GitHub OAuth2 配置。
func InitGitHubOAuth(cfg config.OAuthConfig) {
	if !cfg.GitHubEnabled || cfg.GitHubClientID == "" || cfg.GitHubClientSecret == "" {
		return
	}
	githubOAuthConfig = &oauth2.Config{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		RedirectURL:  cfg.GitHubRedirectURL,
		Endpoint:     github.Endpoint,
		Scopes:       []string{"read:user", "user:email"},
	}
}

// GitHubOAuthEnabled GitHub 登录是否已启用。
func GitHubOAuthEnabled() bool {
	return githubOAuthConfig != nil
}

// GetGitHubAuthURL 生成 GitHub 授权跳转 URL。
func GetGitHubAuthURL(state string) (string, error) {
	if !GitHubOAuthEnabled() {
		return "", fmt.Errorf("GitHub 登录未启用")
	}
	return githubOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

// HandleGitHubCallback 处理 GitHub 回调，返回登录响应。
func HandleGitHubCallback(code string) (*model.LoginResponse, error) {
	if !GitHubOAuthEnabled() {
		return nil, fmt.Errorf("GitHub 登录未启用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token, err := githubOAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("换取 access_token 失败: %w", err)
	}

	userInfo, err := fetchGitHubUser(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}

	// 优先使用已绑定的账号登录
	var oauth model.OAuthAccount
	err = database.DB.Where("provider = ? AND provider_user_id = ?", "github", fmt.Sprintf("%d", userInfo.ID)).First(&oauth).Error
	if err == nil {
		var user model.User
		if err := database.DB.First(&user, oauth.UserID).Error; err != nil {
			return nil, fmt.Errorf("关联用户不存在")
		}
		return generateLoginResponse(&user)
	}

	// 未绑定则创建新用户并绑定
	username := generateOAuthUsername("github", userInfo.Login)
	user := model.User{
		Username: username,
		Password: "",
		Nickname: userInfo.Name,
		Email:    userInfo.Email,
		Avatar:   userInfo.AvatarURL,
		Role:     model.UserRoleUser,
		IsActive: true,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	oauth = model.OAuthAccount{
		UserID:         user.ID,
		Provider:       "github",
		ProviderUserID: fmt.Sprintf("%d", userInfo.ID),
		AccessToken:    token.AccessToken,
	}
	if err := database.DB.Create(&oauth).Error; err != nil {
		return nil, fmt.Errorf("绑定 OAuth 账号失败: %w", err)
	}

	return generateLoginResponse(&user)
}

func generateLoginResponse(user *model.User) (*model.LoginResponse, error) {
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	claims, err := utils.ParseToken(token)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		User:     *user,
		Token:    token,
		ExpireAt: claims.ExpiresAt.Unix(),
	}, nil
}

func generateOAuthUsername(provider, login string) string {
	base := fmt.Sprintf("%s_%s", provider, login)
	username := base
	for i := 1; ; i++ {
		var count int64
		database.DB.Model(&model.User{}).Where("username = ?", username).Count(&count)
		if count == 0 {
			break
		}
		username = fmt.Sprintf("%s_%d", base, i)
	}
	return username
}

// githubUserInfo GitHub 用户信息结构。
type githubUserInfo struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func fetchGitHubUser(ctx context.Context, accessToken string) (*githubUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub 用户信息失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub 用户信息请求失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info githubUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("解析 GitHub 用户信息失败: %w", err)
	}
	return &info, nil
}

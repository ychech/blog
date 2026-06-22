package service

import (
	"blog/config"
	"blog/database"
	"blog/model"
	"blog/utils"
	"fmt"
	"time"
)

// SendNotificationEmail 异步向用户发送通知邮件。
// 用户未绑定邮箱、邮箱未验证、SMTP 未配置或通知邮件开关关闭时直接返回。
func SendNotificationEmail(userID uint, notification *model.Notification) {
	if userID == 0 || notification == nil {
		return
	}
	go sendNotificationEmailSync(userID, notification)
}

func sendNotificationEmailSync(userID uint, notification *model.Notification) {
	if config.C == nil {
		return
	}
	cfg := config.C.Email
	if !cfg.NotificationEmailEnabled {
		return
	}
	if cfg.Host == "" || cfg.Username == "" || cfg.Password == "" {
		utils.Logger.Warn("未配置 SMTP，跳过发送通知邮件")
		return
	}

	var user model.User
	if err := database.DB.Select("id, email, email_verified").First(&user, userID).Error; err != nil {
		utils.Logger.Errorf("发送通知邮件时查询用户失败: %v", err)
		return
	}
	if user.Email == "" || !user.EmailVerified {
		return
	}

	subject, body := renderNotificationEmail(notification)
	if err := sendEmailFunc(user.Email, subject, body); err != nil {
		utils.Logger.Errorf("发送通知邮件失败: %v", err)
	}
}

// sendEmailFunc 用于发送邮件，默认使用 utils.SendEmail；测试可替换为 mock。
var sendEmailFunc = utils.SendEmail

// renderNotificationEmail 根据通知类型渲染邮件标题与正文。
func renderNotificationEmail(n *model.Notification) (string, string) {
	typeName := notificationTypeName(n.Type)
	subject := fmt.Sprintf("【博客】%s", n.Title)
	body := fmt.Sprintf(
		"<p>您好，您收到一条新的%s通知：</p>"+
			"<h3>%s</h3>"+
			"<p>%s</p>"+
			"<p>时间：%s</p>"+
			"<p style=\"color:#999;\">如不想接收此类邮件，请登录博客后在设置中关闭邮件提醒。</p>",
		typeName,
		n.Title,
		n.Content,
		n.CreatedAt.Format(time.RFC3339),
	)
	return subject, body
}

func notificationTypeName(t model.NotificationType) string {
	switch t {
	case model.NotificationTypeCommentReply:
		return "评论回复"
	case model.NotificationTypePostLike:
		return "文章点赞"
	case model.NotificationTypeCommentLike:
		return "评论点赞"
	case model.NotificationTypeFollow:
		return "关注"
	case model.NotificationTypeMessage:
		return "私信"
	case model.NotificationTypeBadgeAward:
		return "勋章"
	default:
		return "系统"
	}
}

// package utils 提供项目通用工具。
//
// 本文件实现基于 net/smtp 的邮件发送功能，支持 SSL/TLS 与普通 SMTP。
package utils

import (
	"blog/config"
	"crypto/tls"
	"fmt"
	"net/smtp"
)

// SendEmail 发送一封邮件。
//
// 参数：
//   - to: 收件人邮箱
//   - subject: 邮件主题
//   - body: 邮件正文，支持 HTML
//
// 注意：发送前必须正确配置 config.C.Email（Host、Port、Username、Password）。
func SendEmail(to, subject, body string) error {
	cfg := config.C.Email
	if cfg.Host == "" || cfg.Username == "" || cfg.Password == "" {
		return fmt.Errorf("邮件配置不完整，请检查 SMTP 配置")
	}

	from := cfg.From
	if from == "" {
		from = cfg.Username
	}

	msg := []byte(fmt.Sprintf(
		"To: %s\r\nFrom: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		to, from, subject, body,
	))

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// SSL/TLS 模式（如 QQ 邮箱、163 邮箱等）
	if cfg.EnableSSL {
		tlsConfig := &tls.Config{
			ServerName:         cfg.Host,
			InsecureSkipVerify: false,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("SMTP TLS 连接失败: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, cfg.Host)
		if err != nil {
			return fmt.Errorf("创建 SMTP 客户端失败: %w", err)
		}
		defer client.Close()

		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP 认证失败: %w", err)
		}
		if err := client.Mail(cfg.Username); err != nil {
			return fmt.Errorf("设置发件人失败: %w", err)
		}
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("设置收件人失败: %w", err)
		}
		wc, err := client.Data()
		if err != nil {
			return fmt.Errorf("打开邮件数据流失败: %w", err)
		}
		if _, err := wc.Write(msg); err != nil {
			return fmt.Errorf("写入邮件内容失败: %w", err)
		}
		if err := wc.Close(); err != nil {
			return fmt.Errorf("关闭邮件数据流失败: %w", err)
		}
		return client.Quit()
	}

	// 普通 SMTP / STARTTLS（如 Gmail 等）
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	if err := smtp.SendMail(addr, auth, cfg.Username, []string{to}, msg); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}
	return nil
}

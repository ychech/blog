package service

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// validateNonEmptyTrimmed 校验字符串去除首尾空格后是否为空。
func validateNonEmptyTrimmed(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return newValidationError(fieldName + "不能为空")
	}
	return nil
}

// validateMaxLength 校验字符串长度是否超过上限（按字符数）。
func validateMaxLength(value, fieldName string, max int) error {
	if len([]rune(value)) > max {
		return newValidationError(fieldName + "不能超过 " + intToString(max) + " 个字符")
	}
	return nil
}

// validateEmail 校验邮箱格式（空值不校验）。
func validateEmail(email string) error {
	if email == "" {
		return nil
	}
	if !emailRegex.MatchString(email) {
		return newValidationError("邮箱格式不正确")
	}
	return nil
}

func newValidationError(msg string) error {
	return &validationError{msg: msg}
}

type validationError struct {
	msg string
}

func (e *validationError) Error() string {
	return e.msg
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

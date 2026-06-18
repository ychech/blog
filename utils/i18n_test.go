package utils

import "testing"

func TestT(t *testing.T) {
	if got := T("zh", "request_too_frequent"); got != "请求过于频繁，请稍后再试" {
		t.Errorf("中文翻译错误: %s", got)
	}
	if got := T("en", "request_too_frequent"); got != "Too many requests, please try again later" {
		t.Errorf("英文翻译错误: %s", got)
	}
	if got := T("unknown", "missing_key"); got != "missing_key" {
		t.Errorf("缺失 key 应返回自身: %s", got)
	}
}

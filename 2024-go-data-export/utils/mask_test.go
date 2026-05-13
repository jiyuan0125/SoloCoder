package utils

import "testing"

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"13812345678", "138****5678"},
		{"13987654321", "139****4321"},
		{"1234567890", "1234567890"},
		{"", ""},
		{"abc", "abc"},
	}

	for _, tt := range tests {
		result := MaskPhone(tt.input)
		if result != tt.expected {
			t.Errorf("MaskPhone(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"zhangsan@example.com", "z******n@example.com"},
		{"lisi@test.com", "l**i@test.com"},
		{"wangwu@company.org", "w****u@company.org"},
		{"a@b.com", "a@b.com"},
		{"", ""},
		{"invalid-email", "invalid-email"},
	}

	for _, tt := range tests {
		result := MaskEmail(tt.input)
		if result != tt.expected {
			t.Errorf("MaskEmail(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestMaskValue(t *testing.T) {
	tests := []struct {
		field    string
		value    string
		expected string
	}{
		{"phone", "13812345678", "138****5678"},
		{"mobile", "13987654321", "139****4321"},
		{"email", "zhangsan@example.com", "z******n@example.com"},
		{"user_email", "lisi@test.com", "l**i@test.com"},
		{"mail_address", "wangwu@org.cn", "w****u@org.cn"},
		{"name", "张三", "张三"},
		{"age", "25", "25"},
	}

	for _, tt := range tests {
		result := MaskValue(tt.field, tt.value)
		if result != tt.expected {
			t.Errorf("MaskValue(%q, %q) = %q, want %q", tt.field, tt.value, result, tt.expected)
		}
	}
}

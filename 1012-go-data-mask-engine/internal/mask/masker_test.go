package mask

import "testing"

func TestMaskIDCard(t *testing.T) {
	rules := NewDefaultRules()
	m := NewMasker(rules)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"18位身份证", "110101199003076618", "110***********6618"},
		{"18位身份证结尾X", "11010119900307661X", "110***********661X"},
		{"18位身份证小写x", "11010119900307661x", "110***********661X"},
		{"15位身份证", "110101900307661", "110********7661"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.MaskIDCard(tt.input, nil)
			if result != tt.expected {
				t.Errorf("MaskIDCard(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMaskPhone(t *testing.T) {
	rules := NewDefaultRules()
	m := NewMasker(rules)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"手机号", "13800138000", "138****8000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.MaskPhone(tt.input, nil)
			if result != tt.expected {
				t.Errorf("MaskPhone(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMaskName(t *testing.T) {
	rules := NewDefaultRules()
	m := NewMasker(rules)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"两字姓名", "张三", "张*"},
		{"三字姓名", "张三丰", "张**"},
		{"四字姓名", "欧阳铁柱", "欧***"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.MaskName(tt.input, nil)
			if result != tt.expected {
				t.Errorf("MaskName(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMaskEmail(t *testing.T) {
	rules := NewDefaultRules()
	m := NewMasker(rules)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"普通邮箱", "zhangsan@example.com", "z*******@example.com"},
		{"单字符用户名", "a@example.com", "a*@example.com"},
		{"长用户名", "firstname.lastname@company.com", "f*****************@company.com"},
		{"无@不脱敏", "invalid-email", "invalid-email"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.MaskEmail(tt.input, nil)
			if result != tt.expected {
				t.Errorf("MaskEmail(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMaskBankCard(t *testing.T) {
	rules := NewDefaultRules()
	m := NewMasker(rules)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"16位银行卡", "6222021234567890", "************7890"},
		{"19位银行卡", "6222021234567890123", "***************0123"},
		{"太短不脱敏", "12345678901", "12345678901"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.MaskBankCard(tt.input, nil)
			if result != tt.expected {
				t.Errorf("MaskBankCard(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

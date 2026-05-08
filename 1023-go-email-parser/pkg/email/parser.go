package email

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/net/idna"
)

// Address 表示解析后的邮箱地址
type Address struct {
	Original       string // 原始邮箱地址
	Local          string // 本地部分（保持原样大小写）
	Domain         string // 域名部分（统一小写）
	HasAlias       bool   // 是否包含+号别名
	BaseLocal      string // 去掉别名后的基础本地部分
	BaseAddress    string // 去掉别名后的基础地址
	IsQuotedLocal  bool   // 本地部分是否被引号包裹
}

// ParseResult 表示单个邮箱的解析结果
type ParseResult struct {
	Address *Address // 解析后的地址，校验失败时为nil
	Valid   bool     // 是否有效
	Error   error    // 错误信息，校验失败时设置
}

// BatchParseResult 表示批量解析结果
type BatchParseResult struct {
	Total   int           // 总数量
	Valid   int           // 有效数量
	Invalid int           // 无效数量
	Results []ParseResult // 每个地址的详细结果
}

// Parse 解析单个邮箱地址
// 注意：本实现暂不支持引号包裹的本地部分中包含@符号的情况
// 例如："foo@bar"@example.com，这种情况需要更复杂的词法分析
func Parse(email string) (*Address, error) {
	// 去除首尾空白
	email = strings.TrimSpace(email)

	if email == "" {
		return nil, fmt.Errorf("邮箱地址不能为空")
	}

	// 查找@符号的位置
	// 注意：这里简单地查找最后一个@符号
	// 对于引号包裹本地部分中包含@的情况，需要更复杂的解析
	atIndex := strings.LastIndex(email, "@")
	if atIndex == -1 {
		return nil, fmt.Errorf("邮箱地址缺少@符号")
	}
	if atIndex == 0 {
		return nil, fmt.Errorf("本地部分不能为空")
	}
	if atIndex == len(email)-1 {
		return nil, fmt.Errorf("域名部分不能为空")
	}

	localPart := email[:atIndex]
	domainPart := email[atIndex+1:]

	// 检查本地部分是否被引号包裹
	isQuoted := false
	if len(localPart) >= 2 && localPart[0] == '"' && localPart[len(localPart)-1] == '"' {
		isQuoted = true
		// 去掉外层引号
		localPart = localPart[1 : len(localPart)-1]
	}

	// 校验本地部分
	if err := validateLocalPart(localPart, isQuoted); err != nil {
		return nil, fmt.Errorf("本地部分校验失败: %w", err)
	}

	// 校验域名部分
	if err := validateDomainPart(domainPart); err != nil {
		return nil, fmt.Errorf("域名部分校验失败: %w", err)
	}

	// 处理+号别名
	hasAlias := false
	baseLocal := localPart
	if !isQuoted {
		// 只对非引号包裹的本地部分处理别名
		plusIndex := strings.Index(localPart, "+")
		if plusIndex != -1 {
			if plusIndex == len(localPart)-1 {
				return nil, fmt.Errorf("本地部分不能以+号结尾")
			}
			if plusIndex == 0 {
				return nil, fmt.Errorf("本地部分不能以+号开头")
			}
			hasAlias = true
			baseLocal = localPart[:plusIndex]
		}
	}

	// 域名统一转小写
	domainLower := strings.ToLower(domainPart)

	// 计算基础地址
	baseAddress := baseLocal + "@" + domainLower

	return &Address{
		Original:      email,
		Local:         localPart,
		Domain:        domainLower,
		HasAlias:      hasAlias,
		BaseLocal:     baseLocal,
		BaseAddress:   baseAddress,
		IsQuotedLocal: isQuoted,
	}, nil
}

// Validate 校验邮箱地址是否合法
func Validate(email string) bool {
	_, err := Parse(email)
	return err == nil
}

// ParseBatch 批量解析邮箱地址列表
// 输入可以是逗号或分号分隔的字符串
func ParseBatch(list string) (*BatchParseResult, error) {
	emails, err := splitEmailList(list)
	if err != nil {
		return nil, err
	}

	result := &BatchParseResult{
		Total:   len(emails),
		Results: make([]ParseResult, len(emails)),
	}

	for i, email := range emails {
		addr, err := Parse(email)
		if err != nil {
			result.Results[i] = ParseResult{
				Address: nil,
				Valid:   false,
				Error:   err,
			}
			result.Invalid++
		} else {
			result.Results[i] = ParseResult{
				Address: addr,
				Valid:   true,
				Error:   nil,
			}
			result.Valid++
		}
	}

	return result, nil
}

// splitEmailList 分割邮箱地址列表
// 处理逗号或分号分隔符，并考虑引号包裹的情况
func splitEmailList(list string) ([]string, error) {
	list = strings.TrimSpace(list)
	if list == "" {
		return nil, fmt.Errorf("邮箱列表不能为空")
	}

	var emails []string
	var current strings.Builder
	inQuotes := false

	for i, r := range list {
		switch {
		case r == '"':
			inQuotes = !inQuotes
			current.WriteRune(r)
		case (r == ',' || r == ';') && !inQuotes:
			email := strings.TrimSpace(current.String())
			if email != "" {
				emails = append(emails, email)
			}
			current.Reset()
		case unicode.IsSpace(r) && !inQuotes:
			// 非引号内的空格忽略，除非在邮箱地址中间
			// 这里简化处理，只保留空格用于后续trim
			current.WriteRune(r)
		default:
			current.WriteRune(r)
		}

		// 检查未闭合的引号
		if i == len(list)-1 && inQuotes {
			return nil, fmt.Errorf("引号未闭合")
		}
	}

	// 添加最后一个邮箱
	email := strings.TrimSpace(current.String())
	if email != "" {
		emails = append(emails, email)
	}

	if len(emails) == 0 {
		return nil, fmt.Errorf("未找到有效的邮箱地址")
	}

	return emails, nil
}

// validateLocalPart 校验本地部分
func validateLocalPart(local string, isQuoted bool) error {
	// 长度限制：RFC 5321规定本地部分最大64个字符
	if len(local) == 0 {
		return fmt.Errorf("本地部分不能为空")
	}
	if len(local) > 64 {
		return fmt.Errorf("本地部分长度超过64个字符")
	}

	if isQuoted {
		// 引号包裹的本地部分规则较宽松
		// 可以包含空格和特殊字符，但需要检查转义
		return validateQuotedLocal(local)
	}

	// 非引号包裹的本地部分校验
	// 不能以点号开头或结尾
	if local[0] == '.' {
		return fmt.Errorf("本地部分不能以点号开头")
	}
	if local[len(local)-1] == '.' {
		return fmt.Errorf("本地部分不能以点号结尾")
	}

	// 不能有连续的点号
	if strings.Contains(local, "..") {
		return fmt.Errorf("本地部分不能包含连续的点号")
	}

	// 检查每个字符
	for i, r := range local {
		switch {
		case r == '.' || r == '+':
			// 点号和+号是允许的，但已经在上面检查了位置
			continue
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			// 字母和数字允许
			continue
		case strings.ContainsRune("!#$%&'*+-/=?^_`{|}~", r):
			// RFC 5322允许的特殊字符
			continue
		default:
			return fmt.Errorf("本地部分包含非法字符: %c (位置%d)", r, i)
		}
	}

	return nil
}

// validateQuotedLocal 校验引号包裹的本地部分
func validateQuotedLocal(local string) error {
	// 引号包裹的本地部分可以包含更多字符
	// 但不能包含未转义的反斜杠或引号
	escaped := false
	for _, r := range local {
		if escaped {
			escaped = false
			continue
		}
		switch r {
		case '\\':
			escaped = true
		case '"':
			return fmt.Errorf("引号包裹的本地部分内部不能包含未转义的引号")
		default:
			// 其他字符都允许
		}
	}

	if escaped {
		return fmt.Errorf("本地部分末尾有未完成的转义")
	}

	return nil
}

// validateDomainPart 校验域名部分
func validateDomainPart(domain string) error {
	// 长度限制：RFC 5321规定域名最大255个字符
	// 但通常实际限制是253个字符（不含末尾的点）
	if len(domain) == 0 {
		return fmt.Errorf("域名部分不能为空")
	}
	if len(domain) > 253 {
		return fmt.Errorf("域名长度超过253个字符")
	}

	// 处理国际化域名：尝试转换为Punycode
	punycodeDomain, err := idna.Lookup.ToASCII(domain)
	if err != nil {
		return fmt.Errorf("国际化域名转换失败: %w", err)
	}

	// 检查转换后的域名长度
	if len(punycodeDomain) > 253 {
		return fmt.Errorf("Punycode转换后的域名长度超过253个字符")
	}

	// 分割标签
	labels := strings.Split(punycodeDomain, ".")
	if len(labels) < 2 {
		return fmt.Errorf("域名至少需要两个标签")
	}

	for i, label := range labels {
		if err := validateLabel(label, i == len(labels)-1); err != nil {
			return err
		}
	}

	return nil
}

// validateLabel 校验单个域名标签
func validateLabel(label string, isTLD bool) error {
	// 标签长度限制：最多63个字符
	if len(label) == 0 {
		return fmt.Errorf("域名标签不能为空")
	}
	if len(label) > 63 {
		return fmt.Errorf("域名标签长度超过63个字符")
	}

	// 不能以连字符开头
	if label[0] == '-' {
		return fmt.Errorf("域名标签不能以连字符开头: %s", label)
	}

	// 不能以连字符结尾
	if label[len(label)-1] == '-' {
		return fmt.Errorf("域名标签不能以连字符结尾: %s", label)
	}

	// 检查是否全部是数字（对TLD的特殊限制）
	if isTLD {
		allDigits := true
		for _, r := range label {
			if !unicode.IsDigit(r) {
				allDigits = false
				break
			}
		}
		if allDigits {
			return fmt.Errorf("顶级域名不能全部是数字: %s", label)
		}
	}

	// 检查每个字符
	for _, r := range label {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			continue
		case r == '-':
			continue
		default:
			return fmt.Errorf("域名标签包含非法字符: %c", r)
		}
	}

	return nil
}

package mask

import (
"strings"
"unicode/utf8"
)

const maskChar = '*'

type Masker struct {
rules *DefaultRules
}

func NewMasker(rules *DefaultRules) *Masker {
if rules == nil {
rules = NewDefaultRules()
}
return &Masker{rules: rules}
}

func (m *Masker) MaskByType(dataType DataType, value string, rule *MaskRule) string {
switch dataType {
case Phone:
return m.MaskPhone(value, rule)
case IDCard:
return m.MaskIDCard(value, rule)
case BankCard:
return m.MaskBankCard(value, rule)
case Email:
return m.MaskEmail(value, rule)
case Name:
return m.MaskName(value, rule)
default:
return value
}
}

func (m *Masker) MaskPhone(phone string, rule *MaskRule) string {
if rule == nil {
rule = m.rules.Phone
}
return maskString(phone, rule.PrefixKeep, rule.SuffixKeep)
}

func (m *Masker) MaskIDCard(idCard string, rule *MaskRule) string {
if rule == nil {
rule = m.rules.IDCard
}
normalized := strings.ToUpper(strings.TrimSpace(idCard))
n := utf8.RuneCountInString(normalized)
if n == 15 || n == 18 {
return maskString(normalized, rule.PrefixKeep, rule.SuffixKeep)
}
return idCard
}

func (m *Masker) MaskBankCard(bankCard string, rule *MaskRule) string {
if rule == nil {
rule = m.rules.BankCard
}
n := len(strings.TrimSpace(bankCard))
if n < 12 || n > 19 {
return bankCard
}
return maskString(bankCard, rule.PrefixKeep, rule.SuffixKeep)
}

func (m *Masker) MaskEmail(email string, rule *MaskRule) string {
if rule == nil {
rule = m.rules.Email
}
atIndex := strings.Index(email, "@")
if atIndex == -1 {
return email
}
username := email[:atIndex]
domain := email[atIndex:]
usernameLen := utf8.RuneCountInString(username)
if usernameLen == 0 {
return email
}
prefixKeep := rule.PrefixKeep
if prefixKeep > usernameLen {
prefixKeep = usernameLen
}
if prefixKeep <= 0 {
prefixKeep = 1
}
maskCount := usernameLen - prefixKeep
if maskCount <= 0 {
maskCount = 1
}
usernameRunes := []rune(username)
result := make([]rune, 0, prefixKeep+maskCount)
for i := 0; i < prefixKeep; i++ {
result = append(result, usernameRunes[i])
}
for i := 0; i < maskCount; i++ {
result = append(result, maskChar)
}
return string(result) + domain
}

func (m *Masker) MaskName(name string, rule *MaskRule) string {
if rule == nil {
rule = m.rules.Name
}
return maskString(name, rule.PrefixKeep, rule.SuffixKeep)
}

func maskString(s string, prefixKeep, suffixKeep int) string {
runes := []rune(s)
n := len(runes)
if n == 0 {
return s
}
if prefixKeep < 0 {
prefixKeep = 0
}
if suffixKeep < 0 {
suffixKeep = 0
}
if prefixKeep >= n || suffixKeep >= n {
return s
}
if prefixKeep+suffixKeep >= n {
suffixKeep = n - prefixKeep - 1
if suffixKeep < 0 {
suffixKeep = 0
}
}
result := make([]rune, 0, n)
for i := 0; i < prefixKeep; i++ {
result = append(result, runes[i])
}
maskCount := n - prefixKeep - suffixKeep
for i := 0; i < maskCount; i++ {
result = append(result, maskChar)
}
for i := n - suffixKeep; i < n; i++ {
result = append(result, runes[i])
}
return string(result)
}

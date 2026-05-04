package password

var builtinWeakPasswords = map[string]bool{
	"123456":     true,
	"password":   true,
	"123456789":  true,
	"12345678":   true,
	"12345":      true,
	"qwerty":     true,
	"abc123":     true,
	"password1":  true,
	"111111":     true,
	"123123":     true,
	"admin":      true,
	"letmein":    true,
	"welcome":    true,
	"monkey":     true,
	"dragon":     true,
	"master":     true,
	"login":      true,
	"princess":   true,
	"sunshine":   true,
	"qwertyuiop": true,
	"password123": true,
	"654321":     true,
	"superman":   true,
	"asdfgh":     true,
	"qazwsx":     true,
	"1234567":    true,
	"1234":       true,
	"1234567890": true,
	"iloveyou":   true,
	"123321":     true,
	"666666":     true,
	"987654321":  true,
	"121212":     true,
	"000000":     true,
	"aaaaaa":     true,
	"abcdef":     true,
	"test":       true,
	"test123":    true,
	"testing":    true,
	"pass":       true,
	"passw0rd":   true,
	"password12":  true,
	"p@ssw0rd":   true,
	"1q2w3e4r":   true,
	"qwe123":     true,
	"asd123":     true,
	"zxc123":     true,
	"123qwe":     true,
	"123asd":     true,
	"123zxc":     true,
	"qweasd":     true,
	"asdzxc":     true,
}

func (e *Evaluator) isWeakPassword(password string) bool {
	lowerPwd := toLower(password)
	if builtinWeakPasswords[lowerPwd] {
		return true
	}
	if e.customWeakPasswords[lowerPwd] {
		return true
	}
	return false
}

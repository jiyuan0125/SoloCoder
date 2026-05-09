package rs

const (
	primitivePolynomial = 0x11D
	fieldSize           = 256
	generator           = 2
)

var (
	logTable  [fieldSize]int
	expTable  [fieldSize * 2]int
)

func init() {
	initTables()
}

func initTables() {
	x := 1
	for i := 0; i < fieldSize-1; i++ {
		expTable[i] = x
		logTable[x] = i
		x = gMul(x, generator)
	}
	for i := fieldSize - 1; i < fieldSize*2; i++ {
		expTable[i] = expTable[i-fieldSize+1]
	}
}

func gMul(a, b int) int {
	result := 0
	for b > 0 {
		if b&1 == 1 {
			result ^= a
		}
		a <<= 1
		if a&fieldSize != 0 {
			a ^= primitivePolynomial
		}
		b >>= 1
	}
	return result
}

func Add(a, b byte) byte {
	return a ^ b
}

func Mul(a, b byte) byte {
	if a == 0 || b == 0 {
		return 0
	}
	return byte(expTable[logTable[a]+logTable[b]])
}

func Div(a, b byte) byte {
	if b == 0 {
		panic("division by zero in GF(256)")
	}
	if a == 0 {
		return 0
	}
	return byte(expTable[(logTable[a]-logTable[b]+fieldSize-1)%(fieldSize-1)])
}

func Inv(a byte) byte {
	if a == 0 {
		panic("inverse of zero in GF(256)")
	}
	return byte(expTable[(fieldSize-1)-logTable[a]])
}

func Pow(a byte, n int) byte {
	if n == 0 {
		return 1
	}
	if a == 0 {
		return 0
	}
	return byte(expTable[(logTable[a]*n)%(fieldSize-1)])
}

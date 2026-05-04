package generator

import (
	"math/rand"
	"time"
)

const (
	charset    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeLength = 6
)

type Generator struct {
	r *rand.Rand
}

func NewGenerator() *Generator {
	return &Generator{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *Generator) Generate() string {
	b := make([]byte, codeLength)
	for i := range b {
		b[i] = charset[g.r.Intn(len(charset))]
	}
	return string(b)
}

type CodeExistsChecker func(code string) bool

func (g *Generator) GenerateUnique(existsChecker CodeExistsChecker) string {
	for {
		code := g.Generate()
		if !existsChecker(code) {
			return code
		}
	}
}

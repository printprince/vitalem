package codegen

import (
	"crypto/rand"
)

type Generator interface {
	GenerateCode() string
}

type CodeGenerator struct{}

func NewCodeGenerator() Generator {
	return &CodeGenerator{}
}

func (g *CodeGenerator) GenerateCode() string {
	const digits = "0123456789"
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		// fallback на простой генератор
		return "000000"
	}
	for i := 0; i < 6; i++ {
		b[i] = digits[int(b[i])%len(digits)]
	}
	return string(b)
}

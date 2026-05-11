package circular

type A struct {
	b *B
}

func NewA(b *B) *A {
	return &A{b: b}
}

package circular

type B struct {
	a *A
}

func NewB(a *A) *B {
	return &B{a: a}
}

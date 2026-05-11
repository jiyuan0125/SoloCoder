package testdata

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type MyStruct struct{}

func (m *MyStruct) Read(p []byte) (int, error) {
	return 0, nil
}

type AnotherStruct struct {
	MyStruct
}

func (a *AnotherStruct) Write(p []byte) (int, error) {
	return 0, nil
}

type IncompleteStruct struct{}

func (i *IncompleteStruct) Read(p []byte) (int, error) {
	return 0, nil
}

type VariadicStruct struct{}

func (v *VariadicStruct) Print(a int, b ...string) {}

type VariadicInterface interface {
	Print(a int, b []string)
}

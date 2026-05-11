package testdata

import "fmt"

type MyInterface interface {
	DoSomething() string
}

type MyStruct struct {
	Name   string
	UnusedField int
}

func (m *MyStruct) DoSomething() string {
	return m.Name
}

type UnusedType struct {
	Value int
}

type UsedAlias = MyStruct

type UnusedAlias = UnusedType

const (
	UsedConst   = "used"
	UnusedConst = "unused"
)

var (
	UsedVar     = "used"
	UnusedVar   = "unused"
	AssignedVar = ""
)

func init() {
	fmt.Println("init")
}

func UsedFunc() {
	fmt.Println(UsedConst, UsedVar)
	AssignedVar = "assigned"
}

func UnusedFunc() {
	fmt.Println("unused")
}

func ExportedUnusedFunc() {
	fmt.Println("exported but unused in package")
}

func main() {
	UsedFunc()
}

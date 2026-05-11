package main

import (
	"fmt"

	"gofmt-tool/formatter"
)

var testCode = `package main

import (
"fmt"
"github.com/gin-gonic/gin"
"github.com/go-sql-driver/mysql"
"net/http"
"os"
"strings"
)

func main() {
	fmt.Println("hello")
}
`

func main() {
	cfg := formatter.DefaultConfig()
	result, stats, err := formatter.Format(testCode, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("=== Original ===")
	fmt.Println(testCode)
	fmt.Println("\n=== Formatted ===")
	fmt.Println(result)
	fmt.Println("\n=== Stats ===")
	fmt.Printf("LinesModified: %d, ImportsMoved: %d, LinesSplit: %d\n",
		stats.LinesModified, stats.ImportsMoved, stats.LinesSplit)
}

package main

import (
	"flag"
	"fmt"
	"go-faq-service/internal/client"
	"os"
	"strings"
)

const defaultServerURL = "http://localhost:8080"

var (
	serverURL string
	lang      string
)

func main() {
	flag.StringVar(&serverURL, "server", defaultServerURL, "FAQ service server URL")
	flag.StringVar(&serverURL, "s", defaultServerURL, "FAQ service server URL (short)")
	flag.StringVar(&lang, "lang", "zh", "Language preference (zh/en)")
	flag.StringVar(&lang, "l", "zh", "Language preference (short)")

	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	apiClient := client.NewAPIClient(serverURL)

	command := strings.ToLower(args[0])

	switch command {
	case "health", "h":
		runHealthCheck(apiClient)
	case "create-faq", "cf":
		runCreateFAQ(apiClient, args[1:])
	case "get-faq", "gf":
		runGetFAQ(apiClient, args[1:])
	case "update-faq", "uf":
		runUpdateFAQ(apiClient, args[1:])
	case "delete-faq", "df":
		runDeleteFAQ(apiClient, args[1:])
	case "enable-faq", "ef":
		runEnableFAQ(apiClient, args[1:], true)
	case "disable-faq", "dfa":
		runEnableFAQ(apiClient, args[1:], false)
	case "pin-faq", "pf":
		runPinFAQ(apiClient, args[1:], true)
	case "unpin-faq", "ufq":
		runPinFAQ(apiClient, args[1:], false)
	case "search", "s":
		runSearch(apiClient, args[1:])
	case "batch-import", "bi":
		runBatchImport(apiClient, args[1:])
	case "record-click", "rc":
		runRecordClick(apiClient, args[1:])
	case "faq-history", "fh":
		runFAQHistory(apiClient, args[1:])
	case "create-category", "cc":
		runCreateCategory(apiClient, args[1:])
	case "list-categories", "lc":
		runListCategories(apiClient)
	case "category-tree", "ct":
		runCategoryTree(apiClient)
	case "update-category", "uc":
		runUpdateCategory(apiClient, args[1:])
	case "delete-category", "dc":
		runDeleteCategory(apiClient, args[1:])
	case "statistics", "stats":
		runStatistics(apiClient)
	case "list-faqs", "lf":
		runListFAQs(apiClient)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runHealthCheck(apiClient *client.APIClient) {
	resp, err := apiClient.HealthCheck()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatResponse(resp))
}

func runCreateFAQ(apiClient *client.APIClient, args []string) {
	fs := flag.NewFlagSet("create-faq", flag.ExitOnError)
	categoryID := fs.String("category", "", "Category ID (required)")
	questions := fs.String("questions", "", "Question(s), comma-separated for multiple languages")
	answers := fs.String("answers", "", "Answer(s), comma-separated for multiple languages")
	langs := fs.String("langs", "zh", "Language(s), comma-separated (zh/en)")

	fs.Usage = func() {
		fmt.Println("Usage: faq-cli create-faq [options]")
		fmt.Println("\nOptions:")
		fs.PrintDefaults()
		fmt.Println("\nExample:")
		fmt.Println("  faq-cli create-faq -category cat-123 -questions \"如何重置密码?,How to reset password?\" -answers \"请点击忘记密码链接,Click forgot password link\" -langs \"zh,en\"")
	}

	fs.Parse(args)

	if *categoryID == "" {
		fmt.Println("Error: -category is required")
		fs.Usage()
		os.Exit(1)
	}

	qList := splitAndTrim(*questions)
	aList := splitAndTrim(*answers)
	lList := splitAndTrim(*langs)

	content, err := client.ParseContentMap(qList, aList, lList)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := apiClient.CreateFAQ(*categoryID, content)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatResponse(resp))
}

func runGetFAQ(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: faq-cli get-faq <faq-id>")
		os.Exit(1)
	}

	faqID := args[0]
	resp, err := apiClient.GetFAQ(faqID, client.ParseLanguage(lang))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatFAQDetail(resp))
}

func runUpdateFAQ(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: faq-cli update-faq <faq-id> [options]")
		fmt.Println("\nOptions:")
		fmt.Println("  -questions   Question(s), comma-separated")
		fmt.Println("  -answers     Answer(s), comma-separated")
		fmt.Println("  -langs       Language(s), comma-separated (zh/en)")
		os.Exit(1)
	}

	faqID := args[0]
	fs := flag.NewFlagSet("update-faq", flag.ExitOnError)
	questions := fs.String("questions", "", "Question(s), comma-separated")
	answers := fs.String("answers", "", "Answer(s), comma-separated")
	langs := fs.String("langs", "zh", "Language(s), comma-separated")

	fs.Parse(args[1:])

	qList := splitAndTrim(*questions)
	aList := splitAndTrim(*answers)
	lList := splitAndTrim(*langs)

	if len(qList) == 0 && len(aList) == 0 {
		fmt.Println("Error: at least one of -questions or -answers is required")
		os.Exit(1)
	}

	content, err := client.ParseContentMap(qList, aList, lList)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := apiClient.UpdateFAQ(faqID, content)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatResponse(resp))
}

func runDeleteFAQ(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: faq-cli delete-faq <faq-id>")
		os.Exit(1)
	}

	resp, err := apiClient.DeleteFAQ(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatResponse(resp))
}

func runEnableFAQ(apiClient *client.APIClient, args []string, enable bool) {
	if len(args) < 1 {
		action := "enable"
		if !enable {
			action = "disable"
		}
		fmt.Printf("Usage: faq-cli %s-faq <faq-id>\n", action)
		os.Exit(1)
	}

	resp, err := apiClient.SetFAQEnabled(args[0], enable)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatResponse(resp))
}

func runPinFAQ(apiClient *client.APIClient, args []string, pin bool) {
	if len(args) < 1 {
		action := "pin"
		if !pin {
			action = "unpin"
		}
		fmt.Printf("Usage: faq-cli %s-faq <faq-id>\n", action)
		os.Exit(1)
	}

	resp, err := apiClient.SetFAQPinned(args[0], pin)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatResponse(resp))
}

func runSearch(apiClient *client.APIClient, args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	keyword := fs.String("keyword", "", "Search keyword")
	page := fs.Int("page", 1, "Page number")
	pageSize := fs.Int("page-size", 10, "Page size")

	fs.Usage = func() {
		fmt.Println("Usage: faq-cli search [options]")
		fmt.Println("\nOptions:")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	resp, err := apiClient.SearchFAQ(*keyword, client.ParseLanguage(lang), *page, *pageSize)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatSearchResult(resp))
}

func runBatchImport(apiClient *client.APIClient, args []string) {
	fmt.Println("Batch import requires JSON input. Example:")
	fmt.Println(`
{
  "items": [
    {
      "category_id": "cat-123",
      "content": {
        "zh": {"question": "问题1", "answer": "答案1"},
        "en": {"question": "Question 1", "answer": "Answer 1"}
      }
    }
  ]
}`)
	fmt.Println("\nUse the API directly for batch import.")
}

func runRecordClick(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: faq-cli record-click <faq-id> [options]")
		fmt.Println("\nOptions:")
		fmt.Println("  -user-id     User ID")
		fmt.Println("  -helpful     Mark as helpful (true/false)")
		os.Exit(1)
	}

	faqID := args[0]
	fs := flag.NewFlagSet("record-click", flag.ExitOnError)
	userID := fs.String("user-id", "anonymous", "User ID")
	helpful := fs.Bool("helpful", true, "Mark as helpful")

	fs.Parse(args[1:])

	resp, err := apiClient.RecordClick(faqID, *userID, *helpful)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatResponse(resp))
}

func runFAQHistory(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: faq-cli faq-history <faq-id>")
		os.Exit(1)
	}

	resp, err := apiClient.GetFAQHistory(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatResponse(resp))
}

func runCreateCategory(apiClient *client.APIClient, args []string) {
	fs := flag.NewFlagSet("create-category", flag.ExitOnError)
	name := fs.String("name", "", "Category name (required)")
	parentID := fs.String("parent", "", "Parent category ID (optional)")

	fs.Usage = func() {
		fmt.Println("Usage: faq-cli create-category [options]")
		fmt.Println("\nOptions:")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	if *name == "" {
		fmt.Println("Error: -name is required")
		fs.Usage()
		os.Exit(1)
	}

	resp, err := apiClient.CreateCategory(*name, *parentID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatResponse(resp))
}

func runListCategories(apiClient *client.APIClient) {
	resp, err := apiClient.ListAllCategories()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatResponse(resp))
}

func runCategoryTree(apiClient *client.APIClient) {
	resp, err := apiClient.GetCategoryTree()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatCategoryTree(resp))
}

func runUpdateCategory(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: faq-cli update-category <category-id> -name <new-name>")
		os.Exit(1)
	}

	catID := args[0]
	fs := flag.NewFlagSet("update-category", flag.ExitOnError)
	name := fs.String("name", "", "New category name")

	fs.Parse(args[1:])

	if *name == "" {
		fmt.Println("Error: -name is required")
		os.Exit(1)
	}

	resp, err := apiClient.UpdateCategory(catID, *name)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatResponse(resp))
}

func runDeleteCategory(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: faq-cli delete-category <category-id>")
		os.Exit(1)
	}

	resp, err := apiClient.DeleteCategory(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatResponse(resp))
}

func runStatistics(apiClient *client.APIClient) {
	resp, err := apiClient.GetStatistics()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatStatistics(resp))
}

func runListFAQs(apiClient *client.APIClient) {
	resp, err := apiClient.ListAllFAQs()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client.FormatResponse(resp))
}

func splitAndTrim(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func printUsage() {
	fmt.Println(`FAQ Service Command Line Client

Usage:
  faq-cli [global-options] <command> [command-options]

Global Options:
  -server, -s  Server URL (default: http://localhost:8080)
  -lang, -l    Language preference: zh or en (default: zh)

Commands:
  Management (Admin):
    create-faq, cf       Create a new FAQ
    update-faq, uf       Update an existing FAQ
    delete-faq, df       Delete an FAQ
    enable-faq, ef       Enable an FAQ
    disable-faq, dfa     Disable an FAQ
    pin-faq, pf          Pin an FAQ (priority in search)
    unpin-faq, ufq       Unpin an FAQ
    batch-import, bi     Batch import FAQs
    faq-history, fh      View FAQ change history
    create-category, cc  Create a category
    update-category, uc  Update a category
    delete-category, dc  Delete a category
    list-categories, lc  List all categories
    category-tree, ct    View category tree
    list-faqs, lf        List all FAQs (admin)
    statistics, stats    View system statistics

  User Queries:
    search, s            Search FAQs
    get-faq, gf          Get FAQ detail
    record-click, rc     Record click feedback
    health, h            Health check

Examples:
  # Search FAQs
  faq-cli search -keyword "密码"

  # Create category
  faq-cli create-category -name "账户问题"

  # Create FAQ with Chinese and English
  faq-cli create-faq -category <cat-id> \
    -questions "如何重置密码?,How to reset password?" \
    -answers "请点击忘记密码,Click forgot password" \
    -langs "zh,en"

  # Get FAQ detail
  faq-cli get-faq <faq-id>

  # View statistics
  faq-cli statistics

Use 'faq-cli <command> -help' for command-specific help.`)
}

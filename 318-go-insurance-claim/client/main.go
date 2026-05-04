package main

import (
	"fmt"
	"os"
)

func main() {
	client := NewAPIClient("http://localhost:8080")
	cli := NewCLI(client)

	for {
		fmt.Println("\n=====================================")
		fmt.Println("   Insurance Claim Management CLI   ")
		fmt.Println("=====================================")
		fmt.Println("\n=== User Operations ===")
		fmt.Println("  1. Submit a Claim")
		fmt.Println("  2. Check Claim Status")
		fmt.Println("  3. View Policy Claims History")
		fmt.Println("\n=== Admin Operations ===")
		fmt.Println("  4. Create New Policy")
		fmt.Println("  5. List Pending Claims")
		fmt.Println("  6. Review a Claim (Approve/Reject)")
		fmt.Println("  7. List Approved Claims")
		fmt.Println("  8. Confirm Payment")
		fmt.Println("\n=== View All Data ===")
		fmt.Println("  9. List All Policies")
		fmt.Println(" 10. List All Claims")
		fmt.Println("\n  0. Exit")
		fmt.Println("=====================================")

		choice := cli.readInput("\nSelect an option (0-10): ")

		switch choice {
		case "1":
			cli.SubmitClaim()
		case "2":
			cli.GetClaimStatus()
		case "3":
			cli.GetPolicyClaims()
		case "4":
			cli.CreatePolicy()
		case "5":
			cli.ListPendingClaims()
		case "6":
			cli.ReviewClaim()
		case "7":
			cli.ListApprovedClaims()
		case "8":
			cli.ConfirmPayment()
		case "9":
			cli.ListAllPolicies()
		case "10":
			cli.ListAllClaims()
		case "0":
			fmt.Println("\nGoodbye!")
			os.Exit(0)
		default:
			fmt.Println("\nInvalid option. Please try again.")
		}

		fmt.Print("\nPress Enter to continue...")
		cli.reader.ReadString('\n')
	}
}

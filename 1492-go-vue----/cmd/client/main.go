package main

import (
	"fmt"
	"os"
)

func printUsage() {
	fmt.Println(`Renovation Management CLI

Usage:
  renovation <command> [arguments]

House Commands:
  create-house <area> <layout> <floor> <orientation>   Create a new house
  list-houses                                          List all houses
  get-house <id>                                       Get house details

Design Commands:
  create-design <house_id> <version> [description]     Create a design scheme
  get-design <id>                                      Get design details
  confirm-design <scheme_id>                           Confirm a design scheme

Quotation Commands:
  create-quotation <scheme_id>                         Create a quotation
  list-quotations                                      List all quotations
  get-quotation <id>                                   Get quotation details

Change Order Commands:
  create-change <quotation_id> <content> <reason> <amount_diff_fen>  Create change order
  confirm-change <change_id>                           Confirm a change order

Construction Commands:
  create-phase <quotation_id> <category> <order_index> <start_date> <end_date>  Create phase
  update-progress <phase_id> <progress> <updated_by>   Update phase progress
  get-phase <id>                                       Get phase details
  list-delayed-phases                                  List delayed phases

Settlement Commands:
  calculate-settlement <quotation_id>                  Calculate settlement
  get-settlement <id>                                  Get settlement details

Environment:
  SERVER_URL   Server base URL (default: http://localhost:8080)
`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	client := NewClient(getBaseURL())

	var err error
	switch cmd {
	case "create-house":
		err = client.cmdCreateHouse(args)
	case "list-houses":
		err = client.cmdListHouses(args)
	case "get-house":
		err = client.cmdGetHouse(args)
	case "create-design":
		err = client.cmdCreateDesign(args)
	case "get-design":
		err = client.cmdGetDesign(args)
	case "confirm-design":
		err = client.cmdConfirmDesign(args)
	case "create-quotation":
		err = client.cmdCreateQuotation(args)
	case "list-quotations":
		err = client.cmdListQuotations(args)
	case "get-quotation":
		err = client.cmdGetQuotation(args)
	case "create-change":
		err = client.cmdCreateChange(args)
	case "confirm-change":
		err = client.cmdConfirmChange(args)
	case "create-phase":
		err = client.cmdCreatePhase(args)
	case "update-progress":
		err = client.cmdUpdateProgress(args)
	case "get-phase":
		err = client.cmdGetPhase(args)
	case "list-delayed-phases":
		err = client.cmdListDelayedPhases(args)
	case "calculate-settlement":
		err = client.cmdCalculateSettlement(args)
	case "get-settlement":
		err = client.cmdGetSettlement(args)
	case "help", "-h", "--help":
		printUsage()
		return
	default:
		fmt.Printf("unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

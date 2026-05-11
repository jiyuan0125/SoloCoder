package main

import (
	"flag"
	"fmt"
	"os"
	"parking-system/common"
	"time"
)

const defaultServerURL = "http://localhost:8904"

func main() {
	serverURL := flag.String("server", defaultServerURL, "Server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	client := NewAPIClient(*serverURL)
	cmd := args[0]

	var err error
	switch cmd {
	case "create-spot":
		err = cmdCreateSpot(client, args[1:])
	case "list-spots":
		err = cmdListSpots(client, args[1:])
	case "update-spot-status":
		err = cmdUpdateSpotStatus(client, args[1:])
	case "guidance":
		err = cmdGuidance(client)
	case "checkin":
		err = cmdCheckIn(client, args[1:])
	case "checkout":
		err = cmdCheckOut(client, args[1:])
	case "query-fee":
		err = cmdQueryFee(client, args[1:])
	case "create-card":
		err = cmdCreateCard(client, args[1:])
	case "renew-card":
		err = cmdRenewCard(client, args[1:])
	case "list-cards":
		err = cmdListCards(client, args[1:])
	case "list-records":
		err = cmdListRecords(client, args[1:])
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Parking System Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [flags] <command> [args]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -server string    Server URL (default: http://localhost:8904)")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create-spot <id> <area> <number> <type>    Create a parking spot (type: normal/charging/accessible)")
	fmt.Println("  list-spots [area] [status]                 List parking spots")
	fmt.Println("  update-spot-status <id> <status>           Update spot status (available/occupied/reserved/maintenance)")
	fmt.Println("  guidance                                   Get parking guidance info")
	fmt.Println("  checkin <plate_number>                     Check in a vehicle")
	fmt.Println("  checkout <plate_number>                    Check out a vehicle")
	fmt.Println("  query-fee <plate_number>                   Query current fee for active parking")
	fmt.Println("  create-card <type> <owner> <plate> [spot]  Create monthly card (type: normal/fixed_spot/charging)")
	fmt.Println("  renew-card <card_id>                       Renew a monthly card")
	fmt.Println("  list-cards [plate] [active_only]           List monthly cards")
	fmt.Println("  list-records [plate] [active_only]         List parking records")
}

func cmdCreateSpot(client *APIClient, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: create-spot <id> <area> <number> <type>")
	}

	spotType := common.ParkingSpotType(args[3])
	switch spotType {
	case common.SpotTypeNormal, common.SpotTypeCharging, common.SpotTypeAccessible:
	default:
		return fmt.Errorf("invalid spot type: %s", args[3])
	}

	req := common.CreateParkingSpotRequest{
		ID:     args[0],
		Area:   args[1],
		Number: args[2],
		Type:   spotType,
	}

	spot, err := client.CreateSpot(req)
	if err != nil {
		return err
	}

	fmt.Printf("Created spot: ID=%s, Area=%s, Number=%s, Type=%s, Status=%s\n",
		spot.ID, spot.Area, spot.Number, spot.Type, spot.Status)
	return nil
}

func cmdListSpots(client *APIClient, args []string) error {
	var area, status string
	if len(args) >= 1 {
		area = args[0]
	}
	if len(args) >= 2 {
		status = args[1]
	}

	spots, err := client.ListSpots(area, status)
	if err != nil {
		return err
	}

	if len(spots) == 0 {
		fmt.Println("No spots found")
		return nil
	}

	fmt.Printf("%-15s %-10s %-8s %-12s %-15s\n", "ID", "Area", "Number", "Type", "Status")
	fmt.Println("------------------------------------------------------------")
	for _, s := range spots {
		fmt.Printf("%-15s %-10s %-8s %-12s %-15s\n",
			s.ID, s.Area, s.Number, s.Type, s.Status)
	}
	return nil
}

func cmdUpdateSpotStatus(client *APIClient, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: update-spot-status <id> <status>")
	}

	status := common.ParkingSpotStatus(args[1])
	switch status {
	case common.SpotStatusAvailable, common.SpotStatusOccupied,
		common.SpotStatusReserved, common.SpotStatusMaintenance:
	default:
		return fmt.Errorf("invalid status: %s", args[1])
	}

	spot, err := client.UpdateSpotStatus(args[0], status)
	if err != nil {
		return err
	}

	fmt.Printf("Updated spot %s status to: %s\n", spot.ID, spot.Status)
	return nil
}

func cmdGuidance(client *APIClient) error {
	guidance, err := client.GetGuidance()
	if err != nil {
		return err
	}

	fmt.Printf("Total Spots: %d\n", guidance.TotalSpots)
	fmt.Printf("Available Spots: %d (%.1f%%)\n", guidance.AvailableSpots, guidance.AvailablePercent)
	fmt.Printf("Is Full: %v\n", guidance.IsFull)
	fmt.Printf("Is Tight: %v\n", guidance.IsTight)
	fmt.Println()
	fmt.Println("Areas:")
	fmt.Printf("%-15s %-12s %-15s %-8s\n", "Area", "Total", "Available", "Is Full")
	fmt.Println("------------------------------------------------")
	for _, a := range guidance.Areas {
		fmt.Printf("%-15s %-12d %-15d %-8v\n",
			a.Name, a.TotalSpots, a.AvailableSpots, a.IsFull)
	}
	return nil
}

func cmdCheckIn(client *APIClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: checkin <plate_number>")
	}

	resp, err := client.CheckIn(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Check-in successful\n")
	fmt.Printf("  Plate: %s\n", resp.PlateNumber)
	fmt.Printf("  Time: %s\n", resp.CheckInTime.Format(time.RFC3339))
	fmt.Printf("  Spot: %s\n", resp.SpotID)
	fmt.Printf("  Monthly Card: %v\n", resp.IsMonthlyCard)
	return nil
}

func cmdCheckOut(client *APIClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: checkout <plate_number>")
	}

	resp, err := client.CheckOut(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Check-out successful\n")
	fmt.Printf("  Plate: %s\n", resp.PlateNumber)
	fmt.Printf("  Check-in: %s\n", resp.CheckInTime.Format(time.RFC3339))
	fmt.Printf("  Check-out: %s\n", resp.CheckOutTime.Format(time.RFC3339))
	fmt.Printf("  Duration: %.1f minutes\n", resp.Duration)
	fmt.Printf("  Monthly Card: %v\n", resp.IsMonthlyCard)
	if !resp.IsMonthlyCard {
		fmt.Printf("  Free Time Used: %.1f minutes\n", resp.FreeTimeUsed)
		fmt.Printf("  Amount: ¥%.2f\n", resp.Amount)
	}
	return nil
}

func cmdQueryFee(client *APIClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: query-fee <plate_number>")
	}

	resp, err := client.QueryFee(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Fee query for %s\n", resp.PlateNumber)
	fmt.Printf("  Check-in: %s\n", resp.CheckInTime.Format(time.RFC3339))
	fmt.Printf("  Current: %s\n", resp.CurrentTime.Format(time.RFC3339))
	fmt.Printf("  Duration: %.1f minutes\n", resp.Duration)
	fmt.Printf("  Monthly Card: %v\n", resp.IsMonthlyCard)
	if !resp.IsMonthlyCard {
		fmt.Printf("  Estimated Amount: ¥%.2f\n", resp.EstimatedAmount)
	}
	return nil
}

func cmdCreateCard(client *APIClient, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: create-card <type> <owner> <plate> [spot]")
	}

	cardType := common.MonthlyCardType(args[0])
	switch cardType {
	case common.CardTypeNormal, common.CardTypeFixedSpot, common.CardTypeCharging:
	default:
		return fmt.Errorf("invalid card type: %s", args[0])
	}

	var spotID string
	if len(args) >= 4 {
		spotID = args[3]
	}

	if cardType == common.CardTypeFixedSpot && spotID == "" {
		return fmt.Errorf("fixed_spot card requires spot id")
	}

	req := common.CreateMonthlyCardRequest{
		CardType:    cardType,
		OwnerName:   args[1],
		PlateNumber: args[2],
		SpotID:      spotID,
	}

	card, err := client.CreateMonthlyCard(req)
	if err != nil {
		return err
	}

	fmt.Printf("Created monthly card\n")
	fmt.Printf("  ID: %s\n", card.ID)
	fmt.Printf("  Type: %s\n", card.CardType)
	fmt.Printf("  Owner: %s\n", card.OwnerName)
	fmt.Printf("  Plate: %s\n", card.PlateNumber)
	if card.SpotID != "" {
		fmt.Printf("  Spot: %s\n", card.SpotID)
	}
	fmt.Printf("  Start: %s\n", card.StartDate.Format(time.RFC3339))
	fmt.Printf("  End: %s\n", card.EndDate.Format(time.RFC3339))
	fmt.Printf("  Grace End: %s\n", card.GraceEndDate.Format(time.RFC3339))
	fmt.Printf("  Status: %s\n", card.Status)
	return nil
}

func cmdRenewCard(client *APIClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: renew-card <card_id>")
	}

	card, err := client.RenewMonthlyCard(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Renewed card %s\n", card.ID)
	fmt.Printf("  New End: %s\n", card.EndDate.Format(time.RFC3339))
	fmt.Printf("  New Grace End: %s\n", card.GraceEndDate.Format(time.RFC3339))
	return nil
}

func cmdListCards(client *APIClient, args []string) error {
	var plate string
	activeOnly := false
	if len(args) >= 1 {
		plate = args[0]
	}
	if len(args) >= 2 && args[1] == "active_only" {
		activeOnly = true
	}

	cards, err := client.ListMonthlyCards(plate, activeOnly)
	if err != nil {
		return err
	}

	if len(cards) == 0 {
		fmt.Println("No cards found")
		return nil
	}

	fmt.Printf("%-20s %-12s %-10s %-12s %-20s %-8s\n",
		"ID", "Type", "Plate", "Spot", "End Date", "Status")
	fmt.Println("--------------------------------------------------------------------------------------")
	for _, c := range cards {
		fmt.Printf("%-20s %-12s %-10s %-12s %-20s %-8s\n",
			c.ID, c.CardType, c.PlateNumber, c.SpotID,
			c.EndDate.Format("2006-01-02"), c.Status)
	}
	return nil
}

func cmdListRecords(client *APIClient, args []string) error {
	var plate string
	activeOnly := false
	if len(args) >= 1 {
		plate = args[0]
	}
	if len(args) >= 2 && args[1] == "active_only" {
		activeOnly = true
	}

	records, err := client.ListParkingRecords(plate, activeOnly)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		fmt.Println("No records found")
		return nil
	}

	fmt.Printf("%-12s %-10s %-20s %-20s %-8s %-8s\n",
		"Plate", "Spot", "Check-in", "Check-out", "Amount", "Active")
	fmt.Println("-------------------------------------------------------------------------------")
	for _, r := range records {
		checkOut := "-"
		if !r.CheckOutTime.IsZero() {
			checkOut = r.CheckOutTime.Format("2006-01-02 15:04")
		}
		amount := "-"
		if r.Amount > 0 {
			amount = fmt.Sprintf("¥%.2f", r.Amount)
		}
		fmt.Printf("%-12s %-10s %-20s %-20s %-8s %-8v\n",
			r.PlateNumber, r.SpotID,
			r.CheckInTime.Format("2006-01-02 15:04"),
			checkOut, amount, r.IsActive)
	}
	return nil
}

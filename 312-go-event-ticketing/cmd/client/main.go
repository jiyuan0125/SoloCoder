package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"event-ticketing/internal/client"
	"event-ticketing/pkg/common"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Parse()

	cmd := flag.Arg(0)
	cli := client.NewClient(*serverURL)

	switch cmd {
	case "create-event":
		handleCreateEvent(cli)
	case "list-events":
		handleListEvents(cli)
	case "purchase":
		handlePurchase(cli)
	case "checkin":
		handleCheckIn(cli)
	case "refund":
		handleRefund(cli)
	case "stats":
		handleStats(cli)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func handleCreateEvent(cli *client.Client) {
	eventName := os.Getenv("EVENT_NAME")
	eventTime := os.Getenv("EVENT_TIME")
	eventLocation := os.Getenv("EVENT_LOCATION")
	tiersInput := os.Getenv("TIERS")

	if eventName == "" || eventTime == "" || eventLocation == "" || tiersInput == "" {
		fmt.Println("Error: EVENT_NAME, EVENT_TIME, EVENT_LOCATION, TIERS environment variables must be set")
		fmt.Println("\nUsage:")
		fmt.Println("  EVENT_NAME=\"演唱会\" EVENT_TIME=\"2026-06-01T20:00:00\" EVENT_LOCATION=\"上海体育馆\" TIERS=\"VIP:500:100,普通票:200:500,学生票:100:200\" client create-event")
		fmt.Println("\nTIERS format: 名称1:价格1:数量1,名称2:价格2:数量2,...")
		os.Exit(1)
	}

	parsedTime, err := time.Parse("2006-01-02T15:04:05", eventTime)
	if err != nil {
		fmt.Printf("Error parsing time: %v\n", err)
		fmt.Println("Time format: 2006-01-02T15:04:05 (e.g., 2026-06-01T20:00:00)")
		os.Exit(1)
	}

	tiers, err := parseTiers(tiersInput)
	if err != nil {
		fmt.Printf("Error parsing tiers: %v\n", err)
		os.Exit(1)
	}

	req := common.CreateEventRequest{
		Name:     eventName,
		Time:     parsedTime,
		Location: eventLocation,
		Tiers:    tiers,
	}

	resp, err := cli.CreateEvent(req)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Println("Event created successfully!")
	fmt.Printf("Event ID: %s\n", resp.EventID)
}

func handleListEvents(cli *client.Client) {
	resp, err := cli.ListEvents()
	if err != nil {
		fmt.Printf("Error listing events: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	if len(resp.Events) == 0 {
		fmt.Println("No events found.")
		return
	}

	fmt.Println("Events:")
	for _, event := range resp.Events {
		fmt.Printf("\n  ID: %s\n", event.ID)
		fmt.Printf("  Name: %s\n", event.Name)
		fmt.Printf("  Time: %s\n", event.Time.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Location: %s\n", event.Location)
		fmt.Println("  Tiers:")
		for _, tier := range event.Tiers {
			fmt.Printf("    - %s: ¥%d (Capacity: %d, Available: %d)\n",
				tier.Name, tier.Price, tier.Capacity, tier.Available)
		}
	}
}

func handlePurchase(cli *client.Client) {
	eventID := os.Getenv("EVENT_ID")
	tierName := os.Getenv("TIER_NAME")
	quantityStr := os.Getenv("QUANTITY")

	if eventID == "" || tierName == "" || quantityStr == "" {
		fmt.Println("Error: EVENT_ID, TIER_NAME, QUANTITY environment variables must be set")
		fmt.Println("\nUsage:")
		fmt.Println("  EVENT_ID=\"EVT123456\" TIER_NAME=\"VIP\" QUANTITY=2 client purchase")
		os.Exit(1)
	}

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		fmt.Printf("Error parsing quantity: %v\n", err)
		os.Exit(1)
	}

	req := common.PurchaseTicketRequest{
		EventID:  eventID,
		TierName: tierName,
		Quantity: quantity,
	}

	resp, err := cli.PurchaseTicket(req)
	if err != nil {
		fmt.Printf("Error purchasing tickets: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Println("Tickets purchased successfully!")
	fmt.Println("Ticket numbers:")
	for _, ticketNumber := range resp.TicketNumbers {
		fmt.Printf("  - %s\n", ticketNumber)
	}
}

func handleCheckIn(cli *client.Client) {
	ticketNumber := os.Getenv("TICKET_NUMBER")

	if ticketNumber == "" {
		fmt.Println("Error: TICKET_NUMBER environment variable must be set")
		fmt.Println("\nUsage:")
		fmt.Println("  TICKET_NUMBER=\"TKT123456\" client checkin")
		os.Exit(1)
	}

	req := common.CheckInRequest{
		TicketNumber: ticketNumber,
	}

	resp, err := cli.CheckIn(req)
	if err != nil {
		fmt.Printf("Error checking in: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Println("Check-in successful!")
	fmt.Println(resp.Message)
}

func handleRefund(cli *client.Client) {
	ticketNumber := os.Getenv("TICKET_NUMBER")

	if ticketNumber == "" {
		fmt.Println("Error: TICKET_NUMBER environment variable must be set")
		fmt.Println("\nUsage:")
		fmt.Println("  TICKET_NUMBER=\"TKT123456\" client refund")
		os.Exit(1)
	}

	req := common.RefundTicketRequest{
		TicketNumber: ticketNumber,
	}

	resp, err := cli.RefundTicket(req)
	if err != nil {
		fmt.Printf("Error refunding ticket: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Println("Refund successful!")
	fmt.Println(resp.Message)
}

func handleStats(cli *client.Client) {
	eventID := os.Getenv("EVENT_ID")

	if eventID == "" {
		fmt.Println("Error: EVENT_ID environment variable must be set")
		fmt.Println("\nUsage:")
		fmt.Println("  EVENT_ID=\"EVT123456\" client stats")
		os.Exit(1)
	}

	resp, err := cli.GetEventStats(eventID)
	if err != nil {
		fmt.Printf("Error getting stats: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	stats := resp.Stats
	fmt.Printf("Event: %s (ID: %s)\n", stats.EventName, stats.EventID)
	fmt.Println("\nTier Statistics:")
	for _, ts := range stats.TierStats {
		fmt.Printf("\n  %s (¥%d):\n", ts.Name, ts.Price)
		fmt.Printf("    Total Capacity: %d\n", ts.Capacity)
		fmt.Printf("    Sold: %d\n", ts.Sold)
		fmt.Printf("    Available: %d\n", ts.Available)
		fmt.Printf("    Checked In: %d\n", ts.CheckedIn)
	}
}

func parseTiers(input string) ([]common.CreateTierRequest, error) {
	var tiers []common.CreateTierRequest
	tierStrings := strings.Split(input, ",")

	for _, tierStr := range tierStrings {
		parts := strings.Split(strings.TrimSpace(tierStr), ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid tier format: %s (expected: name:price:capacity)", tierStr)
		}

		name := strings.TrimSpace(parts[0])
		price, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid price: %s", parts[1])
		}

		capacity, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err != nil {
			return nil, fmt.Errorf("invalid capacity: %s", parts[2])
		}

		tiers = append(tiers, common.CreateTierRequest{
			Name:     name,
			Price:    price,
			Capacity: capacity,
		})
	}

	return tiers, nil
}

func printUsage() {
	fmt.Println("Event Ticketing Client")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create-event   Create a new event (organizer)")
	fmt.Println("  list-events    List all events")
	fmt.Println("  purchase       Purchase tickets (user)")
	fmt.Println("  checkin        Check in with a ticket number (user)")
	fmt.Println("  refund         Refund a ticket (user)")
	fmt.Println("  stats          Get event statistics (organizer)")
	fmt.Println()
	fmt.Println("Global Flags:")
	fmt.Println("  -server        Server URL (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  Create event:")
	fmt.Println("    EVENT_NAME=\"演唱会\" EVENT_TIME=\"2026-06-01T20:00:00\" EVENT_LOCATION=\"上海体育馆\" TIERS=\"VIP:500:100,普通票:200:500\" client create-event")
	fmt.Println()
	fmt.Println("  List events:")
	fmt.Println("    client list-events")
	fmt.Println()
	fmt.Println("  Purchase tickets:")
	fmt.Println("    EVENT_ID=\"EVT123456\" TIER_NAME=\"VIP\" QUANTITY=2 client purchase")
	fmt.Println()
	fmt.Println("  Check in:")
	fmt.Println("    TICKET_NUMBER=\"TKT123456\" client checkin")
	fmt.Println()
	fmt.Println("  Refund ticket:")
	fmt.Println("    TICKET_NUMBER=\"TKT123456\" client refund")
	fmt.Println()
	fmt.Println("  Get stats:")
	fmt.Println("    EVENT_ID=\"EVT123456\" client stats")
}

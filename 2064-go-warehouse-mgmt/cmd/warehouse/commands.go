package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"warehouse-mgmt/internal/report"
)

var inboundCmd = &cobra.Command{
	Use:   "inbound",
	Short: "Manage inbound orders",
}

var inboundCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an inbound order",
	RunE: func(cmd *cobra.Command, args []string) error {
		productID, _ := cmd.Flags().GetString("product")
		quantity, _ := cmd.Flags().GetInt("quantity")
		dateStr, _ := cmd.Flags().GetString("date")

		var date time.Time
		if dateStr != "" {
			var err error
			date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
			}
		} else {
			date = time.Now()
		}

		db, err := store.Load()
		if err != nil {
			return err
		}

		order, err := inboundSvc.CreateOrder(db, productID, quantity, date)
		if err != nil {
			return err
		}

		if err := store.Save(db); err != nil {
			return err
		}

		fmt.Printf("Inbound order created: %s\n", order.ID)
		return nil
	},
}

var inboundProcessCmd = &cobra.Command{
	Use:   "process",
	Short: "Process an inbound order",
	RunE: func(cmd *cobra.Command, args []string) error {
		orderID, _ := cmd.Flags().GetString("id")

		db, err := store.Load()
		if err != nil {
			return err
		}

		order, err := inboundSvc.ProcessOrder(db, orderID)
		if err != nil {
			return err
		}

		if err := store.Save(db); err != nil {
			return err
		}

		fmt.Printf("Inbound order processed: %s\n", order.ID)
		return nil
	},
}

var outboundCmd = &cobra.Command{
	Use:   "outbound",
	Short: "Manage outbound orders",
}

var outboundCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an outbound order",
	RunE: func(cmd *cobra.Command, args []string) error {
		productID, _ := cmd.Flags().GetString("product")
		quantity, _ := cmd.Flags().GetInt("quantity")

		db, err := store.Load()
		if err != nil {
			return err
		}

		order, err := outboundSvc.CreateOrder(db, productID, quantity)
		if err != nil {
			return err
		}

		if err := store.Save(db); err != nil {
			return err
		}

		fmt.Printf("Outbound order created: %s\n", order.ID)
		return nil
	},
}

var outboundProcessCmd = &cobra.Command{
	Use:   "process",
	Short: "Process an outbound order",
	RunE: func(cmd *cobra.Command, args []string) error {
		orderID, _ := cmd.Flags().GetString("id")

		db, err := store.Load()
		if err != nil {
			return err
		}

		order, err := outboundSvc.ProcessOrder(db, orderID)
		if err != nil {
			return err
		}

		if err := store.Save(db); err != nil {
			return err
		}

		fmt.Printf("Outbound order processed: %s\n", order.ID)
		return nil
	},
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query inventory information",
}

var queryLocationCmd = &cobra.Command{
	Use:   "location",
	Short: "Query products in a location",
	RunE: func(cmd *cobra.Command, args []string) error {
		code, _ := cmd.Flags().GetString("code")

		db, err := store.Load()
		if err != nil {
			return err
		}

		records := db.GetLocationProducts(code)
		if len(records) == 0 {
			fmt.Println("[]")
			return nil
		}

		fmt.Printf("Products in location %s:\n", code)
		for _, r := range records {
			fmt.Printf("  Product: %s, Quantity: %d\n", r.ProductID, r.Quantity)
		}
		return nil
	},
}

var queryProductCmd = &cobra.Command{
	Use:   "product",
	Short: "Query locations for a product",
	RunE: func(cmd *cobra.Command, args []string) error {
		productID, _ := cmd.Flags().GetString("id")

		db, err := store.Load()
		if err != nil {
			return err
		}

		records := db.GetProductLocations(productID)
		if len(records) == 0 {
			fmt.Println("[]")
			return nil
		}

		fmt.Printf("Locations for product %s:\n", productID)
		for _, r := range records {
			fmt.Printf("  Location: %s, Quantity: %d\n", r.LocationCode, r.Quantity)
		}
		return nil
	},
}

var countCmd = &cobra.Command{
	Use:   "count",
	Short: "Create stock count report",
	RunE: func(cmd *cobra.Command, args []string) error {
		itemsStr, _ := cmd.Flags().GetString("items")

		items := strings.Split(itemsStr, ",")
		var countItems []report.CountItem
		for _, itemStr := range items {
			parts := strings.Split(strings.TrimSpace(itemStr), ":")
			if len(parts) != 3 {
				continue
			}
			productID := parts[0]
			locationCode := parts[1]
			var qty int
			fmt.Sscanf(parts[2], "%d", &qty)
			countItems = append(countItems, report.CountItem{
				ProductID:    productID,
				LocationCode: locationCode,
				Quantity:     qty,
			})
		}

		db, err := store.Load()
		if err != nil {
			return err
		}

		rpt := reportSvc.CreateCountReport(db, countItems)

		if err := store.Save(db); err != nil {
			return err
		}

		reportSvc.PrintReport(rpt)
		return nil
	},
}

var alertCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Check stock alerts",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := store.Load()
		if err != nil {
			return err
		}

		alerts := inventorySvc.GetAlerts(db)
		if len(alerts) == 0 {
			fmt.Println("No stock alerts")
			return nil
		}

		fmt.Println("Stock alerts:")
		for _, productID := range alerts {
			product := db.GetProductByID(productID)
			total := db.GetTotalStock(productID)
			fmt.Printf("  Product: %s (%s), Current: %d, Safety: %d\n",
				productID, product.Name, total, product.SafetyStock)
		}
		return nil
	},
}

func init() {
	inboundCreateCmd.Flags().String("product", "", "Product ID")
	inboundCreateCmd.Flags().Int("quantity", 0, "Quantity")
	inboundCreateCmd.Flags().String("date", "", "Manufacture date (YYYY-MM-DD)")
	inboundCreateCmd.MarkFlagRequired("product")
	inboundCreateCmd.MarkFlagRequired("quantity")

	inboundProcessCmd.Flags().String("id", "", "Order ID")
	inboundProcessCmd.MarkFlagRequired("id")

	inboundCmd.AddCommand(inboundCreateCmd, inboundProcessCmd)

	outboundCreateCmd.Flags().String("product", "", "Product ID")
	outboundCreateCmd.Flags().Int("quantity", 0, "Quantity")
	outboundCreateCmd.MarkFlagRequired("product")
	outboundCreateCmd.MarkFlagRequired("quantity")

	outboundProcessCmd.Flags().String("id", "", "Order ID")
	outboundProcessCmd.MarkFlagRequired("id")

	outboundCmd.AddCommand(outboundCreateCmd, outboundProcessCmd)

	queryLocationCmd.Flags().String("code", "", "Location code")
	queryLocationCmd.MarkFlagRequired("code")

	queryProductCmd.Flags().String("id", "", "Product ID")
	queryProductCmd.MarkFlagRequired("id")

	queryCmd.AddCommand(queryLocationCmd, queryProductCmd)

	countCmd.Flags().String("items", "", "Count items: product1:loc1:qty1,product2:loc2:qty2")
	countCmd.MarkFlagRequired("items")
}

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var productCmd = &cobra.Command{
	Use:   "product",
	Short: "Manage products",
}

var productAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new product",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")
		safetyStock, _ := cmd.Flags().GetInt("safety-stock")

		db, err := store.Load()
		if err != nil {
			return err
		}

		if err := inventorySvc.AddProduct(db, id, name, safetyStock); err != nil {
			return err
		}

		if err := store.Save(db); err != nil {
			return err
		}

		fmt.Printf("Product added: %s (%s)\n", id, name)
		return nil
	},
}

var productListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all products",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := store.Load()
		if err != nil {
			return err
		}

		if len(db.Products) == 0 {
			fmt.Println("[]")
			return nil
		}

		fmt.Println("Products:")
		for _, p := range db.Products {
			total := db.GetTotalStock(p.ID)
			alert, _ := inventorySvc.CheckStockAlert(db, p.ID)
			alertStr := ""
			if alert {
				alertStr = " [ALERT: Low stock]"
			}
			fmt.Printf("  ID: %s, Name: %s, SafetyStock: %d, TotalStock: %d%s\n",
				p.ID, p.Name, p.SafetyStock, total, alertStr)
		}
		return nil
	},
}

var locationCmd = &cobra.Command{
	Use:   "location",
	Short: "Manage locations",
}

var locationAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new location",
	RunE: func(cmd *cobra.Command, args []string) error {
		code, _ := cmd.Flags().GetString("code")
		capacity, _ := cmd.Flags().GetInt("capacity")

		db, err := store.Load()
		if err != nil {
			return err
		}

		if err := locationSvc.CreateLocation(db, code, capacity); err != nil {
			return err
		}

		if err := store.Save(db); err != nil {
			return err
		}

		fmt.Printf("Location added: %s (capacity: %d)\n", code, capacity)
		return nil
	},
}

var locationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := store.Load()
		if err != nil {
			return err
		}

		if len(db.Locations) == 0 {
			fmt.Println("[]")
			return nil
		}

		fmt.Println("Locations:")
		for _, l := range db.Locations {
			available := l.Capacity - l.Used
			fmt.Printf("  Code: %s, Capacity: %d, Used: %d, Available: %d\n",
				l.Code, l.Capacity, l.Used, available)
		}
		return nil
	},
}

func init() {
	productAddCmd.Flags().String("id", "", "Product ID")
	productAddCmd.Flags().String("name", "", "Product name")
	productAddCmd.Flags().Int("safety-stock", 0, "Safety stock level")
	productAddCmd.MarkFlagRequired("id")
	productAddCmd.MarkFlagRequired("name")
	productCmd.AddCommand(productAddCmd, productListCmd)

	locationAddCmd.Flags().String("code", "", "Location code")
	locationAddCmd.Flags().Int("capacity", 0, "Location capacity")
	locationAddCmd.MarkFlagRequired("code")
	locationAddCmd.MarkFlagRequired("capacity")
	locationCmd.AddCommand(locationAddCmd, locationListCmd)
}

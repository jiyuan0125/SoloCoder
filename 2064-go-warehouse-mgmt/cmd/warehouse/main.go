package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"warehouse-mgmt/internal/inbound"
	"warehouse-mgmt/internal/inventory"
	"warehouse-mgmt/internal/location"
	"warehouse-mgmt/internal/outbound"
	"warehouse-mgmt/internal/report"
	"warehouse-mgmt/internal/storage"
)

var (
	dataFile   = "data/warehouse.json"
	store      *storage.Storage
	locationSvc *location.Service
	inventorySvc *inventory.Service
	inboundSvc *inbound.Service
	outboundSvc *outbound.Service
	reportSvc *report.Service
)

func initServices() {
	store = storage.NewStorage(dataFile)
	locationSvc = location.NewService()
	inventorySvc = inventory.NewService()
	inboundSvc = inbound.NewService(locationSvc)
	outboundSvc = outbound.NewService(locationSvc)
	reportSvc = report.NewService()
}

var rootCmd = &cobra.Command{
	Use:   "warehouse",
	Short: "Warehouse management system",
	Long: `A comprehensive warehouse management system for managing
inventory, inbound/outbound orders, locations, and stock counts.`,
}

func init() {
	initServices()

	rootCmd.AddCommand(productCmd)
	rootCmd.AddCommand(locationCmd)
	rootCmd.AddCommand(inboundCmd)
	rootCmd.AddCommand(outboundCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(countCmd)
	rootCmd.AddCommand(alertCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

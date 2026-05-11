package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

const defaultPort = "9011"

func main() {
	var portFlag string
	flag.StringVar(&portFlag, "port", "", "监听端口")
	flag.Parse()

	port := resolvePort(portFlag)

	app := newApp()

	http.HandleFunc("/api/repairs", app.handleCreateRepair)
	http.HandleFunc("/api/repairs/list", app.handleListRepairs)
	http.HandleFunc("/api/repairs/assign", app.handleAssignTech)
	http.HandleFunc("/api/repairs/apply-spare", app.handleApplySpare)
	http.HandleFunc("/api/repairs/complete", app.handleCompleteRepair)

	http.HandleFunc("/api/spares", app.handleAddSpare)
	http.HandleFunc("/api/spares/list", app.handleListSpares)
	http.HandleFunc("/api/spares/update-price", app.handleUpdateSparePrice)
	http.HandleFunc("/api/spares/purchase-orders", app.handleListPOs)

	http.HandleFunc("/api/techs", app.handleAddTech)
	http.HandleFunc("/api/techs/list", app.handleListTechs)

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("服务端启动，监听 %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "启动失败: %v\n", err)
		os.Exit(1)
	}
}

func resolvePort(cmdPort string) string {
	if cmdPort != "" {
		return cmdPort
	}
	envPort := os.Getenv("PORT")
	if envPort != "" {
		return envPort
	}
	return defaultPort
}

package main

import (
	"flag"

	"carrental/client/api"
	"carrental/client/cli"
)

func main() {
	serverURL := flag.String("server", api.DefaultServerURL, "服务端地址")
	flag.Parse()

	client := api.NewClient(*serverURL)
	c := cli.NewCLI(client)
	c.Run()
}

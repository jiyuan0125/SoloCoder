package main

import (
	"fmt"
	"strings"

	"smart-park/common"
)

func handleAccessCommands(c *Client, args []string) {
	if len(args) == 0 {
		fmt.Println("Access commands: add-point, list-points, add-rule, batch-rules, verify, list-records, fix-point")
		return
	}

	cmd := args[0]
	switch cmd {
	case "add-point":
		if len(args) < 5 {
			fmt.Println("Usage: add-point <id> <name> <area_id> <building_id>")
			return
		}
		req := common.CreateAccessPointRequest{
			ID:         args[1],
			Name:       args[2],
			AreaID:     args[3],
			BuildingID: args[4],
		}
		var resp common.Response
		c.post("/access/point", req, &resp)
		printResponse(resp)

	case "list-points":
		var resp common.Response
		c.get("/access/points", &resp)
		printResponse(resp)

	case "add-rule":
		if len(args) < 5 {
			fmt.Println("Usage: add-rule <point_id> <dept1,dept2> <start> <end>")
			return
		}
		depts := strings.Split(args[2], ",")
		req := common.CreateAccessRuleRequest{
			AccessPointID:      args[1],
			AllowedDepartments:   depts,
			StartTime:          args[3],
			EndTime:            args[4],
		}
		var resp common.Response
		c.post("/access/rule", req, &resp)
		printResponse(resp)

	case "batch-rules":
		if len(args) < 5 {
			fmt.Println("Usage: batch-rules <area_id> <dept1,dept2> <start> <end>")
			return
		}
		depts := strings.Split(args[2], ",")
		req := common.BatchAreaRuleRequest{
			AreaID:             args[1],
			AllowedDepartments: depts,
			StartTime:          args[3],
			EndTime:            args[4],
		}
		var resp common.Response
		c.post("/access/area-rules", req, &resp)
		printResponse(resp)

	case "verify":
		if len(args) < 4 {
			fmt.Println("Usage: verify <employee_id> <point_id> <type>")
			return
		}
		req := common.AccessRequest{
			EmployeeID:    args[1],
			AccessPointID: args[2],
			AccessType:    args[3],
		}
		var resp common.Response
		c.post("/access/verify", req, &resp)
		printResponse(resp)

	case "list-records":
		var resp common.Response
		c.get("/access/records", &resp)
		printResponse(resp)

	case "fix-point":
		if len(args) < 2 {
			fmt.Println("Usage: fix-point <point_id>")
			return
		}
		req := common.FixAccessPointRequest{
			AccessPointID: args[1],
		}
		var resp common.Response
		c.post("/access/fix", req, &resp)
		printResponse(resp)

	default:
		fmt.Printf("Unknown access command: %s\n", cmd)
	}
}

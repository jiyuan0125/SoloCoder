package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"settlement/pkg/common"
)

var baseURL string
var operator string

func httpRequest(method, path string, body interface{}) (*common.APIResponse, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Operator", operator)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp common.APIResponse
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(data, &apiResp)
	return &apiResp, nil
}

func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

func cmdProjectCreate(name string, contractAmount float64) {
	req := common.CreateProjectRequest{Name: name, ContractAmount: contractAmount}
	resp, err := httpRequest("POST", "/api/projects", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func cmdProjectList() {
	resp, err := httpRequest("GET", "/api/projects", nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func cmdProjectGet(id string) {
	resp, err := httpRequest("GET", "/api/projects/"+id, nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func cmdBOQAdd(projectID, itemCode, itemName, unit string, contractQty, actualQty, unitPrice float64) {
	req := common.CreateBOQRequest{
		ProjectID:   projectID,
		ItemCode:    itemCode,
		ItemName:    itemName,
		Unit:        unit,
		ContractQty: contractQty,
		ActualQty:   actualQty,
		UnitPrice:   unitPrice,
	}
	resp, err := httpRequest("POST", "/api/boqs", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func cmdBOQList(projectID string) {
	resp, err := httpRequest("GET", "/api/boqs?project_id="+projectID, nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func cmdBOQUpdate(projectID, boqID string, fields map[string]interface{}) {
	req := common.UpdateBOQRequest{}
	if v, ok := fields["item_name"]; ok {
		req.ItemName = v.(string)
	}
	if v, ok := fields["unit"]; ok {
		req.Unit = v.(string)
	}
	if v, ok := fields["contract_qty"]; ok {
		req.ContractQty = v.(float64)
	}
	if v, ok := fields["actual_qty"]; ok {
		req.ActualQty = v.(float64)
	}
	if v, ok := fields["unit_price"]; ok {
		req.UnitPrice = v.(float64)
	}
	resp, err := httpRequest("PUT", "/api/boqs/"+projectID+"/"+boqID, req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func cmdBOQDelete(projectID, boqID string) {
	resp, err := httpRequest("DELETE", "/api/boqs/"+projectID+"/"+boqID, nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func cmdVisaCreate(projectID, reason, changeContent string, increaseQty, decreaseQty, unitPrice float64) {
	req := common.CreateVisaRequest{
		ProjectID:     projectID,
		Reason:        reason,
		ChangeContent: changeContent,
		IncreaseQty:   increaseQty,
		DecreaseQty:   decreaseQty,
		UnitPrice:     unitPrice,
		Creator:       operator,
	}
	resp, err := httpRequest("POST", "/api/visas", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func cmdVisaList(projectID string) {
	resp, err := httpRequest("GET", "/api/visas?project_id="+projectID, nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func cmdVisaUpdate(projectID, visaID string, fields map[string]interface{}) {
	req := common.UpdateVisaRequest{}
	if v, ok := fields["reason"]; ok {
		req.Reason = v.(string)
	}
	if v, ok := fields["change_content"]; ok {
		req.ChangeContent = v.(string)
	}
	if v, ok := fields["increase_qty"]; ok {
		req.IncreaseQty = v.(float64)
	}
	if v, ok := fields["decrease_qty"]; ok {
		req.DecreaseQty = v.(float64)
	}
	if v, ok := fields["unit_price"]; ok {
		req.UnitPrice = v.(float64)
	}
	resp, err := httpRequest("PUT", "/api/visas/"+projectID+"/"+visaID, req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func cmdVisaConfirm(projectID, visaID, role string) {
	req := common.ConfirmVisaRequest{VisaID: visaID, Operator: operator, Role: role}
	resp, err := httpRequest("POST", "/api/visas/"+projectID+"/confirm", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func cmdVisaReject(projectID, visaID, role, reason string) {
	req := common.RejectVisaRequest{VisaID: visaID, Operator: operator, Role: role, Reason: reason}
	resp, err := httpRequest("POST", "/api/visas/"+projectID+"/reject", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func cmdSettlementGet(projectID string) {
	resp, err := httpRequest("GET", "/api/settlement?project_id="+projectID, nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func cmdAuditPass(projectID, opinion string) {
	req := common.AuditRequest{ProjectID: projectID, Auditor: operator, Opinion: opinion}
	resp, err := httpRequest("POST", "/api/audit/pass", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func cmdAuditLogs(projectID string) {
	resp, err := httpRequest("GET", "/api/audit/logs?project_id="+projectID, nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func printUsage() {
	fmt.Println("竣工结算系统客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  项目管理:")
	fmt.Println("    settlement-client project create <name> <contract_amount>")
	fmt.Println("    settlement-client project list")
	fmt.Println("    settlement-client project get <id>")
	fmt.Println()
	fmt.Println("  工程量管理:")
	fmt.Println("    settlement-client boq add <project_id> <item_code> <item_name> <unit> <contract_qty> <actual_qty> <unit_price>")
	fmt.Println("    settlement-client boq list <project_id>")
	fmt.Println("    settlement-client boq update <project_id> <boq_id> <field>=<value>...")
	fmt.Println("    settlement-client boq delete <project_id> <boq_id>")
	fmt.Println()
	fmt.Println("  签证管理:")
	fmt.Println("    settlement-client visa create <project_id> <reason> <change_content> <increase_qty> <decrease_qty> <unit_price>")
	fmt.Println("    settlement-client visa list <project_id>")
	fmt.Println("    settlement-client visa update <project_id> <visa_id> <field>=<value>...")
	fmt.Println("    settlement-client visa confirm <project_id> <visa_id> <role:supervisor|owner>")
	fmt.Println("    settlement-client visa reject <project_id> <visa_id> <role:supervisor|owner> <reason>")
	fmt.Println()
	fmt.Println("  审计:")
	fmt.Println("    settlement-client settlement get <project_id>")
	fmt.Println("    settlement-client audit pass <project_id> <opinion>")
	fmt.Println("    settlement-client audit logs <project_id>")
	fmt.Println()
	fmt.Println("全局选项:")
	fmt.Println("  --server <url>       服务端地址 (默认: http://localhost:8080)")
	fmt.Println("  --operator <name>    操作人名称 (默认: cli-user)")
}

func parseFieldArgs(args []string) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			switch parts[0] {
			case "contract_qty", "actual_qty", "unit_price", "increase_qty", "decrease_qty":
				if v, err := strconv.ParseFloat(parts[1], 64); err == nil {
					fields[parts[0]] = v
				}
			default:
				fields[parts[0]] = parts[1]
			}
		}
	}
	return fields
}

func main() {
	var serverURL string
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "服务端地址")
	flag.StringVar(&operator, "operator", "cli-user", "操作人名称")
	flag.Parse()

	baseURL = serverURL
	args := flag.Args()

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	subArgs := args[1:]

	switch cmd {
	case "project":
		if len(subArgs) < 1 {
			printUsage()
			return
		}
		switch subArgs[0] {
		case "create":
			if len(subArgs) < 3 {
				fmt.Println("参数不足")
				return
			}
			amount, _ := strconv.ParseFloat(subArgs[2], 64)
			cmdProjectCreate(subArgs[1], amount)
		case "list":
			cmdProjectList()
		case "get":
			if len(subArgs) < 2 {
				fmt.Println("参数不足")
				return
			}
			cmdProjectGet(subArgs[1])
		default:
			printUsage()
		}

	case "boq":
		if len(subArgs) < 1 {
			printUsage()
			return
		}
		switch subArgs[0] {
		case "add":
			if len(subArgs) < 8 {
				fmt.Println("参数不足")
				return
			}
			cq, _ := strconv.ParseFloat(subArgs[5], 64)
			aq, _ := strconv.ParseFloat(subArgs[6], 64)
			up, _ := strconv.ParseFloat(subArgs[7], 64)
			cmdBOQAdd(subArgs[1], subArgs[2], subArgs[3], subArgs[4], cq, aq, up)
		case "list":
			if len(subArgs) < 2 {
				fmt.Println("参数不足")
				return
			}
			cmdBOQList(subArgs[1])
		case "update":
			if len(subArgs) < 3 {
				fmt.Println("参数不足")
				return
			}
			fields := parseFieldArgs(subArgs[3:])
			cmdBOQUpdate(subArgs[1], subArgs[2], fields)
		case "delete":
			if len(subArgs) < 3 {
				fmt.Println("参数不足")
				return
			}
			cmdBOQDelete(subArgs[1], subArgs[2])
		default:
			printUsage()
		}

	case "visa":
		if len(subArgs) < 1 {
			printUsage()
			return
		}
		switch subArgs[0] {
		case "create":
			if len(subArgs) < 7 {
				fmt.Println("参数不足")
				return
			}
			iq, _ := strconv.ParseFloat(subArgs[4], 64)
			dq, _ := strconv.ParseFloat(subArgs[5], 64)
			up, _ := strconv.ParseFloat(subArgs[6], 64)
			cmdVisaCreate(subArgs[1], subArgs[2], subArgs[3], iq, dq, up)
		case "list":
			if len(subArgs) < 2 {
				fmt.Println("参数不足")
				return
			}
			cmdVisaList(subArgs[1])
		case "update":
			if len(subArgs) < 3 {
				fmt.Println("参数不足")
				return
			}
			fields := parseFieldArgs(subArgs[3:])
			cmdVisaUpdate(subArgs[1], subArgs[2], fields)
		case "confirm":
			if len(subArgs) < 4 {
				fmt.Println("参数不足")
				return
			}
			cmdVisaConfirm(subArgs[1], subArgs[2], subArgs[3])
		case "reject":
			if len(subArgs) < 5 {
				fmt.Println("参数不足")
				return
			}
			cmdVisaReject(subArgs[1], subArgs[2], subArgs[3], subArgs[4])
		default:
			printUsage()
		}

	case "settlement":
		if len(subArgs) < 2 {
			printUsage()
			return
		}
		if subArgs[0] == "get" {
			cmdSettlementGet(subArgs[1])
		} else {
			printUsage()
		}

	case "audit":
		if len(subArgs) < 1 {
			printUsage()
			return
		}
		switch subArgs[0] {
		case "pass":
			if len(subArgs) < 3 {
				fmt.Println("参数不足")
				return
			}
			cmdAuditPass(subArgs[1], subArgs[2])
		case "logs":
			if len(subArgs) < 2 {
				fmt.Println("参数不足")
				return
			}
			cmdAuditLogs(subArgs[1])
		default:
			printUsage()
		}

	default:
		printUsage()
	}
}

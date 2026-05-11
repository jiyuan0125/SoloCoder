package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"oacms/internal/api"
)

var baseURL = "http://localhost:9020"

func main() {
	if envURL := os.Getenv("OACMS_URL"); envURL != "" {
		baseURL = envURL
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "owner":
		handleOwner(args)
	case "election":
		handleElection(args)
	case "proposal":
		handleProposal(args)
	case "announcement":
		handleAnnouncement(args)
	case "delegate":
		handleDelegate(args)
	case "help":
		printUsage()
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("小区业主委员会管理系统 CLI")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  oacms-cli <command> [options]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  owner        业主管理")
	fmt.Println("  delegate     委托管理")
	fmt.Println("  election     选举管理")
	fmt.Println("  proposal     提案管理")
	fmt.Println("  announcement 公告管理")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  OACMS_URL    服务端地址 (默认: http://localhost:9020)")
}

func handleOwner(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: oacms-cli owner <list|create|get> [options]")
		return
	}

	switch args[0] {
	case "list":
		listOwners()
	case "create":
		fs := flag.NewFlagSet("create", flag.ExitOnError)
		name := fs.String("name", "", "业主姓名")
		phone := fs.String("phone", "", "手机号")
		room := fs.String("room", "", "房号")
		area := fs.Int64("area", 0, "建筑面积(平方米)")
		status := fs.String("status", "resident", "入住状态(resident/vacant/rented)")
		fs.Parse(args[1:])
		createOwner(*name, *phone, *room, *area, *status)
	case "get":
		if len(args) < 2 {
			fmt.Println("用法: oacms-cli owner get <id>")
			return
		}
		getOwner(args[1])
	default:
		fmt.Printf("未知的 owner 命令: %s\n", args[0])
	}
}

func handleDelegate(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: oacms-cli delegate <list|create|accept|reject> [options]")
		return
	}

	switch args[0] {
	case "list":
		listDelegates()
	case "create":
		fs := flag.NewFlagSet("create", flag.ExitOnError)
		principal := fs.String("principal", "", "委托人ID")
		agent := fs.String("agent", "", "受托人ID")
		fs.Parse(args[1:])
		createDelegate(*principal, *agent)
	case "accept":
		fs := flag.NewFlagSet("accept", flag.ExitOnError)
		delegateID := fs.String("id", "", "委托关系ID")
		agentID := fs.String("agent", "", "受托人ID")
		fs.Parse(args[1:])
		acceptDelegate(*delegateID, *agentID)
	case "reject":
		fs := flag.NewFlagSet("reject", flag.ExitOnError)
		delegateID := fs.String("id", "", "委托关系ID")
		agentID := fs.String("agent", "", "受托人ID")
		fs.Parse(args[1:])
		rejectDelegate(*delegateID, *agentID)
	default:
		fmt.Printf("未知的 delegate 命令: %s\n", args[0])
	}
}

func handleElection(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: oacms-cli election <list|create|get|candidate|vote|finalize> [options]")
		return
	}

	switch args[0] {
	case "list":
		listElections()
	case "create":
		fs := flag.NewFlagSet("create", flag.ExitOnError)
		title := fs.String("title", "", "选举标题")
		desc := fs.String("desc", "", "选举描述")
		rules := fs.String("rules", "", "选举规则")
		size := fs.Int("size", 5, "业委会委员人数")
		fs.Parse(args[1:])
		createElection(*title, *desc, *rules, *size)
	case "get":
		if len(args) < 2 {
			fmt.Println("用法: oacms-cli election get <id>")
			return
		}
		getElection(args[1])
	case "candidate":
		fs := flag.NewFlagSet("candidate", flag.ExitOnError)
		electionID := fs.String("election", "", "选举ID")
		ownerID := fs.String("owner", "", "候选人业主ID")
		reason := fs.String("reason", "", "参选理由")
		fs.Parse(args[1:])
		addCandidate(*electionID, *ownerID, *reason)
	case "vote":
		fs := flag.NewFlagSet("vote", flag.ExitOnError)
		electionID := fs.String("election", "", "选举ID")
		voterID := fs.String("voter", "", "投票人ID")
		candidateID := fs.String("candidate", "", "候选人ID")
		abstain := fs.Bool("abstain", false, "是否弃权")
		fs.Parse(args[1:])
		castVote(*electionID, *voterID, *candidateID, *abstain)
	case "finalize":
		if len(args) < 2 {
			fmt.Println("用法: oacms-cli election finalize <election_id>")
			return
		}
		finalizeElection(args[1])
	default:
		fmt.Printf("未知的 election 命令: %s\n", args[0])
	}
}

func handleProposal(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: oacms-cli proposal <list|create|get|second|review> [options]")
		return
	}

	switch args[0] {
	case "list":
		listProposals()
	case "create":
		fs := flag.NewFlagSet("create", flag.ExitOnError)
		title := fs.String("title", "", "提案标题")
		content := fs.String("content", "", "提案内容")
		proposer := fs.String("proposer", "", "提案人ID")
		fs.Parse(args[1:])
		createProposal(*title, *content, *proposer)
	case "get":
		if len(args) < 2 {
			fmt.Println("用法: oacms-cli proposal get <id>")
			return
		}
		getProposal(args[1])
	case "second":
		fs := flag.NewFlagSet("second", flag.ExitOnError)
		proposalID := fs.String("id", "", "提案ID")
		seconder := fs.String("seconder", "", "附议人ID")
		fs.Parse(args[1:])
		secondProposal(*proposalID, *seconder)
	case "review":
		fs := flag.NewFlagSet("review", flag.ExitOnError)
		proposalID := fs.String("id", "", "提案ID")
		decision := fs.String("decision", "", "处理意见(accepted/partially/postponed/rejected)")
		opinion := fs.String("opinion", "", "意见内容")
		dept := fs.String("dept", "", "负责部门")
		days := fs.Int("deadline", 0, "完成期限(天)")
		fs.Parse(args[1:])
		reviewProposal(*proposalID, *decision, *opinion, *dept, *days)
	default:
		fmt.Printf("未知的 proposal 命令: %s\n", args[0])
	}
}

func handleAnnouncement(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: oacms-cli announcement <list|create|get|archive> [options]")
		return
	}

	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("list", flag.ExitOnError)
		all := fs.Bool("all", false, "包含已归档公告")
		fs.Parse(args[1:])
		listAnnouncements(*all)
	case "create":
		fs := flag.NewFlagSet("create", flag.ExitOnError)
		title := fs.String("title", "", "公告标题")
		content := fs.String("content", "", "公告内容")
		days := fs.Int("days", 7, "有效期(天)")
		fs.Parse(args[1:])
		createAnnouncement(*title, *content, *days)
	case "get":
		if len(args) < 2 {
			fmt.Println("用法: oacms-cli announcement get <id>")
			return
		}
		getAnnouncement(args[1])
	case "archive":
		if len(args) < 2 {
			fmt.Println("用法: oacms-cli announcement archive <id>")
			return
		}
		archiveAnnouncement(args[1])
	default:
		fmt.Printf("未知的 announcement 命令: %s\n", args[0])
	}
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func httpPost(url string, body interface{}) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func printJSON(data []byte) {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err == nil {
		fmt.Println(pretty.String())
	} else {
		fmt.Println(string(data))
	}
}

func listOwners() {
	body, err := httpGet(baseURL + "/api/owners")
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func createOwner(name, phone, room string, area int64, status string) {
	req := api.CreateOwnerRequest{
		Name:       name,
		Phone:      phone,
		RoomNumber: room,
		Area:       area,
		Status:     status,
	}
	body, err := httpPost(baseURL+"/api/owners", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func getOwner(id string) {
	body, err := httpGet(baseURL + "/api/owners/" + id)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func listDelegates() {
	body, err := httpGet(baseURL + "/api/delegates")
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func createDelegate(principal, agent string) {
	req := api.CreateDelegateRequest{PrincipalID: principal, AgentID: agent}
	body, err := httpPost(baseURL+"/api/delegates", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func acceptDelegate(delegateID, agentID string) {
	req := api.AcceptDelegateRequest{DelegateID: delegateID, AgentID: agentID}
	body, err := httpPost(baseURL+"/api/delegates/accept", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func rejectDelegate(delegateID, agentID string) {
	req := api.AcceptDelegateRequest{DelegateID: delegateID, AgentID: agentID}
	body, err := httpPost(baseURL+"/api/delegates/reject", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func listElections() {
	body, err := httpGet(baseURL + "/api/elections")
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func createElection(title, desc, rules string, size int) {
	req := api.CreateElectionRequest{Title: title, Description: desc, Rules: rules, CommitteeSize: size}
	body, err := httpPost(baseURL+"/api/elections", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func getElection(id string) {
	body, err := httpGet(baseURL + "/api/elections/" + id)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func addCandidate(electionID, ownerID, reason string) {
	req := api.AddCandidateRequest{ElectionID: electionID, OwnerID: ownerID, Reason: reason}
	body, err := httpPost(baseURL+"/api/elections/candidate", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func castVote(electionID, voterID, candidateID string, abstain bool) {
	req := api.CastVoteRequest{ElectionID: electionID, VoterID: voterID, CandidateID: candidateID, Abstain: abstain}
	body, err := httpPost(baseURL+"/api/elections/vote", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func finalizeElection(id string) {
	req := map[string]string{"election_id": id}
	body, err := httpPost(baseURL+"/api/elections/finalize", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func listProposals() {
	body, err := httpGet(baseURL + "/api/proposals")
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func createProposal(title, content, proposer string) {
	req := api.CreateProposalRequest{Title: title, Content: content, ProposerID: proposer}
	body, err := httpPost(baseURL+"/api/proposals", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func getProposal(id string) {
	body, err := httpGet(baseURL + "/api/proposals/" + id)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func secondProposal(proposalID, seconder string) {
	req := api.SecondProposalRequest{ProposalID: proposalID, SeconderID: seconder}
	body, err := httpPost(baseURL+"/api/proposals/second", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func reviewProposal(proposalID, decision, opinion, dept string, days int) {
	req := api.ReviewProposalRequest{ProposalID: proposalID, Decision: decision, Opinion: opinion, ResponsibleDept: dept, DeadlineDays: days}
	body, err := httpPost(baseURL+"/api/proposals/review", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func listAnnouncements(all bool) {
	url := baseURL + "/api/announcements"
	if all {
		url += "?include_archived=true"
	}
	body, err := httpGet(url)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func createAnnouncement(title, content string, days int) {
	req := api.CreateAnnouncementRequest{Title: title, Content: content, ValidityDays: days}
	body, err := httpPost(baseURL+"/api/announcements", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func getAnnouncement(id string) {
	body, err := httpGet(baseURL + "/api/announcements/" + id)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

func archiveAnnouncement(id string) {
	req := struct{}{}
	body, err := httpPost(baseURL+"/api/announcements/"+id+"/archive", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(body)
}

var _ = strings.Repeat

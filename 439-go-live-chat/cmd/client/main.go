package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go-live-chat/common"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const serverURL = "http://localhost:8080/api"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "user-session":
		handleUserSession(args)
	case "agent-login":
		handleAgentLogin(args)
	case "agent-logout":
		handleAgentLogout(args)
	case "send-message":
		handleSendMessage(args)
	case "close-session":
		handleCloseSession(args)
	case "submit-satisfaction":
		handleSubmitSatisfaction(args)
	case "transfer-session":
		handleTransferSession(args)
	case "create-quick-reply":
		handleCreateQuickReply(args)
	case "list-quick-replies":
		handleListQuickReplies(args)
	case "add-blacklist":
		handleAddBlacklist(args)
	case "remove-blacklist":
		handleRemoveBlacklist(args)
	case "list-agent-sessions":
		handleListAgentSessions(args)
	case "get-session-history":
		handleGetSessionHistory(args)
	case "get-statistics":
		handleGetStatistics(args)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  user-session          Create a new user session")
	fmt.Println("    -user-id <id>       User ID")
	fmt.Println("    -username <name>    User name (optional)")
	fmt.Println()
	fmt.Println("  agent-login           Login an agent")
	fmt.Println("    -agent-id <id>      Agent ID")
	fmt.Println("    -username <name>    Agent name")
	fmt.Println()
	fmt.Println("  agent-logout          Logout an agent")
	fmt.Println("    -agent-id <id>      Agent ID")
	fmt.Println()
	fmt.Println("  send-message          Send a message in a session")
	fmt.Println("    -session-id <id>    Session ID")
	fmt.Println("    -sender-id <id>     Sender ID (user or agent)")
	fmt.Println("    -sender-type <type> Sender type: 'user' or 'agent'")
	fmt.Println("    -content <text>     Message content")
	fmt.Println()
	fmt.Println("  close-session         Close a session")
	fmt.Println("    -session-id <id>    Session ID")
	fmt.Println("    -agent-id <id>      Agent ID (optional)")
	fmt.Println()
	fmt.Println("  submit-satisfaction   Submit satisfaction rating")
	fmt.Println("    -session-id <id>    Session ID")
	fmt.Println("    -user-id <id>       User ID")
	fmt.Println("    -rating <1-5>       Rating 1-5")
	fmt.Println("    -comment <text>     Comment (optional)")
	fmt.Println()
	fmt.Println("  transfer-session      Transfer a session to another agent")
	fmt.Println("    -session-id <id>    Session ID")
	fmt.Println("    -from-agent <id>    Current agent ID")
	fmt.Println("    -to-agent <id>      Target agent ID")
	fmt.Println()
	fmt.Println("  create-quick-reply    Create a quick reply template")
	fmt.Println("    -agent-id <id>      Agent ID")
	fmt.Println("    -title <text>       Template title")
	fmt.Println("    -content <text>     Template content")
	fmt.Println()
	fmt.Println("  list-quick-replies    List quick reply templates")
	fmt.Println("    -agent-id <id>      Agent ID")
	fmt.Println()
	fmt.Println("  add-blacklist         Add user to blacklist")
	fmt.Println("    -user-id <id>       User ID")
	fmt.Println("    -reason <text>      Reason (optional)")
	fmt.Println()
	fmt.Println("  remove-blacklist      Remove user from blacklist")
	fmt.Println("    -user-id <id>       User ID")
	fmt.Println()
	fmt.Println("  list-agent-sessions   List active sessions for an agent")
	fmt.Println("    -agent-id <id>      Agent ID")
	fmt.Println()
	fmt.Println("  get-session-history   Get session history")
	fmt.Println("    -session-id <id>    Session ID")
	fmt.Println()
	fmt.Println("  get-statistics        Get system statistics")
}

func handleUserSession(args []string) {
	fs := flag.NewFlagSet("user-session", flag.ExitOnError)
	userID := fs.String("user-id", "", "User ID")
	username := fs.String("username", "", "User name")
	fs.Parse(args)

	if *userID == "" {
		fmt.Println("Error: -user-id is required")
		os.Exit(1)
	}

	request := common.CreateSessionRequest{
		UserID:   *userID,
		Username: *username,
	}

	var response common.CreateSessionResponse
	if err := sendRequest("POST", serverURL+"/sessions", request, &response); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Session created successfully!")
	fmt.Printf("  Session ID: %s\n", response.SessionID)
	fmt.Printf("  Status: %s\n", response.Status)
	if response.AgentID != "" {
		fmt.Printf("  Assigned to Agent: %s\n", response.AgentID)
	}
	if response.QueuePosition > 0 {
		fmt.Printf("  Queue Position: %d\n", response.QueuePosition)
	}
}

func handleAgentLogin(args []string) {
	fs := flag.NewFlagSet("agent-login", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID")
	username := fs.String("username", "", "Agent name")
	fs.Parse(args)

	if *agentID == "" {
		fmt.Println("Error: -agent-id is required")
		os.Exit(1)
	}

	request := common.AgentLoginRequest{
		AgentID:  *agentID,
		Username: *username,
	}

	var response common.AgentLoginResponse
	if err := sendRequest("POST", serverURL+"/agents/login", request, &response); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if response.Success {
		fmt.Println("Agent logged in successfully!")
	}
}

func handleAgentLogout(args []string) {
	fs := flag.NewFlagSet("agent-logout", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID")
	fs.Parse(args)

	if *agentID == "" {
		fmt.Println("Error: -agent-id is required")
		os.Exit(1)
	}

	request := common.AgentLogoutRequest{
		AgentID: *agentID,
	}

	var response common.AgentLogoutResponse
	if err := sendRequest("POST", serverURL+"/agents/logout", request, &response); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if response.Success {
		fmt.Println("Agent logged out successfully!")
	}
}

func handleSendMessage(args []string) {
	fs := flag.NewFlagSet("send-message", flag.ExitOnError)
	sessionID := fs.String("session-id", "", "Session ID")
	senderID := fs.String("sender-id", "", "Sender ID")
	senderType := fs.String("sender-type", "", "Sender type (user/agent)")
	content := fs.String("content", "", "Message content")
	fs.Parse(args)

	if *sessionID == "" || *senderID == "" || *senderType == "" || *content == "" {
		fmt.Println("Error: -session-id, -sender-id, -sender-type, and -content are required")
		os.Exit(1)
	}

	var st common.MessageSenderType
	switch strings.ToLower(*senderType) {
	case "user":
		st = common.SenderTypeUser
	case "agent":
		st = common.SenderTypeAgent
	default:
		fmt.Println("Error: -sender-type must be 'user' or 'agent'")
		os.Exit(1)
	}

	request := common.SendMessageRequest{
		SessionID:  *sessionID,
		SenderID:   *senderID,
		SenderType: st,
		Content:    *content,
	}

	var response common.SendMessageResponse
	if err := sendRequest("POST", serverURL+"/messages", request, &response); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Message sent successfully!")
	fmt.Printf("  Message ID: %s\n", response.MessageID)
	fmt.Printf("  Timestamp: %s\n", response.Timestamp.Format(time.RFC3339))
}

func handleCloseSession(args []string) {
	fs := flag.NewFlagSet("close-session", flag.ExitOnError)
	sessionID := fs.String("session-id", "", "Session ID")
	agentID := fs.String("agent-id", "", "Agent ID")
	fs.Parse(args)

	if *sessionID == "" {
		fmt.Println("Error: -session-id is required")
		os.Exit(1)
	}

	request := common.CloseSessionRequest{
		SessionID: *sessionID,
		AgentID:   *agentID,
	}

	var response common.CloseSessionResponse
	if err := sendRequest("POST", serverURL+"/sessions/close", request, &response); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if response.Closed {
		fmt.Println("Session closed successfully!")
	}
}

func handleSubmitSatisfaction(args []string) {
	fs := flag.NewFlagSet("submit-satisfaction", flag.ExitOnError)
	sessionID := fs.String("session-id", "", "Session ID")
	userID := fs.String("user-id", "", "User ID")
	rating := fs.Int("rating", 0, "Rating 1-5")
	comment := fs.String("comment", "", "Comment")
	fs.Parse(args)

	if *sessionID == "" || *userID == "" || *rating == 0 {
		fmt.Println("Error: -session-id, -user-id, and -rating are required")
		os.Exit(1)
	}

	if *rating < 1 || *rating > 5 {
		fmt.Println("Error: -rating must be between 1 and 5")
		os.Exit(1)
	}

	request := common.SubmitSatisfactionRequest{
		SessionID: *sessionID,
		UserID:    *userID,
		Rating:    *rating,
		Comment:   *comment,
	}

	var response common.SubmitSatisfactionResponse
	if err := sendRequest("POST", serverURL+"/satisfaction", request, &response); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if response.Success {
		fmt.Println("Satisfaction submitted successfully!")
	}
}

func handleTransferSession(args []string) {
	fs := flag.NewFlagSet("transfer-session", flag.ExitOnError)
	sessionID := fs.String("session-id", "", "Session ID")
	fromAgent := fs.String("from-agent", "", "From agent ID")
	toAgent := fs.String("to-agent", "", "To agent ID")
	fs.Parse(args)

	if *sessionID == "" || *fromAgent == "" || *toAgent == "" {
		fmt.Println("Error: -session-id, -from-agent, and -to-agent are required")
		os.Exit(1)
	}

	request := common.TransferSessionRequest{
		SessionID:   *sessionID,
		FromAgentID: *fromAgent,
		ToAgentID:   *toAgent,
	}

	var response common.TransferSessionResponse
	if err := sendRequest("POST", serverURL+"/sessions/transfer", request, &response); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if response.Success {
		fmt.Println("Session transferred successfully!")
		fmt.Printf("  New Agent: %s\n", response.NewAgentID)
	}
}

func handleCreateQuickReply(args []string) {
	fs := flag.NewFlagSet("create-quick-reply", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID")
	title := fs.String("title", "", "Template title")
	content := fs.String("content", "", "Template content")
	fs.Parse(args)

	if *agentID == "" || *title == "" || *content == "" {
		fmt.Println("Error: -agent-id, -title, and -content are required")
		os.Exit(1)
	}

	request := common.CreateQuickReplyRequest{
		AgentID: *agentID,
		Title:   *title,
		Content: *content,
	}

	var response common.CreateQuickReplyResponse
	if err := sendRequest("POST", serverURL+"/quick-replies/create", request, &response); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Quick reply created successfully!")
	fmt.Printf("  Template ID: %s\n", response.TemplateID)
}

func handleListQuickReplies(args []string) {
	fs := flag.NewFlagSet("list-quick-replies", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID")
	fs.Parse(args)

	if *agentID == "" {
		fmt.Println("Error: -agent-id is required")
		os.Exit(1)
	}

	var response common.GetQuickRepliesResponse
	if err := sendRequest("GET", serverURL+"/quick-replies?agent_id="+*agentID, nil, &response); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(response.Templates) == 0 {
		fmt.Println("No quick reply templates found.")
		return
	}

	fmt.Println("Quick Reply Templates:")
	for i, t := range response.Templates {
		fmt.Printf("\n  [%d] %s\n", i+1, t.Title)
		fmt.Printf("      Content: %s\n", t.Content)
		fmt.Printf("      Created: %s\n", t.CreatedAt.Format(time.RFC3339))
	}
}

func handleAddBlacklist(args []string) {
	fs := flag.NewFlagSet("add-blacklist", flag.ExitOnError)
	userID := fs.String("user-id", "", "User ID")
	reason := fs.String("reason", "", "Reason")
	fs.Parse(args)

	if *userID == "" {
		fmt.Println("Error: -user-id is required")
		os.Exit(1)
	}

	request := common.AddToBlacklistRequest{
		UserID: *userID,
		Reason: *reason,
	}

	var response common.AddToBlacklistResponse
	if err := sendRequest("POST", serverURL+"/blacklist", request, &response); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if response.Success {
		fmt.Println("User added to blacklist successfully!")
	}
}

func handleRemoveBlacklist(args []string) {
	fs := flag.NewFlagSet("remove-blacklist", flag.ExitOnError)
	userID := fs.String("user-id", "", "User ID")
	fs.Parse(args)

	if *userID == "" {
		fmt.Println("Error: -user-id is required")
		os.Exit(1)
	}

	var response common.RemoveFromBlacklistResponse
	if err := sendRequest("DELETE", serverURL+"/blacklist/remove?user_id="+*userID, nil, &response); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if response.Success {
		fmt.Println("User removed from blacklist successfully!")
	}
}

func handleListAgentSessions(args []string) {
	fs := flag.NewFlagSet("list-agent-sessions", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID")
	fs.Parse(args)

	if *agentID == "" {
		fmt.Println("Error: -agent-id is required")
		os.Exit(1)
	}

	var response common.GetAgentSessionsResponse
	if err := sendRequest("GET", serverURL+"/agents/sessions?agent_id="+*agentID, nil, &response); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(response.Sessions) == 0 {
		fmt.Println("No active sessions for this agent.")
		return
	}

	fmt.Println("Active Sessions:")
	for i, s := range response.Sessions {
		fmt.Printf("\n  [%d] Session ID: %s\n", i+1, s.ID)
		fmt.Printf("      User ID: %s\n", s.UserID)
		fmt.Printf("      Status: %s\n", s.Status)
		fmt.Printf("      Created: %s\n", s.CreatedAt.Format(time.RFC3339))
		fmt.Printf("      Messages: %d\n", len(s.Messages))
		if s.Satisfaction != nil {
			fmt.Printf("      Satisfaction: %d/5\n", s.Satisfaction.Rating)
		}
	}
}

func handleGetSessionHistory(args []string) {
	fs := flag.NewFlagSet("get-session-history", flag.ExitOnError)
	sessionID := fs.String("session-id", "", "Session ID")
	fs.Parse(args)

	if *sessionID == "" {
		fmt.Println("Error: -session-id is required")
		os.Exit(1)
	}

	var response common.GetSessionHistoryResponse
	if err := sendRequest("GET", serverURL+"/sessions/history?session_id="+*sessionID, nil, &response); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Session ID: %s\n", response.Session.ID)
	fmt.Printf("User ID: %s\n", response.Session.UserID)
	fmt.Printf("Agent ID: %s\n", response.Session.AgentID)
	fmt.Printf("Status: %s\n", response.Session.Status)
	fmt.Printf("Created: %s\n", response.Session.CreatedAt.Format(time.RFC3339))

	if response.Session.Satisfaction != nil {
		fmt.Printf("\nSatisfaction Rating: %d/5\n", response.Session.Satisfaction.Rating)
		if response.Session.Satisfaction.Comment != "" {
			fmt.Printf("Comment: %s\n", response.Session.Satisfaction.Comment)
		}
	}

	fmt.Println("\nMessages:")
	for i, m := range response.Messages {
		fmt.Printf("\n  [%d] [%s] %s (%s):\n", i+1, m.Timestamp.Format(time.RFC3339), m.SenderID, m.SenderType)
		fmt.Printf("      %s\n", m.Content)
	}
}

func handleGetStatistics(args []string) {
	var response common.GetStatisticsResponse
	if err := sendRequest("GET", serverURL+"/statistics", nil, &response); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("System Statistics:")
	fmt.Printf("  Total Sessions: %d\n", response.Statistics.TotalSessions)
	fmt.Printf("  Average Wait Time: %.2f seconds\n", response.Statistics.AverageWaitTime)
	fmt.Printf("  Average Session Time: %.2f seconds\n", response.Statistics.AverageSessionTime)
	fmt.Printf("  Average Satisfaction: %.2f/5\n", response.Statistics.AverageSatisfaction)
}

func sendRequest(method, url string, request interface{}, response interface{}) error {
	var body []byte
	var err error

	if request != nil {
		body, err = json.Marshal(request)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp common.ErrorResponse
		json.Unmarshal(respBody, &errorResp)
		if errorResp.Error != "" {
			return fmt.Errorf("server error: %s", errorResp.Error)
		}
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	if response != nil {
		err = json.Unmarshal(respBody, response)
		if err != nil {
			return err
		}
	}

	return nil
}

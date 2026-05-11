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

	"community-activity-platform/pkg/common"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func (c *Client) do(method, endpoint string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return http.DefaultClient.Do(req)
}

func (c *Client) printResponse(resp *http.Response) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp common.APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Println(string(body))
		return err
	}

	prettyJSON, err := json.MarshalIndent(apiResp, "", "  ")
	if err != nil {
		fmt.Println(string(body))
		return err
	}

	fmt.Println(string(prettyJSON))
	return nil
}

func (c *Client) registerParticipant(phone, room, name string, age int) error {
	req := common.RegisterParticipantRequest{
		Phone:      phone,
		RoomNumber: room,
		Name:       name,
		Age:        age,
	}

	resp, err := c.do("POST", "/participant/register", req)
	if err != nil {
		return err
	}

	return c.printResponse(resp)
}

func (c *Client) getParticipant(id, phone string) error {
	endpoint := "/participant?"
	if id != "" {
		endpoint += "id=" + id
	} else if phone != "" {
		endpoint += "phone=" + phone
	} else {
		return fmt.Errorf("must provide id or phone")
	}

	resp, err := c.do("GET", endpoint, nil)
	if err != nil {
		return err
	}

	return c.printResponse(resp)
}

func (c *Client) createActivity(req common.CreateActivityRequest) error {
	resp, err := c.do("POST", "/activity/create", req)
	if err != nil {
		return err
	}

	return c.printResponse(resp)
}

func (c *Client) getActivity(id string) error {
	resp, err := c.do("GET", "/activity?id="+id, nil)
	if err != nil {
		return err
	}

	return c.printResponse(resp)
}

func (c *Client) listActivities(status, activityType string) error {
	endpoint := "/activity/list"
	params := []string{}
	if status != "" {
		params = append(params, "status="+status)
	}
	if activityType != "" {
		params = append(params, "type="+activityType)
	}
	if len(params) > 0 {
		endpoint += "?" + strings.Join(params, "&")
	}

	resp, err := c.do("GET", endpoint, nil)
	if err != nil {
		return err
	}

	return c.printResponse(resp)
}

func (c *Client) cancelActivity(activityID string) error {
	req := common.CancelActivityRequest{
		ActivityID: activityID,
	}

	resp, err := c.do("POST", "/activity/cancel", req)
	if err != nil {
		return err
	}

	return c.printResponse(resp)
}

func (c *Client) registerActivity(participantID, activityID string, adults, children int) error {
	req := common.RegisterActivityRequest{
		ParticipantID: participantID,
		ActivityID:    activityID,
		AdultsCount:   adults,
		ChildrenCount: children,
	}

	resp, err := c.do("POST", "/registration/register", req)
	if err != nil {
		return err
	}

	return c.printResponse(resp)
}

func (c *Client) cancelRegistration(registrationID string) error {
	req := common.CancelRegistrationRequest{
		RegistrationID: registrationID,
	}

	resp, err := c.do("POST", "/registration/cancel", req)
	if err != nil {
		return err
	}

	return c.printResponse(resp)
}

func (c *Client) checkIn(activityID, phoneLast4 string) error {
	req := common.CheckInRequest{
		ActivityID: activityID,
		PhoneLast4: phoneLast4,
	}

	resp, err := c.do("POST", "/checkin", req)
	if err != nil {
		return err
	}

	return c.printResponse(resp)
}

func (c *Client) submitFeedback(activityID, participantID string, rating int, comments string) error {
	req := common.SubmitFeedbackRequest{
		ActivityID:    activityID,
		ParticipantID: participantID,
		Rating:        rating,
		Comments:      comments,
	}

	resp, err := c.do("POST", "/feedback/submit", req)
	if err != nil {
		return err
	}

	return c.printResponse(resp)
}

func (c *Client) getActivitySummary(activityID string) error {
	resp, err := c.do("GET", "/activity/summary?activity_id="+activityID, nil)
	if err != nil {
		return err
	}

	return c.printResponse(resp)
}

func (c *Client) listAnalysisTodos() error {
	resp, err := c.do("GET", "/analysis/todos", nil)
	if err != nil {
		return err
	}

	return c.printResponse(resp)
}

func printUsage() {
	fmt.Println("Community Activity Platform CLI Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [command] [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  participant-register    Register a new participant")
	fmt.Println("  participant-get         Get participant info")
	fmt.Println("  activity-create         Create a new activity")
	fmt.Println("  activity-get            Get activity details")
	fmt.Println("  activity-list           List activities")
	fmt.Println("  activity-cancel         Cancel an activity")
	fmt.Println("  register                Register for an activity")
	fmt.Println("  cancel-registration     Cancel a registration")
	fmt.Println("  checkin                 Check in to an activity")
	fmt.Println("  feedback-submit         Submit feedback for an activity")
	fmt.Println("  activity-summary        Get activity summary")
	fmt.Println("  analysis-todos          List analysis todos")
	fmt.Println()
	fmt.Println("Use 'client [command] --help' for more information about a command")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
	}

	baseURL := "http://localhost:9019"
	if envURL := os.Getenv("APP_URL"); envURL != "" {
		baseURL = envURL
	}

	client := NewClient(baseURL)

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "participant-register":
		fs := flag.NewFlagSet("participant-register", flag.ExitOnError)
		phone := fs.String("phone", "", "Phone number (required)")
		room := fs.String("room", "", "Room number (required)")
		name := fs.String("name", "", "Name (required)")
		age := fs.Int("age", 0, "Age")
		fs.Parse(args)

		if *phone == "" || *room == "" || *name == "" {
			fmt.Println("Error: phone, room, and name are required")
			fs.PrintDefaults()
			os.Exit(1)
		}

		if err := client.registerParticipant(*phone, *room, *name, *age); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	case "participant-get":
		fs := flag.NewFlagSet("participant-get", flag.ExitOnError)
		id := fs.String("id", "", "Participant ID")
		phone := fs.String("phone", "", "Phone number")
		fs.Parse(args)

		if *id == "" && *phone == "" {
			fmt.Println("Error: must provide id or phone")
			fs.PrintDefaults()
			os.Exit(1)
		}

		if err := client.getParticipant(*id, *phone); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	case "activity-create":
		fs := flag.NewFlagSet("activity-create", flag.ExitOnError)
		title := fs.String("title", "", "Activity title (required)")
		desc := fs.String("description", "", "Activity description")
		location := fs.String("location", "", "Activity location (required)")
		maxPart := fs.Int("max-participants", 10, "Maximum participants")
		activityType := fs.String("type", "other", "Activity type: art_performance, health_lecture, parent_child, sports, volunteer, festival, other")
		frequency := fs.String("frequency", "single", "Frequency: single or recurring")
		episodes := fs.Int("episodes", 1, "Number of episodes for recurring activities")
		fs.Parse(args)

		if *title == "" || *location == "" {
			fmt.Println("Error: title and location are required")
			fs.PrintDefaults()
			os.Exit(1)
		}

		req := common.CreateActivityRequest{
			Title:                *title,
			Description:          *desc,
			Location:             *location,
			MaxParticipants:      *maxPart,
			ActivityType:         common.ActivityType(*activityType),
			Frequency:            common.ActivityFrequency(*frequency),
			EpisodeCount:         *episodes,
		}

		if err := client.createActivity(req); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	case "activity-get":
		fs := flag.NewFlagSet("activity-get", flag.ExitOnError)
		id := fs.String("id", "", "Activity ID (required)")
		fs.Parse(args)

		if *id == "" {
			fmt.Println("Error: id is required")
			fs.PrintDefaults()
			os.Exit(1)
		}

		if err := client.getActivity(*id); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	case "activity-list":
		fs := flag.NewFlagSet("activity-list", flag.ExitOnError)
		status := fs.String("status", "", "Filter by status: draft, active, cancelled, completed")
		activityType := fs.String("type", "", "Filter by type")
		fs.Parse(args)

		if err := client.listActivities(*status, *activityType); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	case "activity-cancel":
		fs := flag.NewFlagSet("activity-cancel", flag.ExitOnError)
		id := fs.String("id", "", "Activity ID (required)")
		fs.Parse(args)

		if *id == "" {
			fmt.Println("Error: id is required")
			fs.PrintDefaults()
			os.Exit(1)
		}

		if err := client.cancelActivity(*id); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	case "register":
		fs := flag.NewFlagSet("register", flag.ExitOnError)
		participantID := fs.String("participant-id", "", "Participant ID (required)")
		activityID := fs.String("activity-id", "", "Activity ID (required)")
		adults := fs.Int("adults", 0, "Number of adults (for parent-child activities)")
		children := fs.Int("children", 0, "Number of children (for parent-child activities)")
		fs.Parse(args)

		if *participantID == "" || *activityID == "" {
			fmt.Println("Error: participant-id and activity-id are required")
			fs.PrintDefaults()
			os.Exit(1)
		}

		if err := client.registerActivity(*participantID, *activityID, *adults, *children); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	case "cancel-registration":
		fs := flag.NewFlagSet("cancel-registration", flag.ExitOnError)
		regID := fs.String("registration-id", "", "Registration ID (required)")
		fs.Parse(args)

		if *regID == "" {
			fmt.Println("Error: registration-id is required")
			fs.PrintDefaults()
			os.Exit(1)
		}

		if err := client.cancelRegistration(*regID); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	case "checkin":
		fs := flag.NewFlagSet("checkin", flag.ExitOnError)
		activityID := fs.String("activity-id", "", "Activity ID (required)")
		phoneLast4 := fs.String("phone-last4", "", "Last 4 digits of phone number (required)")
		fs.Parse(args)

		if *activityID == "" || *phoneLast4 == "" {
			fmt.Println("Error: activity-id and phone-last4 are required")
			fs.PrintDefaults()
			os.Exit(1)
		}

		if err := client.checkIn(*activityID, *phoneLast4); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	case "feedback-submit":
		fs := flag.NewFlagSet("feedback-submit", flag.ExitOnError)
		activityID := fs.String("activity-id", "", "Activity ID (required)")
		participantID := fs.String("participant-id", "", "Participant ID (required)")
		rating := fs.Int("rating", 0, "Rating 1-5 (required)")
		comments := fs.String("comments", "", "Feedback comments")
		fs.Parse(args)

		if *activityID == "" || *participantID == "" || *rating == 0 {
			fmt.Println("Error: activity-id, participant-id, and rating are required")
			fs.PrintDefaults()
			os.Exit(1)
		}

		if err := client.submitFeedback(*activityID, *participantID, *rating, *comments); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	case "activity-summary":
		fs := flag.NewFlagSet("activity-summary", flag.ExitOnError)
		id := fs.String("id", "", "Activity ID (required)")
		fs.Parse(args)

		if *id == "" {
			fmt.Println("Error: id is required")
			fs.PrintDefaults()
			os.Exit(1)
		}

		if err := client.getActivitySummary(*id); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	case "analysis-todos":
		if err := client.listAnalysisTodos(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	default:
		fmt.Println("Unknown command:", cmd)
		printUsage()
	}
}

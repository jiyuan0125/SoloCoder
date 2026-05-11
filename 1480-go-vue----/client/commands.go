package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"vehicle-inspection/common"
)

type Commander struct {
	client *APIClient
}

func NewCommander(client *APIClient) *Commander {
	return &Commander{client: client}
}

func (c *Commander) RegisterVehicle(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: register-vehicle <plate> <type> <register-date> [last-inspection-date]")
	}

	plate := args[0]
	vehicleType := common.VehicleType(args[1])
	registerDate, err := time.Parse("2006-01-02", args[2])
	if err != nil {
		return fmt.Errorf("invalid register date format (use YYYY-MM-DD): %v", err)
	}

	var lastInspectionDate time.Time
	if len(args) >= 4 {
		lastInspectionDate, err = time.Parse("2006-01-02", args[3])
		if err != nil {
			return fmt.Errorf("invalid last inspection date format (use YYYY-MM-DD): %v", err)
		}
	}

	req := common.RegisterVehicleRequest{
		PlateNumber:        plate,
		VehicleType:        vehicleType,
		RegisterDate:       registerDate,
		LastInspectionDate: lastInspectionDate,
	}

	data, err := c.client.Post("/api/vehicles", req)
	if err != nil {
		printJSON(data)
		return err
	}

	printJSON(data)
	return nil
}

func (c *Commander) GetVehicleInfo(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: vehicle-info <plate>")
	}

	plate := args[0]
	data, err := c.client.Get("/api/vehicles/" + plate)
	if err != nil {
		printJSON(data)
		return err
	}

	printJSON(data)
	return nil
}

func (c *Commander) ListVehicles(args []string) error {
	data, err := c.client.Get("/api/vehicles")
	if err != nil {
		printJSON(data)
		return err
	}

	printJSON(data)
	return nil
}

func (c *Commander) CreateStation(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: create-station <name>")
	}

	name := args[0]
	req := common.CreateStationRequest{
		Name:      name,
		Workshops: []common.Workshop{},
	}

	data, err := c.client.Post("/api/stations", req)
	if err != nil {
		printJSON(data)
		return err
	}

	printJSON(data)
	return nil
}

func (c *Commander) AddSchedule(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: add-schedule <station-id> <date> <time-slot> <max-vehicles>")
	}

	stationID := args[0]
	date, err := time.Parse("2006-01-02", args[1])
	if err != nil {
		return fmt.Errorf("invalid date format (use YYYY-MM-DD): %v", err)
	}
	timeSlot := args[2]
	var maxVehicles int
	_, err = fmt.Sscanf(args[3], "%d", &maxVehicles)
	if err != nil {
		return fmt.Errorf("invalid max vehicles: %v", err)
	}

	req := common.AddScheduleRequest{
		StationID:   stationID,
		Date:        date,
		TimeSlot:    timeSlot,
		MaxVehicles: maxVehicles,
	}

	data, err := c.client.Post("/api/schedules", req)
	if err != nil {
		printJSON(data)
		return err
	}

	printJSON(data)
	return nil
}

func (c *Commander) ListStations(args []string) error {
	data, err := c.client.Get("/api/stations")
	if err != nil {
		printJSON(data)
		return err
	}

	printJSON(data)
	return nil
}

func (c *Commander) CreateAppointment(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: create-appointment <plate> <station-id> <date> <time-slot>")
	}

	plate := args[0]
	stationID := args[1]
	date, err := time.Parse("2006-01-02", args[2])
	if err != nil {
		return fmt.Errorf("invalid date format (use YYYY-MM-DD): %v", err)
	}
	timeSlot := args[3]

	req := common.CreateAppointmentRequest{
		PlateNumber:     plate,
		StationID:       stationID,
		AppointmentDate: date,
		TimeSlot:        timeSlot,
	}

	data, err := c.client.Post("/api/appointments", req)
	if err != nil {
		printJSON(data)
		return err
	}

	printJSON(data)
	return nil
}

func (c *Commander) ListAppointments(args []string) error {
	data, err := c.client.Get("/api/appointments")
	if err != nil {
		printJSON(data)
		return err
	}

	printJSON(data)
	return nil
}

func (c *Commander) StartInspection(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: start-inspection <appointment-id>")
	}

	appointmentID := args[0]
	req := common.StartInspectionRequest{
		AppointmentID: appointmentID,
	}

	data, err := c.client.Post("/api/inspections/start", req)
	if err != nil {
		printJSON(data)
		return err
	}

	printJSON(data)
	return nil
}

func (c *Commander) CompleteStep(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: complete-step <process-id> <result> [fail-items-json]")
	}

	processID := args[0]
	result := common.InspectionResult(args[1])
	var failItems []common.FailItem

	if len(args) >= 4 {
		if err := json.Unmarshal([]byte(args[3]), &failItems); err != nil {
			return fmt.Errorf("invalid fail items JSON: %v", err)
		}
	}

	req := common.CompleteStepRequest{
		ProcessID: processID,
		Result:    result,
		FailItems: failItems,
	}

	data, err := c.client.Post("/api/inspections/complete-step", req)
	if err != nil {
		printJSON(data)
		return err
	}

	printJSON(data)
	return nil
}

func (c *Commander) StartRecheck(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: start-recheck <process-id>")
	}

	processID := args[0]
	req := common.StartRecheckRequest{
		ProcessID: processID,
	}

	data, err := c.client.Post("/api/inspections/recheck", req)
	if err != nil {
		printJSON(data)
		return err
	}

	printJSON(data)
	return nil
}

func (c *Commander) ListProcesses(args []string) error {
	data, err := c.client.Get("/api/processes")
	if err != nil {
		printJSON(data)
		return err
	}

	printJSON(data)
	return nil
}

func printJSON(data []byte) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		fmt.Println(string(data))
		return
	}
	fmt.Println(prettyJSON.String())
}

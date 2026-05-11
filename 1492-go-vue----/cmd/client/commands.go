package main

import (
	"fmt"

	"renovation-management/internal/api"
)

func (c *Client) cmdCreateHouse(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: create-house <area> <layout> <floor> <orientation>")
	}

	var area float64
	fmt.Sscanf(args[0], "%f", &area)

	var floor int
	fmt.Sscanf(args[2], "%d", &floor)

	req := &api.CreateHouseRequest{
		Area:        area,
		Layout:      args[1],
		Floor:       floor,
		Orientation: args[3],
	}

	data, err := c.doRequest("POST", "/houses", req)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdListHouses(args []string) error {
	data, err := c.doRequest("GET", "/houses", nil)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdGetHouse(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-house <id>")
	}
	data, err := c.doRequest("GET", "/houses/"+args[0], nil)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdCreateDesign(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: create-design <house_id> <version> [description]")
	}

	var version int
	fmt.Sscanf(args[1], "%d", &version)

	req := &api.CreateDesignRequest{
		HouseID:     args[0],
		Version:     version,
		Description: "",
		Spaces:      []api.SpaceDesign{},
	}

	if len(args) >= 3 {
		req.Description = args[2]
	}

	data, err := c.doRequest("POST", "/designs", req)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdGetDesign(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-design <id>")
	}
	data, err := c.doRequest("GET", "/designs/"+args[0], nil)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdConfirmDesign(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: confirm-design <scheme_id>")
	}
	req := &api.ConfirmDesignRequest{SchemeID: args[0]}
	data, err := c.doRequest("POST", "/designs/confirm", req)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdCreateQuotation(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: create-quotation <scheme_id>")
	}

	req := &api.CreateQuotationRequest{
		SchemeID:  args[0],
		WorkItems:  []api.WorkItemInput{},
		Materials:  []api.MaterialInput{},
	}

	data, err := c.doRequest("POST", "/quotations", req)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdListQuotations(args []string) error {
	data, err := c.doRequest("GET", "/quotations", nil)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdGetQuotation(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-quotation <id>")
	}
	data, err := c.doRequest("GET", "/quotations/"+args[0], nil)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdCreateChange(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: create-change <quotation_id> <content> <reason> <amount_diff_fen>")
	}

	var amountDiff int64
	fmt.Sscanf(args[3], "%d", &amountDiff)

	req := &api.CreateChangeOrderRequest{
		QuotationID:   args[0],
		Content:       args[1],
		Reason:        args[2],
		AmountDiffFen: amountDiff,
	}

	data, err := c.doRequest("POST", "/changes", req)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdConfirmChange(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: confirm-change <change_id>")
	}
	req := &api.ConfirmChangeOrderRequest{ChangeOrderID: args[0]}
	data, err := c.doRequest("POST", "/changes/confirm", req)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdCreatePhase(args []string) error {
	if len(args) < 5 {
		return fmt.Errorf("usage: create-phase <quotation_id> <category> <order_index> <start_date> <end_date>")
	}

	var orderIndex int
	fmt.Sscanf(args[2], "%d", &orderIndex)

	req := &api.CreatePhaseRequest{
		QuotationID:   args[0],
		Category:      args[1],
		OrderIndex: orderIndex,
	}

	data, err := c.doRequest("POST", "/phases", req)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdUpdateProgress(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: update-progress <phase_id> <progress> <updated_by>")
	}

	var progress int
	fmt.Sscanf(args[1], "%d", &progress)

	req := &api.UpdateProgressRequest{
		PhaseID:   args[0],
		Progress:  progress,
		UpdatedBy: args[2],
	}

	data, err := c.doRequest("POST", "/phases/progress", req)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdGetPhase(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-phase <id>")
	}
	data, err := c.doRequest("GET", "/phases/"+args[0], nil)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdListDelayedPhases(args []string) error {
	data, err := c.doRequest("GET", "/phases/delayed", nil)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdCalculateSettlement(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: calculate-settlement <quotation_id>")
	}
	data, err := c.doRequest("POST", "/settlements/calculate/"+args[0], nil)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

func (c *Client) cmdGetSettlement(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-settlement <id>")
	}
	data, err := c.doRequest("GET", "/settlements/"+args[0], nil)
	if err != nil {
		return err
	}
	return prettyPrint(data)
}

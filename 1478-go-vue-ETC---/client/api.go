package main

import (
	"bytes"
	"encoding/json"
	"etc-system/common"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type APIError struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func CreateAccount(server, licensePlate, bankCard string, balance float64, vehicleType int, seats int, loadWeight float64) error {
	req := common.CreateAccountRequest{
		LicensePlate: licensePlate,
		BankCard:     bankCard,
		Balance:      balance,
		VehicleType:  common.VehicleType(vehicleType),
		Seats:        seats,
		LoadWeight:   loadWeight,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(server+"/api/account/create", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		json.NewDecoder(resp.Body).Decode(&apiErr)
		return fmt.Errorf(apiErr.Message)
	}

	return nil
}

func GetAccount(server, licensePlate string) (*common.Account, error) {
	u, _ := url.Parse(server + "/api/account/get")
	q := u.Query()
	q.Set("license_plate", licensePlate)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		json.NewDecoder(resp.Body).Decode(&apiErr)
		return nil, fmt.Errorf(apiErr.Message)
	}

	var result common.GetAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Account, nil
}

func Recharge(server, licensePlate string, amount float64) (float64, error) {
	req := common.RechargeRequest{
		LicensePlate: licensePlate,
		Amount:       amount,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(server+"/api/account/recharge", "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		json.NewDecoder(resp.Body).Decode(&apiErr)
		return 0, fmt.Errorf(apiErr.Message)
	}

	var result common.RechargeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.Balance, nil
}

func Pass(server, licensePlate, entryStation, exitStation string, mileage float64, isFree bool) (*common.PassRecord, error) {
	req := common.PassRequest{
		LicensePlate: licensePlate,
		EntryStation: entryStation,
		ExitStation:  exitStation,
		PassDate:     time.Now(),
		Mileage:      mileage,
		IsFree:       isFree,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(server+"/api/pass", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		json.NewDecoder(resp.Body).Decode(&apiErr)
		return nil, fmt.Errorf(apiErr.Message)
	}

	var result common.PassResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Record, nil
}

func ExportRecords(server, startDate, endDate, outputFile string) error {
	u, _ := url.Parse(server + "/api/export")
	q := u.Query()
	q.Set("start_date", startDate)
	q.Set("end_date", endDate)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		json.NewDecoder(resp.Body).Decode(&apiErr)
		return fmt.Errorf(apiErr.Message)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return writeFile(outputFile, data)
}

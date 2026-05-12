package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const registryURL = "http://localhost:8080"

type ServiceInstance struct {
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Weight  int    `json:"weight"`
}

type WatchEvent struct {
	Action    string           `json:"action"`
	Service   string           `json:"service"`
	Instances []ServiceInstance `json:"instances"`
}

func register(inst ServiceInstance) error {
	data, _ := json.Marshal(inst)
	resp, err := http.Post(registryURL+"/register", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Register response: %s\n", body)
	return nil
}

func heartbeat(name, ip string, port int) error {
	req := map[string]interface{}{"name": name, "ip": ip, "port": port}
	data, _ := json.Marshal(req)
	resp, err := http.Post(registryURL+"/heartbeat", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Heartbeat response: %s\n", body)
	return nil
}

func discover(name, version string) ([]ServiceInstance, error) {
	url := registryURL + "/discover?name=" + name
	if version != "" {
		url += "&version=" + version
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var instances []ServiceInstance
	if err := json.NewDecoder(resp.Body).Decode(&instances); err != nil {
		return nil, err
	}
	return instances, nil
}

func main() {
	fmt.Println("=== Service Registry Demo ===")

	inst := ServiceInstance{
		IP:      "127.0.0.1",
		Port:    9001,
		Name:    "user-service",
		Version: "v1.0.0",
		Weight:  100,
	}

	fmt.Println("\n1. Registering service...")
	if err := register(inst); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("\n2. Discovering services...")
	instances, err := discover("user-service", "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Found %d instances:\n", len(instances))
	for _, i := range instances {
		fmt.Printf("  - %s:%d (%s, weight=%d)\n", i.IP, i.Port, i.Version, i.Weight)
	}

	fmt.Println("\n3. Starting heartbeat every 5 seconds...")
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				heartbeat(inst.Name, inst.IP, inst.Port)
			case <-stop:
				return
			}
		}
	}()

	fmt.Println("\n4. Press Ctrl+C to exit, or wait to see health check behavior")
	time.Sleep(2 * time.Minute)
	close(stop)
}

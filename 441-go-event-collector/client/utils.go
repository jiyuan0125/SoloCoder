package client

import (
	"encoding/json"
	"fmt"
)

func ParseJSON(jsonStr string, result interface{}) error {
	return json.Unmarshal([]byte(jsonStr), result)
}

func PrettyPrintJSON(data interface{}) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("%+v\n", data)
		return
	}
	fmt.Println(string(jsonBytes))
}

func MustJSON(data interface{}) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}

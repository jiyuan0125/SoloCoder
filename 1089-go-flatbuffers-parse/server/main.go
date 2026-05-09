package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/example/flatbuffers-parser/common"
	"github.com/example/flatbuffers-parser/flatbuffers"
)

func main() {
	http.HandleFunc("/inspect", handleInspect)
	http.HandleFunc("/health", handleHealth)

	fmt.Println("FlatBuffers server starting on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleInspect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req common.InspectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.InspectResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid request: %v", err),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	buf, err := req.DecodeBuffer()
	if err != nil {
		resp := common.InspectResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid buffer: %v", err),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	schema, err := convertSchema(req.Schema)
	if err != nil {
		resp := common.InspectResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid schema: %v", err),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	node, err := flatbuffers.Inspect(buf, schema)
	if err != nil {
		resp := common.InspectResponse{
			Success: false,
			Message: fmt.Sprintf("Inspect error: %v", err),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.InspectResponse{
		Success: true,
		Result:  node.ToMap(),
	}
	json.NewEncoder(w).Encode(resp)
}

func convertSchema(s *common.SchemaDefinition) (*flatbuffers.TableSchema, error) {
	if s == nil {
		return nil, fmt.Errorf("schema is nil")
	}
	fields := make([]flatbuffers.FieldSchema, len(s.Fields))
	for i, f := range s.Fields {
		ft, err := parseFieldType(f.Type)
		if err != nil {
			return nil, err
		}
		fields[i] = flatbuffers.FieldSchema{
			Name: f.Name,
			Type: ft,
		}
		if f.ElemType != "" {
			et, err := parseFieldType(f.ElemType)
			if err != nil {
				return nil, err
			}
			fields[i].ElemType = et
		}
		if f.Nested != nil {
			nested, err := convertSchema(f.Nested)
			if err != nil {
				return nil, err
			}
			fields[i].Nested = nested
		}
	}
	return &flatbuffers.TableSchema{
		Name:   s.Name,
		Fields: fields,
	}, nil
}

func parseFieldType(t string) (flatbuffers.FieldType, error) {
	switch t {
	case "int8":
		return flatbuffers.TypeInt8, nil
	case "uint8":
		return flatbuffers.TypeUint8, nil
	case "int16":
		return flatbuffers.TypeInt16, nil
	case "uint16":
		return flatbuffers.TypeUint16, nil
	case "int32":
		return flatbuffers.TypeInt32, nil
	case "uint32":
		return flatbuffers.TypeUint32, nil
	case "int64":
		return flatbuffers.TypeInt64, nil
	case "uint64":
		return flatbuffers.TypeUint64, nil
	case "float32":
		return flatbuffers.TypeFloat32, nil
	case "float64":
		return flatbuffers.TypeFloat64, nil
	case "bool":
		return flatbuffers.TypeBool, nil
	case "string":
		return flatbuffers.TypeString, nil
	case "table":
		return flatbuffers.TypeTable, nil
	case "scalar_array":
		return flatbuffers.TypeScalarArray, nil
	case "table_array":
		return flatbuffers.TypeTableArray, nil
	default:
		return 0, fmt.Errorf("unknown field type: %s", t)
	}
}

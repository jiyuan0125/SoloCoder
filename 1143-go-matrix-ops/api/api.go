package api

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/matrix-ops/matrix-service/matrix"
)

type MatrixJSON [][]float64

const (
	epsilon = 1e-10
)

type MultiplyRequest struct {
	MatrixA   MatrixJSON `json:"a"`
	MatrixB   MatrixJSON `json:"b"`
	Compress  *bool      `json:"compress,omitempty"`
}

type UnaryRequest struct {
	Matrix   MatrixJSON `json:"matrix"`
	Compress *bool      `json:"compress,omitempty"`
}

type MatrixResponse struct {
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type DeterminantResponse struct {
	Success bool    `json:"success"`
	Result  float64 `json:"result,omitempty"`
	Error   string  `json:"error,omitempty"`
}

func formatMatrixResponse(m *matrix.Matrix, preferCompressed bool) (MatrixResponse, error) {
	shouldCompress := false
	if preferCompressed {
		shouldCompress = true
	}

	if shouldCompress || m.ShouldCompress() {
		sparse := m.ToSparse()
		data, err := json.Marshal(sparse)
		if err != nil {
			return MatrixResponse{}, err
		}
		return MatrixResponse{
			Success: true,
			Result:  data,
		}, nil
	}

	dense := m.To2D()
	data, err := json.Marshal(dense)
	if err != nil {
		return MatrixResponse{}, err
	}
	return MatrixResponse{
		Success: true,
		Result:  data,
	}, nil
}

func HandleMultiply(req MultiplyRequest) ([]byte, error) {
	matA, err := matrix.New([][]float64(req.MatrixA))
	if err != nil {
		return errorResponse(err.Error())
	}

	matB, err := matrix.New([][]float64(req.MatrixB))
	if err != nil {
		return errorResponse(err.Error())
	}

	result, err := matrix.Multiply(matA, matB)
	if err != nil {
		return errorResponse(err.Error())
	}

	compress := req.Compress != nil && *req.Compress
	resp, err := formatMatrixResponse(result, compress)
	if err != nil {
		return errorResponse(err.Error())
	}

	return json.Marshal(resp)
}

func HandleTranspose(req UnaryRequest) ([]byte, error) {
	mat, err := matrix.New([][]float64(req.Matrix))
	if err != nil {
		return errorResponse(err.Error())
	}

	result, err := mat.Transpose()
	if err != nil {
		return errorResponse(err.Error())
	}

	compress := req.Compress != nil && *req.Compress
	resp, err := formatMatrixResponse(result, compress)
	if err != nil {
		return errorResponse(err.Error())
	}

	return json.Marshal(resp)
}

func HandleDeterminant(req UnaryRequest) ([]byte, error) {
	mat, err := matrix.New([][]float64(req.Matrix))
	if err != nil {
		return errorResponse(err.Error())
	}

	det, err := mat.Determinant()
	if err != nil {
		return json.Marshal(DeterminantResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if math.Abs(det) < epsilon {
		det = 0
	}

	return json.Marshal(DeterminantResponse{
		Success: true,
		Result:  det,
	})
}

func errorResponse(message string) ([]byte, error) {
	return json.Marshal(MatrixResponse{
		Success: false,
		Error:   message,
	})
}

func FormatMatrixCompact(m [][]float64) string {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("%v", m)
	}
	return string(data)
}

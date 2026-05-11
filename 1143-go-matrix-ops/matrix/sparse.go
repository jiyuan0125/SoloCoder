package matrix

import (
	"encoding/json"
	"math"
)

type SparseEntry struct {
	Row   int     `json:"r"`
	Col   int     `json:"c"`
	Value float64 `json:"v"`
}

type SparseMatrix struct {
	Rows    int           `json:"rows"`
	Cols    int           `json:"cols"`
	Entries []SparseEntry `json:"entries"`
}

func (m *Matrix) ShouldCompress() bool {
	total := m.Rows * m.Cols
	if total == 0 {
		return false
	}

	zeroCount := 0
	for _, row := range m.Data {
		for _, v := range row {
			if math.Abs(v) < epsilon {
				zeroCount++
			}
		}
	}

	return float64(zeroCount)/float64(total) >= sparseRatio
}

func (m *Matrix) ToSparse() *SparseMatrix {
	entries := make([]SparseEntry, 0)
	for i, row := range m.Data {
		for j, v := range row {
			if math.Abs(v) >= epsilon {
				entries = append(entries, SparseEntry{
					Row:   i,
					Col:   j,
					Value: v,
				})
			}
		}
	}
	return &SparseMatrix{
		Rows:    m.Rows,
		Cols:    m.Cols,
		Entries: entries,
	}
}

func (s *SparseMatrix) ToDense() (*Matrix, error) {
	data := make([][]float64, s.Rows)
	for i := range data {
		data[i] = make([]float64, s.Cols)
	}

	for _, entry := range s.Entries {
		if entry.Row >= 0 && entry.Row < s.Rows && entry.Col >= 0 && entry.Col < s.Cols {
			data[entry.Row][entry.Col] = entry.Value
		}
	}

	return New(data)
}

type MatrixRepresentation struct {
	Format string          `json:"format"`
	Data   json.RawMessage `json:"data"`
}

func SerializeMatrix(m *Matrix, preferCompressed bool) ([]byte, error) {
	shouldCompress := preferCompressed || m.ShouldCompress()

	if shouldCompress {
		sparse := m.ToSparse()
		rep := MatrixRepresentation{
			Format: "sparse",
		}
		var err error
		rep.Data, err = json.Marshal(sparse)
		if err != nil {
			return nil, err
		}
		return json.Marshal(rep)
	}

	rep := MatrixRepresentation{
		Format: "dense",
	}
	var err error
	rep.Data, err = json.Marshal(m.To2D())
	if err != nil {
		return nil, err
	}
	return json.Marshal(rep)
}

func DeserializeMatrix(data []byte) (*Matrix, error) {
	var rep MatrixRepresentation
	if err := json.Unmarshal(data, &rep); err != nil {
		var dense [][]float64
		if err2 := json.Unmarshal(data, &dense); err2 == nil {
			return New(dense)
		}
		return nil, err
	}

	switch rep.Format {
	case "sparse":
		var sparse SparseMatrix
		if err := json.Unmarshal(rep.Data, &sparse); err != nil {
			return nil, err
		}
		return sparse.ToDense()
	case "dense":
		var dense [][]float64
		if err := json.Unmarshal(rep.Data, &dense); err != nil {
			return nil, err
		}
		return New(dense)
	default:
		var dense [][]float64
		if err := json.Unmarshal(rep.Data, &dense); err == nil {
			return New(dense)
		}
		return nil, nil
	}
}

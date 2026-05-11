package matrix

import (
	"errors"
	"fmt"
	"math"
)

type Matrix struct {
	Rows    int
	Cols    int
	Data    [][]float64
}

const (
	epsilon     = 1e-10
	sparseRatio = 0.7
)

func New(data [][]float64) (*Matrix, error) {
	if len(data) == 0 || (len(data) == 1 && len(data[0]) == 0) {
		return nil, errors.New("空矩阵不允许")
	}

	cols := len(data[0])
	for i, row := range data {
		if len(row) != cols {
			return nil, fmt.Errorf("矩阵行长度不一致，第%d行有%d个元素，预期%d个", i+1, len(row), cols)
		}
	}

	m := &Matrix{
		Rows: len(data),
		Cols: cols,
		Data: make([][]float64, len(data)),
	}
	for i, row := range data {
		m.Data[i] = make([]float64, len(row))
		copy(m.Data[i], row)
	}

	return m, nil
}

func (m *Matrix) To2D() [][]float64 {
	result := make([][]float64, m.Rows)
	for i := range m.Data {
		result[i] = make([]float64, m.Cols)
		copy(result[i], m.Data[i])
	}
	return result
}

func (m *Matrix) IsZeroMatrix() bool {
	for _, row := range m.Data {
		for _, v := range row {
			if math.Abs(v) > epsilon {
				return false
			}
		}
	}
	return true
}

func (m *Matrix) IsSquare() bool {
	return m.Rows == m.Cols
}

func Multiply(a, b *Matrix) (*Matrix, error) {
	if a.Cols != b.Rows {
		return nil, fmt.Errorf("矩阵维度不匹配，不能相乘：%dx%d 和 %dx%d", a.Rows, a.Cols, b.Rows, b.Cols)
	}

	if a.IsZeroMatrix() || b.IsZeroMatrix() {
		m, _ := New(makeZeroMatrix(a.Rows, b.Cols))
		return m, nil
	}

	result := make([][]float64, a.Rows)
	for i := range result {
		result[i] = make([]float64, b.Cols)
	}

	for i := 0; i < a.Rows; i++ {
		for j := 0; j < b.Cols; j++ {
			var sum float64
			for k := 0; k < a.Cols; k++ {
				sum += a.Data[i][k] * b.Data[k][j]
			}
			result[i][j] = sum
		}
	}

	return New(result)
}

func makeZeroMatrix(rows, cols int) [][]float64 {
	m := make([][]float64, rows)
	for i := range m {
		m[i] = make([]float64, cols)
	}
	return m
}

func (m *Matrix) Transpose() (*Matrix, error) {
	result := make([][]float64, m.Cols)
	for i := range result {
		result[i] = make([]float64, m.Rows)
	}

	for i := 0; i < m.Rows; i++ {
		for j := 0; j < m.Cols; j++ {
			result[j][i] = m.Data[i][j]
		}
	}

	return New(result)
}

func (m *Matrix) Determinant() (float64, error) {
	if !m.IsSquare() {
		return 0, errors.New("只有方阵才能计算行列式")
	}

	if m.Rows == 1 {
		return m.Data[0][0], nil
	}

	if m.Rows == 2 {
		return m.Data[0][0]*m.Data[1][1] - m.Data[0][1]*m.Data[1][0], nil
	}

	mat := make([][]float64, m.Rows)
	for i := range mat {
		mat[i] = make([]float64, m.Cols)
		copy(mat[i], m.Data[i])
	}

	var det float64 = 1.0
	n := m.Rows

	for i := 0; i < n; i++ {
		pivotRow := i
		for j := i + 1; j < n; j++ {
			if math.Abs(mat[j][i]) > math.Abs(mat[pivotRow][i]) {
				pivotRow = j
			}
		}

		if math.Abs(mat[pivotRow][i]) < epsilon {
			return 0, nil
		}

		if pivotRow != i {
			mat[i], mat[pivotRow] = mat[pivotRow], mat[i]
			det = -det
		}

		pivot := mat[i][i]
		det *= pivot

		for j := i + 1; j < n; j++ {
			factor := mat[j][i] / pivot
			for k := i; k < n; k++ {
				mat[j][k] -= factor * mat[i][k]
			}
		}
	}

	if math.Abs(det) < epsilon {
		return 0, nil
	}

	return det, nil
}

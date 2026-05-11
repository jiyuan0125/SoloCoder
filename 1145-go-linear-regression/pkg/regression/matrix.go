package regression

import (
	"errors"
)

func transpose(matrix [][]float64) [][]float64 {
	rows := len(matrix)
	cols := len(matrix[0])
	result := make([][]float64, cols)
	for i := range result {
		result[i] = make([]float64, rows)
		for j := 0; j < rows; j++ {
			result[i][j] = matrix[j][i]
		}
	}
	return result
}

func multiply(a, b [][]float64) [][]float64 {
	rowsA := len(a)
	colsA := len(a[0])
	colsB := len(b[0])
	result := make([][]float64, rowsA)
	for i := range result {
		result[i] = make([]float64, colsB)
		for j := 0; j < colsB; j++ {
			for k := 0; k < colsA; k++ {
				result[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return result
}

func multiplyVector(matrix [][]float64, vector []float64) []float64 {
	rows := len(matrix)
	cols := len(matrix[0])
	result := make([]float64, rows)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			result[i] += matrix[i][j] * vector[j]
		}
	}
	return result
}

func inverse(matrix [][]float64) ([][]float64, error) {
	n := len(matrix)
	augmented := make([][]float64, n)
	for i := range augmented {
		augmented[i] = make([]float64, 2*n)
		copy(augmented[i], matrix[i])
		augmented[i][n+i] = 1
	}
	
	for col := 0; col < n; col++ {
		pivotRow := col
		for row := col + 1; row < n; row++ {
			if augmented[row][col] > augmented[pivotRow][col] {
				pivotRow = row
			}
		}
		augmented[col], augmented[pivotRow] = augmented[pivotRow], augmented[col]
		
		if augmented[col][col] == 0 {
			return nil, errors.New("singular matrix")
		}
		
		div := augmented[col][col]
		for j := 0; j < 2*n; j++ {
			augmented[col][j] /= div
		}
		
		for row := 0; row < n; row++ {
			if row != col && augmented[row][col] != 0 {
				factor := augmented[row][col]
				for j := 0; j < 2*n; j++ {
					augmented[row][j] -= factor * augmented[col][j]
				}
			}
		}
	}
	
	result := make([][]float64, n)
	for i := range result {
		result[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			result[i][j] = augmented[i][j+n]
		}
	}
	return result, nil
}

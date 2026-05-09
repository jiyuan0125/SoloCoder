package rs

type Matrix [][]byte

func NewMatrix(rows, cols int) Matrix {
	m := make(Matrix, rows)
	for i := range m {
		m[i] = make([]byte, cols)
	}
	return m
}

func (m Matrix) Rows() int {
	return len(m)
}

func (m Matrix) Cols() int {
	if len(m) == 0 {
		return 0
	}
	return len(m[0])
}

func (m Matrix) Get(row, col int) byte {
	return m[row][col]
}

func (m Matrix) Set(row, col int, val byte) {
	m[row][col] = val
}

func (m Matrix) Row(row int) []byte {
	return m[row]
}

func Vandermonde(rows, cols int) Matrix {
	m := NewMatrix(rows, cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			m[i][j] = Pow(byte(i), j)
		}
	}
	return m
}

func Identity(n int) Matrix {
	m := NewMatrix(n, n)
	for i := 0; i < n; i++ {
		m[i][i] = 1
	}
	return m
}

func (m Matrix) SubMatrix(rowStart, colStart, rowEnd, colEnd int) Matrix {
	rows := rowEnd - rowStart
	cols := colEnd - colStart
	result := NewMatrix(rows, cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			result[i][j] = m[rowStart+i][colStart+j]
		}
	}
	return result
}

func (m Matrix) Append(cols int) Matrix {
	rows := m.Rows()
	result := NewMatrix(rows, m.Cols()+cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < m.Cols(); j++ {
			result[i][j] = m[i][j]
		}
	}
	return result
}

func (m Matrix) SwapRows(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m Matrix) Invert() (Matrix, error) {
	size := m.Rows()
	if size != m.Cols() {
		return nil, ErrNonSquareMatrix
	}

	aug := NewMatrix(size, size*2)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			aug[i][j] = m[i][j]
		}
		aug[i][size+i] = 1
	}

	for col := 0; col < size; col++ {
		pivotRow := -1
		for row := col; row < size; row++ {
			if aug[row][col] != 0 {
				pivotRow = row
				break
			}
		}
		if pivotRow == -1 {
			return nil, ErrSingularMatrix
		}

		aug.SwapRows(col, pivotRow)

		pivotVal := aug[col][col]
		invPivot := Inv(pivotVal)
		for j := 0; j < size*2; j++ {
			aug[col][j] = Mul(aug[col][j], invPivot)
		}

		for row := 0; row < size; row++ {
			if row != col && aug[row][col] != 0 {
				factor := aug[row][col]
				for j := 0; j < size*2; j++ {
					aug[row][j] = Add(aug[row][j], Mul(factor, aug[col][j]))
				}
			}
		}
	}

	result := NewMatrix(size, size)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			result[i][j] = aug[i][size+j]
		}
	}
	return result, nil
}

func (m Matrix) Multiply(other Matrix) (Matrix, error) {
	if m.Cols() != other.Rows() {
		return nil, ErrMatrixMismatch
	}

	rows := m.Rows()
	cols := other.Cols()
	k := m.Cols()
	result := NewMatrix(rows, cols)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			sum := byte(0)
			for idx := 0; idx < k; idx++ {
				sum = Add(sum, Mul(m[i][idx], other[idx][j]))
			}
			result[i][j] = sum
		}
	}
	return result, nil
}

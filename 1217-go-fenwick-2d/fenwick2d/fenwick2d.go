package fenwick2d

import "errors"

var (
	ErrOutOfBounds = errors.New("fenwick2d: coordinates out of bounds")
)

type Fenwick2D struct {
	rows int
	cols int
	data [][]int64
}

func New(rows, cols int) (*Fenwick2D, error) {
	if rows <= 0 || cols <= 0 {
		return nil, ErrOutOfBounds
	}
	data := make([][]int64, rows+1)
	for i := 0; i <= rows; i++ {
		data[i] = make([]int64, cols+1)
	}
	return &Fenwick2D{
		rows: rows,
		cols: cols,
		data: data,
	}, nil
}

func NewWithMatrix(matrix [][]int64) (*Fenwick2D, error) {
	if len(matrix) == 0 {
		return nil, ErrOutOfBounds
	}
	rows := len(matrix)
	cols := len(matrix[0])
	if cols == 0 {
		return nil, ErrOutOfBounds
	}
	for _, row := range matrix {
		if len(row) != cols {
			return nil, ErrOutOfBounds
		}
	}

	f := &Fenwick2D{
		rows: rows,
		cols: cols,
		data: make([][]int64, rows+1),
	}

	for i := 0; i <= rows; i++ {
		f.data[i] = make([]int64, cols+1)
	}

	for i := 1; i <= rows; i++ {
		for j := 1; j <= cols; j++ {
			f.data[i][j] = matrix[i-1][j-1]
		}
	}

	for i := 1; i <= rows; i++ {
		for j := 1; j <= cols; j++ {
			if pj := j + lsb(j); pj <= cols {
				f.data[i][pj] += f.data[i][j]
			}
		}
	}

	for j := 1; j <= cols; j++ {
		for i := 1; i <= rows; i++ {
			if pi := i + lsb(i); pi <= rows {
				f.data[pi][j] += f.data[i][j]
			}
		}
	}

	return f, nil
}

func (f *Fenwick2D) Rows() int {
	return f.rows
}

func (f *Fenwick2D) Cols() int {
	return f.cols
}

func (f *Fenwick2D) Update(x, y int, delta int64) error {
	if x < 1 || x > f.rows || y < 1 || y > f.cols {
		return ErrOutOfBounds
	}

	for i := x; i <= f.rows; i += lsb(i) {
		for j := y; j <= f.cols; j += lsb(j) {
			f.data[i][j] += delta
		}
	}
	return nil
}

func (f *Fenwick2D) Query(x, y int) (int64, error) {
	if x <= 0 || y <= 0 {
		return 0, nil
	}
	if x > f.rows || y > f.cols {
		return 0, ErrOutOfBounds
	}

	var sum int64
	for i := x; i > 0; i -= lsb(i) {
		for j := y; j > 0; j -= lsb(j) {
			sum += f.data[i][j]
		}
	}
	return sum, nil
}

func (f *Fenwick2D) QueryRange(lx, ly, rx, ry int) (int64, error) {
	if lx > rx || ly > ry {
		return 0, nil
	}

	if lx < 1 {
		lx = 1
	}
	if ly < 1 {
		ly = 1
	}
	if rx > f.rows {
		return 0, ErrOutOfBounds
	}
	if ry > f.cols {
		return 0, ErrOutOfBounds
	}

	sum1, _ := f.Query(rx, ry)
	sum2, _ := f.Query(lx-1, ry)
	sum3, _ := f.Query(rx, ly-1)
	sum4, _ := f.Query(lx-1, ly-1)

	return sum1 - sum2 - sum3 + sum4, nil
}

func lsb(x int) int {
	return x & -x
}

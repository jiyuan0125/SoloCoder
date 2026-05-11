package tableparser

type Alignment int

const (
	AlignmentLeft Alignment = iota
	AlignmentRight
	AlignmentCenter
)

func (a Alignment) String() string {
	switch a {
	case AlignmentLeft:
		return "left"
	case AlignmentRight:
		return "right"
	case AlignmentCenter:
		return "center"
	default:
		return "left"
	}
}

type Table struct {
	Headers    []string
	Rows       [][]string
	Alignments []Alignment
}

func NewTable(headers []string, rows [][]string, alignments []Alignment) *Table {
	if alignments == nil {
		alignments = make([]Alignment, len(headers))
		for i := range alignments {
			alignments[i] = AlignmentLeft
		}
	}
	return &Table{
		Headers:    headers,
		Rows:       rows,
		Alignments: alignments,
	}
}

package mdparser

type BlockType int

const (
	BlockParagraph BlockType = iota
	BlockHeading
	BlockCode
	BlockQuote
	BlockList
	BlockListItem
	BlockTable
	BlockTableHeader
	BlockTableRow
	BlockTableCell
)

type Block struct {
	Type         BlockType
	Content      string
	Children     []*Block
	StartNum     int
	Ordered      bool
	Level        int
	Language     string
	Checked      bool
	HasCheckbox  bool
	Alignments   []string
	CellAlign    string
	IsHeader     bool
	IsCompleted  bool
}

type InlineNodeType int

const (
	InlineText InlineNodeType = iota
	InlineBold
	InlineItalic
	InlineLink
	InlineCode
	InlineStrikethrough
	InlineImage
)

type InlineNode struct {
	Type     InlineNodeType
	Content  string
	URL      string
	Alt      string
	Title    string
	Children []*InlineNode
}

type Parser struct {
	lines  []string
	index  int
}

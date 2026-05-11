package externalsort

type chunk struct {
	records []int64
	pos     int
}

func newChunk(records []int64) *chunk {
	return &chunk{
		records: records,
		pos:     0,
	}
}

func (c *chunk) peek() (int64, bool) {
	if c.pos >= len(c.records) {
		return 0, false
	}
	return c.records[c.pos], true
}

func (c *chunk) next() (int64, bool) {
	if c.pos >= len(c.records) {
		return 0, false
	}
	val := c.records[c.pos]
	c.pos++
	return val, true
}

func (c *chunk) exhausted() bool {
	return c.pos >= len(c.records)
}

func (c *chunk) size() int {
	return len(c.records)
}

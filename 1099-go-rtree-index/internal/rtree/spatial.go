package rtree

type SpatialObject interface {
	ID() string
	Bounds() (minX, minY, maxX, maxY float64)
	Metadata() map[string]interface{}
}

type MBR struct {
	MinX, MinY, MaxX, MaxY float64
}

func NewMBR(minX, minY, maxX, maxY float64) MBR {
	return MBR{
		MinX: minX,
		MinY: minY,
		MaxX: maxX,
		MaxY: maxY,
	}
}

func (m MBR) Width() float64 {
	return m.MaxX - m.MinX
}

func (m MBR) Height() float64 {
	return m.MaxY - m.MinY
}

func (m MBR) Area() float64 {
	return m.Width() * m.Height()
}

func (m MBR) Contains(other MBR) bool {
	return m.MinX <= other.MinX && m.MinY <= other.MinY && m.MaxX >= other.MaxX && m.MaxY >= other.MaxY
}

func (m MBR) Intersects(other MBR) bool {
	return !(m.MaxX < other.MinX || m.MinX > other.MaxX || m.MaxY < other.MinY || m.MinY > other.MaxY)
}

func (m MBR) Union(other MBR) MBR {
	return MBR{
		MinX: min(m.MinX, other.MinX),
		MinY: min(m.MinY, other.MinY),
		MaxX: max(m.MaxX, other.MaxX),
		MaxY: max(m.MaxY, other.MaxY),
	}
}

func (m MBR) Expansion(other MBR) float64 {
	union := m.Union(other)
	return union.Area() - m.Area()
}

func (m MBR) DistanceToPoint(x, y float64) float64 {
	if x >= m.MinX && x <= m.MaxX && y >= m.MinY && y <= m.MaxY {
		return 0
	}

	dx := 0.0
	if x < m.MinX {
		dx = m.MinX - x
	} else if x > m.MaxX {
		dx = x - m.MaxX
	}

	dy := 0.0
	if y < m.MinY {
		dy = m.MinY - y
	} else if y > m.MaxY {
		dy = y - m.MaxY
	}

	return dx*dx + dy*dy
}

type SimpleObject struct {
	id       string
	bounds   MBR
	metadata map[string]interface{}
}

func NewSimpleObject(id string, minX, minY, maxX, maxY float64, metadata map[string]interface{}) *SimpleObject {
	return &SimpleObject{
		id:       id,
		bounds:   NewMBR(minX, minY, maxX, maxY),
		metadata: metadata,
	}
}

func (s *SimpleObject) ID() string {
	return s.id
}

func (s *SimpleObject) Bounds() (minX, minY, maxX, maxY float64) {
	return s.bounds.MinX, s.bounds.MinY, s.bounds.MaxX, s.bounds.MaxY
}

func (s *SimpleObject) Metadata() map[string]interface{} {
	return s.metadata
}

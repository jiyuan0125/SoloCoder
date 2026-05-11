package geom

import "math"

const Epsilon = 1e-9

type Point struct {
	X float64
	Y float64
}

func NewPoint(x, y float64) Point {
	return Point{X: x, Y: y}
}

func (p Point) Add(q Point) Point {
	return Point{X: p.X + q.X, Y: p.Y + q.Y}
}

func (p Point) Sub(q Point) Point {
	return Point{X: p.X - q.X, Y: p.Y - q.Y}
}

func (p Point) Mul(s float64) Point {
	return Point{X: p.X * s, Y: p.Y * s}
}

func (p Point) Dot(q Point) float64 {
	return p.X*q.X + p.Y*q.Y
}

func (p Point) Cross(q Point) float64 {
	return p.X*q.Y - p.Y*q.X
}

func (p Point) Length() float64 {
	return math.Sqrt(p.Dot(p))
}

func (p Point) Distance(q Point) float64 {
	return p.Sub(q).Length()
}

func (p Point) DistanceSq(q Point) float64 {
	d := p.Sub(q)
	return d.Dot(d)
}

func (p Point) Normalize() Point {
	len := p.Length()
	if len < Epsilon {
		return Point{0, 0}
	}
	return Point{X: p.X / len, Y: p.Y / len}
}

func (p Point) Equals(q Point) bool {
	return math.Abs(p.X-q.X) < Epsilon && math.Abs(p.Y-q.Y) < Epsilon
}

type Line struct {
	A, B, C float64
}

func NewLine(a, b, c float64) Line {
	return Line{A: a, B: b, C: c}
}

func LineFromPoints(p1, p2 Point) Line {
	a := p2.Y - p1.Y
	b := p1.X - p2.X
	c := p2.X*p1.Y - p1.X*p2.Y
	return Line{A: a, B: b, C: c}
}

func PerpendicularBisector(p1, p2 Point) Line {
	mid := Point{X: (p1.X + p2.X) / 2, Y: (p1.Y + p2.Y) / 2}
	dir := p2.Sub(p1)
	a := dir.X
	b := dir.Y
	c := -(a*mid.X + b*mid.Y)
	return Line{A: a, B: b, C: c}
}

func (l Line) Intersect(m Line) (Point, bool) {
	det := l.A*m.B - l.B*m.A
	if math.Abs(det) < Epsilon {
		return Point{}, false
	}
	x := (l.B*m.C - l.C*m.B) / det
	y := (l.C*m.A - l.A*m.C) / det
	return Point{X: x, Y: y}, true
}

func (l Line) DistanceToPoint(p Point) float64 {
	return math.Abs(l.A*p.X+l.B*p.Y+l.C) / math.Sqrt(l.A*l.A+l.B*l.B)
}

type Segment struct {
	Start Point
	End   Point
}

func NewSegment(start, end Point) Segment {
	return Segment{Start: start, End: end}
}

func (s Segment) Line() Line {
	return LineFromPoints(s.Start, s.End)
}

func (s Segment) Length() float64 {
	return s.Start.Distance(s.End)
}

func (s Segment) ContainsPoint(p Point) bool {
	if !OnLine(s.Start, s.End, p) {
		return false
	}
	return Between(s.Start, p, s.End)
}

type Circle struct {
	Center Point
	Radius float64
}

func NewCircle(center Point, radius float64) Circle {
	return Circle{Center: center, Radius: radius}
}

func Circumcircle(a, b, c Point) (Circle, bool) {
	d := 2 * (a.X*(b.Y-c.Y) + b.X*(c.Y-a.Y) + c.X*(a.Y-b.Y))
	if math.Abs(d) < Epsilon {
		return Circle{}, false
	}

	sa2 := a.X*a.X + a.Y*a.Y
	sb2 := b.X*b.X + b.Y*b.Y
	sc2 := c.X*c.X + c.Y*c.Y

	ux := (sa2*(b.Y-c.Y) + sb2*(c.Y-a.Y) + sc2*(a.Y-b.Y)) / d
	uy := (sa2*(c.X-b.X) + sb2*(a.X-c.X) + sc2*(b.X-a.X)) / d

	center := Point{X: ux, Y: uy}
	radius := center.Distance(a)
	return Circle{Center: center, Radius: radius}, true
}

func Sign(x float64) int {
	if x < -Epsilon {
		return -1
	}
	if x > Epsilon {
		return 1
	}
	return 0
}

func CrossProduct(a, b, c Point) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

func OnLine(a, b, p Point) bool {
	return math.Abs(CrossProduct(a, b, p)) < Epsilon
}

func Between(a, p, b Point) bool {
	return math.Min(a.X, b.X)-Epsilon <= p.X && p.X <= math.Max(a.X, b.X)+Epsilon &&
		math.Min(a.Y, b.Y)-Epsilon <= p.Y && p.Y <= math.Max(a.Y, b.Y)+Epsilon
}

func Colinear(a, b, c Point) bool {
	return math.Abs(CrossProduct(a, b, c)) < Epsilon
}

func Concyclic(a, b, c, d Point) bool {
	circle, ok := Circumcircle(a, b, c)
	if !ok {
		return false
	}
	return math.Abs(circle.Center.Distance(d)-circle.Radius) < Epsilon
}

type Rectangle struct {
	Min Point
	Max Point
}

func NewRectangle(minX, minY, maxX, maxY float64) Rectangle {
	return Rectangle{
		Min: Point{X: minX, Y: minY},
		Max: Point{X: maxX, Y: maxY},
	}
}

func (r Rectangle) Width() float64 {
	return r.Max.X - r.Min.X
}

func (r Rectangle) Height() float64 {
	return r.Max.Y - r.Min.Y
}

func (r Rectangle) ContainsPoint(p Point) bool {
	return p.X >= r.Min.X-Epsilon && p.X <= r.Max.X+Epsilon &&
		p.Y >= r.Min.Y-Epsilon && p.Y <= r.Max.Y+Epsilon
}

func (r Rectangle) Center() Point {
	return Point{
		X: (r.Min.X + r.Max.X) / 2,
		Y: (r.Min.Y + r.Max.Y) / 2,
	}
}

func (r Rectangle) Expand(factor float64) Rectangle {
	center := r.Center()
	halfW := r.Width() / 2
	halfH := r.Height() / 2
	return Rectangle{
		Min: Point{X: center.X - halfW*factor, Y: center.Y - halfH*factor},
		Max: Point{X: center.X + halfW*factor, Y: center.Y + halfH*factor},
	}
}

func PolygonArea(points []Point) float64 {
	n := len(points)
	if n < 3 {
		return 0
	}
	area := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += points[i].X*points[j].Y - points[j].X*points[i].Y
	}
	return math.Abs(area) / 2
}

func ClipSegmentToRect(seg Segment, rect Rectangle) (Segment, bool) {
	x1, y1 := seg.Start.X, seg.Start.Y
	x2, y2 := seg.End.X, seg.End.Y
	xmin, ymin := rect.Min.X, rect.Min.Y
	xmax, ymax := rect.Max.X, rect.Max.Y

	inside1 := x1 >= xmin-Epsilon && x1 <= xmax+Epsilon && y1 >= ymin-Epsilon && y1 <= ymax+Epsilon
	inside2 := x2 >= xmin-Epsilon && x2 <= xmax+Epsilon && y2 >= ymin-Epsilon && y2 <= ymax+Epsilon

	if inside1 && inside2 {
		return seg, true
	}

	if inside1 {
		if clipped, ok := clipPoint(x2, y2, x1, y1, xmin, ymin, xmax, ymax); ok {
			return Segment{Start: seg.Start, End: clipped}, true
		}
		return Segment{}, false
	}

	if inside2 {
		if clipped, ok := clipPoint(x1, y1, x2, y2, xmin, ymin, xmax, ymax); ok {
			return Segment{Start: clipped, End: seg.End}, true
		}
		return Segment{}, false
	}

	clipped1, ok1 := clipPoint(x1, y1, x2, y2, xmin, ymin, xmax, ymax)
	clipped2, ok2 := clipPoint(x2, y2, x1, y1, xmin, ymin, xmax, ymax)

	if ok1 && ok2 {
		return Segment{Start: clipped1, End: clipped2}, true
	}
	return Segment{}, false
}

func clipPoint(x, y, refX, refY, xmin, ymin, xmax, ymax float64) (Point, bool) {
	dx := x - refX
	dy := y - refY

	if math.Abs(dx) < Epsilon && math.Abs(dy) < Epsilon {
		return Point{}, false
	}

	if x < xmin-Epsilon {
		t := (xmin - refX) / dx
		ny := refY + t*dy
		if ny >= ymin-Epsilon && ny <= ymax+Epsilon {
			return Point{X: xmin, Y: ny}, true
		}
	}

	if x > xmax+Epsilon {
		t := (xmax - refX) / dx
		ny := refY + t*dy
		if ny >= ymin-Epsilon && ny <= ymax+Epsilon {
			return Point{X: xmax, Y: ny}, true
		}
	}

	if y < ymin-Epsilon {
		t := (ymin - refY) / dy
		nx := refX + t*dx
		if nx >= xmin-Epsilon && nx <= xmax+Epsilon {
			return Point{X: nx, Y: ymin}, true
		}
	}

	if y > ymax+Epsilon {
		t := (ymax - refY) / dy
		nx := refX + t*dx
		if nx >= xmin-Epsilon && nx <= xmax+Epsilon {
			return Point{X: nx, Y: ymax}, true
		}
	}

	return Point{}, false
}

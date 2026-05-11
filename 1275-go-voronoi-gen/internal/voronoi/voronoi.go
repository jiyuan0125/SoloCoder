package voronoi

import (
	"math"

	"voronoi/internal/geom"
)

const eps = geom.Epsilon

type HalfEdge struct {
	Start  geom.Point
	End    geom.Point
	Site1  geom.Point
	Site2  geom.Point
	HasEnd bool
}

type Diagram struct {
	Seeds []geom.Point
	Edges []HalfEdge
	Cells []*VoronoiCell
}

type VoronoiCell struct {
	Seed   geom.Point
	Points []geom.Point
	Edges  []HalfEdge
	Area   float64
}

func Generate(seeds []geom.Point) *Diagram {
	if len(seeds) < 2 {
		return &Diagram{Seeds: seeds}
	}

	unique := removeDuplicates(seeds)
	if len(unique) < 2 {
		return &Diagram{Seeds: unique}
	}

	if areAllColinear(unique) {
		return generateColinear(unique)
	}

	return generateVoronoiEdges(unique)
}

func generateVoronoiEdges(seeds []geom.Point) *Diagram {
	diagram := &Diagram{Seeds: seeds}
	diagram.Edges = []HalfEdge{}

	n := len(seeds)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			s1, s2 := seeds[i], seeds[j]

			mid := geom.Point{X: (s1.X + s2.X) / 2, Y: (s1.Y + s2.Y) / 2}
			dir := geom.Point{X: s2.Y - s1.Y, Y: s1.X - s2.X}
			dir = dir.Normalize()

			vertices := []geom.Point{}

			for k := 0; k < n; k++ {
				if k == i || k == j {
					continue
				}

				s3 := seeds[k]
				circle, ok := geom.Circumcircle(s1, s2, s3)
				if !ok {
					continue
				}

				center := circle.Center
				radius := circle.Radius

				allOutside := true
				for l := 0; l < n; l++ {
					if l == i || l == j || l == k {
						continue
					}
					dist := seeds[l].Distance(center)
					if dist < radius-eps {
						allOutside = false
						break
					}
				}

				if allOutside {
					vertices = append(vertices, center)
				}
			}

			if len(vertices) >= 2 {
				minDist := math.Inf(1)
				maxDist := math.Inf(-1)
				var minPt, maxPt geom.Point

				for _, v := range vertices {
					vec := v.Sub(mid)
					dist := vec.Dot(dir)
					if dist < minDist {
						minDist = dist
						minPt = v
					}
					if dist > maxDist {
						maxDist = dist
						maxPt = v
					}
				}

				if !minPt.Equals(maxPt) {
					diagram.Edges = append(diagram.Edges, HalfEdge{
						Start:  minPt,
						End:    maxPt,
						Site1:  s1,
						Site2:  s2,
						HasEnd: true,
					})
				}
			} else if len(vertices) == 1 {
				diagram.Edges = append(diagram.Edges, HalfEdge{
					Start:  vertices[0],
					Site1:  s1,
					Site2:  s2,
					HasEnd: false,
				})
			}
		}
	}

	return diagram
}

func removeDuplicates(seeds []geom.Point) []geom.Point {
	seen := make(map[[2]float64]bool)
	var unique []geom.Point
	for _, seed := range seeds {
		key := [2]float64{
			float64(int64(seed.X*1e9)) / 1e9,
			float64(int64(seed.Y*1e9)) / 1e9,
		}
		if !seen[key] {
			seen[key] = true
			unique = append(unique, seed)
		}
	}
	return unique
}

func areAllColinear(seeds []geom.Point) bool {
	if len(seeds) <= 2 {
		return true
	}
	for i := 2; i < len(seeds); i++ {
		if !geom.Colinear(seeds[0], seeds[1], seeds[i]) {
			return false
		}
	}
	return true
}

func generateColinear(seeds []geom.Point) *Diagram {
	diagram := &Diagram{Seeds: seeds}
	if len(seeds) < 2 {
		return diagram
	}

	sorted := make([]geom.Point, len(seeds))
	copy(sorted, seeds)

	dir := sorted[1].Sub(sorted[0])
	if dir.Length() < eps {
		return diagram
	}

	isHorizontal := math.Abs(dir.Y) < math.Abs(dir.X)

	if isHorizontal {
		for i := 1; i < len(sorted); i++ {
			for j := i; j > 0 && sorted[j-1].X > sorted[j].X; j-- {
				sorted[j-1], sorted[j] = sorted[j], sorted[j-1]
			}
		}
	} else {
		for i := 1; i < len(sorted); i++ {
			for j := i; j > 0 && sorted[j-1].Y > sorted[j].Y; j-- {
				sorted[j-1], sorted[j] = sorted[j], sorted[j-1]
			}
		}
	}

	for i := 0; i < len(sorted)-1; i++ {
		s1 := sorted[i]
		s2 := sorted[i+1]
		mid := geom.Point{X: (s1.X + s2.X) / 2, Y: (s1.Y + s2.Y) / 2}

		diagram.Edges = append(diagram.Edges, HalfEdge{
			Start:  mid,
			Site1:  s1,
			Site2:  s2,
			HasEnd: false,
		})
	}

	return diagram
}

func ClipDiagram(diagram *Diagram, bounds geom.Rectangle) *Diagram {
	if diagram == nil {
		return nil
	}

	clipped := &Diagram{Seeds: diagram.Seeds}
	clipped.Edges = make([]HalfEdge, 0, len(diagram.Edges))

	for _, edge := range diagram.Edges {
		if edge.HasEnd {
			seg := geom.NewSegment(edge.Start, edge.End)
			clippedSeg, ok := geom.ClipSegmentToRect(seg, bounds)
			if ok {
				clipped.Edges = append(clipped.Edges, HalfEdge{
					Start:  clippedSeg.Start,
					End:    clippedSeg.End,
					Site1:  edge.Site1,
					Site2:  edge.Site2,
					HasEnd: true,
				})
			}
			continue
		}

		dir := geom.Point{X: edge.Site2.Y - edge.Site1.Y, Y: edge.Site1.X - edge.Site2.X}
		dir = dir.Normalize()

		p1 := geom.Point{X: edge.Start.X + dir.X*1000000, Y: edge.Start.Y + dir.Y*1000000}
		p2 := geom.Point{X: edge.Start.X - dir.X*1000000, Y: edge.Start.Y - dir.Y*1000000}

		seg1 := geom.NewSegment(edge.Start, p1)
		seg2 := geom.NewSegment(edge.Start, p2)

		clipped1, ok1 := geom.ClipSegmentToRect(seg1, bounds)
		clipped2, ok2 := geom.ClipSegmentToRect(seg2, bounds)

		if ok1 && ok2 {
			clipped.Edges = append(clipped.Edges, HalfEdge{
				Start:  clipped1.End,
				End:    clipped2.End,
				Site1:  edge.Site1,
				Site2:  edge.Site2,
				HasEnd: true,
			})
		} else if ok1 {
			if bounds.ContainsPoint(edge.Start) {
				clipped.Edges = append(clipped.Edges, HalfEdge{
					Start:  edge.Start,
					End:    clipped1.End,
					Site1:  edge.Site1,
					Site2:  edge.Site2,
					HasEnd: true,
				})
			}
		} else if ok2 {
			if bounds.ContainsPoint(edge.Start) {
				clipped.Edges = append(clipped.Edges, HalfEdge{
					Start:  edge.Start,
					End:    clipped2.End,
					Site1:  edge.Site1,
					Site2:  edge.Site2,
					HasEnd: true,
				})
			}
		}
	}

	return clipped
}

func BuildCells(diagram *Diagram, bounds geom.Rectangle) []*VoronoiCell {
	if diagram == nil || len(diagram.Seeds) == 0 {
		return nil
	}

	if len(diagram.Seeds) == 1 {
		cell := &VoronoiCell{
			Seed: diagram.Seeds[0],
			Points: []geom.Point{
				{X: bounds.Min.X, Y: bounds.Min.Y},
				{X: bounds.Max.X, Y: bounds.Min.Y},
				{X: bounds.Max.X, Y: bounds.Max.Y},
				{X: bounds.Min.X, Y: bounds.Max.Y},
			},
		}
		cell.Area = geom.PolygonArea(cell.Points)
		return []*VoronoiCell{cell}
	}

	cells := make([]*VoronoiCell, len(diagram.Seeds))
	for i, seed := range diagram.Seeds {
		cells[i] = &VoronoiCell{Seed: seed}
		cells[i].Points = buildCellByClipping(seed, bounds, diagram.Seeds)
		if len(cells[i].Points) >= 3 {
			cells[i].Area = geom.PolygonArea(cells[i].Points)
		}
	}

	clipped := ClipDiagram(diagram, bounds)
	for _, edge := range clipped.Edges {
		idx1 := findSeedIndex(diagram.Seeds, edge.Site1)
		idx2 := findSeedIndex(diagram.Seeds, edge.Site2)

		if idx1 >= 0 {
			cells[idx1].Edges = append(cells[idx1].Edges, edge)
		}
		if idx2 >= 0 {
			cells[idx2].Edges = append(cells[idx2].Edges, edge)
		}
	}

	return cells
}

func buildCellByClipping(seed geom.Point, bounds geom.Rectangle, allSeeds []geom.Point) []geom.Point {
	polygon := []geom.Point{
		{X: bounds.Min.X, Y: bounds.Min.Y},
		{X: bounds.Max.X, Y: bounds.Min.Y},
		{X: bounds.Max.X, Y: bounds.Max.Y},
		{X: bounds.Min.X, Y: bounds.Max.Y},
	}

	for _, other := range allSeeds {
		if other.Equals(seed) {
			continue
		}

		bisector := geom.PerpendicularBisector(seed, other)
		polygon = clipPolygonByLine(polygon, bisector, seed, other)

		if len(polygon) < 3 {
			break
		}
	}

	return polygon
}

func clipPolygonByLine(polygon []geom.Point, line geom.Line, keepSide, otherSide geom.Point) []geom.Point {
	if len(polygon) < 2 {
		return polygon
	}

	result := []geom.Point{}

	keepVal := line.A*keepSide.X + line.B*keepSide.Y + line.C

	for i := 0; i < len(polygon); i++ {
		current := polygon[i]
		next := polygon[(i+1)%len(polygon)]

		currentVal := line.A*current.X + line.B*current.Y + line.C
		nextVal := line.A*next.X + line.B*next.Y + line.C

		currentInside := math.Abs(currentVal) < eps || math.Signbit(currentVal) == math.Signbit(keepVal)
		nextInside := math.Abs(nextVal) < eps || math.Signbit(nextVal) == math.Signbit(keepVal)

		if currentInside {
			if len(result) == 0 || !result[len(result)-1].Equals(current) {
				result = append(result, current)
			}
		}

		if currentInside != nextInside && math.Abs(currentVal) > eps && math.Abs(nextVal) > eps {
			segLine := geom.LineFromPoints(current, next)
			intersect, ok := line.Intersect(segLine)
			if ok {
				if len(result) == 0 || !result[len(result)-1].Equals(intersect) {
					result = append(result, intersect)
				}
			}
		}
	}

	return result
}

func findSeedIndex(seeds []geom.Point, target geom.Point) int {
	for i, seed := range seeds {
		if seed.Equals(target) {
			return i
		}
	}
	return -1
}

func FindNearestNeighbor(diagram *Diagram, query geom.Point) (geom.Point, float64, int) {
	if diagram == nil || len(diagram.Seeds) == 0 {
		return geom.Point{}, 0, -1
	}

	minDist := query.Distance(diagram.Seeds[0])
	minIndex := 0

	for i := 1; i < len(diagram.Seeds); i++ {
		dist := query.Distance(diagram.Seeds[i])
		if dist < minDist {
			minDist = dist
			minIndex = i
		}
	}

	return diagram.Seeds[minIndex], minDist, minIndex
}

func ClipEdgesToRegion(diagram *Diagram, region geom.Rectangle) []HalfEdge {
	if diagram == nil {
		return nil
	}

	clipped := ClipDiagram(diagram, region)
	return clipped.Edges
}

func CalculateCellAreas(diagram *Diagram, bounds geom.Rectangle) []float64 {
	if diagram == nil || len(diagram.Seeds) == 0 {
		return nil
	}

	cells := BuildCells(diagram, bounds)
	areas := make([]float64, len(cells))
	for i, cell := range cells {
		areas[i] = cell.Area
	}
	return areas
}

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(seeds []geom.Point) *Diagram {
	return Generate(seeds)
}

func ClipToBounds(diagram *Diagram, bounds geom.Rectangle) *Diagram {
	return ClipDiagram(diagram, bounds)
}

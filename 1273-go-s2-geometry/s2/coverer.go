package s2

import (
	"math"
)

func (id CellID) AllNeighbors(level int) []CellID {
	if level < 0 || level > MaxLevel {
		return nil
	}
	face, i, j := id.FaceIJ()
	currentLevel := id.Level()
	if level > currentLevel {
		level = currentLevel
	}
	neighbors := []CellID{}
	for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		ni := i + d[0]
		nj := j + d[1]
		nCell := getNeighborIJAtLevel(face, ni, nj, level)
		if nCell != 0 {
			neighbors = append(neighbors, nCell)
		}
	}
	return neighbors
}

func getNeighborIJAtLevel(face, i, j, level int) CellID {
	maxIJAtLevel := 1 << level
	if i < 0 || i >= maxIJAtLevel || j < 0 || j >= maxIJAtLevel {
		return 0
	}
	scale := 1 << (MaxLevel - level)
	return CellIDFromFaceIJ(face, i*scale, j*scale).Parent(level)
}

type RegionCoverer struct {
	MinLevel int
	MaxLevel int
	MaxCells int
}

func (rc RegionCoverer) Covering(rect Rect) []CellID {
	result := []CellID{}
	candidates := []CellID{}
	for face := 0; face < NumFaces; face++ {
		id := CellIDFromFaceIJ(face, 0, 0).Parent(0)
		if rc.intersectsCell(id, rect) {
			candidates = append(candidates, id)
		}
	}
	processed := 0
	for len(candidates) > 0 {
		if len(result) >= rc.MaxCells {
			break
		}
		cell := candidates[len(candidates)-1]
		candidates = candidates[:len(candidates)-1]
		processed++
		if processed > 10000 {
			break
		}
		children := rc.getChildren(cell)
		intersecting := 0
		for _, child := range children {
			if rc.intersectsCell(child, rect) {
				intersecting++
			}
		}
		if intersecting == 0 {
			continue
		}
		cellLevel := cell.Level()
		if cellLevel >= rc.MinLevel && (intersecting == 1 || cellLevel >= rc.MaxLevel) {
			result = append(result, cell)
			continue
		}
		for _, child := range children {
			if rc.intersectsCell(child, rect) {
				candidates = append(candidates, child)
			}
		}
	}
	return result
}

func (rc RegionCoverer) getChildren(id CellID) []CellID {
	level := id.Level()
	if level >= MaxLevel {
		return nil
	}
	face, i, j := id.FaceIJ()
	nextLevel := level + 1
	children := []CellID{
		CellIDFromFaceIJ(face, i*2, j*2).Parent(nextLevel),
		CellIDFromFaceIJ(face, i*2+1, j*2).Parent(nextLevel),
		CellIDFromFaceIJ(face, i*2, j*2+1).Parent(nextLevel),
		CellIDFromFaceIJ(face, i*2+1, j*2+1).Parent(nextLevel),
	}
	return children
}

func (rc RegionCoverer) intersectsCell(id CellID, rect Rect) bool {
	cellRect := id.RectBound()
	return rect.Intersects(cellRect)
}

func RectFromDegrees(latLo, latHi, lngLo, lngHi float64) Rect {
	return Rect{
		LatLo: DegToRad(latLo),
		LatHi: DegToRad(latHi),
		LngLo: DegToRad(lngLo),
		LngHi: DegToRad(lngHi),
	}
}

func NormalizeLng(lng float64) float64 {
	for lng > math.Pi {
		lng -= 2 * math.Pi
	}
	for lng < -math.Pi {
		lng += 2 * math.Pi
	}
	return lng
}

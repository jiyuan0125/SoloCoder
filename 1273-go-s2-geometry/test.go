package main

import (
	"fmt"
	"s2geometry/s2"
)

func main() {
	ll := s2.LatLng{
		Lat: s2.DegToRad(39.9042),
		Lng: s2.DegToRad(116.4074),
	}
	
	p := s2.PointFromLatLng(ll)
	face := s2.Face(p)
	u, v := s2.FaceUV(face, p)
	si := s2.STToIJ(s2.UvToST(u))
	sj := s2.STToIJ(s2.UvToST(v))
	
	fmt.Printf("Original IJ: i=%d (0x%08x), j=%d (0x%08x)\n", si, si, sj, sj)
	
	cell := s2.CellIDFromLatLng(ll)
	fmt.Printf("\nCellID at Level 30: %d\n", uint64(cell))
	fmt.Printf("  binary: %064b\n", uint64(cell))
	fmt.Printf("  Face: %d\n", cell.Face())
	fmt.Printf("  Level: %d\n", cell.Level())
	
	f, i, j := cell.FaceIJ()
	fmt.Printf("\nFaceIJ decode: face=%d, i=%d, j=%d\n", f, i, j)
	fmt.Printf("  Match: i=%v, j=%v\n", i == si, j == sj)
	
	fmt.Println("\n=== Parent tests ===")
	for level := 30; level >= 25; level-- {
		parent := cell.Parent(level)
		f2, i2, j2 := parent.FaceIJ()
		mask := ^((1 << (30 - level)) - 1)
		expectedI := si & mask
		expectedJ := sj & mask
		fmt.Printf("  Level %d: CellID=%d, face=%d, i=%d, j=%d\n", 
			level, uint64(parent), f2, i2, j2)
		fmt.Printf("    Expected i=%d, j=%d\n", expectedI, expectedJ)
		fmt.Printf("    Match: i=%v, j=%v, level=%v\n", 
			i2 == expectedI, j2 == expectedJ, parent.Level() == level)
	}
	
	fmt.Println("\n=== Contains tests ===")
	cell0 := cell.Parent(0)
	cell5 := cell.Parent(5)
	cell10 := cell.Parent(10)
	cell15 := cell.Parent(15)
	fmt.Printf("  cell0 contains cell5: %v (expected: true)\n", cell0.Contains(cell5))
	fmt.Printf("  cell5 contains cell10: %v (expected: true)\n", cell5.Contains(cell10))
	fmt.Printf("  cell10 contains cell15: %v (expected: true)\n", cell10.Contains(cell15))
	fmt.Printf("  cell15 contains cell5: %v (expected: false)\n", cell15.Contains(cell5))
	
	fmt.Println("\n=== RectBound debug ===")
	f10, i10, j10 := cell10.FaceIJ()
	level10 := cell10.Level()
	scale10 := 1 << (30 - level10)
	iScaled := i10 << (30 - level10)
	jScaled := j10 << (30 - level10)
	fmt.Printf("  Level=%d, face=%d, i=%d, j=%d, scale=%d\n", level10, f10, i10, j10, scale10)
	fmt.Printf("  Scaled: i=%d, j=%d, i+scale=%d, j+scale=%d\n", iScaled, jScaled, iScaled+scale10, jScaled+scale10)
	
	loPt := s2.FaceIJToSTFaceUV(f10, iScaled, jScaled)
	hiPt := s2.FaceIJToSTFaceUV(f10, iScaled+scale10, jScaled+scale10)
	loLL := s2.LatLngFromPoint(loPt)
	hiLL := s2.LatLngFromPoint(hiPt)
	fmt.Printf("  loPoint: (%.6f, %.6f, %.6f) -> LatLng: (%.6f, %.6f)\n", loPt.X, loPt.Y, loPt.Z, s2.RadToDeg(loLL.Lat), s2.RadToDeg(loLL.Lng))
	fmt.Printf("  hiPoint: (%.6f, %.6f, %.6f) -> LatLng: (%.6f, %.6f)\n", hiPt.X, hiPt.Y, hiPt.Z, s2.RadToDeg(hiLL.Lat), s2.RadToDeg(hiLL.Lng))
	
	cell10Rect := cell10.RectBound()
	fmt.Printf("  Lat: [%.6f, %.6f]\n", s2.RadToDeg(cell10Rect.LatLo), s2.RadToDeg(cell10Rect.LatHi))
	fmt.Printf("  Lng: [%.6f, %.6f]\n", s2.RadToDeg(cell10Rect.LngLo), s2.RadToDeg(cell10Rect.LngHi))
	center := cell10.LatLng()
	fmt.Printf("  Center: (%.6f, %.6f)\n", s2.RadToDeg(center.Lat), s2.RadToDeg(center.Lng))
	
	fmt.Println("\n=== Neighbors test (level 10) ===")
	neighbors := cell10.AllNeighbors(10)
	fmt.Printf("  Found %d neighbors\n", len(neighbors))
	for idx, n := range neighbors {
		fn, _, _ := n.FaceIJ()
		fmt.Printf("    Neighbor %d: %d (face=%d, level=%d)\n", 
			idx, uint64(n), fn, n.Level())
	}
	
	fmt.Println("\n=== Covering test ===")
	rect := s2.RectFromDegrees(39, 40, 116, 117)
	fmt.Printf("  Query rect: Lat[%.6f, %.6f], Lng[%.6f, %.6f]\n",
		s2.RadToDeg(rect.LatLo), s2.RadToDeg(rect.LatHi),
		s2.RadToDeg(rect.LngLo), s2.RadToDeg(rect.LngHi))
	
	for face := 0; face < 6; face++ {
		faceCell := s2.CellIDFromFaceIJ(face, 0, 0).Parent(0)
		faceRect := faceCell.RectBound()
		fmt.Printf("  Face %d: Lat[%.6f, %.6f], Lng[%.6f, %.6f]\n",
			face, s2.RadToDeg(faceRect.LatLo), s2.RadToDeg(faceRect.LatHi),
			s2.RadToDeg(faceRect.LngLo), s2.RadToDeg(faceRect.LngHi))
		fmt.Printf("    Intersects: %v\n", rect.Intersects(faceRect))
	}
	
	coverer := s2.RegionCoverer{
		MinLevel: 5,
		MaxLevel: 10,
		MaxCells: 100,
	}
	cells := coverer.Covering(rect)
	fmt.Printf("  Found %d covering cells\n", len(cells))
	for idx, c := range cells {
		fmt.Printf("    Cell %d: %d (level=%d)\n", idx, uint64(c), c.Level())
	}
}

package main

import (
	"fmt"
	"s2geometry/s2"
)

func main() {
	llBeijing := s2.LatLng{Lat: s2.DegToRad(39.9042), Lng: s2.DegToRad(116.4074)}
	ptBeijing := s2.PointFromLatLng(llBeijing)
	cellBeijing := s2.CellIDFromLatLng(llBeijing)
	faceBeijing, iBeijing, jBeijing := cellBeijing.FaceIJ()
	
	fmt.Println("=== Beijing ===")
	fmt.Printf("  LatLng: (%.4f, %.4f)\n", s2.RadToDeg(llBeijing.Lat), s2.RadToDeg(llBeijing.Lng))
	fmt.Printf("  Point: (%.6f, %.6f, %.6f)\n", ptBeijing.X, ptBeijing.Y, ptBeijing.Z)
	fmt.Printf("  CellID: %d, face=%d, level=%d\n", uint64(cellBeijing), faceBeijing, cellBeijing.Level())
	fmt.Printf("  IJ: i=%d, j=%d\n", iBeijing, jBeijing)
	
	scale30 := 1 << (30 - cellBeijing.Level())
	i30 := iBeijing * scale30
	j30 := jBeijing * scale30
	fmt.Printf("  IJ (scaled to level 30): i=%d, j=%d\n", i30, j30)
	
	reconstructedPt := s2.FaceIJToSTFaceUV(faceBeijing, i30, j30)
	reconstructedLL := s2.LatLngFromPoint(reconstructedPt)
	fmt.Printf("  Reconstructed from FaceIJ: (%.6f, %.6f, %.6f)\n", reconstructedPt.X, reconstructedPt.Y, reconstructedPt.Z)
	fmt.Printf("  Reconstructed LatLng: (%.6f, %.6f)\n", s2.RadToDeg(reconstructedLL.Lat), s2.RadToDeg(reconstructedLL.Lng))
	
	fmt.Println("\n=== Face 1 boundary points ===")
	boundaryIJ := []struct{ name string; i, j int }{
		{"low-lat, low-lng", 0, 0},
		{"low-lat, high-lng", 1 << 30, 0},
		{"high-lat, low-lng", 0, 1 << 30},
		{"high-lat, high-lng", 1 << 30, 1 << 30},
		{"Beijing i,j", i30, j30},
	}
	
	for _, b := range boundaryIJ {
		pt := s2.FaceIJToSTFaceUV(1, b.i, b.j)
		ll := s2.LatLngFromPoint(pt)
		fmt.Printf("  %s: i=%d, j=%d -> LatLng=(%.6f, %.6f)\n",
			b.name, b.i, b.j, s2.RadToDeg(ll.Lat), s2.RadToDeg(ll.Lng))
	}
	
	fmt.Println("\n=== Face 2 (Z positive, North Pole) ===")
	for i := 0; i <= 1; i++ {
		for j := 0; j <= 1; j++ {
			pt := s2.FaceIJToSTFaceUV(2, i*(1<<30), j*(1<<30))
			ll := s2.LatLngFromPoint(pt)
			fmt.Printf("  Corner (%d,%d): (%.6f, %.6f)\n", i, j, s2.RadToDeg(ll.Lat), s2.RadToDeg(ll.Lng))
		}
	}
	ptCenter2 := s2.FaceIJToSTFaceUV(2, 1<<29, 1<<29)
	llCenter2 := s2.LatLngFromPoint(ptCenter2)
	fmt.Printf("  Center: (%.6f, %.6f)\n", s2.RadToDeg(llCenter2.Lat), s2.RadToDeg(llCenter2.Lng))
}

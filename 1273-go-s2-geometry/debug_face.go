package main

import (
	"fmt"
	"s2geometry/s2"
)

func main() {
	face := 1
	maxIJ := 1 << 30
	
	fmt.Println("Face 1 corners:")
	corners := []struct{ name string; i, j int }{
		{"(0,0)", 0, 0},
		{"(max,0)", maxIJ, 0},
		{"(0,max)", 0, maxIJ},
		{"(max,max)", maxIJ, maxIJ},
		{"center", maxIJ / 2, maxIJ / 2},
	}
	
	for _, c := range corners {
		pt := s2.FaceIJToSTFaceUV(face, c.i, c.j)
		ll := s2.LatLngFromPoint(pt)
		fmt.Printf("  %s: point=(%.6f, %.6f, %.6f) -> LatLng=(%.6f, %.6f)\n",
			c.name, pt.X, pt.Y, pt.Z, s2.RadToDeg(ll.Lat), s2.RadToDeg(ll.Lng))
	}
	
	fmt.Println("\n=== Test Beijing point from before ===")
	ll := s2.LatLng{Lat: s2.DegToRad(39.9042), Lng: s2.DegToRad(116.4074)}
	pt := s2.PointFromLatLng(ll)
	face2 := s2.Face(pt)
	fmt.Printf("  Beijing: face=%d, point=(%.6f, %.6f, %.6f)\n", face2, pt.X, pt.Y, pt.Z)
	
	fmt.Println("\n=== Face 1 cells level 0 ===")
	face1Cell := s2.CellIDFromFaceIJ(1, 0, 0).Parent(0)
	face1Rect := face1Cell.RectBound()
	fmt.Printf("  Face 1 Cell Level 0: cell_id=%d, Lat[%.6f, %.6f], Lng[%.6f, %.6f]\n",
		uint64(face1Cell), s2.RadToDeg(face1Rect.LatLo), s2.RadToDeg(face1Rect.LatHi),
		s2.RadToDeg(face1Rect.LngLo), s2.RadToDeg(face1Rect.LngHi))
	
	fmt.Println("\n=== Beijing cell ===")
	cell := s2.CellIDFromLatLng(ll)
	cell10 := cell.Parent(10)
	cell10Rect := cell10.RectBound()
	fmt.Printf("  Beijing Cell Level 30: cell_id=%d, face=%d, level=%d\n",
		uint64(cell), cell.Face(), cell.Level())
	fmt.Printf("  Beijing Cell Level 10: cell_id=%d, face=%d, level=%d\n",
		uint64(cell10), cell10.Face(), cell10.Level())
	fmt.Printf("  Beijing Cell Level 10 rect: Lat[%.6f, %.6f], Lng[%.6f, %.6f]\n",
		s2.RadToDeg(cell10Rect.LatLo), s2.RadToDeg(cell10Rect.LatHi),
		s2.RadToDeg(cell10Rect.LngLo), s2.RadToDeg(cell10Rect.LngHi))
}

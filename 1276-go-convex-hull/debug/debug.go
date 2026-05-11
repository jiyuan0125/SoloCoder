package main

import (
	"fmt"

	"convex-hull/convexhull"
)

func main() {
	fmt.Println("=== Test 1: Convex Hull Intersection Area ===")
	square1 := []convexhull.Point{
		{0, 0}, {10, 0}, {10, 10}, {0, 10},
	}
	square2 := []convexhull.Point{
		{5, 5}, {15, 5}, {15, 15}, {5, 15},
	}
	h1 := convexhull.Compute(square1)
	h2 := convexhull.Compute(square2)
	fmt.Printf("Hull 1: type=%v, points=%v\n", h1.Type, h1.Points)
	fmt.Printf("Hull 2: type=%v, points=%v\n", h2.Type, h2.Points)

	intersection := convexhull.Intersection(h1, h2)
	fmt.Printf("Intersection: type=%v\n", intersection.Type)
	fmt.Printf("Intersection points: %v\n", intersection.Points)
	area := convexhull.Area(intersection)
	fmt.Printf("Intersection area: %.4f (expected: 25)\n", area)
	fmt.Printf("Perimeter: %.4f\n", convexhull.Perimeter(intersection))

	fmt.Println("\n=== Test 2: Dynamic Convex Hull ===")
	dh := convexhull.NewDynamicConvexHull()
	dh.Add(convexhull.Point{0, 0})
	dh.Add(convexhull.Point{10, 0})
	dh.Add(convexhull.Point{10, 10})
	hull1 := dh.Hull()
	fmt.Printf("After adding (0,0), (10,0), (10,10):\n")
	fmt.Printf("  Hull type: %v\n", hull1.Type)
	fmt.Printf("  Hull points: %v\n", hull1.Points)
	fmt.Printf("  Hull area: %.4f\n", convexhull.Area(hull1))

	dh.Add(convexhull.Point{0, 10})
	hull2 := dh.Hull()
	fmt.Printf("\nAfter adding (0,10):\n")
	fmt.Printf("  Hull type: %v\n", hull2.Type)
	fmt.Printf("  Hull points: %v\n", hull2.Points)
	fmt.Printf("  Hull area: %.4f (expected: 100)\n", convexhull.Area(hull2))

	fmt.Println("\n=== Test 3: Point-in-hull check for (0,10) in triangle ===")
	triangle := []convexhull.Point{{0, 0}, {10, 0}, {10, 10}}
	hTri := convexhull.Compute(triangle)
	testPt := convexhull.Point{0, 10}
	fmt.Printf("Triangle hull points: %v\n", hTri.Points)
	fmt.Printf("Point (0,10) in triangle hull? %v (expected: false)\n",
		convexhull.PointInConvexHull(testPt, hTri))
}

package kmeans

import (
	"errors"
	"math"
	"math/rand"
)

type Point []float64
type Cluster struct {
	Center Point
	Points []Point
}

const (
	DefaultMaxIterations = 100
	DefaultEpsilon       = 1e-6
)

func EuclideanDistance(p1, p2 Point) float64 {
	if len(p1) != len(p2) {
		return math.Inf(1)
	}
	sum := 0.0
	for i := range p1 {
		diff := p1[i] - p2[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

func SquaredDistance(p1, p2 Point) float64 {
	if len(p1) != len(p2) {
		return math.Inf(1)
	}
	sum := 0.0
	for i := range p1 {
		diff := p1[i] - p2[i]
		sum += diff * diff
	}
	return sum
}

func Mean(points []Point) Point {
	if len(points) == 0 {
		return nil
	}
	dim := len(points[0])
	mean := make(Point, dim)
	for _, p := range points {
		for i := range p {
			mean[i] += p[i]
		}
	}
	n := float64(len(points))
	for i := range mean {
		mean[i] /= n
	}
	return mean
}

func KMeansPlusPlusInit(points []Point, k int) ([]Point, error) {
	if k <= 0 {
		return nil, errors.New("k must be positive")
	}
	if len(points) == 0 {
		return nil, errors.New("no data points")
	}
	if k > len(points) {
		return nil, errors.New("k cannot be larger than number of points")
	}

	dim := len(points[0])
	for _, p := range points {
		if len(p) != dim {
			return nil, errors.New("all points must have same dimension")
		}
	}

	centers := make([]Point, 0, k)
	firstIdx := rand.Intn(len(points))
	centers = append(centers, append(Point(nil), points[firstIdx]...))

	allSame := true
	for i := 1; i < len(points); i++ {
		if SquaredDistance(points[0], points[i]) > 0 {
			allSame = false
			break
		}
	}

	if allSame {
		for i := 1; i < k; i++ {
			centers = append(centers, append(Point(nil), points[0]...))
		}
		return centers, nil
	}

	for len(centers) < k {
		distances := make([]float64, len(points))
		for i, p := range points {
			minDist := math.Inf(1)
			for _, c := range centers {
				d := SquaredDistance(p, c)
				if d < minDist {
					minDist = d
				}
			}
			distances[i] = minDist
		}

		var total float64
		for _, d := range distances {
			total += d
		}

		if total == 0 {
			for len(centers) < k {
				centers = append(centers, append(Point(nil), centers[0]...))
			}
			break
		}

		r := rand.Float64() * total
		cumulative := 0.0
		selectedIdx := 0
		for i, d := range distances {
			cumulative += d
			if cumulative >= r {
				selectedIdx = i
				break
			}
		}
		centers = append(centers, append(Point(nil), points[selectedIdx]...))
	}

	return centers, nil
}

func AssignToClusters(points []Point, centers []Point) ([]int, error) {
	if len(centers) == 0 {
		return nil, errors.New("no centers")
	}
	labels := make([]int, len(points))
	for i, p := range points {
		minDist := math.Inf(1)
		label := 0
		for j, c := range centers {
			d := SquaredDistance(p, c)
			if d < minDist {
				minDist = d
				label = j
			}
		}
		labels[i] = label
	}
	return labels, nil
}

func UpdateCenters(points []Point, labels []int, k int, oldCenters []Point) ([]Point, error) {
	if len(points) != len(labels) {
		return nil, errors.New("points and labels length mismatch")
	}

	clusters := make([][]Point, k)
	for i := range clusters {
		clusters[i] = make([]Point, 0)
	}

	for i, label := range labels {
		if label < 0 || label >= k {
			return nil, errors.New("invalid label")
		}
		clusters[label] = append(clusters[label], points[i])
	}

	newCenters := make([]Point, k)
	for i := range clusters {
		if len(clusters[i]) == 0 {
			newCenters[i] = append(Point(nil), oldCenters[i]...)
		} else {
			newCenters[i] = Mean(clusters[i])
		}
	}

	return newCenters, nil
}

func CentersConverged(oldCenters, newCenters []Point, epsilon float64) bool {
	if len(oldCenters) != len(newCenters) {
		return false
	}
	for i := range oldCenters {
		if EuclideanDistance(oldCenters[i], newCenters[i]) > epsilon {
			return false
		}
	}
	return true
}

type KMeansResult struct {
	Labels     []int
	Centers    []Point
	Iterations int
	Converged  bool
}

func KMeans(points []Point, k int, maxIterations int, epsilon float64) (*KMeansResult, error) {
	if maxIterations <= 0 {
		maxIterations = DefaultMaxIterations
	}
	if epsilon <= 0 {
		epsilon = DefaultEpsilon
	}

	if k <= 0 {
		return nil, errors.New("k must be positive")
	}
	if len(points) == 0 {
		return nil, errors.New("no data points")
	}
	if k > len(points) {
		return nil, errors.New("k cannot be larger than number of points")
	}

	centers, err := KMeansPlusPlusInit(points, k)
	if err != nil {
		return nil, err
	}

	var labels []int
	iterations := 0
	converged := false

	for iterations < maxIterations {
		labels, err = AssignToClusters(points, centers)
		if err != nil {
			return nil, err
		}

		newCenters, err := UpdateCenters(points, labels, k, centers)
		if err != nil {
			return nil, err
		}

		iterations++

		if CentersConverged(centers, newCenters, epsilon) {
			converged = true
			centers = newCenters
			break
		}

		centers = newCenters
	}

	if !converged {
		labels, err = AssignToClusters(points, centers)
		if err != nil {
			return nil, err
		}
	}

	return &KMeansResult{
		Labels:     labels,
		Centers:    centers,
		Iterations: iterations,
		Converged:  converged,
	}, nil
}

func WCSS(points []Point, labels []int, centers []Point) (float64, error) {
	if len(points) != len(labels) {
		return 0, errors.New("points and labels length mismatch")
	}
	if len(centers) == 0 {
		return 0, errors.New("no centers")
	}

	sum := 0.0
	for i, p := range points {
		label := labels[i]
		if label < 0 || label >= len(centers) {
			return 0, errors.New("invalid label")
		}
		sum += SquaredDistance(p, centers[label])
	}
	return sum, nil
}

func ElbowMethod(points []Point, maxK int, maxIterations int, epsilon float64) (map[int]float64, error) {
	if maxK <= 0 {
		return nil, errors.New("maxK must be positive")
	}
	if len(points) == 0 {
		return nil, errors.New("no data points")
	}
	if maxK > len(points) {
		maxK = len(points)
	}

	wcssMap := make(map[int]float64)
	for k := 1; k <= maxK; k++ {
		result, err := KMeans(points, k, maxIterations, epsilon)
		if err != nil {
			return nil, err
		}
		wcss, err := WCSS(points, result.Labels, result.Centers)
		if err != nil {
			return nil, err
		}
		wcssMap[k] = wcss
	}
	return wcssMap, nil
}

func FindOptimalK(wcssMap map[int]float64) int {
	if len(wcssMap) == 0 {
		return 1
	}

	maxK := 0
	for k := range wcssMap {
		if k > maxK {
			maxK = k
		}
	}

	if maxK <= 1 {
		return 1
	}

	prevDiff := math.Inf(-1)
	optimalK := 1

	for k := 2; k <= maxK; k++ {
		diff := wcssMap[k-1] - wcssMap[k]
		if k > 2 {
			ratio := prevDiff / diff
			if ratio > 1.5 {
				optimalK = k - 1
				break
			}
		}
		prevDiff = diff
		optimalK = k
	}

	return optimalK
}

func ClusterStats(points []Point, labels []int, centers []Point) ([]*ClusterInfo, error) {
	if len(points) != len(labels) {
		return nil, errors.New("points and labels length mismatch")
	}
	if len(centers) == 0 {
		return nil, errors.New("no centers")
	}

	k := len(centers)
	clusterPoints := make([][]Point, k)
	for i := range clusterPoints {
		clusterPoints[i] = make([]Point, 0)
	}

	for i, label := range labels {
		if label < 0 || label >= k {
			return nil, errors.New("invalid label")
		}
		clusterPoints[label] = append(clusterPoints[label], points[i])
	}

	stats := make([]*ClusterInfo, k)
	for i := range stats {
		stats[i] = &ClusterInfo{
			Index:      i,
			Center:     centers[i],
			PointCount: len(clusterPoints[i]),
		}

		if len(clusterPoints[i]) > 0 {
			var totalDist float64
			for _, p := range clusterPoints[i] {
				totalDist += EuclideanDistance(p, centers[i])
			}
			stats[i].AvgDistance = totalDist / float64(len(clusterPoints[i]))
		}
	}

	return stats, nil
}

type ClusterInfo struct {
	Index       int
	Center      Point
	PointCount  int
	AvgDistance float64
}

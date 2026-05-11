package rtree

import (
	"math"
	"sort"

	"github.com/spatial-index/pkg/geometry"
)

const (
	DefaultMinEntries = 2
	DefaultMaxEntries = 8
)

type Entry struct {
	MBR     geometry.MBR
	Feature *geometry.Feature
}

type Node struct {
	MBR      geometry.MBR
	Entries  []Entry
	Children []*Node
	IsLeaf   bool
}

type RTree struct {
	Root       *Node
	MinEntries int
	MaxEntries int
	Size       int
}

func NewRTree() *RTree {
	return &RTree{
		Root:       newLeafNode(),
		MinEntries: DefaultMinEntries,
		MaxEntries: DefaultMaxEntries,
		Size:       0,
	}
}

func NewRTreeWithParams(minEntries, maxEntries int) *RTree {
	if minEntries < 1 {
		minEntries = 1
	}
	if maxEntries <= minEntries*2 {
		maxEntries = minEntries * 2
	}
	return &RTree{
		Root:       newLeafNode(),
		MinEntries: minEntries,
		MaxEntries: maxEntries,
		Size:       0,
	}
}

func newLeafNode() *Node {
	return &Node{
		IsLeaf:  true,
		Entries: make([]Entry, 0),
	}
}

func newInternalNode() *Node {
	return &Node{
		IsLeaf:   false,
		Children: make([]*Node, 0),
	}
}

func (n *Node) updateMBR() {
	if n.IsLeaf {
		if len(n.Entries) == 0 {
			n.MBR = geometry.MBR{}
			return
		}
		mbr := n.Entries[0].MBR
		for i := 1; i < len(n.Entries); i++ {
			mbr = mbr.Union(n.Entries[i].MBR)
		}
		n.MBR = mbr
	} else {
		if len(n.Children) == 0 {
			n.MBR = geometry.MBR{}
			return
		}
		mbr := n.Children[0].MBR
		for i := 1; i < len(n.Children); i++ {
			mbr = mbr.Union(n.Children[i].MBR)
		}
		n.MBR = mbr
	}
}

func (t *RTree) Insert(feature *geometry.Feature) {
	entry := Entry{
		MBR:     geometry.FeatureMBR(*feature),
		Feature: feature,
	}

	newRoot, splitNode := t.insert(t.Root, entry)

	if splitNode != nil {
		newRootNode := newInternalNode()
		newRootNode.Children = append(newRootNode.Children, t.Root, splitNode)
		newRootNode.updateMBR()
		t.Root = newRootNode
	} else if newRoot != nil {
		t.Root = newRoot
	}

	t.Size++
}

func (t *RTree) insert(node *Node, entry Entry) (*Node, *Node) {
	if node.IsLeaf {
		node.Entries = append(node.Entries, entry)
		node.updateMBR()

		if len(node.Entries) > t.MaxEntries {
			left, right := t.splitLeafNode(node)
			return left, right
		}
		return node, nil
	}

	bestChild := t.chooseSubtree(node, entry.MBR)
	newChild, splitChild := t.insert(bestChild, entry)

	if splitChild != nil {
		for i, c := range node.Children {
			if c == bestChild {
				node.Children[i] = newChild
				node.Children = append(node.Children, splitChild)
				break
			}
		}
		node.updateMBR()

		if len(node.Children) > t.MaxEntries {
			left, right := t.splitInternalNode(node)
			return left, right
		}
	} else {
		for i, c := range node.Children {
			if c == bestChild {
				node.Children[i] = newChild
				break
			}
		}
		node.updateMBR()
	}

	return node, nil
}

func (t *RTree) chooseSubtree(node *Node, mbr geometry.MBR) *Node {
	bestIndex := 0
	bestEnlargement := math.Inf(1)
	bestArea := math.Inf(1)

	for i, child := range node.Children {
		childArea := child.MBR.Area()
		enlargedMBR := child.MBR.Union(mbr)
		enlargement := enlargedMBR.Area() - childArea

		if enlargement < bestEnlargement {
			bestEnlargement = enlargement
			bestArea = childArea
			bestIndex = i
		} else if enlargement == bestEnlargement {
			if childArea < bestArea {
				bestArea = childArea
				bestIndex = i
			}
		}
	}

	return node.Children[bestIndex]
}

func (t *RTree) splitLeafNode(node *Node) (*Node, *Node) {
	entries := node.Entries
	idx1, idx2 := t.pickSeedsEntries(entries)

	group1 := newLeafNode()
	group2 := newLeafNode()

	group1.Entries = append(group1.Entries, entries[idx1])
	group2.Entries = append(group2.Entries, entries[idx2])

	remaining := make([]Entry, 0, len(entries)-2)
	for i, e := range entries {
		if i != idx1 && i != idx2 {
			remaining = append(remaining, e)
		}
	}

	group1.updateMBR()
	group2.updateMBR()

	for len(remaining) > 0 {
		if len(remaining)+len(group1.Entries) <= t.MinEntries {
			group1.Entries = append(group1.Entries, remaining...)
			group1.updateMBR()
			break
		}
		if len(remaining)+len(group2.Entries) <= t.MinEntries {
			group2.Entries = append(group2.Entries, remaining...)
			group2.updateMBR()
			break
		}

		entryIdx := t.pickNextEntryEntries(remaining, group1.MBR, group2.MBR)
		entry := remaining[entryIdx]
		remaining = append(remaining[:entryIdx], remaining[entryIdx+1:]...)

		area1 := group1.MBR.Area()
		area2 := group2.MBR.Area()
		enlarged1 := group1.MBR.Union(entry.MBR).Area()
		enlarged2 := group2.MBR.Union(entry.MBR).Area()
		enlargement1 := enlarged1 - area1
		enlargement2 := enlarged2 - area2

		if enlargement1 < enlargement2 {
			group1.Entries = append(group1.Entries, entry)
			group1.updateMBR()
		} else if enlargement2 < enlargement1 {
			group2.Entries = append(group2.Entries, entry)
			group2.updateMBR()
		} else {
			if area1 < area2 {
				group1.Entries = append(group1.Entries, entry)
				group1.updateMBR()
			} else if area2 < area1 {
				group2.Entries = append(group2.Entries, entry)
				group2.updateMBR()
			} else {
				if len(group1.Entries) <= len(group2.Entries) {
					group1.Entries = append(group1.Entries, entry)
					group1.updateMBR()
				} else {
					group2.Entries = append(group2.Entries, entry)
					group2.updateMBR()
				}
			}
		}
	}

	return group1, group2
}

func (t *RTree) splitInternalNode(node *Node) (*Node, *Node) {
	children := node.Children
	idx1, idx2 := t.pickSeedsNodes(children)

	group1 := newInternalNode()
	group2 := newInternalNode()

	group1.Children = append(group1.Children, children[idx1])
	group2.Children = append(group2.Children, children[idx2])

	remaining := make([]*Node, 0, len(children)-2)
	for i, c := range children {
		if i != idx1 && i != idx2 {
			remaining = append(remaining, c)
		}
	}

	group1.updateMBR()
	group2.updateMBR()

	for len(remaining) > 0 {
		if len(remaining)+len(group1.Children) <= t.MinEntries {
			group1.Children = append(group1.Children, remaining...)
			group1.updateMBR()
			break
		}
		if len(remaining)+len(group2.Children) <= t.MinEntries {
			group2.Children = append(group2.Children, remaining...)
			group2.updateMBR()
			break
		}

		entryIdx := t.pickNextEntryNodes(remaining, group1.MBR, group2.MBR)
		child := remaining[entryIdx]
		remaining = append(remaining[:entryIdx], remaining[entryIdx+1:]...)

		area1 := group1.MBR.Area()
		area2 := group2.MBR.Area()
		enlarged1 := group1.MBR.Union(child.MBR).Area()
		enlarged2 := group2.MBR.Union(child.MBR).Area()
		enlargement1 := enlarged1 - area1
		enlargement2 := enlarged2 - area2

		if enlargement1 < enlargement2 {
			group1.Children = append(group1.Children, child)
			group1.updateMBR()
		} else if enlargement2 < enlargement1 {
			group2.Children = append(group2.Children, child)
			group2.updateMBR()
		} else {
			if area1 < area2 {
				group1.Children = append(group1.Children, child)
				group1.updateMBR()
			} else if area2 < area1 {
				group2.Children = append(group2.Children, child)
				group2.updateMBR()
			} else {
				if len(group1.Children) <= len(group2.Children) {
					group1.Children = append(group1.Children, child)
					group1.updateMBR()
				} else {
					group2.Children = append(group2.Children, child)
					group2.updateMBR()
				}
			}
		}
	}

	return group1, group2
}

func (t *RTree) pickSeedsEntries(entries []Entry) (int, int) {
	bestWaste := math.Inf(-1)
	bestI := 0
	bestJ := 1

	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			combined := entries[i].MBR.Union(entries[j].MBR)
			waste := combined.Area() - entries[i].MBR.Area() - entries[j].MBR.Area()
			if waste > bestWaste {
				bestWaste = waste
				bestI = i
				bestJ = j
			}
		}
	}

	return bestI, bestJ
}

func (t *RTree) pickSeedsNodes(nodes []*Node) (int, int) {
	bestWaste := math.Inf(-1)
	bestI := 0
	bestJ := 1

	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			combined := nodes[i].MBR.Union(nodes[j].MBR)
			waste := combined.Area() - nodes[i].MBR.Area() - nodes[j].MBR.Area()
			if waste > bestWaste {
				bestWaste = waste
				bestI = i
				bestJ = j
			}
		}
	}

	return bestI, bestJ
}

func (t *RTree) pickNextEntryEntries(entries []Entry, mbr1, mbr2 geometry.MBR) int {
	bestDiff := math.Inf(-1)
	bestIdx := 0

	for i, e := range entries {
		area1 := mbr1.Area()
		area2 := mbr2.Area()
		enlarged1 := mbr1.Union(e.MBR).Area()
		enlarged2 := mbr2.Union(e.MBR).Area()
		diff := math.Abs((enlarged1 - area1) - (enlarged2 - area2))

		if diff > bestDiff {
			bestDiff = diff
			bestIdx = i
		}
	}

	return bestIdx
}

func (t *RTree) pickNextEntryNodes(nodes []*Node, mbr1, mbr2 geometry.MBR) int {
	bestDiff := math.Inf(-1)
	bestIdx := 0

	for i, n := range nodes {
		area1 := mbr1.Area()
		area2 := mbr2.Area()
		enlarged1 := mbr1.Union(n.MBR).Area()
		enlarged2 := mbr2.Union(n.MBR).Area()
		diff := math.Abs((enlarged1 - area1) - (enlarged2 - area2))

		if diff > bestDiff {
			bestDiff = diff
			bestIdx = i
		}
	}

	return bestIdx
}

func (t *RTree) Search(rect geometry.MBR) []*geometry.Feature {
	results := make([]*geometry.Feature, 0)
	t.search(t.Root, rect, &results)
	return results
}

func (t *RTree) search(node *Node, rect geometry.MBR, results *[]*geometry.Feature) {
	if node.IsLeaf {
		for _, entry := range node.Entries {
			if entry.MBR.Intersects(rect) {
				*results = append(*results, entry.Feature)
			}
		}
	} else {
		for _, child := range node.Children {
			if child.MBR.Intersects(rect) {
				t.search(child, rect, results)
			}
		}
	}
}

func (t *RTree) PointQuery(p geometry.Point) []*geometry.Feature {
	results := make([]*geometry.Feature, 0)
	candidates := t.Search(geometry.PointMBR(p))

	for _, f := range candidates {
		if pointInGeometry(p, f.Geometry) {
			results = append(results, f)
		}
	}

	return results
}

func pointInGeometry(p geometry.Point, g geometry.Geometry) bool {
	switch g.Type {
	case geometry.GeometryTypePoint:
		if g.Point != nil {
			return geometry.Distance(p, *g.Point) < 1e-9
		}
	case geometry.GeometryTypeLineString:
		if g.LineString != nil {
			return geometry.PointDistanceToLineString(p, *g.LineString) < 1e-9
		}
	case geometry.GeometryTypePolygon:
		if g.Polygon != nil {
			return geometry.PointInPolygon(p, *g.Polygon)
		}
	case geometry.GeometryTypeMultiPoint:
		if g.MultiPoint != nil {
			for _, pt := range g.MultiPoint.Points {
				if geometry.Distance(p, pt) < 1e-9 {
					return true
				}
			}
		}
	case geometry.GeometryTypeMultiLineString:
		if g.MultiLineString != nil {
			for _, ls := range g.MultiLineString.LineStrings {
				if geometry.PointDistanceToLineString(p, ls) < 1e-9 {
					return true
				}
			}
		}
	case geometry.GeometryTypeMultiPolygon:
		if g.MultiPolygon != nil {
			for _, poly := range g.MultiPolygon.Polygons {
				if geometry.PointInPolygon(p, poly) {
					return true
				}
			}
		}
	}
	return false
}

type knnItem struct {
	feature  *geometry.Feature
	distance float64
}

func (t *RTree) NearestNeighbors(p geometry.Point, k int) []*geometry.Feature {
	if k <= 0 || t.Size == 0 {
		return []*geometry.Feature{}
	}

	items := t.knnSearch(p)

	if len(items) == 0 {
		return []*geometry.Feature{}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].distance < items[j].distance
	})

	resultCount := k
	if resultCount > len(items) {
		resultCount = len(items)
	}

	results := make([]*geometry.Feature, resultCount)
	for i := 0; i < resultCount; i++ {
		results[i] = items[i].feature
	}

	return results
}

func (t *RTree) knnSearch(p geometry.Point) []knnItem {
	results := make([]knnItem, 0)
	t.knnSearchNode(t.Root, p, &results)
	return results
}

func (t *RTree) knnSearchNode(node *Node, p geometry.Point, results *[]knnItem) {
	if node.IsLeaf {
		for _, entry := range node.Entries {
			dist := geometry.PointDistanceToGeometry(p, entry.Feature.Geometry)
			*results = append(*results, knnItem{
				feature:  entry.Feature,
				distance: dist,
			})
		}
	} else {
		type nodeDist struct {
			node     *Node
			distance float64
		}

		children := make([]nodeDist, 0, len(node.Children))
		for _, child := range node.Children {
			dist := pointToMBRDistance(p, child.MBR)
			children = append(children, nodeDist{node: child, distance: dist})
		}

		sort.Slice(children, func(i, j int) bool {
			return children[i].distance < children[j].distance
		})

		for _, cd := range children {
			t.knnSearchNode(cd.node, p, results)
		}
	}
}

func pointToMBRDistance(p geometry.Point, mbr geometry.MBR) float64 {
	if mbr.Contains(p) {
		return 0
	}

	if mbr.CrossesDateLine() {
		dist1 := pointToMBRDistanceSimple(p, geometry.NewMBR(mbr.MinLon, mbr.MinLat, 180, mbr.MaxLat))
		dist2 := pointToMBRDistanceSimple(p, geometry.NewMBR(-180, mbr.MinLat, mbr.MaxLon, mbr.MaxLat))
		return math.Min(dist1, dist2)
	}

	return pointToMBRDistanceSimple(p, mbr)
}

func pointToMBRDistanceSimple(p geometry.Point, mbr geometry.MBR) float64 {
	var dx, dy float64

	if p.Lon < mbr.MinLon {
		dx = mbr.MinLon - p.Lon
	} else if p.Lon > mbr.MaxLon {
		dx = p.Lon - mbr.MaxLon
	} else {
		dx = 0
	}

	if p.Lat < mbr.MinLat {
		dy = mbr.MinLat - p.Lat
	} else if p.Lat > mbr.MaxLat {
		dy = p.Lat - mbr.MaxLat
	} else {
		dy = 0
	}

	if dx == 0 && dy == 0 {
		return 0
	}

	return math.Sqrt(dx*dx + dy*dy)
}

func (t *RTree) Len() int {
	return t.Size
}

func (t *RTree) InsertFeatureCollection(fc *geometry.FeatureCollection) {
	for i := range fc.Features {
		t.Insert(&fc.Features[i])
	}
}

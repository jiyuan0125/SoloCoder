package rtree

import (
	"container/heap"
	"errors"
	"math"
	"sort"
	"strconv"
)

const (
	DefaultMinEntries = 2
	DefaultMaxEntries = 4
)

type nodeType int

const (
	leafNode nodeType = iota
	internalNode
)

type node struct {
	typ      nodeType
	entries  []*entry
	parent   *node
	level    int
}

type entry struct {
	mbr      MBR
	child    *node
	object   SpatialObject
}

type Rtree struct {
	root       *node
	minEntries int
	maxEntries int
	count      int
	idIndex    map[string]*entry
}

func New() *Rtree {
	return NewWithOptions(DefaultMinEntries, DefaultMaxEntries)
}

func NewWithOptions(minEntries, maxEntries int) *Rtree {
	return &Rtree{
		root: &node{
			typ: leafNode,
			level: 0,
		},
		minEntries: minEntries,
		maxEntries: maxEntries,
		idIndex:    make(map[string]*entry),
	}
}

func (r *Rtree) Insert(obj SpatialObject) error {
	if obj == nil {
		return errors.New("object is nil")
	}
	if _, exists := r.idIndex[obj.ID()]; exists {
		return errors.New("object with same ID already exists")
	}

	minX, minY, maxX, maxY := obj.Bounds()
	e := &entry{
		mbr:    NewMBR(minX, minY, maxX, maxY),
		object: obj,
	}

	leaf := r.chooseLeaf(r.root, e.mbr)
	leaf.entries = append(leaf.entries, e)
	r.idIndex[obj.ID()] = e
	r.count++

	if len(leaf.entries) > r.maxEntries {
		split := r.splitNode(leaf)
		r.adjustTree(leaf, split)
	} else {
		r.adjustMBRs(leaf)
	}

	return nil
}

func (r *Rtree) Delete(id string) error {
	e, exists := r.idIndex[id]
	if !exists {
		return errors.New("object not found")
	}

	leaf := r.findLeaf(r.root, e)
	if leaf == nil {
		return errors.New("object not found")
	}

	for i, entry := range leaf.entries {
		if entry == e {
			leaf.entries = append(leaf.entries[:i], leaf.entries[i+1:]...)
			break
		}
	}

	delete(r.idIndex, id)
	r.count--

	if len(leaf.entries) < r.minEntries {
		r.condenseTree(leaf)
	} else {
		r.adjustMBRs(leaf)
	}

	return nil
}

func (r *Rtree) Search(queryMinX, queryMinY, queryMaxX, queryMaxY float64) []SpatialObject {
	query := NewMBR(queryMinX, queryMinY, queryMaxX, queryMaxY)
	var results []SpatialObject

	var search func(n *node)
	search = func(n *node) {
		for _, e := range n.entries {
			if !e.mbr.Intersects(query) {
				continue
			}

			if n.typ == leafNode {
				results = append(results, e.object)
			} else {
				search(e.child)
			}
		}
	}

	search(r.root)
	return results
}

func (r *Rtree) KNN(x, y float64, k int) []SpatialObject {
	if k <= 0 || r.count == 0 {
		return nil
	}

	if k >= r.count {
		all := r.Search(math.Inf(-1), math.Inf(-1), math.Inf(1), math.Inf(1))
		return r.sortByDistance(all, x, y)
	}

	pq := &nodeHeap{}
	heap.Init(pq)

	for _, e := range r.root.entries {
		dist := e.mbr.DistanceToPoint(x, y)
		heap.Push(pq, &heapItem{
			mbr:    e.mbr,
			node:   e.child,
			object: e.object,
			dist:  dist,
			isLeaf: r.root.typ == leafNode,
		})
	}

	results := make([]SpatialObject, 0, k)

	for pq.Len() > 0 && len(results) < k {
		item := heap.Pop(pq).(*heapItem)

		if item.isLeaf {
			results = append(results, item.object)
			continue
		}

		for _, e := range item.node.entries {
			dist := e.mbr.DistanceToPoint(x, y)
			heap.Push(pq, &heapItem{
				mbr:   e.mbr,
				node:  e.child,
				object: e.object,
				dist:  dist,
				isLeaf: item.node.typ == leafNode,
			})
		}
	}

	return results
}

func (r *Rtree) Count() int {
	return r.count
}

func (r *Rtree) Height() int {
	return r.root.level + 1
}

func (r *Rtree) NodeCount() int {
	count := 0
	var traverse func(n *node)
	traverse = func(n *node) {
		count++
		for _, e := range n.entries {
			if e.child != nil {
				traverse(e.child)
			}
		}
	}
	traverse(r.root)
	return count
}

func (r *Rtree) AllNodesMBR() []MBR {
	var result []MBR
	var traverse func(n *node)
	traverse = func(n *node) {
		result = append(result, r.nodeMBR(n))
		for _, e := range n.entries {
			if e.child != nil {
				traverse(e.child)
			}
		}
	}
	traverse(r.root)
	return result
}

func (r *Rtree) Visualize() string {
	if len(r.root.entries) == 0 {
		return "R-tree is empty\n"
	}

	result := ""
	var traverse func(n *node, level int, prefix string, isLast bool)
	traverse = func(n *node, level int, prefix string, isLast bool) {
		mbr := r.nodeMBR(n)
		nodeType := "Leaf"
		if n.typ == internalNode {
			nodeType = "Internal"
		}

		connector := "└── "
		if !isLast {
			connector = "├── "
		}

		result += prefix + connector
		result += nodeType + " ["
		result += formatMBR(mbr) + "]"
		if n.typ == leafNode {
			result += " (" + strconv.Itoa(len(n.entries)) + " objects)\n"
		} else {
			result += " (" + strconv.Itoa(len(n.entries)) + " children)\n"
		}

		newPrefix := prefix
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}

		for i, e := range n.entries {
			lastChild := i == len(n.entries)-1
			if n.typ == leafNode {
				result += newPrefix
				if lastChild {
					result += "└── "
				} else {
					result += "├── "
				}
				result += "Object: " + e.object.ID() + " [" + formatMBR(e.mbr) + "]\n"
			} else {
				traverse(e.child, level+1, newPrefix, lastChild)
			}
		}
	}

	traverse(r.root, 0, "", true)
	return result
}

func (r *Rtree) chooseLeaf(n *node, mbr MBR) *node {
	if n.typ == leafNode {
		return n
	}

	best := n.entries[0]
	bestExpansion := best.mbr.Expansion(mbr)

	for _, e := range n.entries[1:] {
		expansion := e.mbr.Expansion(mbr)
		if expansion < bestExpansion || (expansion == bestExpansion && e.mbr.Area() < best.mbr.Area()) {
			best = e
			bestExpansion = expansion
		}
	}

	return r.chooseLeaf(best.child, mbr)
}

func (r *Rtree) splitNode(n *node) *node {
	entries := n.entries
	groupA, groupB := r.quadraticSplit(entries)

	n.entries = groupA
	r.adjustMBRs(n)

	split := &node{
		typ:    n.typ,
		level:  n.level,
		parent: n.parent,
	}
	split.entries = groupB
	r.adjustMBRs(split)

	for _, e := range split.entries {
		if e.child != nil {
			e.child.parent = split
		}
	}

	return split
}

func (r *Rtree) quadraticSplit(entries []*entry) ([]*entry, []*entry) {
	if len(entries) == 2 {
		return entries[:1], entries[1:]
	}

	seedA, seedB := r.pickSeeds(entries)
	remaining := make([]*entry, 0, len(entries)-2)
	for _, e := range entries {
		if e != seedA && e != seedB {
			remaining = append(remaining, e)
		}
	}

	groupA := []*entry{seedA}
	groupB := []*entry{seedB}

	mbrA := seedA.mbr
	mbrB := seedB.mbr

	for len(remaining) > 0 {
		if len(remaining)+len(groupA) <= r.minEntries {
			groupA = append(groupA, remaining...)
			break
		}
		if len(remaining)+len(groupB) <= r.minEntries {
			groupB = append(groupB, remaining...)
			break
		}

		bestAIdx := -1
		bestBDiff := math.Inf(1)

		for i, e := range remaining {
			expA := mbrA.Expansion(e.mbr)
			expB := mbrB.Expansion(e.mbr)
			diff := math.Abs(expA - expB)

			if diff < bestBDiff {
				bestBDiff = diff
				bestAIdx = i
			}
		}

		chosen := remaining[bestAIdx]
		remaining = append(remaining[:bestAIdx], remaining[bestAIdx+1:]...)

		expA := mbrA.Expansion(chosen.mbr)
		expB := mbrB.Expansion(chosen.mbr)

		if expA < expB || (expA == expB && mbrA.Area() < mbrB.Area()) || (expA == expB && mbrA.Area() == mbrB.Area() && len(groupA) <= len(groupB)) {
			groupA = append(groupA, chosen)
			mbrA = mbrA.Union(chosen.mbr)
		} else {
			groupB = append(groupB, chosen)
			mbrB = mbrB.Union(chosen.mbr)
		}
	}

	return groupA, groupB
}

func (r *Rtree) pickSeeds(entries []*entry) (*entry, *entry) {
	bestDiff := math.Inf(-1)
	var bestA, bestB *entry

	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			a := entries[i]
			b := entries[j]
			union := a.mbr.Union(b.mbr)
			diff := union.Area() - a.mbr.Area() - b.mbr.Area()

			if diff > bestDiff {
				bestDiff = diff
				bestA = a
				bestB = b
			}
		}
	}

	return bestA, bestB
}

func (r *Rtree) adjustTree(n, split *node) {
	if n == r.root {
		newRoot := &node{
			typ:   internalNode,
			level: n.level + 1,
		}

		entryN := &entry{
			mbr:   r.nodeMBR(n),
			child: n,
		}
		entrySplit := &entry{
			mbr:   r.nodeMBR(split),
			child: split,
		}

		newRoot.entries = []*entry{entryN, entrySplit}
		n.parent = newRoot
		split.parent = newRoot
		r.root = newRoot
		return
	}

	parent := n.parent

	for _, e := range parent.entries {
		if e.child == n {
			e.mbr = r.nodeMBR(n)
			break
		}
	}

	if split != nil {
		entrySplit := &entry{
			mbr:   r.nodeMBR(split),
			child: split,
		}
		parent.entries = append(parent.entries, entrySplit)

		if len(parent.entries) > r.maxEntries {
			newSplit := r.splitNode(parent)
			r.adjustTree(parent, newSplit)
		} else {
			r.adjustMBRs(parent)
		}
	} else {
		r.adjustMBRs(parent)
	}
}

func (r *Rtree) findLeaf(n *node, target *entry) *node {
	if n.typ == leafNode {
		for _, e := range n.entries {
			if e == target {
				return n
			}
		}
		return nil
	}

	for _, e := range n.entries {
		if e.mbr.Intersects(target.mbr) {
			found := r.findLeaf(e.child, target)
			if found != nil {
				return found
			}
		}
	}

	return nil
}

func (r *Rtree) condenseTree(n *node) {
	if n == r.root {
		if len(n.entries) == 0 {
			r.root.typ = leafNode
			r.root.level = 0
		} else if n.typ == internalNode && len(n.entries) == 1 {
			r.root = n.entries[0].child
			r.root.parent = nil
		}
		return
	}

	parent := n.parent
	var parentEntry *entry
	for _, e := range parent.entries {
		if e.child == n {
			parentEntry = e
			break
		}
	}

	if len(n.entries) >= r.minEntries {
		parentEntry.mbr = r.nodeMBR(n)
		r.adjustMBRs(parent)
		return
	}

	leafEntries := r.collectAllLeafEntries(n)

	for i, e := range parent.entries {
		if e.child == n {
			parent.entries = append(parent.entries[:i], parent.entries[i+1:]...)
			break
		}
	}

	if len(parent.entries) < r.minEntries {
		r.condenseTree(parent)
	} else {
		r.adjustMBRs(parent)
	}

	for _, e := range leafEntries {
		leaf := r.chooseLeaf(r.root, e.mbr)
		leaf.entries = append(leaf.entries, e)
		if len(leaf.entries) > r.maxEntries {
			split := r.splitNode(leaf)
			r.adjustTree(leaf, split)
		} else {
			r.adjustMBRs(leaf)
		}
	}
}

func (r *Rtree) collectAllLeafEntries(n *node) []*entry {
	var result []*entry
	if n.typ == leafNode {
		result = append(result, n.entries...)
	} else {
		for _, e := range n.entries {
			result = append(result, r.collectAllLeafEntries(e.child)...)
		}
	}
	return result
}

func (r *Rtree) adjustMBRs(n *node) {
	for n != nil {
		for _, e := range n.entries {
			if e.child != nil {
				e.mbr = r.nodeMBR(e.child)
			}
		}
		n = n.parent
	}
}

func (r *Rtree) nodeMBR(n *node) MBR {
	if len(n.entries) == 0 {
		return MBR{}
	}

	mbr := n.entries[0].mbr
	for _, e := range n.entries[1:] {
		mbr = mbr.Union(e.mbr)
	}
	return mbr
}

func (r *Rtree) sortByDistance(objects []SpatialObject, x, y float64) []SpatialObject {
	type distObj struct {
		obj  SpatialObject
		dist float64
	}

	items := make([]distObj, 0, len(objects))
	for _, obj := range objects {
		minX, minY, maxX, maxY := obj.Bounds()
		mbr := NewMBR(minX, minY, maxX, maxY)
		items = append(items, distObj{obj, mbr.DistanceToPoint(x, y)})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].dist < items[j].dist
	})

	result := make([]SpatialObject, 0, len(items))
	for _, item := range items {
		result = append(result, item.obj)
	}

	return result
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func formatMBR(mbr MBR) string {
	return "(" + strconv.FormatFloat(mbr.MinX, 'f', 2, 64) + ", " + strconv.FormatFloat(mbr.MinY, 'f', 2, 64) + ", " +
		strconv.FormatFloat(mbr.MaxX, 'f', 2, 64) + ", " + strconv.FormatFloat(mbr.MaxY, 'f', 2, 64) + ")"
}

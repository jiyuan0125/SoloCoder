package spf

type Node struct {
	ID string
}

type Link struct {
	From   string
	To     string
	Cost   uint32
	SeqNum uint32
}

type Adjacency struct {
	Neighbor string
	Cost     uint32
	SeqNum   uint32
}

type ShortestPath struct {
	Destination string
	Paths       [][]string
	TotalCost   uint32
	Valid       bool
	Reachable   bool
}

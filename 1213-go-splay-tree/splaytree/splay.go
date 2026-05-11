package splaytree

func (t *SplayTree) rotate(x *Node) {
	if x == nil || x.Parent == nil {
		return
	}
	y := x.Parent
	z := y.Parent
	if z != nil {
		if z.Left == y {
			z.Left = x
		} else {
			z.Right = x
		}
	}
	x.Parent = z
	if y.Left == x {
		y.Left = x.Right
		if x.Right != nil {
			x.Right.Parent = y
		}
		x.Right = y
	} else {
		y.Right = x.Left
		if x.Left != nil {
			x.Left.Parent = y
		}
		x.Left = y
	}
	y.Parent = x
	y.update()
	x.update()
}

func (t *SplayTree) splay(x *Node) {
	if x == nil {
		return
	}
	for x.Parent != nil {
		y := x.Parent
		z := y.Parent
		if z != nil {
			if (z.Left == y) == (y.Left == x) {
				t.rotate(y)
			} else {
				t.rotate(x)
			}
		}
		t.rotate(x)
	}
	t.root = x
}

func (t *SplayTree) splayToRoot(x *Node) {
	if x == nil {
		return
	}
	t.splay(x)
}

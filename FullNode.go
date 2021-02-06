package main

type FullNode struct {
	gen      uint64
	children [valuesPerLevel]interface{}
}

func (n *FullNode) width() int {
	w := 0

	for _, v := range n.children {
		if v != nil {
			w++
		}
	}

	return w
}

func (n *FullNode) childAtIndex(i int) interface{} {
	return n.children[i]
}

func (n *FullNode) setChildAtIndex(i int, c interface{}) {
	n.children[i] = c
}

func (n *FullNode) iterableChildren() []interface{} {
	return n.children[:]
}

var fullCopies [valuesPerLevel + 1]int

func (n *FullNode) copy() HamtNode {
	fullCopies[n.width()]++
	newNode := *n
	return &newNode
}

func (n *FullNode) copyForGrowth() HamtNode {
	return n.copy()
}

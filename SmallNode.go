package main

import (
	"math/bits"
)

const smallWidth = 4

type SmallNode struct {
	gen      uint64
	bitmap   uint64
	children [smallWidth]interface{}
}

func elementForIndex(b uint64, i int) int {
	// The element that a particular index maps to in the sparse array
	// is equal the number of set bits _to the right_ of the bit representing
	// the index in question. So, if the right mostset bit is at index i, then
	// index i maps to the first element of the sparse array.
	// We can count this value by shifting the bitmap to the left by (64-i)
	// and counting the resulting ones. The bit for this particular index,
	// and all bits to the left of it will be shifted off the end
	// debug.Assert(b&(1<<i) != 0, "Asking for the index of something that doesn't exist!")
	return bits.OnesCount64(b << (64 - i))
}

func (n *SmallNode) width() int {
	return bits.OnesCount64(n.bitmap)
}

func (n *SmallNode) childAtIndex(i int) interface{} {
	if n.bitmap&(1<<i) != 0 {
		return n.children[elementForIndex(n.bitmap, i)]
	} else {
		return nil
	}
}

func (n *SmallNode) setChildAtIndex(i int, c interface{}) {
	if n.bitmap&(1<<i) != 0 {
		// replacing an existing value is easy
		n.children[elementForIndex(n.bitmap, i)] = c
	} else {
		// debug.Assert(n.width() < 4, "Attempt to insert a child into a full SmallNode")
		// Update the bitmap to include the new value
		n.bitmap |= 1 << i
		insertionOffset := elementForIndex(n.bitmap, i)
		copy(n.children[insertionOffset+1:], n.children[insertionOffset:])
		n.children[insertionOffset] = c
	}
}

func (n *SmallNode) iterableChildren() []interface{} {
	return n.children[:]
}

var growCount = 0
var smallCopies = 0

func (n *SmallNode) copy() HamtNode {
	smallCopies++
	newNode := *n
	return &newNode
}

func (n *SmallNode) copyForGrowth() HamtNode {
	if n.width() < smallWidth {
		return n.copy()
	} else {
		newNode := FullNode{}
		growCount++
		// TODO: This is wildly inefficient!
		for i := 0; i < valuesPerLevel; i++ {
			newNode.setChildAtIndex(i, n.childAtIndex(i))
		}

		return &newNode
	}
}

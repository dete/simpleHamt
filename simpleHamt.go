package main

import (
	"fmt"
)

type HamtKey interface {
	Equal(HamtKey) bool
	Hash() uint64
}

const bitsPerLevel = 6
const valuesPerLevel = 1 << bitsPerLevel
const maxDepth = 64 / bitsPerLevel

type CowSet interface {
	insert(HamtKey)
	// delete(HamtKey)
	find(HamtKey) HamtKey
	contains(HamtKey) bool
	// iterate(func (HamtKey) bool) bool

	debugPrint()
	dumpStats()

	// copy() CowSet
}

func NewCowSet() CowSet {
	return new(HamtRoot)
}

type HamtRoot struct {
	root HamtNode
	gen  uint64
}

func (r *HamtRoot) insert(v HamtKey) {
	if r.root == nil {
		r.root = new(SmallNode)
	}

	r.root = internalInsert(r.root, v, v.Hash(), 0, true)
}

func (r *HamtRoot) find(v HamtKey) HamtKey {
	if r.root != nil {
		return internalFind(r.root, v, v.Hash(), 0)
	} else {
		return nil
	}
}

func (n *HamtRoot) contains(v HamtKey) bool {
	return n.find(v) != nil
}

type HamtNode interface {
	// generation() uint64
	width() int
	childAtIndex(int) interface{}
	setChildAtIndex(int, interface{})
	iterableChildren() []interface{}
	copy() HamtNode
	copyForGrowth() HamtNode
}

func indexForDepth(h uint64, d int) int {
	return int((h >> (d * bitsPerLevel)) % valuesPerLevel)
}

func newNodeWithValues(v1 HamtKey, v2 HamtKey, d int) HamtNode {
	result := new(SmallNode)

	i1 := indexForDepth(v1.Hash(), d)
	i2 := indexForDepth(v2.Hash(), d)

	if i1 != i2 {
		result.setChildAtIndex(i1, v1)
		result.setChildAtIndex(i2, v2)
	} else if d < maxDepth {
		result.setChildAtIndex(i1, newNodeWithValues(v1, v2, d+1))
	} else {
		// At max depth, we map colliding values into an array
		result.setChildAtIndex(i1, []HamtKey{v1, v2})
	}

	return result
}

func internalInsert(n HamtNode, v HamtKey, h uint64, d int, overwrite bool) HamtNode {
	index := indexForDepth(h, d)

	switch c := n.childAtIndex(index).(type) {
	case HamtNode:
		// There is a node in the spot we need, we ask that node to insert this value as a child
		newChild := internalInsert(c, v, h, d+1, overwrite)

		if newChild == c {
			// The new child was same as the old child, the insert must not have been
			// necessary, just return the same Node object we started with.
			return n
		} else {
			// We've got a new node for this child slot, create a new node at this
			// level, and insert the new child node.
			newNode := n.copy()
			newNode.setChildAtIndex(index, newChild)

			return newNode
		}
	case HamtKey:
		// There's already a value in the slot we want, first check to see if it's equal to the existing value
		if v.Equal(c) {
			if overwrite {
				// Create a new node with the new value replacing the old value
				newNode := n.copy()
				newNode.setChildAtIndex(index, v)

				return newNode
			} else {
				// We aren't overwriting existing values, so just return the same
				// node we started with
				return n
			}
		} else {
			// There is a value in the slot we want, but it's not equal to the new
			// value, we need to create a new node that has both values as children
			newNode := n.copy()
			newNode.setChildAtIndex(index, newNodeWithValues(v, c, d+1))

			return newNode
		}
	case []HamtKey:
		fmt.Println("COLLISION")
		return nil
	case nil:
		// There is no value in the slot we need, so we can just
		// store the value as a child
		newNode := n.copyForGrowth()
		newNode.setChildAtIndex(index, v)

		return newNode
	default:
		fmt.Println("PANIC")
		return nil
	}

	//     // We have a hash collision at the maximum depth in the tree. We use an array instead
	//     // of a Node to collect all of the colliding children. This seems like it might be
	//     // expensive, but remember that it only happens in the case of a 64-bit hash collision...
	//     if n.nodeMap&(1<<index) != 0 {
	//         // At max depth, a set bit in the nodeMap means we've already had a collision before
	//         // and there is a slice stored where we'd normally put a node
	//         previousCollisions := n.children[index].([]HamtKey)
	//         duplicateIndex := -1
	//
	//         for ci, cv := range previousCollisions {
	//             if cv.Equal(v) {
	//                 duplicateIndex = ci
	//             }
	//         }
	//
	//         if duplicateIndex > 0 {
	//             // we found the value we were looking for
	//             if overwrite {
	//                 // Return a new node with a new slice in the hash slot, but with
	//                 // the new value in place of the old value
	//                 newNode := *n
	//                 newSlice := make([]HamtKey, len(previousCollisions))
	//                 copy(newSlice, previousCollisions)
	//                 newSlice[duplicateIndex] = v
	//                 newNode.children[index] = newSlice
	//
	//                 totalNodesByWidth[bits.OnesCount64(newNode.childMap)]++
	//                 return &newNode
	//             } else {
	//                 // Value's already where we expect it, no updates needed, just return the same node
	//                 return n
	//             }
	//         } else {
	//             // The new value isn't already in the array, we can just append it
	//             newNode := *n
	//             newNode.children[index] = append(previousCollisions, v)
	//
	//             totalNodesByWidth[bits.OnesCount64(newNode.childMap)]++
	//             return &newNode
	//         }
	//     } else {
	//         // There is already a value in the slot we want
	//         existingValue := n.children[index].(HamtKey)
	//
	//         // First check to see if it's equal to the existing value
	//         if v.Equal(existingValue) {
	//             if overwrite {
	//                 // Create a new node with the new value replacing the old value
	//                 newNode := *n
	//                 newNode.children[index] = v
	//
	//                 totalNodesByWidth[bits.OnesCount64(newNode.childMap)]++
	//                 return &newNode
	//             } else {
	//                 // We aren't overwriting existing values, so just return the same
	//                 // node we started with
	//                 return n
	//             }
	//         } else {
	//             // There is a value in the slot we want, but it's not equal to the new
	//             // value, we need to create a slice that has both values as children
	//             newNode := *n
	//             newNode.nodeMap |= (1 << index)
	//             newNode.children[index] = []HamtKey{existingValue, v}
	//
	//             totalNodesByWidth[bits.OnesCount64(newNode.childMap)]++
	//             return &newNode
	//         }
	//     }
	// }
}

// func (n *HamtNode) delete(v HamtKey) *HamtNode {
//     result := n.internalDelete(v, 0)
//
//     if result == nil {
//         result = new(HamtNode)
//     }
//
//     return result
// }
//
// func (n *HamtNode) internalDelete(v HamtKey, d uint64) *HamtNode {
//     index := indexForDepth(v, d)
//
//     if n.childMap&(1<<index) == 0 {
//         // The value in question doesn't exist, just return the unchanged node
//         return n
//     } else if d < maxDepth {
//         if n.nodeMap&(1<<index) != 0 {
//             // The value is in a subnode, ask it to handle the deletion
//             newChild := n.children[index].(*HamtNode).internalDelete(v, d+1)
//
//             if n.children[index] == newChild {
//                 // The new child was same as the old child, the delete must not have been
//                 // necessary, just return the same Node object we started with.
//                 return n
//             } else {
//                 if n.children[index] == newChild {
//                     // The new child was same as the old child, the insert must not have been
//                     // necessary, just return the same Node object we started with.
//                     return n
//                 } else {
//                     // We've got a new node for this index, create a new node at this
//                     // level, and insert the new node.
//                     newNode := *n
//                     newNode.children[index] = newChild
//
//                     return &newNode
//                 }
//
//
// }

func internalFind(n HamtNode, v HamtKey, h uint64, d int) HamtKey {
	index := indexForDepth(h, d)

	switch c := n.childAtIndex(index).(type) {
	case HamtNode:
		return internalFind(c, v, h, d+1)
	case HamtKey:
		if c.Equal(v) {
			return c
		} else {
			return nil
		}
	case []HamtKey:
		for _, cv := range c {
			if cv.Equal(v) {
				return cv
			}
		}

		return nil
	default:
		return nil
	}
}

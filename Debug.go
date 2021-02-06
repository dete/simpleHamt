package main

import (
	"fmt"
)

func (s HamtRoot) debugPrint() {
	if s.root != nil {
		debugPrint(s.root)
	} else {
		fmt.Print("()")
	}
}

func debugPrint(n HamtNode) {
	fmt.Print("(")
	for i := 0; i < valuesPerLevel; i++ {
		switch c := n.childAtIndex(i).(type) {
		case HamtNode:
			debugPrint(c)
		case HamtKey:
			fmt.Print(c, ", ")
		case []HamtKey:
			fmt.Print("{")
			for _, v := range c {
				fmt.Print(v, ", ")
			}
			fmt.Print("}")
		}
	}
	fmt.Print(")")
}

type HamtStats struct {
	nodes           uint64
	nodesByWidth    [valuesPerLevel + 1]uint64
	nodesByDepth    [maxDepth + 1]uint64
	leavesByDepth   [maxDepth + 1]uint64
	collidingLeaves uint64
}

func (s HamtRoot) dumpStats() {
	stats := HamtStats{}

	internalDumpStats(s.root, &stats, 0)

	fmt.Println("Width:  ", stats.nodesByWidth)
	fmt.Println("Depth:  ", stats.nodesByDepth)
	fmt.Println("Leaves: ", stats.leavesByDepth)
	fmt.Println("Nodes: ", stats.nodes, " Collisions: ", stats.collidingLeaves)
}

func internalDumpStats(n HamtNode, stats *HamtStats, d int) {
	stats.nodes++
	stats.nodesByDepth[d]++
	stats.nodesByWidth[n.width()]++

	for i := 0; i < valuesPerLevel; i++ {
		switch c := n.childAtIndex(i).(type) {
		case HamtNode:
			internalDumpStats(c, stats, d+1)
		case HamtKey:
			stats.leavesByDepth[d]++
		case []HamtKey:
			stats.leavesByDepth[d] += uint64(len(c))
			stats.collidingLeaves += uint64(len(c))
		}
	}
}

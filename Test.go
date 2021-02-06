package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

func ByteCount(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func testInsert(n CowSet, v string) {
	n.insert(StringKey(v))
	n.debugPrint()
	n.dumpStats()
}

func testFind(n CowSet, v string) {
	foundObject := n.find(StringKey(v))

	if foundObject == nil {
		fmt.Println(v, "not found")
	} else if foundObject.Equal(StringKey(v)) {
		fmt.Println(v, "found")
	} else {
		fmt.Println("!! Looking for ", v, ", found ", foundObject)
	}
}

func smallTest() {
	set := NewCowSet()
	set.debugPrint()
	fmt.Println()

	testInsert(set, "bob")
	testInsert(set, "sally")
	testInsert(set, "bob")

	testInsert(set, "hank")
	testInsert(set, "sue")
	testInsert(set, "billy")
	testInsert(set, "ned")
	testInsert(set, "zippy")
	testInsert(set, "stu")
	testInsert(set, "fred")
	testInsert(set, "leopold")
	testInsert(set, "groucho")

	testFind(set, "bob")
	testFind(set, "sally")
	testFind(set, "hank")
	testFind(set, "sue")
	testFind(set, "billy")
	testFind(set, "ned")
	testFind(set, "zippy")
	testFind(set, "stu")
	testFind(set, "fred")
	testFind(set, "leopold")
	testFind(set, "groucho")

	testFind(set, "boopsy")
	testFind(set, "snucko")
	testFind(set, "flash")
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func bigTest() {
	testArray, _ := readLines("testData.txt")

	fmt.Println("Input size: ", len(testArray))

	runtime.GC()
	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	fmt.Println("Baseline: ", ByteCount(ms.HeapAlloc))

	set := NewCowSet()

	profFile, _ := os.Create("insert.prof")
	defer profFile.Close()

	pprof.StartCPUProfile(profFile)
	start := time.Now()

	for _, v := range testArray {
		set.insert(StringKey(v))
	}

	elapsed := time.Since(start)
	pprof.StopCPUProfile()

	set.dumpStats()
	fmt.Println("Insert: ", elapsed)
	runtime.GC()
	ms = runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	fmt.Println("Mem: ", ByteCount(ms.HeapAlloc))

	start = time.Now()
	for _, v := range testArray {
		set.contains(StringKey(v))
	}

	elapsed = time.Since(start)
	fmt.Println("Contains: ", elapsed)

	fmt.Println("Grown: ", growCount)
	fmt.Println("Small Copies: ", smallCopies)
	fmt.Println("Full Copies: ", fullCopies)
}

func bigTestNative() {
	testArray, _ := readLines("testData.txt")

	fmt.Println("Input size: ", len(testArray))

	runtime.GC()
	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	fmt.Println("Baseline: ", ByteCount(ms.HeapAlloc))

	set := make(map[string]bool)

	start := time.Now()

	for _, v := range testArray {
		set[v] = true
	}

	elapsed := time.Since(start)
	fmt.Println("Insert: ", elapsed)

	runtime.GC()
	ms = runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	fmt.Println("Mem: ", ByteCount(ms.HeapAlloc))

	start = time.Now()
	for _, v := range testArray {
		if !set[v] == true {
			fmt.Println("Didn't find ", v)
		}
	}

	elapsed = time.Since(start)
	fmt.Println("Contains: ", elapsed)

	start = time.Now()
	newSet := make(map[string]bool)

	for key, value := range set {
		newSet[key] = value
	}

	elapsed = time.Since(start)
	fmt.Println("Copy: ", elapsed)

}

func main() {
	// smallTest()
	bigTest()
	//	bigTestNative()
}

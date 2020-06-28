package main

import (
	"fmt"	
	"os"
	"io/ioutil"
	"container/heap"
	"sort"
)


// Filesize
type filesize int64

func (h *filesize) String() string {
	return "1234"
}

// Record
type record struct {
	size filesize
	path string
	par *record
	child []record
}

func newRecord(size filesize, path string, par *record, isDir bool) record {
	ret := record{}
	ret.size = size
	ret.path = path
	ret.par = par
	if isDir {
		ret.child = make([]record, 0)
	} else {
		ret.child = nil
	}

	return ret
}


// Priority Queue
type priority_queue []*record

func (h priority_queue) Len() int           { return len(h) }
func (h priority_queue) Less(i, j int) bool { return h[i].size > h[j].size }
func (h priority_queue) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *priority_queue) Push(x interface{}) {
	*h = append(*h, x.(*record))
}

func (h *priority_queue) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func traverse(path string, par *record) (record, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return record{}, err
	}

	mode := stat.Mode()
	if !mode.IsDir() {
		return newRecord(filesize(stat.Size()), path, par, false), nil
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return record{}, err
	}

	cur_file := newRecord(0, path, par, true)

	for _, file := range files {
		chi_file, err := traverse(file.Name(), &cur_file)

		if err != nil {
			continue
		}

		cur_file.size += chi_file.size
		cur_file.child = append(cur_file.child, chi_file)
	}

	return cur_file, nil
}

var list_size = 0
var result = make(map[*record]bool)
var pq = &priority_queue{}

func main() {
	root, err := traverse("/Users/soultch/Project", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	heap.Push(pq, root)

	for len(result) < list_size {
		item := heap.Pop(pq).(*record)
		if _, ok := result[item.par]; ok {
			list_size -= 1
			delete(result, item.par)
		}
		
		result[item] = true
		for _, file := range item.child {
			heap.Push(pq, file)
		}
	}

	fin := make([]*record, 0)
	for key, _ := range result {
		fin = append(fin, key)
	}

	sort.Slice(fin, func(i, j int) bool {
		return fin[i].size > fin[j].size
	})

	for _, val := range fin {
		fmt.Println(val.size, val.path)
	}
}
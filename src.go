package main

import (
	"fmt"	
	"os"
	"flag"
	"io/ioutil"
	"container/heap"
	"sort"
	"path"
)

// Filesize
type filesize int64

func (h filesize) String() string {
	names := []string{"B", "KiB", "MiB", "GiB"}
	
	size := float64(h)
	for i := 0; i < len(names); i++ {
		if size < 1000 {
			numbs := 0
			switch {
			case i == 0: numbs = 0
			case size < 10: numbs = 2
			case size < 100: numbs = 1
			case size < 1000: numbs = 0
			}

			return fmt.Sprintf("%.*f %s", numbs, size, names[i])
		}
		
		size /= 1024
	}

	return fmt.Sprintf("%.0f GiB", size*1024)
}

// Record
type record struct {
	size filesize
	cpath string
	par *record
	child []*record
}

func newRecord(size filesize, cpath string, par *record, isDir bool) *record {
	ret := new(record)
	ret.size = size
	ret.cpath = cpath
	ret.par = par
	if isDir {
		ret.child = make([]*record, 0)
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

func traverse(cpath string, par *record) (*record, error) {
	stat, err := os.Stat(cpath)
	if err != nil {
		return nil, err
	}

	mode := stat.Mode()
	if !mode.IsDir() {
		return newRecord(filesize(stat.Size()), cpath, par, false), nil
	}

	files, err := ioutil.ReadDir(cpath)
	if err != nil {
		return nil, err
	}

	cur_file := newRecord(filesize(stat.Size()), cpath, par, true)

	for _, file := range files {
		chi_file, err := traverse(path.Join(cpath, file.Name()), cur_file)

		if err != nil {
			continue
		}

		cur_file.size += chi_file.size
		cur_file.child = append(cur_file.child, chi_file)
	}

	return cur_file, nil
}

var list_size *int
var target_path string

var result = make(map[*record]bool)
var pq = &priority_queue{}

func main() {
	list_size = flag.Int("size", 20, "Number of maximum items to find out")

	flag.Parse()
	if target_path = "."; len(flag.Args()) > 0 {
		target_path = flag.Args()[0]
	}

	root, err := traverse(target_path, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	heap.Init(pq)
	heap.Push(pq, root)

	for len(result) < *list_size && pq.Len() > 0 {
		item := heap.Pop(pq).(*record)

		if _, ok := result[item.par]; ok {
			delete(result, item.par)
		}
		
		result[item] = true
		if item.child != nil {
			for _, file := range item.child {
				heap.Push(pq, file)
			}
		}
	}

	fin := make([]*record, 0, len(result))
	for key, _ := range result {
		fin = append(fin, key)
	}

	sort.Slice(fin, func(i, j int) bool {
		return fin[i].size > fin[j].size
	})

	for _, val := range fin {
		fmt.Printf("%-70s %10s\n", val.cpath, val.size)
	}
}
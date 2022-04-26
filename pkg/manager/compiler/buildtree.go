package compiler

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"encoding/json"
	"sort"
)

func rankMapStringInt(values map[string]int) []string {
	type kv struct {
		Key   string
		Value int
	}
	var ss []kv
	for k, v := range values {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	ranked := make([]string, len(values))
	for i, kv := range ss {
		ranked[i] = kv.Key
	}
	return ranked
}

type BuildTree struct {
	order map[string]int
}

func (bt *BuildTree) Increase(s string) {
	if bt.order == nil {
		bt.order = make(map[string]int)
	}
	if _, ok := bt.order[s]; !ok {
		bt.order[s] = 0
	}

	bt.order[s]++
}

func (bt *BuildTree) Reset(s string) {
	if bt.order == nil {
		bt.order = make(map[string]int)
	}
	bt.order[s] = 0
}

func (bt *BuildTree) Level(s string) int {
	return bt.order[s]
}

func ints(input []int) []int {
	u := make([]int, 0, len(input))
	m := make(map[int]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}

func (bt *BuildTree) AllInLevel(l int) []string {
	var all []string
	for k, v := range bt.order {
		if v == l {
			all = append(all, k)
		}
	}
	return all
}

func (bt *BuildTree) Order(compilationTree map[string]map[string]interface{}) {
	sentinel := false
	for !sentinel {
		sentinel = true

	LEVEL:
		for _, l := range bt.AllLevels() {

			for _, j := range bt.AllInLevel(l) {
				for _, k := range bt.AllInLevel(l) {
					if j == k {
						continue
					}
					if _, ok := compilationTree[j][k]; ok {
						if bt.Level(j) == bt.Level(k) {
							bt.Increase(j)
							sentinel = false
							break LEVEL
						}
					}
				}
			}
		}
	}
}

func (bt *BuildTree) AllLevels() []int {
	var all []int
	for _, v := range bt.order {
		all = append(all, v)
	}

	sort.Sort(sort.IntSlice(all))

	return ints(all)
}

func (bt *BuildTree) JSON() (string, error) {
	type buildjob struct {
		Jobs []string `json:"packages"`
	}

	result := []buildjob{}
	for _, l := range bt.AllLevels() {
		result = append(result, buildjob{Jobs: bt.AllInLevel(l)})
	}
	dat, err := json.Marshal(&result)
	return string(dat), err
}

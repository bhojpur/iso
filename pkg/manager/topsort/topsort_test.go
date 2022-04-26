package topsort

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
	"strings"
	"testing"
)

func TestTopSort(t *testing.T) {
	graph := initGraph()

	// a -> b -> c
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")

	results, err := graph.TopSort("a")
	if err != nil {
		t.Error(err)
		return
	}
	if results[0] != "c" || results[1] != "b" || results[2] != "a" {
		t.Errorf("Wrong sort order: %v", results)
	}
}

func TestTopSort2(t *testing.T) {
	graph := initGraph()

	// a -> c
	// a -> b
	// b -> c
	graph.AddEdge("a", "c")
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")

	results, err := graph.TopSort("a")
	if err != nil {
		t.Error(err)
		return
	}
	if results[0] != "c" || results[1] != "b" || results[2] != "a" {
		t.Errorf("Wrong sort order: %v", results)
	}
}

func TestTopSort3(t *testing.T) {
	graph := initGraph()

	// a -> b
	// a -> d
	// d -> c
	// c -> b
	graph.AddEdge("a", "b")
	graph.AddEdge("a", "d")
	graph.AddEdge("d", "c")
	graph.AddEdge("c", "b")

	results, err := graph.TopSort("a")
	if err != nil {
		t.Error(err)
		return
	}
	if len(results) != 4 {
		t.Errorf("Wrong number of results: %v", results)
		return
	}
	expected := [4]string{"b", "c", "d", "a"}
	for i := 0; i < 4; i++ {
		if results[i] != expected[i] {
			t.Errorf("Wrong sort order: %v", results)
			break
		}
	}
}

func TestTopSortCycleError(t *testing.T) {
	graph := initGraph()

	// a -> b
	// b -> a
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "a")

	_, err := graph.TopSort("a")
	if err == nil {
		t.Errorf("Expected cycle error")
		return
	}
	if !strings.Contains(err.Error(), "a -> b -> a") {
		t.Errorf("Error doesn't print cycle: %q", err)
	}
}

func TestTopSortCycleError2(t *testing.T) {
	graph := initGraph()

	// a -> b
	// b -> c
	// c -> a
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")
	graph.AddEdge("c", "a")

	_, err := graph.TopSort("a")
	if err == nil {
		t.Errorf("Expected cycle error")
		return
	}
	if !strings.Contains(err.Error(), "a -> b -> c -> a") {
		t.Errorf("Error doesn't print cycle: %q", err)
	}
}

func TestTopSortCycleError3(t *testing.T) {
	graph := initGraph()

	// a -> b
	// b -> c
	// c -> b
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")
	graph.AddEdge("c", "b")

	_, err := graph.TopSort("a")
	if err == nil {
		t.Errorf("Expected cycle error")
		return
	}
	if !strings.Contains(err.Error(), "b -> c -> b") {
		t.Errorf("Error doesn't print cycle: %q", err)
	}
}

func initGraph() *Graph {
	graph := NewGraph()
	graph.AddNode("a")
	graph.AddNode("b")
	graph.AddNode("c")
	graph.AddNode("d")
	return graph
}

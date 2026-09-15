// -*- mode: go; coding: utf-8 -*-
//
// SPDX-FileCopyrightText: 2026 Frederico Martins
// SPDX-License-Identifier: GPL-3.0-only

package xdu

import (
	"math"
	"reflect"
	"testing"
)

func n(path string, size, inodes int64) *Node {
	return &Node{Path: path, Size: size, Inodes: inodes}
}

func paths(nodes []*Node) []string {
	out := make([]string, len(nodes))
	for i, n := range nodes {
		out[i] = n.Path
	}
	return out
}

func TestLessBySize(t *testing.T) {
	tests := []struct {
		name       string
		a, b       *Node
		reverse    bool
		inodes     bool
		wantBefore bool // a should sort before b
	}{
		{"size a<b", n("a", 10, 100), n("b", 20, 50), false, false, true},
		{"size a>b", n("a", 20, 50), n("b", 10, 100), false, false, false},
		{
			"size a==b tie path a<b",
			n("a", 10, 5),
			n("b", 10, 99),
			false,
			false,
			true,
		},
		{
			"size a==b tie path a>b",
			n("b", 10, 5),
			n("a", 10, 99),
			false,
			false,
			false,
		},
		{
			"size a==b tie path equal",
			n("a", 10, 5),
			n("a", 10, 5),
			false,
			false,
			false,
		}, // a.Path < b.Path false when equal
		{
			"size reverse a<b",
			n("a", 10, 100),
			n("b", 20, 50),
			true,
			false,
			false,
		},
		{
			"size reverse a>b",
			n("a", 20, 50),
			n("b", 10, 100),
			true,
			false,
			true,
		},
		{
			"size reverse tie",
			n("a", 10, 5),
			n("b", 10, 99),
			true,
			false,
			true,
		},
		{"inodes a<b", n("a", 999, 10), n("b", 111, 20), false, true, true},
		{"inodes a>b", n("a", 111, 20), n("b", 999, 10), false, true, false},
		{"inodes tie path", n("a", 5, 10), n("b", 99, 10), false, true, true},
		{"inodes reverse a<b", n("a", 0, 10), n("b", 0, 20), true, true, false},
		{"inodes reverse a>b", n("a", 0, 20), n("b", 0, 10), true, true, true},
		{
			"inodes reverse tie",
			n("a", 50, 10),
			n("b", 99, 10),
			true,
			true,
			true,
		},
		{"negative size a<b", n("a", -10, 0), n("b", 5, 0), false, false, true},
		{
			"max int64 tie",
			n("a", math.MaxInt64, 0),
			n("b", math.MaxInt64, 0),
			false,
			false,
			true,
		},
		{
			"max int64 diff reverse",
			n("a", math.MaxInt64, 0),
			n("b", math.MaxInt64-1, 0),
			true,
			false,
			true,
		},
	}
	for _, tc := range tests {
		got := sortBySize(tc.a, tc.b, tc.reverse, tc.inodes)
		if got != tc.wantBefore {
			t.Errorf(
				"%s: lessBySize(%v,%v, reverse=%v,inodes=%v)=%v want %v",
				tc.name,
				tc.a,
				tc.b,
				tc.reverse,
				tc.inodes,
				got,
				tc.wantBefore,
			)
		}
	}
}

func TestSortList_Lexical(t *testing.T) {
	cases := []struct {
		name    string
		in      []*Node
		reverse bool
		inodes  bool
		want    []string
	}{
		{"empty", nil, false, false, []string{}},
		{"single", []*Node{n("b", 10, 1)}, false, false, []string{"b"}},
		{
			"already sorted asc",
			[]*Node{n("a", 1, 1), n("b", 2, 2), n("c", 3, 3)},
			false,
			false,
			[]string{"a", "b", "c"},
		},
		{
			"already sorted desc",
			[]*Node{n("a", 1, 1), n("b", 2, 2), n("c", 3, 3)},
			true,
			false,
			[]string{"c", "b", "a"},
		},
		{
			"reverse lex desc",
			[]*Node{n("c", 3, 3), n("b", 2, 2), n("a", 1, 1)},
			true,
			false,
			[]string{"c", "b", "a"},
		},
		{
			"unsorted asc",
			[]*Node{n("z", 1, 1), n("m", 2, 2), n("a", 3, 3)},
			false,
			false,
			[]string{"a", "m", "z"},
		},
		{
			"unsorted desc",
			[]*Node{n("z", 1, 1), n("m", 2, 2), n("a", 3, 3)},
			true,
			false,
			[]string{"z", "m", "a"},
		},
		{
			"inodes ignored when bySize false asc",
			[]*Node{n("b", 100, 1), n("a", 1, 100)},
			false,
			true,
			[]string{"a", "b"},
		},
		{
			"inodes ignored when bySize false desc",
			[]*Node{n("b", 100, 1), n("a", 1, 100)},
			true,
			true,
			[]string{"b", "a"},
		},
		{
			"case sensitive asc",
			[]*Node{n("B", 1, 1), n("a", 1, 1)},
			false,
			false,
			[]string{"B", "a"},
		},
		{
			"case sensitive desc",
			[]*Node{n("B", 1, 1), n("a", 1, 1)},
			true,
			false,
			[]string{"a", "B"},
		},
	}
	for _, tc := range cases {
		nodes := append([]*Node(nil), tc.in...)
		SortList(nodes, false, tc.reverse, tc.inodes)
		got := paths(nodes)
		if len(got) == 0 && len(tc.want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

func TestSortList_BySize(t *testing.T) {
	tests := []struct {
		name    string
		in      []*Node
		reverse bool
		inodes  bool
		want    []string
	}{
		{
			"size asc",
			[]*Node{n("c", 30, 3), n("a", 10, 9), n("b", 20, 2)},
			false,
			false,
			[]string{"a", "b", "c"},
		},
		{
			"size desc",
			[]*Node{n("a", 10, 9), n("b", 20, 2), n("c", 30, 3)},
			true,
			false,
			[]string{"c", "b", "a"},
		},
		{
			"size tie break path asc even when reverse",
			[]*Node{n("b", 10, 1), n("a", 10, 2), n("c", 10, 3)},
			true,
			false,
			[]string{"a", "b", "c"},
		},
		{
			"size tie break path asc",
			[]*Node{n("b", 10, 1), n("a", 10, 2)},
			false,
			false,
			[]string{"a", "b"},
		},
		{
			"inodes asc",
			[]*Node{n("c", 999, 30), n("a", 111, 10), n("b", 222, 20)},
			false,
			true,
			[]string{"a", "b", "c"},
		},
		{
			"inodes desc",
			[]*Node{n("a", 111, 10), n("b", 222, 20), n("c", 999, 30)},
			true,
			true,
			[]string{"c", "b", "a"},
		},
		{
			"inodes tie path",
			[]*Node{n("b", 5, 10), n("a", 99, 10)},
			false,
			true,
			[]string{"a", "b"},
		},
		{"empty", nil, false, false, []string{}},
		{"single", []*Node{n("x", 5, 5)}, true, true, []string{"x"}},
	}
	for _, tc := range tests {
		nodes := append([]*Node(nil), tc.in...)
		SortList(nodes, true, tc.reverse, tc.inodes)
		got := paths(nodes)
		if len(got) == 0 && len(tc.want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf(
				"%s: got %v want %v (reverse=%v inodes=%v)",
				tc.name,
				got,
				tc.want,
				tc.reverse,
				tc.inodes,
			)
		}
	}
}

func TestSortList_Stable(t *testing.T) {
	a1 := n("same", 10, 1)
	a2 := n("same", 10, 1)
	a3 := n("same", 10, 1)
	nodes := []*Node{a3, a1, a2}
	SortList(nodes, true, false, false)
	if nodes[0] != a3 || nodes[1] != a1 || nodes[2] != a2 {
		t.Errorf(
			"stable: got %p %p %p want %p %p %p",
			nodes[0],
			nodes[1],
			nodes[2],
			a3,
			a1,
			a2,
		)
	}
	b1 := n("same", 1, 1)
	b2 := n("same", 2, 2)
	nodes2 := []*Node{b2, b1}
	SortList(nodes2, false, false, false)
	if nodes2[0] != b2 || nodes2[1] != b1 {
		t.Errorf(
			"lexical stable: got %p %p want %p %p",
			nodes2[0],
			nodes2[1],
			b2,
			b1,
		)
	}
}

func TestSortTree_NilAndEmpty(t *testing.T) {
	SortTree(nil, false, false, false)
	SortTree(nil, true, true, true)
	SortTree(&Node{}, false, false, false)
	SortTree(&Node{}, true, false, false)
	leaf := n("leaf", 10, 1)
	parent := &Node{Path: "root", Children: []*Node{leaf}}
	SortTree(parent, false, false, false)
	if len(parent.Children) != 1 || parent.Children[0] != leaf {
		t.Error("single child should remain")
	}
}

func TestSortTree_Lexical(t *testing.T) {
	root := &Node{
		Path: "root",
		Children: []*Node{
			n("z", 3, 3),
			n("a", 1, 1),
			{Path: "m", Size: 2, Inodes: 2, Children: []*Node{
				n("z2", 1, 1),
				n("a2", 2, 2),
			}},
		},
	}
	SortTree(root, false, false, false)
	if got := paths(root.Children); !reflect.DeepEqual(
		got,
		[]string{"a", "m", "z"},
	) {
		t.Errorf("root children lexical got %v want [a m z]", got)
	}
	var mNode *Node
	for _, c := range root.Children {
		if c.Path == "m" {
			mNode = c
		}
	}
	if mNode == nil {
		t.Fatal("missing m")
	}
	if got := paths(mNode.Children); !reflect.DeepEqual(
		got,
		[]string{"a2", "z2"},
	) {
		t.Errorf("m children lexical got %v want [a2 z2]", got)
	}
	root2 := &Node{
		Path:     "root",
		Children: []*Node{n("c", 1, 1), n("a", 1, 1), n("b", 1, 1)},
	}
	SortTree(root2, false, true, false)
	if got := paths(root2.Children); !reflect.DeepEqual(
		got,
		[]string{"c", "b", "a"},
	) {
		t.Errorf("lexical with reverse true got %v want [c b a]", got)
	}
}

func TestSortTree_BySize(t *testing.T) {
	root := &Node{
		Path: "root",
		Children: []*Node{
			n("big", 100, 1),
			n("small", 10, 9),
			n("mid", 50, 5),
		},
	}
	SortTree(root, true, false, false)
	if got := paths(root.Children); !reflect.DeepEqual(
		got,
		[]string{"small", "mid", "big"},
	) {
		t.Errorf("bySize asc got %v want [small mid big]", got)
	}
	SortTree(root, true, true, false)
	if got := paths(root.Children); !reflect.DeepEqual(
		got,
		[]string{"big", "mid", "small"},
	) {
		t.Errorf("bySize desc got %v want [big mid small]", got)
	}
}

func TestSortTree_ByInodes(t *testing.T) {
	root := &Node{
		Path: "root",
		Children: []*Node{
			n("a", 999, 1),
			n("b", 1, 10),
			n("c", 500, 5),
		},
	}
	SortTree(root, true, false, true)
	if got := paths(root.Children); !reflect.DeepEqual(
		got,
		[]string{"a", "c", "b"},
	) {
		t.Errorf("byInodes asc got %v want [a c b]", got)
	}
	SortTree(root, true, true, true)
	if got := paths(root.Children); !reflect.DeepEqual(
		got,
		[]string{"b", "c", "a"},
	) {
		t.Errorf("byInodes desc got %v want [b c a]", got)
	}
}

func TestSortTree_BySizeTieBreak(t *testing.T) {
	root := &Node{
		Path:     "root",
		Children: []*Node{n("b", 10, 1), n("a", 10, 2)},
	}
	SortTree(root, true, false, false)
	if got := paths(root.Children); !reflect.DeepEqual(
		got,
		[]string{"a", "b"},
	) {
		t.Errorf("tie break got %v want [a b]", got)
	}
	SortTree(root, true, true, false)
	if got := paths(root.Children); !reflect.DeepEqual(
		got,
		[]string{"a", "b"},
	) {
		t.Errorf(
			"tie break reverse got %v want [a b] (tie still path asc)",
			got,
		)
	}
}

func TestSortTree_Recursive(t *testing.T) {
	leaf1 := n("leaf1", 30, 1)
	leaf2 := n("leaf2", 10, 1)
	leaf3 := n("leaf3", 20, 1)
	mid := &Node{Path: "mid", Size: 50, Children: []*Node{leaf1, leaf2, leaf3}}
	root := &Node{Path: "root", Children: []*Node{
		{Path: "z", Size: 100, Children: []*Node{n("z1", 5, 1)}},
		mid,
		n("a", 1, 1),
	}}
	SortTree(root, true, false, false)
	if got := paths(root.Children); !reflect.DeepEqual(
		got,
		[]string{"a", "mid", "z"},
	) {
		t.Errorf("recursive root got %v want [a mid z]", got)
	}
	if got := paths(mid.Children); !reflect.DeepEqual(
		got,
		[]string{"leaf2", "leaf3", "leaf1"},
	) {
		t.Errorf("mid leaves got %v want [leaf2 leaf3 leaf1]", got)
	}
	mid2 := &Node{Path: "mid", Children: []*Node{n("z", 1, 1), n("a", 1, 1)}}
	root2 := &Node{Path: "root", Children: []*Node{mid2}}
	SortTree(root2, false, false, false)
	if got := paths(mid2.Children); !reflect.DeepEqual(
		got,
		[]string{"a", "z"},
	) {
		t.Errorf("lexical recursion got %v want [a z]", got)
	}
}

func TestSortTree_DeepRecursion(t *testing.T) {
	c := n("c", 3, 1)
	b := &Node{Path: "b", Children: []*Node{c}}
	a := &Node{Path: "a", Children: []*Node{b}}
	root := &Node{Path: "root", Children: []*Node{a}}
	SortTree(root, true, true, false)
	if len(root.Children) != 1 || root.Children[0].Path != "a" {
		t.Error("deep chain root")
	}
	if len(a.Children) != 1 || a.Children[0].Path != "b" {
		t.Error("deep chain a")
	}
}

func TestSortList_InPlace(t *testing.T) {
	nodes := []*Node{n("b", 2, 2), n("a", 1, 1)}
	orig := nodes
	SortList(nodes, false, false, false)
	if paths(orig) != nil && paths(orig)[0] != "a" {
		t.Error("SortList should sort in place")
	}
}

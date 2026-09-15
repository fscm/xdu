// -*- mode: go; coding: utf-8 -*-
//
// SPDX-FileCopyrightText: 2026 Frederico Martins
// SPDX-License-Identifier: GPL-3.0-only

package xdu

import (
	"math"
	"strings"
	"testing"
)

func node(
	path string,
	name string,
	isDir bool,
	size int64,
	inodes int64,
	children ...*Node,
) *Node {
	return &Node{
		Path:     path,
		Name:     name,
		IsDir:    isDir,
		Size:     size,
		Inodes:   inodes,
		Children: children,
	}
}

func TestFlattenDirs(t *testing.T) {
	if got := flattenDirs(nil, -1); len(got) != 0 {
		t.Fatalf("nil: got %v want 0", got)
	}
	if got := flattenDirs([]*Node{}, -1); len(got) != 0 {
		t.Fatalf("empty: got %v", got)
	}
	file := node("/a/file", "file", false, 10, 1)
	if got := flattenDirs([]*Node{file}, 0); len(got) != 1 || got[0] != file {
		t.Fatalf("file root maxDepth 0: got %v", got)
	}
	dir := node("/root", "root", true, 0, 1)
	if got := flattenDirs([]*Node{dir}, -1); len(got) != 1 || got[0] != dir {
		t.Fatalf("empty dir: got %v", got)
	}
	subdir := node("/root/sub", "sub", true, 5, 2)
	fileChild := node("/root/file", "file", false, 10, 1)
	dir.Children = []*Node{fileChild, subdir}
	got := flattenDirs([]*Node{dir}, -1)
	if len(got) != 2 {
		t.Fatalf(
			"dir with file+subdir: got %d want 2, %v",
			len(got),
			pathsOf(got),
		)
	}
	if got[0] != dir || got[1] != subdir {
		t.Fatalf("got %v want [dir subdir]", pathsOf(got))
	}
	a := node("/root/a", "a", true, 0, 1)
	b := node("/root/a/b", "b", true, 0, 1)
	c := node("/root/a/b/c", "c", true, 0, 1)
	a.Children = []*Node{b}
	b.Children = []*Node{c}
	root := node("/root", "root", true, 0, 1, a)
	got = flattenDirs([]*Node{root}, -1)
	if len(got) != 4 {
		t.Fatalf("nested unlimited: got %v want 4", pathsOf(got))
	}
	got = flattenDirs([]*Node{root}, 0)
	if len(got) != 1 || got[0] != root {
		t.Fatalf("maxDepth 0: got %v want [root]", pathsOf(got))
	}
	got = flattenDirs([]*Node{root}, 1)
	if len(got) != 2 {
		t.Fatalf("maxDepth 1: got %v want 2", pathsOf(got))
	}
	if got[0] != root || got[1] != a {
		t.Fatalf("maxDepth1 got %v", pathsOf(got))
	}
	got = flattenDirs([]*Node{root}, 2)
	if len(got) != 3 {
		t.Fatalf("maxDepth2 got %v", pathsOf(got))
	}
	got = flattenDirs([]*Node{file, root}, -1)
	if len(got) != 5 {
		t.Fatalf("multi roots: got %v want 5", pathsOf(got))
	}
	if got[0] != file {
		t.Errorf("multi roots order: first should be file")
	}
	got = flattenDirs([]*Node{file, root}, 0)
	if len(got) != 2 {
		t.Fatalf("multi maxDepth0: got %v want 2", pathsOf(got))
	}
}

func pathsOf(nodes []*Node) []string {
	out := make([]string, len(nodes))
	for i, n := range nodes {
		out[i] = n.Path
	}
	return out
}

func TestWriteEntry(t *testing.T) {
	var sb strings.Builder
	n := node("/tmp/foo", "foo", false, 2048, 5)
	opts := Options{HumanReadable: true, Inodes: false}
	writeEntry(&sb, n, "display", opts, false)
	want := FormatUsage(n, true, false) + "\t" + "display" + "\n"
	if sb.String() != want {
		t.Errorf("writeEntry no color: got %q want %q", sb.String(), want)
	}
	sb.Reset()
	writeEntry(&sb, n, "display", opts, true)
	want = FormatUsage(n, true, false) + "\t" + ColorizeName(n) + "\n"
	if sb.String() != want {
		t.Errorf("writeEntry color: got %q want %q", sb.String(), want)
	}
	sb.Reset()
	opts.Inodes = true
	writeEntry(&sb, n, "x", opts, false)
	want = FormatUsage(n, false, true) + "\tx\n"
	if sb.String() != want {
		t.Errorf("inodes: got %q want %q", sb.String(), want)
	}
	sb.Reset()
	dir := node("/dir", "dir", true, 0, 1)
	writeEntry(&sb, dir, "ignored", Options{}, true)
	if sb.String() !=
		FormatUsage(dir, false, false)+"\t"+ansiBlue+"/dir"+ansiReset+"\n" {
		t.Errorf(
			"color should use node.Path not displayName: got %q",
			sb.String(),
		)
	}
}

func TestWriteList(t *testing.T) {
	var sb strings.Builder
	WriteList(&sb, nil, Options{MaxDepth: -1}, false)
	if sb.String() != "" {
		t.Errorf("empty WriteList: got %q", sb.String())
	}
	sb.Reset()
	file := node("/a", "a", false, 100, 2)
	WriteList(&sb, []*Node{file}, Options{MaxDepth: -1}, false)
	want := FormatUsage(file, false, false) + "\t" + "/a\n"
	if sb.String() != want {
		t.Errorf("single file: got %q want %q", sb.String(), want)
	}
	sb.Reset()
	dir1 := node("/b", "b", true, 10, 2)
	dir2 := node("/a", "a", true, 20, 3)
	sub := node("/b/sub", "sub", true, 5, 1)
	dir1.Children = []*Node{sub}
	fileInDir := node("/a/file", "file", false, 99, 1)
	dir2.Children = []*Node{fileInDir}
	nodes := []*Node{dir1, dir2}
	opts := Options{MaxDepth: -1, Sort: false}
	WriteList(&sb, nodes, opts, false)
	wantPaths := []string{"/a", "/b", "/b/sub"}
	paths := flattenDirs(nodes, -1)
	SortList(paths, false, false, false)
	var exp strings.Builder
	for _, p := range paths {
		exp.WriteString(FormatUsage(p, false, false))
		exp.WriteString("\t")
		exp.WriteString(p.Path)
		exp.WriteString("\n")
	}
	if sb.String() != exp.String() {
		t.Errorf(
			"WriteList lexical: got %q want %q, wantPaths %v",
			sb.String(),
			exp.String(),
			wantPaths,
		)
	}
	sb.Reset()
	opts = Options{MaxDepth: -1, Sort: true, Reverse: true, Inodes: false}
	dir1.Size = 10
	dir2.Size = 50
	sub.Size = 5
	WriteList(&sb, nodes, opts, false)
	paths = flattenDirs(nodes, -1)
	SortList(paths, true, true, false)
	exp.Reset()
	for _, p := range paths {
		exp.WriteString(FormatUsage(p, false, false))
		exp.WriteString("\t")
		exp.WriteString(p.Path)
		exp.WriteString("\n")
	}
	if sb.String() != exp.String() {
		t.Errorf(
			"WriteList bySize reverse: got %q want %q",
			sb.String(),
			exp.String(),
		)
	}
	sb.Reset()
	opts = Options{MaxDepth: 0, Sort: false}
	WriteList(&sb, nodes, opts, false)
	paths = flattenDirs(nodes, 0)
	if !strings.Contains(sb.String(), "/a") ||
		!strings.Contains(sb.String(), "/b") {
		t.Errorf("maxDepth 0: got %q", sb.String())
	}
	if strings.Contains(sb.String(), "/b/sub") {
		t.Errorf("maxDepth 0 should not contain sub: %q", sb.String())
	}
	sb.Reset()
	opts = Options{MaxDepth: -1}
	WriteList(&sb, []*Node{node("/dir", "dir", true, 0, 1)}, opts, true)
	if !strings.Contains(sb.String(), ansiBlue) {
		t.Errorf("color true should contain ansiBlue: %q", sb.String())
	}
	sb.Reset()
	WriteList(&sb, []*Node{node("/plain", "plain", false, 0, 1)}, opts, true)
	if strings.Contains(sb.String(), ansiBlue) ||
		strings.Contains(sb.String(), ansiCyan) ||
		strings.Contains(sb.String(), ansiGreen) {
		t.Errorf(
			"plain file with color should not contain ansi: %q",
			sb.String(),
		)
	}
	sb.Reset()
	file.Size = 999
	file.Inodes = 42
	opts = Options{MaxDepth: -1, Inodes: true}
	WriteList(&sb, []*Node{file}, opts, false)
	if !strings.Contains(sb.String(), "42") {
		t.Errorf("inodes mode: got %q want 42", sb.String())
	}
	sb.Reset()
	file.Size = 1536
	opts = Options{MaxDepth: -1, HumanReadable: true}
	WriteList(&sb, []*Node{file}, opts, false)
	if !strings.Contains(sb.String(), "1.5 KiB") {
		t.Errorf("human: got %q want 1.5 KiB", sb.String())
	}
}

func TestWriteTree(t *testing.T) {
	var sb strings.Builder
	WriteTree(&sb, nil, Options{}, false)
	if sb.String() != "" {
		t.Errorf("empty WriteTree: got %q", sb.String())
	}
	sb.Reset()
	root := node("/root", "root", true, 100, 3)
	WriteTree(&sb, []*Node{root}, Options{MaxDepth: -1}, false)
	want := FormatUsage(root, false, false) + "\t" + "/root\n"
	if sb.String() != want {
		t.Errorf("single root: got %q want %q", sb.String(), want)
	}
	sb.Reset()
	child1 := node("/root/b", "b", false, 10, 1)
	child2 := node("/root/a", "a", false, 20, 1)
	child3 := node(
		"/root/c",
		"c",
		true,
		30,
		2,
		node("/root/c/sub", "sub", false, 5, 1),
	)
	root.Children = []*Node{child1, child2, child3}
	opts := Options{MaxDepth: -1, Sort: false}
	WriteTree(&sb, []*Node{root}, opts, false)
	out := sb.String()
	if !strings.HasPrefix(out, FormatUsage(root, false, false)+"\t/root\n") {
		t.Errorf("root line: %q", out)
	}
	if !strings.Contains(out, "├── ") || !strings.Contains(out, "└── ") {
		t.Errorf("branch glyphs missing: %q", out)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 4 {
		t.Fatalf("lines %v", lines)
	}
	idxA := strings.Index(out, "\ta\n")
	idxB := strings.Index(out, "\tb\n")
	idxC := strings.Index(out, "`?") // not needed
	_ = idxC
	if idxA == -1 || idxB == -1 || idxA > idxB {
		t.Errorf(
			"lexical order a before b: idxA=%d idxB=%d out=%q",
			idxA,
			idxB,
			out,
		)
	}
	sb.Reset()
	opts = Options{MaxDepth: -1, Sort: true, Reverse: false}
	WriteTree(&sb, []*Node{root}, opts, false)
	out = sb.String()
	idxB = strings.Index(out, "\tb\n")
	idxA = strings.Index(out, "\ta\n")
	if idxB == -1 || idxA == -1 || idxB > idxA {
		t.Errorf(
			"bySize asc b before a: idxB=%d idxA=%d out=%q",
			idxB,
			idxA,
			out,
		)
	}
	sb.Reset()
	opts = Options{MaxDepth: -1, Sort: true, Reverse: true}
	WriteTree(&sb, []*Node{root}, opts, false)
	out = sb.String()
	idxA = strings.Index(out, "\ta\n")
	idxB = strings.Index(out, "\tb\n")
	idxC = strings.Index(out, "\tc\n")
	if !(idxC < idxA && idxA < idxB) {
		t.Errorf(
			"bySize desc order c<a<b: c=%d a=%d b=%d out=%q",
			idxC,
			idxA,
			idxB,
			out,
		)
	}
	sb.Reset()
	opts = Options{MaxDepth: -1}
	WriteTree(&sb, []*Node{root}, opts, true)
	if !strings.Contains(sb.String(), ansiBlue) {
		t.Errorf("color tree should contain ansiBlue: %q", sb.String())
	}
	sb.Reset()
	plainRoot := node("/plain", "plain", false, 10, 1)
	WriteTree(&sb, []*Node{plainRoot}, Options{}, true)
	if strings.Contains(sb.String(), ansiBlue) {
		t.Errorf("plain root with color should not be blue: %q", sb.String())
	}
	sb.Reset()
	r1 := node("/r1", "r1", true, 10, 1)
	r2 := node("/r2", "r2", true, 20, 1)
	WriteTree(&sb, []*Node{r1, r2}, Options{MaxDepth: -1}, false)
	out = sb.String()
	parts := strings.Split(strings.TrimSpace(out), "\n")
	foundR1, foundR2 := false, false
	for _, l := range parts {
		if strings.Contains(l, "/r1") {
			foundR1 = true
		}
		if strings.Contains(l, "/r2") {
			foundR2 = true
		}
	}
	if !foundR1 || !foundR2 {
		t.Errorf("multi roots: got %q", out)
	}
	sb.Reset()
	r1.Children = []*Node{node("/r1/a", "a", false, 1, 1)}
	r2.Children = []*Node{node("/r2/b", "b", false, 1, 1)}
	WriteTree(&sb, []*Node{r1, r2}, Options{MaxDepth: -1}, false)
	out = sb.String()
	if strings.Count(out, "/r1") != 1 || strings.Count(out, "/r2") != 1 {
		t.Errorf("multi roots counts: %q", out)
	}
	sb.Reset()
	deepLeaf := node("/root/a/b/c", "c", false, 1, 1)
	midC := node("/root/a/b", "b", true, 1, 1, deepLeaf)
	midB := node("/root/a", "a", true, 1, 1, midC)
	rootDeep := node("/root", "root", true, 1, 1, midB)
	opts = Options{MaxDepth: 1}
	WriteTree(&sb, []*Node{rootDeep}, opts, false)
	out = sb.String()
	if !strings.Contains(out, "/root") {
		t.Error("maxDepth 1: should contain root")
	}
	if !strings.Contains(out, "a") {
		t.Error("maxDepth 1: should contain a (depth1)")
	}
	if strings.Contains(out, "b") || strings.Contains(out, "c") {
		t.Errorf("maxDepth 1: should not contain b/c: %q", out)
	}
	sb.Reset()
	opts = Options{MaxDepth: 0}
	WriteTree(&sb, []*Node{rootDeep}, opts, false)
	out = sb.String()
	if strings.Contains(out, "a") {
		t.Errorf("maxDepth 0: should not contain children: %q", out)
	}
	sb.Reset()
	root.Size = 2048
	root.Inodes = 99
	opts = Options{MaxDepth: -1, HumanReadable: true, Inodes: false}
	WriteTree(&sb, []*Node{root}, opts, false)
	if !strings.Contains(sb.String(), "2.0 KiB") {
		t.Errorf("human readable: got %q", sb.String())
	}
	sb.Reset()
	opts = Options{MaxDepth: -1, Inodes: true}
	WriteTree(&sb, []*Node{root}, opts, false)
	if !strings.Contains(sb.String(), "99") {
		t.Errorf("inodes: got %q", sb.String())
	}
	sb.Reset()
	execNode := node("/exec", "exec", false, 10, 1)
	execNode.IsExec = true
	linkNode := node("/link", "link", false, 10, 1)
	linkNode.IsLink = true
	root.Children = []*Node{execNode, linkNode}
	WriteTree(&sb, []*Node{root}, Options{MaxDepth: -1}, true)
	out = sb.String()
	if !strings.Contains(out, ansiGreen) {
		t.Errorf("exec color green missing: %q", out)
	}
	if !strings.Contains(out, ansiCyan) {
		t.Errorf("link color cyan missing: %q", out)
	}
}

func TestWriteChildren_Prefix(t *testing.T) {
	root := node(
		"/root", "root", true, 0, 1,
		node(
			"/root/a", "a", true, 0, 1,
			node("/root/a/1", "1", false, 1, 1),
			node("/root/a/2", "2", false, 1, 1),
		),
		node(
			"/root/b", "b", true, 0, 1,
			node("/root/b/3", "3", false, 1, 1),
		),
	)
	var sb strings.Builder
	WriteTree(&sb, []*Node{root}, Options{MaxDepth: -1, Sort: false}, false)
	out := sb.String()
	if !strings.Contains(out, "├── ") {
		t.Error("missing ├──")
	}
	if !strings.Contains(out, "└── ") {
		t.Error("missing └──")
	}
	if !strings.Contains(out, "│") {
		t.Errorf("missing │ continuation: %q", out)
	}
	if !strings.Contains(out, "1") || !strings.Contains(out, "2") {
		t.Errorf("missing nested files: %q", out)
	}
}

func TestGrandTotal(t *testing.T) {
	if got := GrandTotal(nil, false); got != 0 {
		t.Errorf("nil: got %d want 0", got)
	}
	if got := GrandTotal([]*Node{}, true); got != 0 {
		t.Errorf("empty: got %d", got)
	}
	n1 := node("/a", "a", false, 10, 3)
	n2 := node("/b", "b", false, 20, 5)
	if got := GrandTotal([]*Node{n1, n2}, false); got != 30 {
		t.Errorf("size total: got %d want 30", got)
	}
	if got := GrandTotal([]*Node{n1, n2}, true); got != 8 {
		t.Errorf("inodes total: got %d want 8", got)
	}
	m := &Node{Size: math.MaxInt64, Inodes: math.MaxInt64}
	if got := GrandTotal([]*Node{m, m}, false); got != math.MaxInt64 {
		t.Errorf("saturating size: got %d want MaxInt64", got)
	}
	if got := GrandTotal([]*Node{m, m}, true); got != math.MaxInt64 {
		t.Errorf("saturating inodes: got %d", got)
	}
	if got := GrandTotal([]*Node{n1}, false); got != 10 {
		t.Errorf("single size: got %d", got)
	}
}

func TestFlattenDirs_Coverage(t *testing.T) {
	subfile := node("/root/sub/file", "file", false, 1, 1)
	sub := node("/root/sub", "sub", true, 0, 1, subfile)
	root := node("/root", "root", true, 0, 1, sub)
	got := flattenDirs([]*Node{root}, -1)
	for _, n := range got {
		if n.Path == "/root/sub/file" {
			t.Error("file inside subdir should not be in flattenDirs")
		}
	}
	if len(got) != 2 {
		t.Errorf("got %v want 2", pathsOf(got))
	}
}

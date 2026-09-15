// -*- mode: go; coding: utf-8 -*-
//
// SPDX-FileCopyrightText: 2026 Frederico Martins
// SPDX-License-Identifier: GPL-3.0-only

package xdu

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestSaturatingAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b int64
		want int64
	}{
		{"zero", 0, 0, 0},
		{"simple", 1, 2, 3},
		{"max+0", math.MaxInt64, 0, math.MaxInt64},
		{"0+max", 0, math.MaxInt64, math.MaxInt64},
		{"just below overflow", math.MaxInt64 - 1, 1, math.MaxInt64},
		{"overflow by 1", math.MaxInt64, 1, math.MaxInt64},
		{"overflow large", math.MaxInt64 - 5, 10, math.MaxInt64},
		{"negative b", 10, -5, 5},
		{"negative a", -5, 10, 5},
		{"both negative", -5, -10, -15},
		{"a max b negative", math.MaxInt64, -1, math.MaxInt64 - 1},
	}
	for _, tc := range tests {
		got := saturatingAdd(tc.a, tc.b)
		if got != tc.want {
			t.Errorf(
				"%s: saturatingAdd(%d,%d)=%d want %d",
				tc.name,
				tc.a,
				tc.b,
				got,
				tc.want,
			)
		}
	}
}

func TestScanPaths_Empty(t *testing.T) {
	res := ScanPaths(nil, false)
	if len(res.Paths) != 0 {
		t.Fatalf("nil paths: got %d paths want 0", len(res.Paths))
	}
	if len(res.Errors) != 0 {
		t.Fatalf("nil paths: got %d errors want 0", len(res.Errors))
	}
	res = ScanPaths([]string{}, false)
	if len(res.Paths) != 0 || len(res.Errors) != 0 {
		t.Fatalf(
			"empty slice: paths=%d errors=%d want 0,0",
			len(res.Paths),
			len(res.Errors),
		)
	}
}

func TestScanPaths_Nonexistent(t *testing.T) {
	tDir := t.TempDir()
	missing := filepath.Join(tDir, "no-such-file-12345")
	res := ScanPaths([]string{missing}, false)
	if len(res.Paths) != 0 {
		t.Fatalf("nonexistent: paths=%d want 0", len(res.Paths))
	}
	if len(res.Errors) != 1 {
		t.Fatalf("nonexistent: errors=%d want 1", len(res.Errors))
	}
	if !strings.Contains(res.Errors[0].Error(), missing) {
		t.Errorf(
			"error %q should contain path %q",
			res.Errors[0].Error(),
			missing,
		)
	}
}

func TestScanPaths_SingleFile(t *testing.T) {
	tDir := t.TempDir()
	fPath := filepath.Join(tDir, "hello.txt")
	content := []byte("hello world")
	if err := os.WriteFile(fPath, content, 0o644); err != nil {
		t.Fatal(err)
	}
	res := ScanPaths([]string{fPath}, false)
	if len(res.Errors) != 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if len(res.Paths) != 1 {
		t.Fatalf("paths=%d want 1", len(res.Paths))
	}
	n := res.Paths[0]
	if n.Path != fPath {
		t.Errorf("Path=%q want %q", n.Path, fPath)
	}
	if n.Name != "hello.txt" {
		t.Errorf("Name=%q want hello.txt", n.Name)
	}
	if n.IsDir {
		t.Error("IsDir should be false")
	}
	if n.IsLink {
		t.Error("IsLink should be false")
	}
	if n.IsExec {
		t.Error("IsExec should be false for 0644")
	}
	if n.Size != int64(len(content)) {
		t.Errorf("Size=%d want %d", n.Size, len(content))
	}
	if n.Inodes != 1 {
		t.Errorf("Inodes=%d want 1", n.Inodes)
	}
	if len(n.Children) != 0 {
		t.Errorf("Children=%d want 0", len(n.Children))
	}
}

func TestScanPaths_SingleFile_Executable(t *testing.T) {
	tDir := t.TempDir()
	fPath := filepath.Join(tDir, "run.sh")
	if err := os.WriteFile(fPath, []byte("#!/bin/sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	res := ScanPaths([]string{fPath}, false)
	if len(res.Errors) != 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	n := res.Paths[0]
	if !n.IsExec {
		t.Error("IsExec should be true for 0755")
	}
	fPath2 := filepath.Join(tDir, "plain.txt")
	if err := os.WriteFile(fPath2, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	res2 := ScanPaths([]string{fPath2}, false)
	if res2.Paths[0].IsExec {
		t.Error("IsExec should be false for 0644")
	}
	fPath3 := filepath.Join(tDir, "g-exec")
	if err := os.WriteFile(fPath3, []byte("x"), 0o010); err != nil {
		t.Fatal(err)
	}
	res3 := ScanPaths([]string{fPath3}, false)
	if !res3.Paths[0].IsExec {
		t.Error("IsExec should be true for 0010")
	}
}

func TestScanPaths_EmptyDir(t *testing.T) {
	tDir := t.TempDir()
	empty := filepath.Join(tDir, "empty")
	if err := os.Mkdir(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	res := ScanPaths([]string{empty}, false)
	if len(res.Errors) != 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if len(res.Paths) != 1 {
		t.Fatalf("paths=%d want 1", len(res.Paths))
	}
	n := res.Paths[0]
	if !n.IsDir {
		t.Error("IsDir want true")
	}
	if n.IsLink {
		t.Error("IsLink want false")
	}
	if n.Size != 0 {
		t.Errorf("Size=%d want 0", n.Size)
	}
	if n.Inodes != 1 {
		t.Errorf("Inodes=%d want 1", n.Inodes)
	}
	if len(n.Children) != 0 {
		t.Errorf("Children=%d want 0", len(n.Children))
	}
}

func TestScanPaths_DirWithFiles_SortedAndAggregate(t *testing.T) {
	tDir := t.TempDir()
	root := filepath.Join(tDir, "root")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	files := []struct {
		name string
		size int
	}{
		{"z.txt", 10},
		{"a.txt", 5},
		{"m.txt", 20},
	}
	var total int64
	for _, f := range files {
		p := filepath.Join(root, f.name)
		if err := os.WriteFile(p, make([]byte, f.size), 0o644); err != nil {
			t.Fatal(err)
		}
		total += int64(f.size)
	}
	res := ScanPaths([]string{root}, false)
	if len(res.Errors) != 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	n := res.Paths[0]
	if n.Size != total {
		t.Errorf("Size=%d want %d", n.Size, total)
	}
	if n.Inodes != 1+int64(len(files)) {
		t.Errorf("Inodes=%d want %d", n.Inodes, 1+len(files))
	}
	if len(n.Children) != len(files) {
		t.Fatalf("Children=%d want %d", len(n.Children), len(files))
	}
	names := []string{"a.txt", "m.txt", "z.txt"}
	for i, want := range names {
		if n.Children[i].Name != want {
			t.Errorf(
				"Children[%d].Name=%q want %q",
				i,
				n.Children[i].Name,
				want,
			)
		}
	}
	sorted := map[string]int{"a.txt": 5, "m.txt": 20, "z.txt": 10}
	for _, ch := range n.Children {
		if ch.Size != int64(sorted[ch.Name]) {
			t.Errorf("%s Size=%d want %d", ch.Name, ch.Size, sorted[ch.Name])
		}
	}
}

func TestScanPaths_HiddenFiles(t *testing.T) {
	tDir := t.TempDir()
	root := filepath.Join(tDir, "hiddendir")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".hidden"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "visible"), []byte("pub"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".hiddendir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".hiddendir", "inside"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := ScanPaths([]string{root}, false)
	n := res.Paths[0]
	if len(n.Children) != 1 {
		t.Fatalf(
			"without hidden: Children=%d want 1 (visible)",
			len(n.Children),
		)
	}
	if n.Children[0].Name != "visible" {
		t.Errorf("without hidden: got %q want visible", n.Children[0].Name)
	}
	if n.Size != 3 {
		t.Errorf("without hidden: Size=%d want 3", n.Size)
	}
	if n.Inodes != 2 {
		t.Errorf("without hidden: Inodes=%d want 2", n.Inodes)
	}
	res2 := ScanPaths([]string{root}, true)
	n2 := res2.Paths[0]
	if len(n2.Children) != 3 {
		t.Fatalf("with hidden: Children=%d want 3", len(n2.Children))
	}
	wantNames := []string{".hidden", ".hiddendir", "visible"}
	for i, w := range wantNames {
		if n2.Children[i].Name != w {
			t.Errorf(
				"with hidden: Children[%d]=%q want %q",
				i,
				n2.Children[i].Name,
				w,
			)
		}
	}
	var hiddenDir *Node
	for _, ch := range n2.Children {
		if ch.Name == ".hiddendir" {
			hiddenDir = ch
		}
	}
	if hiddenDir == nil {
		t.Fatal("missing .hiddendir")
	}
	if len(hiddenDir.Children) != 1 || hiddenDir.Children[0].Name != "inside" {
		t.Errorf(".hiddendir children=%v want [inside]", hiddenDir.Children)
	}
	if n2.Inodes != 1+1+2+1 { // root + .hidden + .hiddendir(+inside) + visible
		t.Errorf("with hidden: Inodes=%d want 5", n2.Inodes)
	}
}

func TestScanPaths_Symlink(t *testing.T) {
	tDir := t.TempDir()
	target := filepath.Join(tDir, "target.txt")
	if err := os.WriteFile(target, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(tDir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symlink not supported", err)
	}
	res := ScanPaths([]string{link}, false)
	if len(res.Errors) != 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	n := res.Paths[0]
	if !n.IsLink {
		t.Error("IsLink want true")
	}
	if n.IsDir {
		t.Error("IsDir want false for symlink to file")
	}
	if n.IsExec {
		t.Error("IsExec want false for symlink (even if target exec)")
	}
	if n.Inodes != 1 {
		t.Errorf("Inodes=%d want 1", n.Inodes)
	}
	if len(n.Children) != 0 {
		t.Errorf("Children=%d want 0", len(n.Children))
	}
	dirTarget := filepath.Join(tDir, "realdir")
	if err := os.Mkdir(dirTarget, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dirTarget, "file"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(tDir, "linkdir")
	if err := os.Symlink(dirTarget, linkDir); err != nil {
		t.Fatal(err)
	}
	res2 := ScanPaths([]string{linkDir}, false)
	n2 := res2.Paths[0]
	if !n2.IsLink {
		t.Error("link to dir: IsLink want true")
	}
	if n2.IsDir {
		t.Error("link to dir: IsDir want false (don't follow)")
	}
	if len(n2.Children) != 0 {
		t.Errorf("link to dir: Children=%d want 0", len(n2.Children))
	}
	root := filepath.Join(tDir, "root2")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	linkInside := filepath.Join(root, "link_inside")
	if err := os.Symlink(target, linkInside); err != nil {
		t.Fatal(err)
	}
	res3 := ScanPaths([]string{root}, false)
	n3 := res3.Paths[0]
	if len(n3.Children) != 2 {
		t.Fatalf("dir with symlink: Children=%d want 2", len(n3.Children))
	}
	var linkChild *Node
	for _, ch := range n3.Children {
		if ch.Name == "link_inside" {
			linkChild = ch
		}
	}
	if linkChild == nil {
		t.Fatal("missing link_inside")
	}
	if !linkChild.IsLink {
		t.Error("link_inside: IsLink want true")
	}
	expected := n3.Children[0].Size + n3.Children[1].Size
	if n3.Size != expected {
		t.Errorf("root Size=%d want %d", n3.Size, expected)
	}
}

func TestScanPaths_ExecutableSymlinkNotExec(t *testing.T) {
	tDir := t.TempDir()
	target := filepath.Join(tDir, "exec_target")
	if err := os.WriteFile(target, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(tDir, "exec_link")
	if err := os.Symlink(target, link); err != nil {
		t.Skip(err)
	}
	res := ScanPaths([]string{link}, false)
	n := res.Paths[0]
	if n.IsExec {
		t.Error("symlink to exec should not be IsExec")
	}
}

func TestScanPaths_NestedDirs(t *testing.T) {
	tDir := t.TempDir()
	root := filepath.Join(tDir, "nested")
	if err := os.MkdirAll(filepath.Join(root, "a", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a", "b", "deep.txt"), []byte("deep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a", "file.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := ScanPaths([]string{root}, false)
	if len(res.Errors) != 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	n := res.Paths[0]
	if len(n.Children) != 1 || n.Children[0].Name != "a" {
		t.Fatalf("root children=%v want [a]", n.Children)
	}
	a := n.Children[0]
	if len(a.Children) != 2 {
		t.Fatalf("a children=%d want 2", len(a.Children))
	}
	if a.Children[0].Name != "b" || a.Children[1].Name != "file.txt" {
		t.Errorf(
			"a children names=%v want [b file.txt]",
			[]string{a.Children[0].Name, a.Children[1].Name},
		)
	}
	if a.Size != 6 {
		t.Errorf("a Size=%d want 6", a.Size)
	}
	if n.Size != 6 {
		t.Errorf("root Size=%d want 6", n.Size)
	}
	if n.Inodes != 5 {
		t.Errorf("root Inodes=%d want 5", n.Inodes)
	}
}

func TestScanPaths_MultiplePaths(t *testing.T) {
	tDir := t.TempDir()
	p1 := filepath.Join(tDir, "p1.txt")
	p2 := filepath.Join(tDir, "p2.txt")
	if err := os.WriteFile(p1, []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p2, []byte("22"), 0o644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(tDir, "missing")
	res := ScanPaths([]string{p1, missing, p2}, false)
	if len(res.Paths) != 2 {
		t.Fatalf("Paths=%d want 2", len(res.Paths))
	}
	if len(res.Errors) != 1 {
		t.Fatalf("Errors=%d want 1", len(res.Errors))
	}
	if !strings.Contains(res.Errors[0].Error(), missing) {
		t.Errorf("error %q should contain %q", res.Errors[0].Error(), missing)
	}
	if res.Paths[0].Path != p1 || res.Paths[1].Path != p2 {
		t.Errorf("Paths=%v want [%q %q]", res.Paths, p1, p2)
	}
}

func TestScanPaths_SortDeterministic(t *testing.T) {
	tDir := t.TempDir()
	root := filepath.Join(tDir, "sorttest")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	names := []string{"z", "a", "m", "0", "B"}
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(root, n), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	res := ScanPaths([]string{root}, false)
	sort.Strings(names)
	for i, want := range names {
		if res.Paths[0].Children[i].Name != want {
			t.Errorf(
				"Children[%d]=%q want %q",
				i,
				res.Paths[0].Children[i].Name,
				want,
			)
		}
	}
}

func TestScanPaths_ReadDirError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod not reliable on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root, permission denied test would not fail")
	}
	tDir := t.TempDir()
	root := filepath.Join(tDir, "no_read")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(root, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(root, 0o755) })
	res := ScanPaths([]string{root}, false)
	if len(res.Paths) != 1 {
		t.Fatalf("Paths=%d want 1", len(res.Paths))
	}
	n := res.Paths[0]
	if !n.IsDir {
		t.Error("IsDir want true")
	}
	if len(n.Children) != 0 {
		t.Errorf("Children=%d want 0 for unreadable dir", len(n.Children))
	}
	if len(res.Errors) == 0 {
		t.Error("want error for unreadable dir")
	} else if !strings.Contains(res.Errors[0].Error(), root) {
		t.Errorf("error %q should contain %q", res.Errors[0].Error(), root)
	}
	if n.Inodes != 1 {
		t.Errorf("Inodes=%d want 1", n.Inodes)
	}
	if n.Size != 0 {
		t.Errorf("Size=%d want 0", n.Size)
	}
}

func TestScanPaths_SaturatingAddAggregation(t *testing.T) {
	a := int64(math.MaxInt64 - 10)
	b := int64(20)
	if got := saturatingAdd(a, b); got != math.MaxInt64 {
		t.Errorf("saturatingAdd(%d,%d)=%d want %d", a, b, got, math.MaxInt64)
	}
	tDir := t.TempDir()
	root := filepath.Join(tDir, "sat")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := range 2 {
		p := filepath.Join(root, string(rune('a'+i)))
		if err := os.WriteFile(p, []byte("12345"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	res := ScanPaths([]string{root}, false)
	if res.Paths[0].Size != 10 {
		t.Errorf("Size=%d want 10", res.Paths[0].Size)
	}
}

func TestScanPaths_PathIsDirWithHiddenInside(t *testing.T) {
	tDir := t.TempDir()
	root := filepath.Join(tDir, "hiddenDirTest")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	hiddenDir := filepath.Join(root, ".hiddenDir")
	if err := os.Mkdir(hiddenDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := ScanPaths([]string{root}, false)
	if len(res.Paths[0].Children) != 0 {
		t.Errorf(
			"without hidden: Children=%d want 0",
			len(res.Paths[0].Children),
		)
	}
	res2 := ScanPaths([]string{root}, true)
	if len(res2.Paths[0].Children) != 1 {
		t.Fatalf("with hidden: Children=%d want 1", len(res2.Paths[0].Children))
	}
	if res2.Paths[0].Children[0].Name != ".hiddenDir" {
		t.Errorf("Name=%q want .hiddenDir", res2.Paths[0].Children[0].Name)
	}
}

func TestScanPaths_DirectFileHiddenName(t *testing.T) {
	tDir := t.TempDir()
	hiddenFile := filepath.Join(tDir, ".hiddenFile")
	if err := os.WriteFile(hiddenFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, showHidden := range []bool{true, false} {
		res := ScanPaths([]string{hiddenFile}, showHidden)
		if len(res.Paths) != 1 {
			t.Fatalf(
				"showHidden=%v: Paths=%d want 1",
				showHidden,
				len(res.Paths),
			)
		}
		if res.Paths[0].Name != ".hiddenFile" {
			t.Errorf("Name=%q want .hiddenFile", res.Paths[0].Name)
		}
		if len(res.Errors) != 0 {
			t.Errorf("showHidden=%v: errors=%v want 0", showHidden, res.Errors)
		}
	}
}

func TestScanPaths_BrokenSymlinkInside(t *testing.T) {
	tDir := t.TempDir()
	root := filepath.Join(tDir, "brokenRoot")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(tDir, "nonexistent-target")
	link := filepath.Join(root, "broken")
	if err := os.Symlink(target, link); err != nil {
		t.Skip(err)
	}
	res := ScanPaths([]string{root}, false)
	if len(res.Errors) != 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if len(res.Paths[0].Children) != 1 {
		t.Fatalf("Children=%d want 1", len(res.Paths[0].Children))
	}
	ch := res.Paths[0].Children[0]
	if !ch.IsLink {
		t.Error("IsLink want true for broken symlink")
	}
	if ch.IsDir {
		t.Error("IsDir want false for broken symlink")
	}
}

func TestScanPath_LstatError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "definitely-missing-xyz")
	res := ScanPaths([]string{missing}, false)
	if len(res.Paths) != 0 {
		t.Error("want no paths for missing")
	}
	if len(res.Errors) != 1 {
		t.Fatalf("want 1 error")
	}
	if !strings.Contains(res.Errors[0].Error(), missing) {
		t.Errorf("error %q want contain %q", res.Errors[0].Error(), missing)
	}
	node, errs := scanPath(missing, false)
	if node != nil {
		t.Error("node should be nil for missing")
	}
	if len(errs) != 1 {
		t.Fatalf("errs=%d want 1", len(errs))
	}
}

func TestBuildNode_ReadDirErrorDirect(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod not reliable on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root can still read")
	}
	tDir := t.TempDir()
	dir := filepath.Join(tDir, "errdir")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })
	info, err := os.Lstat(dir)
	if err != nil {
		t.Fatal(err)
	}
	var errs []error
	n := buildNode(dir, info, false, &errs)
	if !n.IsDir {
		t.Error("IsDir want true")
	}
	if len(n.Children) != 0 {
		t.Error("Children want 0")
	}
	if len(errs) == 0 {
		t.Error("want error for unreadable ReadDir")
	}
}

func TestScanPaths_ChildLstatRace(t *testing.T) {
	for iter := range 100 {
		tDir := t.TempDir()
		root := filepath.Join(tDir, "race")
		if err := os.Mkdir(root, 0o755); err != nil {
			t.Fatal(err)
		}
		for i := range 20 {
			p := filepath.Join(root, fmt.Sprintf("f%02d", i))
			if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		done := make(chan struct{})
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					_ = os.Remove(filepath.Join(root, "f10"))
				}
			}
		}()
		res := ScanPaths([]string{root}, false)
		close(done)
		for _, e := range res.Errors {
			if !strings.Contains(e.Error(), root) {
				t.Errorf("race error %q should contain root", e)
			}
		}
		if len(res.Errors) > 0 {
			t.Logf("child Lstat race hit iter %d: %v", iter, res.Errors)
			return
		}
	}
	t.Log("child Lstat race branch not hit (expected, it's timing-dependent)")
}

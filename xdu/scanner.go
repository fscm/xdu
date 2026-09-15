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
	"sort"
	"strings"
)

// Node  is  one  filesystem  entry.  For directories Size/Inodes aggregate all
// included descendants. Children holds direct included children.
type Node struct {
	Path     string
	Name     string
	Children []*Node
	Size     int64 // logical bytes (aggregate for dirs)
	Inodes   int64 // entry count, self included (aggregate for dirs)
	IsDir    bool
	IsLink   bool
	IsExec   bool
}

// Result  groups  scanned paths with accumulated filesystem errors. Errors are
// kept so partial output can still be rendered.
type Result struct {
	Paths  []*Node
	Errors []error
}

// ScanPaths  scans  all  given  paths and returns a Result containing the root
// nodes and any errors encountered.
func ScanPaths(paths []string, showHidden bool) Result {
	result := Result{}
	for _, path := range paths {
		node, errs := scanPath(path, showHidden)
		if node != nil {
			result.Paths = append(result.Paths, node)
		}
		result.Errors = append(result.Errors, errs...)
	}
	return result
}

// scanPath performs an Lstat on the path and builds the node tree.
func scanPath(path string, showHidden bool) (*Node, []error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, []error{fmt.Errorf("%s: %w", path, err)}
	}
	var errs []error
	node := buildNode(path, info, showHidden, &errs)
	return node, errs
}

// buildNode recursively builds a Node for the given file/directory.
func buildNode(
	path string, info os.FileInfo, showHidden bool, errs *[]error,
) *Node {
	mode := info.Mode()
	isLink := mode&os.ModeSymlink != 0
	isDir := info.IsDir() && !isLink
	base := filepath.Base(path)
	node := &Node{
		Path:   path,
		Name:   base,
		IsDir:  isDir,
		IsLink: isLink,
	}
	if !isDir {
		node.Size = info.Size()
		node.Inodes = 1
		if !isLink && mode.Perm()&0o111 != 0 {
			node.IsExec = true
		}
		return node
	}
	node.Inodes = 1 // Directory: count self, then recursively process children.
	entries, err := os.ReadDir(path)
	if err != nil {
		*errs = append(*errs, fmt.Errorf("%s: %w", path, err))
		return node
	}
	sort.Slice( // Sort entries for deterministic output.
		entries,
		func(a, b int) bool { return entries[a].Name() < entries[b].Name() },
	)
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") && !showHidden {
			continue
		}
		childPath := filepath.Join(path, name)
		childInfo, err := os.Lstat(childPath)
		if err != nil {
			*errs = append(*errs, fmt.Errorf("%s: %w", childPath, err))
			continue
		}
		child := buildNode(childPath, childInfo, showHidden, errs)
		node.Children = append(node.Children, child)
		node.Size = saturatingAdd(node.Size, child.Size)
		node.Inodes = saturatingAdd(node.Inodes, child.Inodes)
	}
	return node
}

// saturatingAdd performs addition with saturation at math.MaxInt64.
func saturatingAdd(a, b int64) int64 {
	if b > 0 && a > math.MaxInt64-b {
		return math.MaxInt64
	}
	return a + b
}

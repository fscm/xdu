// -*- mode: go; coding: utf-8 -*-
//
// SPDX-FileCopyrightText: 2026 Frederico Martins
// SPDX-License-Identifier: GPL-3.0-only

package xdu

import (
	"strings"
)

// flattenDirs returns a flat slice of nodes representing directories (plus any
// non-directory  roots)  up  to the given max depth. A negative maxDepth means
// unlimited depth. Roots are considered to be at depth 0.
func flattenDirs(nodes []*Node, maxDepth int) []*Node {
	paths := make([]*Node, 0, len(nodes))
	var collect func(node *Node, depth int)
	collect = func(node *Node, depth int) {
		if maxDepth >= 0 && depth > maxDepth {
			return
		}
		paths = append(paths, node)
		for _, child := range node.Children {
			if child.IsDir {
				collect(child, depth+1)
			}
		}
	}
	for _, node := range nodes {
		if !node.IsDir {
			paths = append(paths, node)
			continue
		}
		collect(node, 0)
	}
	return paths
}

// writeEntry  writes  a  single  line with the formatted usage, a tab, and the
// (optionally colorized) display name, followed by a newline.
func writeEntry(
	strBuilder *strings.Builder,
	node *Node,
	displayName string,
	options Options,
	color bool,
) {
	strBuilder.WriteString(
		FormatUsage(node, options.HumanReadable, options.Inodes),
	)
	strBuilder.WriteByte('\t')
	if color {
		displayName = ColorizeName(node)
	}
	strBuilder.WriteString(displayName)
	strBuilder.WriteByte('\n')
}

// WriteList  writes  a  flat  listing of directories (and non-directory roots)
// found in nodes, sorted according to the given options.
func WriteList(
	strBuilder *strings.Builder,
	nodes []*Node,
	options Options,
	color bool,
) {
	paths := flattenDirs(nodes, options.MaxDepth)
	SortList(paths, options.Sort, options.Reverse, options.Inodes)
	for _, path := range paths {
		writeEntry(strBuilder, path, path.Path, options, color)
	}
}

// WriteTree writes a tree representation of the given nodes, sorted according
// to the given options.
func WriteTree(
	strBuilder *strings.Builder,
	nodes []*Node,
	options Options,
	color bool,
) {
	for i, node := range nodes {
		if i > 0 {
			strBuilder.WriteByte('\n')
		}
		SortTree(node, options.Sort, options.Reverse, options.Inodes)
		// Root line carries its total so single-root trees show usage.
		writeEntry(strBuilder, node, node.Path, options, color)
		writeChildren(strBuilder, node, "", options, color, 1)
	}
}

// writeChildren  recursively  writes  the  children  of node using tree branch
// characters. Depth starts at 1 for direct children of the root.
func writeChildren(
	strBuilder *strings.Builder,
	node *Node,
	prefix string,
	options Options,
	color bool,
	depth int,
) {
	if options.MaxDepth >= 0 && depth > options.MaxDepth {
		return
	}
	children := node.Children
	for i, child := range children {
		last := i == len(children)-1
		branch, cont := "├── ", "│   "
		if last {
			branch, cont = "└── ", "    "
		}
		strBuilder.WriteString(prefix)
		strBuilder.WriteString(branch)
		writeEntry(strBuilder, child, child.Name, options, color)
		if child.IsDir && len(child.Children) > 0 {
			writeChildren(
				strBuilder,
				child,
				prefix+cont,
				options,
				color,
				depth+1,
			)
		}
	}
}

// GrandTotal  returns  the  sum  of  sizes (or inode counts if inodes is true)
// across all given nodes, saturating at math.MaxInt64 on overflow.
func GrandTotal(nodes []*Node, inodes bool) int64 {
	var total int64
	for _, node := range nodes {
		if inodes {
			total = saturatingAdd(total, node.Inodes)
		} else {
			total = saturatingAdd(total, node.Size)
		}
	}
	return total
}

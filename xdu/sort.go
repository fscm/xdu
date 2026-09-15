package xdu

import (
	"sort"
)

// sortBySize  reports whether node a should sort before node b when sorting by
// size.  It  compares  Size  or Inodes (if inodes is true). Ties are broken by
// Path  (ascending).  If  reverse  is  true,  the  size comparison is inverted
// (descending).
func sortBySize(a, b *Node, reverse, inodes bool) bool {
	var aSize, bSize int64
	if inodes {
		aSize, bSize = a.Inodes, b.Inodes
	} else {
		aSize, bSize = a.Size, b.Size
	}
	if aSize == bSize {
		return a.Path < b.Path
	}
	if reverse {
		return aSize > bSize
	}
	return aSize < bSize
}

// SortList  sorts  a slice of nodes in place using a stable sort. If bySize is
// false, nodes are sorted lexically by Path (reverse is ignored). If bySize is
// true,  nodes  are  sorted  by  Size or Inodes (if inodes is true), with ties
// broken by Path. The reverse flag inverts the size comparison.
func SortList(nodes []*Node, bySize, reverse, inodes bool) {
	if !bySize { // Path-only sort.
		sort.SliceStable(nodes, func(a, b int) bool {
			if !reverse {
				return nodes[a].Path < nodes[b].Path
			}
			return nodes[a].Path > nodes[b].Path
		})
		return
	}
	sort.SliceStable(nodes, func(a, b int) bool {
		return sortBySize(nodes[a], nodes[b], reverse, inodes)
	})
}

// SortTree  recursively  sorts  the  children  of  the  given node and all its
// descendants. It uses the same sorting rules as SortList.
func SortTree(node *Node, bySize, reverse, inodes bool) {
	if node == nil || len(node.Children) == 0 {
		return
	}
	if !bySize { // Path-only sort.
		sort.SliceStable(node.Children, func(a, b int) bool {
			if !reverse {
				return node.Children[a].Path < node.Children[b].Path
			}
			return node.Children[a].Path > node.Children[b].Path
		})
	} else {
		sort.SliceStable(node.Children, func(a, b int) bool {
			return sortBySize(
				node.Children[a],
				node.Children[b],
				reverse,
				inodes,
			)
		})
	}
	for _, c := range node.Children { // Recurse into children.
		SortTree(c, bySize, reverse, inodes)
	}
}

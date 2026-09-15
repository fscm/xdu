// -*- mode: go; coding: utf-8 -*-
//
// SPDX-FileCopyrightText: 2026 Frederico Martins
// SPDX-License-Identifier: GPL-3.0-only

package xdu

import (
	"strconv"
)

// sizeUnits  holds the human-readable size unit suffixes, starting with bytes.
// The  index  corresponds  to  the power of 1024 (0 = bytes, 1 = KiB, ...).
var sizeUnits = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}

// FormatSize  formats  a byte count as a string. If human is false, it returns
// the   raw   integer  as  a  decimal  string.  If  human  is  true,  it  uses
// human-readable  units  (B, KiB, MiB, ...) with one decimal place, except for
// values less than 1024 which are printed as whole bytes (e.g., "512 B").
func FormatSize(bytes int64, human bool) string {
	if !human {
		return strconv.FormatInt(bytes, 10)
	}
	if bytes < 1024 {
		return strconv.FormatInt(bytes, 10) + " B"
	}
	value := float64(bytes)
	unitIndex := 0
	// Divide by 1024 until the value is less than 1024 or we reach the largest
	// unit.
	for value >= 1024 && unitIndex < len(sizeUnits)-1 {
		value /= 1024
		unitIndex++
	}
	// Format with one decimal place (e.g., "1.5 MiB").
	return strconv.FormatFloat(value, 'f', 1, 64) + " " + sizeUnits[unitIndex]
}

// FormatUsage  returns  the formatted size or inode count of a node. If inodes
// is  true,  it returns the node's inode count as a decimal string. Otherwise,
// it returns the node's size formatted according to the human flag.
func FormatUsage(node *Node, human bool, inodes bool) string {
	if inodes {
		return strconv.FormatInt(node.Inodes, 10)
	}
	return FormatSize(node.Size, human)
}

// -*- mode: go; coding: utf-8 -*-
//
// SPDX-FileCopyrightText: 2026 Frederico Martins
// SPDX-License-Identifier: GPL-3.0-only

package xdu

import (
	"os"
)

const (
	ansiReset = "\x1b[0m"
	ansiBold  = "\x1b[1m"
	ansiBlue  = "\x1b[1;34m"
	ansiCyan  = "\x1b[36m"
	ansiGreen = "\x1b[32m"
)

// ColorEnabled  reports whether color output should be used based on the given
// mode.  Mode  "always" returns true, "never" returns false, and "auto" checks
// the  NO_COLOR  environment variable and whether stdout is a character device
// (terminal).
func ColorEnabled(mode string) bool {
	switch mode {
	case "always":
		return true
	case "never":
		return false
	default: // auto
		if os.Getenv("NO_COLOR") != "" {
			return false
		}
		if info, err := os.Stdout.Stat(); err == nil &&
			info.Mode()&os.ModeCharDevice != 0 {
			return true
		}
		return false
	}
}

// ColorizeName  returns  the node path with ANSI color codes based on the node
// type.
func ColorizeName(node *Node) string {
	switch {
	case node.IsDir:
		return ansiBlue + node.Path + ansiReset
	case node.IsLink:
		return ansiCyan + node.Path + ansiReset
	case node.IsExec:
		return ansiGreen + node.Path + ansiReset
	default:
		return node.Path
	}
}

// ColorBold wraps the given string with ANSI bold codes.
func ColorBold(str string) string {
	return ansiBold + str + ansiReset
}

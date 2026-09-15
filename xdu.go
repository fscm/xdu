// -*- mode: go; coding: utf-8 -*-
//
// SPDX-FileCopyrightText: 2026 Frederico Martins
// SPDX-License-Identifier: GPL-3.0-only

/*
xdu

Synopsis:

xdu  merges the behavior of 'du', 'sort', and 'tree' into one static binary. It
walks  one  or  more  paths,  sums  apparent  sizes  or  inode  counts for each
directory, and can show the result as a flat list or as a hierarchical tree.

For  each  path given on the command line (or '.' when no path is given), 'xdu'
calls  'os.Lstat'  and  builds  a  tree  of  nodes.  Directories  are read with
'os.ReadDir'  and  walked  recursively. Each node stores the apparent size from
'Lstat', an inode count, and flags for directory, symlink, and executable.

A  directory  node  aggregates  its  descendants.  In  byte  mode the size of a
directory  is  the sum of the sizes of everything under it. The directory entry
itself  does  not add extra bytes. In inode mode ('-i') a directory counts as 1
plus  the  count  of  all  descendants.  Symlinks are never followed. A symlink
contributes its 'Lstat' size and one inode, even when it points to a directory.
Hard  links  are counted once per directory entry. Permission or I/O errors are
kept as warnings and do not stop the scan of other paths.

After  the  scan  'xdu'  either  prints  a  flat  list  of directories (and any
non-directory  paths)  or renders a tree. Both outputs can be sorted by size or
inode  count,  limited  by  depth,  printed  with  binary human-readable units,
colored  by  file  type, and followed by a grand total. Output is deterministic
because  directory  entries are sorted lexically before aggregation, and totals
still include pruned or filtered descendants.
Usage:

xdu [-a | --all] [-c | --total] [-C mode | --color mode] [-d depth | --depth
depth] [-h | --human-readable] [--help] [-i | --inodes] [-r | --reverse] [-s |
--sort] [-t | --tree] [-v | --version] [path...]

-a, --all

	Include hidden files and directories (basename starts with `.'). Explicitly
	supplied hidden paths are always scanned.

-c, --total

	Print grand total after all results as "<value><TAB>total".

-C mode, --color mode

	Color mode: auto, always, never. Default auto.  --color=mode is accepted.

-d depth, --depth depth

	Max visible depth (0 = root only, -1 unlimited). Display-only; totals still
	include pruned descendants. In list mode limits shown directory depth.

-h, --human-readable

	Human-readable sizes (binary units B, KiB, MiB, ...). Rejected with -i.

--help Print usage and exit 0.

-i, --inodes

	Show inode counts instead of bytes.

-r, --reverse

	Reverse sort. With -s: largest first (desc).

-s, --sort

	Sort smallest first (asc). Without sorting flags output is lexical.

-t, --tree

	Render as a tree instead of a flat list.

-v, --version

	Print version and exit 0.

Examples:

List the current directory (like du):

	$ xdu

List with human-readable sizes for a specific path:

	$ xdu -h .

List inode counts and print a grand total:

	$ xdu -i -c my-folder

Render a tree:

	$ xdu -t my-folder

Render a tree that includes hidden entries:

	$ xdu -t -a my-folder

Limit tree depth to one level:

	$ xdu -t -d 1 my-folder

Sort by size in the tree, largest first, with human-readable sizes and a grand
total:

	$ xdu -t -s -r -c -h my-folder

Control color:

	$ xdu --color never -t my-folder
	$ xdu --color always -t my-folder
	$ NO_COLOR=1 xdu -t my-folder

Hidden file handling:

	$ xdu -a my-folder
	$ xdu .hidden-folder

Exit Status:

	0   Success, no scan errors.
	1   One  or  more  entries  skipped  (permission,  vanishing file). Partial
	results printed; warnings go to stderr as "xdu: <path>: <error>".
	2   Usage error (unknown flag, invalid value, illegal combination).

Notes:

xdu  reports  apparent  sizes  from  Lstat,  not allocated blocks on disk. This
matches  du  --apparent-size  rather than the default du block count. Directory
totals  are the sum of included descendants. Symlinked directories are shown as
links and are not recursed, so a symlink loop cannot occur.
*/
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"fscm/xdu/xdu"
)

const Version string = "0.1.0"

func run(args []string, stdout, stderr io.Writer, version string) int {
	options, err := xdu.ParseArgs(args, version)
	if err != nil {
		if err == xdu.ErrHelp || err == xdu.ErrVersion {
			return 0
		}
		fmt.Fprintf(stderr, "invalid option: %v\n", err)
		return 1
	}
	result := xdu.ScanPaths(options.Paths, options.ShowHidden)
	colorEnabled := xdu.ColorEnabled(options.ColorMode)
	var strBuilder strings.Builder
	if options.Tree {
		xdu.WriteTree(&strBuilder, result.Paths, options, colorEnabled)
	} else {
		xdu.WriteList(&strBuilder, result.Paths, options, colorEnabled)
	}
	if options.ShowTotal {
		total := xdu.GrandTotal(result.Paths, options.Inodes)
		line := xdu.FormatSize(total, options.HumanReadable) + "\ttotal\n"
		if colorEnabled {
			line = xdu.ColorBold(line)
		}
		strBuilder.WriteString(line)
	}
	fmt.Fprint(stdout, strBuilder.String())
	for _, err := range result.Errors {
		fmt.Fprintf(stderr, "error: %v\n", err)
	}
	if len(result.Errors) > 0 {
		return 2
	}
	return 0
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr, Version))
}

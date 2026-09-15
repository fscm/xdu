// -*- mode: go; coding: utf-8 -*-
//
// SPDX-FileCopyrightText: 2026 Frederico Martins
// SPDX-License-Identifier: GPL-3.0-only

package xdu

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
)

const usage string = `Usage: %s [-a | --all] [-c | --total] [-C | --color MODE] [-d | --depth DEPTH] [-h | --human-readable] [--help] [-i | --inodes] [-r | --reverse] [-s | --sort] [-t | --tree] [-v | --version] [path]...
  -a, --all      Include hidden files and directories.
  -c, --total    Print grand total after all results.
  -C MODE, --color=MODE
                 Color the output ('auto', 'always', 'never').
  -d, depth=DEPTH
                 Max visible depth in tree mode (0 = root only).
  -h, --human-readable
                 Print sizes in human readable format.
  --help         Show this help message and exit.
  -i, --inodes   Show inode counts instead of bytes.
  -r, --reverse  Reverse sort order ('desc').
  -s, --sort     Sort by size ('asc') instead of lexically.
  -t, --tree     Render as a tree instead of a flat list.
  -v, --version  Show the program's version number and exit.
`

var (
	ErrHelp    = fmt.Errorf("'help' option error")
	ErrVersion = fmt.Errorf("'version' option error")
)

type Options struct {
	ColorMode     string // --color, -C
	Paths         []string
	MaxDepth      int  // -d
	HumanReadable bool // -h
	Inodes        bool // -i
	Reverse       bool // -r
	ShowHidden    bool // -a
	ShowTotal     bool // -c
	Sort          bool // -s
	Tree          bool // -t
}

// ParseArgs parses the program arguments given at args.
func ParseArgs(args []string, version string) (Options, error) {
	if len(args) == 0 { // Guard against empty args (no program name).
		args = []string{filepath.Base("xdu")} // fallback name
	}
	options := Options{ // Initialize options with defaults.
		ColorMode: "auto",
		MaxDepth:  -1,
	}
	var showHelp, showVersion bool
	flags := flag.NewFlagSet(filepath.Base(args[0]), flag.ContinueOnError)
	flags.SetOutput(io.Discard) // suppress automatic error/help output.
	flags.BoolVar(&options.ShowHidden, "a", false, "")
	flags.BoolVar(&options.ShowHidden, "all", false, "")
	flags.BoolVar(&options.ShowTotal, "c", false, "")
	flags.BoolVar(&options.ShowTotal, "total", false, "")
	flags.StringVar(&options.ColorMode, "C", options.ColorMode, "")
	flags.StringVar(&options.ColorMode, "color", options.ColorMode, "")
	flags.IntVar(&options.MaxDepth, "d", options.MaxDepth, "")
	flags.IntVar(&options.MaxDepth, "depth", options.MaxDepth, "")
	flags.BoolVar(&options.HumanReadable, "h", false, "")
	flags.BoolVar(&options.HumanReadable, "human-readable", false, "")
	flags.BoolVar(&showHelp, "help", false, "")
	flags.BoolVar(&options.Inodes, "i", false, "")
	flags.BoolVar(&options.Inodes, "inodes", false, "")
	flags.BoolVar(&options.Reverse, "r", false, "")
	flags.BoolVar(&options.Reverse, "reverse", false, "")
	flags.BoolVar(&options.Sort, "s", false, "")
	flags.BoolVar(&options.Sort, "sort", false, "")
	flags.BoolVar(&options.Tree, "t", false, "")
	flags.BoolVar(&options.Tree, "tree", false, "")
	flags.BoolVar(&showVersion, "v", false, "")
	flags.BoolVar(&showVersion, "version", false, "")
	if err := flags.Parse(args[1:]); err != nil {
		return Options{}, err
	}
	switch {
	case showHelp:
		fmt.Printf(usage, filepath.Base(args[0]))
		return Options{}, ErrHelp
	case showVersion:
		fmt.Println(version)
		return Options{}, ErrVersion
	case options.MaxDepth < -1:
		return Options{}, fmt.Errorf("depth must be zero or greater")
	}
	switch options.ColorMode {
	case "auto", "always", "never":
		// valid
	default:
		return Options{}, fmt.Errorf(
			"invalid color mode %q; expected 'auto', 'always', or 'never'",
			options.ColorMode,
		)
	}
	options.Paths = flags.Args()
	if len(options.Paths) == 0 {
		options.Paths = []string{"."}
	}
	return options, nil
}

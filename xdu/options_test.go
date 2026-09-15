// -*- mode: go; coding: utf-8 -*-
//
// SPDX-FileCopyrightText: 2026 Frederico Martins
// SPDX-License-Identifier: GPL-3.0-only

package xdu

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func captureStdout(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()
	return buf.String()
}

func TestVersionConstant(t *testing.T) {
	for _, v := range []string{"0.1.0", "9.9.9", "test-version", ""} {
		out := captureStdout(func() {
			_, _ = ParseArgs([]string{"xdu", "-v"}, v)
		})
		if !strings.Contains(out, v) {
			t.Errorf("version %q: output %q should contain version", v, out)
		}
	}
}

func TestUsageConstant(t *testing.T) {
	if usage == "" {
		t.Fatal("usage must not be empty")
	}
	mustContain := []string{
		"-a, --all",
		"-c, --total",
		"-C MODE, --color=MODE",
		"-d, depth=DEPTH",
		"-h, --human-readable",
		"--help",
		"-i, --inodes",
		"-r, --reverse",
		"-s, --sort",
		"-t, --tree",
		"-v, --version",
		"[path]",
	}
	for _, s := range mustContain {
		if !strings.Contains(usage, s) {
			t.Errorf("usage missing %q", s)
		}
	}
	if !strings.Contains(usage, "%s") {
		t.Error("usage must contain placeholder for program name")
	}
}

func TestErrSentinels(t *testing.T) {
	if ErrHelp == nil || ErrVersion == nil {
		t.Fatal("sentinels must not be nil")
	}
	if errors.Is(ErrHelp, ErrVersion) {
		t.Error("ErrHelp and ErrVersion must be distinct")
	}
	if !strings.Contains(ErrHelp.Error(), "help") {
		t.Errorf("ErrHelp = %q, should mention help", ErrHelp)
	}
	if !strings.Contains(ErrVersion.Error(), "version") {
		t.Errorf("ErrVersion = %q, should mention version", ErrVersion)
	}
}

func TestIoDiscard(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"hello", []byte("hello")},
		{"large", bytes.Repeat([]byte("x"), 8192)},
		{"binary", []byte{0, 1, 2, 255}},
	}
	for _, tc := range tests {
		n, err := io.Discard.Write(tc.data)
		if err != nil {
			t.Errorf("%s: unexpected error %v", tc.name, err)
		}
		if n != len(tc.data) {
			t.Errorf("%s: n=%d want %d", tc.name, n, len(tc.data))
		}
	}
	if io.Discard == nil {
		t.Error("io.Discard should not be nil")
	}
	var _ io.Writer = io.Discard
	flagsErr := func() error {
		_, err := ParseArgs([]string{"prog", "--unknown-flag"}, "0.1.0")
		return err
	}()
	if flagsErr == nil {
		t.Error("ParseArgs with unknown flag must return error")
	}
}

func assertOptions(t *testing.T, got, want Options) {
	t.Helper()
	if got.ColorMode != want.ColorMode {
		t.Errorf("ColorMode = %q, want %q", got.ColorMode, want.ColorMode)
	}
	if got.HumanReadable != want.HumanReadable {
		t.Errorf(
			"HumanReadable = %v, want %v",
			got.HumanReadable,
			want.HumanReadable,
		)
	}
	if got.Inodes != want.Inodes {
		t.Errorf("Inodes = %v, want %v", got.Inodes, want.Inodes)
	}
	if got.MaxDepth != want.MaxDepth {
		t.Errorf("MaxDepth = %d, want %d", got.MaxDepth, want.MaxDepth)
	}
	if got.Reverse != want.Reverse {
		t.Errorf("Reverse = %v, want %v", got.Reverse, want.Reverse)
	}
	if got.ShowHidden != want.ShowHidden {
		t.Errorf("ShowHidden = %v, want %v", got.ShowHidden, want.ShowHidden)
	}
	if got.ShowTotal != want.ShowTotal {
		t.Errorf("ShowTotal = %v, want %v", got.ShowTotal, want.ShowTotal)
	}
	if got.Sort != want.Sort {
		t.Errorf("Sort = %v, want %v", got.Sort, want.Sort)
	}
	if got.Tree != want.Tree {
		t.Errorf("Tree = %v, want %v", got.Tree, want.Tree)
	}
	if len(got.Paths) != len(want.Paths) {
		t.Fatalf(
			"Paths len = %d (%v), want %d (%v)",
			len(got.Paths),
			got.Paths,
			len(want.Paths),
			want.Paths,
		)
	}
	for i := range got.Paths {
		if got.Paths[i] != want.Paths[i] {
			t.Errorf("Paths[%d] = %q, want %q", i, got.Paths[i], want.Paths[i])
		}
	}
}

func TestParseArgs_Defaults(t *testing.T) {
	got, err := ParseArgs([]string{"xdu"}, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Options{
		ColorMode: "auto",
		MaxDepth:  -1,
		Paths:     []string{"."},
	}
	assertOptions(t, got, want)
}

func TestParseArgs_ProgNameOnly(t *testing.T) {
	for _, prog := range []string{"xdu", "/usr/bin/xdu", "my-xdu"} {
		got, err := ParseArgs([]string{prog}, "0.1.0")
		if err != nil {
			t.Fatalf("[%s] unexpected error: %v", prog, err)
		}
		if len(got.Paths) != 1 || got.Paths[0] != "." {
			t.Errorf("[%s] Paths = %v, want [.]", prog, got.Paths)
		}
	}
}

func TestParseArgs_ExplicitDotPath(t *testing.T) {
	got, err := ParseArgs([]string{"xdu", "."}, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Paths) != 1 || got.Paths[0] != "." {
		t.Fatalf("Paths = %v, want [.]", got.Paths)
	}
}

func TestParseArgs_SinglePath(t *testing.T) {
	got, err := ParseArgs([]string{"xdu", "/tmp"}, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Options{ColorMode: "auto", MaxDepth: -1, Paths: []string{"/tmp"}}
	assertOptions(t, got, want)
}

func TestParseArgs_MultiplePaths(t *testing.T) {
	got, err := ParseArgs([]string{"xdu", "/a", "/b", "/c"}, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Paths) != 3 {
		t.Fatalf("Paths = %v, want 3 entries", got.Paths)
	}
	for i, p := range []string{"/a", "/b", "/c"} {
		if got.Paths[i] != p {
			t.Errorf("Paths[%d]=%q want %q", i, got.Paths[i], p)
		}
	}
}

func TestParseArgs_AllFlagsShort(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want Options
	}{
		{
			name: "-a",
			args: []string{"xdu", "-a"},
			want: Options{
				ShowHidden: true,
				ColorMode:  "auto",
				MaxDepth:   -1,
				Paths:      []string{"."},
			},
		},
		{
			name: "-c",
			args: []string{"xdu", "-c"},
			want: Options{
				ShowTotal: true,
				ColorMode: "auto",
				MaxDepth:  -1,
				Paths:     []string{"."},
			},
		},
		{
			name: "-h",
			args: []string{"xdu", "-h"},
			want: Options{
				HumanReadable: true,
				ColorMode:     "auto",
				MaxDepth:      -1,
				Paths:         []string{"."},
			},
		},
		{
			name: "-i",
			args: []string{"xdu", "-i"},
			want: Options{
				Inodes:    true,
				ColorMode: "auto",
				MaxDepth:  -1,
				Paths:     []string{"."},
			},
		},
		{
			name: "-r",
			args: []string{"xdu", "-r"},
			want: Options{
				Reverse:   true,
				ColorMode: "auto",
				MaxDepth:  -1,
				Paths:     []string{"."},
			},
		},
		{
			name: "-s",
			args: []string{"xdu", "-s"},
			want: Options{
				Sort:      true,
				ColorMode: "auto",
				MaxDepth:  -1,
				Paths:     []string{"."},
			},
		},
		{
			name: "-t",
			args: []string{"xdu", "-t"},
			want: Options{
				Tree:      true,
				ColorMode: "auto",
				MaxDepth:  -1,
				Paths:     []string{"."},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseArgs(tc.args, "0.1.0")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertOptions(t, got, tc.want)
		})
	}
}

func TestParseArgs_AllFlagsLong(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want Options
	}{
		{
			name: "--all",
			args: []string{"xdu", "--all"},
			want: Options{
				ShowHidden: true,
				ColorMode:  "auto",
				MaxDepth:   -1,
				Paths:      []string{"."},
			},
		},
		{
			name: "--total",
			args: []string{"xdu", "--total"},
			want: Options{
				ShowTotal: true,
				ColorMode: "auto",
				MaxDepth:  -1,
				Paths:     []string{"."},
			},
		},
		{
			name: "--human-readable",
			args: []string{"xdu", "--human-readable"},
			want: Options{
				HumanReadable: true,
				ColorMode:     "auto",
				MaxDepth:      -1,
				Paths:         []string{"."},
			},
		},
		{
			name: "--inodes",
			args: []string{"xdu", "--inodes"},
			want: Options{
				Inodes:    true,
				ColorMode: "auto",
				MaxDepth:  -1,
				Paths:     []string{"."},
			},
		},
		{
			name: "--reverse",
			args: []string{"xdu", "--reverse"},
			want: Options{
				Reverse:   true,
				ColorMode: "auto",
				MaxDepth:  -1,
				Paths:     []string{"."},
			},
		},
		{
			name: "--sort",
			args: []string{"xdu", "--sort"},
			want: Options{
				Sort:      true,
				ColorMode: "auto",
				MaxDepth:  -1,
				Paths:     []string{"."},
			},
		},
		{
			name: "--tree",
			args: []string{"xdu", "--tree"},
			want: Options{
				Tree:      true,
				ColorMode: "auto",
				MaxDepth:  -1,
				Paths:     []string{"."},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseArgs(tc.args, "0.1.0")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertOptions(t, got, tc.want)
		})
	}
}

func TestParseArgs_ColorMode(t *testing.T) {
	tests := []struct {
		name string
		want string
		args []string
	}{
		{"default", "auto", []string{"xdu"}},
		{"-C always", "always", []string{"xdu", "-C", "always"}},
		{"-C never", "never", []string{"xdu", "-C", "never"}},
		{"-C auto", "auto", []string{"xdu", "-C", "auto"}},
		{"-C=always", "always", []string{"xdu", "-C=always"}},
		{"--color always", "always", []string{"xdu", "--color", "always"}},
		{"--color=never", "never", []string{"xdu", "--color=never"}},
		{"--color auto", "auto", []string{"xdu", "--color", "auto"}},
		{"--color=always", "always", []string{"xdu", "--color=always"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseArgs(tc.args, "0.1.0")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ColorMode != tc.want {
				t.Errorf("ColorMode = %q, want %q", got.ColorMode, tc.want)
			}
			if len(got.Paths) == 0 || got.Paths[0] != "." {
				t.Errorf("Paths = %v, want [.]", got.Paths)
			}
		})
	}
	for _, args := range [][]string{
		{"xdu", "-C", "custom"},
		{"xdu", "-C", ""},
		{"xdu", "--color", "AUTO"},
		{"xdu", "--color=Always"},
		{"xdu", "--color", "invalid"},
	} {
		opts, err := ParseArgs(args, "0.1.0")
		if err == nil {
			t.Fatalf(
				"args %v: want error for invalid color, got opts %+v",
				args,
				opts,
			)
		}
		if !strings.Contains(err.Error(), "invalid color mode") {
			t.Errorf(
				"args %v: err %q should contain %q",
				args,
				err,
				"invalid color mode",
			)
		}
		if opts.ColorMode != "" || opts.Paths != nil {
			t.Errorf(
				"args %v: on color error, options should be zero, got %+v",
				args,
				opts,
			)
		}
	}
}

func TestParseArgs_ColorModeWithPath(t *testing.T) {
	got, err := ParseArgs([]string{"xdu", "-C", "never", "/tmp"}, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ColorMode != "never" {
		t.Errorf("ColorMode = %q, want never", got.ColorMode)
	}
	if len(got.Paths) != 1 || got.Paths[0] != "/tmp" {
		t.Errorf("Paths = %v, want [/tmp]", got.Paths)
	}
}

func TestParseArgs_Depth(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"default", []string{"xdu"}, -1},
		{"-d 0", []string{"xdu", "-d", "0"}, 0},
		{"-d 1", []string{"xdu", "-d", "1"}, 1},
		{"-d 5", []string{"xdu", "-d", "5"}, 5},
		{"-d=3", []string{"xdu", "-d=3"}, 3},
		{"--depth 2", []string{"xdu", "--depth", "2"}, 2},
		{"--depth=4", []string{"xdu", "--depth=4"}, 4},
		{"--depth=0", []string{"xdu", "--depth=0"}, 0},
		{"large depth", []string{"xdu", "-d", "100"}, 100},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseArgs(tc.args, "0.1.0")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.MaxDepth != tc.want {
				t.Errorf("MaxDepth = %d, want %d", got.MaxDepth, tc.want)
			}
		})
	}
}

func TestParseArgs_DepthNegative(t *testing.T) {
	for _, args := range [][]string{
		{"xdu", "-d", "-5"},
		{"xdu", "-d", "-2"},
		{"xdu", "--depth", "-10"},
		{"xdu", "--depth=-5"},
		{"xdu", "-d=-2"},
	} {
		opts, err := ParseArgs(args, "0.1.0")
		if err == nil {
			t.Fatalf(
				"args %v: want error for depth < -1, got opts %+v",
				args,
				opts,
			)
		}
		if !strings.Contains(err.Error(), "depth must be zero or greater") {
			t.Errorf(
				"args %v: err %q want to contain %q",
				args,
				err,
				"depth must be zero or greater",
			)
		}
		if opts.ColorMode != "" || opts.MaxDepth != 0 || opts.Paths != nil {
			t.Errorf(
				"args %v: on depth error, options should be zero, got %+v",
				args,
				opts,
			)
		}
	}
	got, err := ParseArgs([]string{"xdu", "-d", "-1"}, "0.1.0")
	if err != nil {
		t.Fatalf("explicit -1 should be allowed, got err %v", err)
	}
	if got.MaxDepth != -1 {
		t.Errorf("MaxDepth = %d, want -1", got.MaxDepth)
	}
}

func TestParseArgs_DepthWithPath(t *testing.T) {
	got, err := ParseArgs([]string{"xdu", "-d", "2", "/some/path"}, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MaxDepth != 2 {
		t.Errorf("MaxDepth = %d, want 2", got.MaxDepth)
	}
	if len(got.Paths) != 1 || got.Paths[0] != "/some/path" {
		t.Errorf("Paths = %v, want [/some/path]", got.Paths)
	}
}

func TestParseArgs_CombinedFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want Options
	}{
		{
			name: "all flags together",
			args: []string{"xdu", "-a", "-c", "-h", "-i", "-r", "-s", "-t"},
			want: Options{
				ShowHidden:    true,
				ShowTotal:     true,
				HumanReadable: true,
				Inodes:        true,
				Reverse:       true,
				Sort:          true,
				Tree:          true,
				ColorMode:     "auto",
				MaxDepth:      -1,
				Paths:         []string{"."},
			},
		},
		{
			name: "long flags together",
			args: []string{
				"xdu",
				"--all",
				"--total",
				"--human-readable",
				"--inodes",
				"--reverse",
				"--sort",
				"--tree",
			},
			want: Options{
				ShowHidden:    true,
				ShowTotal:     true,
				HumanReadable: true,
				Inodes:        true,
				Reverse:       true,
				Sort:          true,
				Tree:          true,
				ColorMode:     "auto",
				MaxDepth:      -1,
				Paths:         []string{"."},
			},
		},
		{
			name: "mixed short/long with values",
			args: []string{
				"xdu",
				"-a",
				"--total",
				"-C", "always",
				"-d", "2",
				"--human-readable",
				"-i",
				"-r",
				"--sort",
				"--tree",
				"/path1", "/path2",
			},
			want: Options{
				ShowHidden:    true,
				ShowTotal:     true,
				HumanReadable: true,
				Inodes:        true,
				Reverse:       true,
				Sort:          true,
				Tree:          true,
				ColorMode:     "always",
				MaxDepth:      2,
				Paths:         []string{"/path1", "/path2"},
			},
		},
		{
			name: "color and depth combined",
			args: []string{"xdu", "--color=never", "--depth=0", "-t", "/tmp"},
			want: Options{
				ColorMode: "never",
				MaxDepth:  0,
				Tree:      true,
				Paths:     []string{"/tmp"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseArgs(tc.args, "0.1.0")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertOptions(t, got, tc.want)
		})
	}
}

func TestParseArgs_FlagEqualsSyntax(t *testing.T) {
	cases := []struct {
		args []string
		want Options
	}{
		{
			[]string{"xdu", "-C=never"},
			Options{ColorMode: "never", MaxDepth: -1, Paths: []string{"."}},
		},
		{
			[]string{"xdu", "--color=always"},
			Options{ColorMode: "always", MaxDepth: -1, Paths: []string{"."}},
		},
		{
			[]string{"xdu", "-d=10"},
			Options{ColorMode: "auto", MaxDepth: 10, Paths: []string{"."}},
		},
		{
			[]string{"xdu", "--depth=10"},
			Options{ColorMode: "auto", MaxDepth: 10, Paths: []string{"."}},
		},
	}
	for _, tc := range cases {
		got, err := ParseArgs(tc.args, "0.1.0")
		if err != nil {
			t.Fatalf("args %v: unexpected error %v", tc.args, err)
		}
		assertOptions(t, got, tc.want)
	}
}

func TestParseArgs_Help(t *testing.T) {
	tests := []struct {
		name     string
		progName string
		args     []string
	}{
		{"basic", "xdu", []string{"xdu", "--help"}},
		{"with path ignored", "xdu", []string{"xdu", "--help", "/some/path"}},
		{"with other flags", "xdu", []string{"xdu", "-a", "--help", "-t"}},
		{
			"basename extraction",
			"/usr/local/bin/xdu",
			[]string{"/usr/local/bin/xdu", "--help"},
		},
		{"dot prog", "./xdu", []string{"./xdu", "--help"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotErr error
			var gotOpts Options
			out := captureStdout(func() {
				var err error
				gotOpts, err = ParseArgs(tc.args, "0.1.0")
				gotErr = err
			})
			if !errors.Is(gotErr, ErrHelp) {
				t.Fatalf("err = %v, want ErrHelp", gotErr)
			}
			if gotOpts.ColorMode != "" || gotOpts.Paths != nil {
				t.Errorf("on help, options should be zero, got %+v", gotOpts)
			}
			base := filepath.Base(tc.progName)
			if !strings.Contains(out, base) {
				t.Errorf(
					"help output %q must contain prog basename %q",
					out,
					base,
				)
			}
			if !strings.Contains(out, "Usage:") {
				t.Errorf("help output %q must contain Usage:", out)
			}
			if !strings.Contains(out, "--help") {
				t.Errorf("help output must contain --help")
			}
		})
	}
}

func TestParseArgs_HelpPrecedenceOverVersion(t *testing.T) {
	var gotErr error
	out := captureStdout(func() {
		_, gotErr = ParseArgs([]string{"xdu", "--help", "-v"}, "0.1.0")
	})
	if !errors.Is(gotErr, ErrHelp) {
		t.Fatalf("err = %v, want ErrHelp (help takes precedence)", gotErr)
	}
	if !strings.Contains(out, "Usage:") {
		t.Error(
			"help output should contain Usage when both help and version present",
		)
	}
	out2 := captureStdout(func() {
		_, gotErr = ParseArgs([]string{"xdu", "-v", "--help"}, "0.1.0")
	})
	if !errors.Is(gotErr, ErrHelp) {
		t.Fatalf("err = %v, want ErrHelp when version before help", gotErr)
	}
	if !strings.Contains(out2, "Usage:") {
		t.Error("help output should contain Usage when version before help")
	}
}

func TestParseArgs_Version(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"short", []string{"xdu", "-v"}},
		{"long", []string{"xdu", "--version"}},
		{"with path", []string{"xdu", "-v", "/tmp"}},
		{"with multiple paths", []string{"xdu", "--version", "/a", "/b"}},
		{"with flags", []string{"xdu", "-a", "--version", "-t", "/tmp"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotErr error
			var gotOpts Options
			out := captureStdout(func() {
				var err error
				gotOpts, err = ParseArgs(tc.args, "0.1.0")
				gotErr = err
			})
			if !errors.Is(gotErr, ErrVersion) {
				t.Fatalf("err = %v, want ErrVersion", gotErr)
			}
			if !strings.Contains(out, "0.1.0") {
				t.Errorf("version output %q must contain %q", out, "0.1.0")
			}
			if gotOpts.ColorMode != "" || gotOpts.Paths != nil {
				t.Errorf("on version, options should be zero, got %+v", gotOpts)
			}
			if gotOpts.MaxDepth != 0 {
				t.Errorf(
					"on version, MaxDepth should be 0, got %d",
					gotOpts.MaxDepth,
				)
			}
		})
	}
}

func TestParseArgs_VersionNoDefaultPath(t *testing.T) {
	var opts Options
	var err error
	out := captureStdout(func() {
		opts, err = ParseArgs([]string{"xdu", "-v"}, "0.1.0")
	})
	if !errors.Is(err, ErrVersion) {
		t.Fatalf("err = %v, want ErrVersion", err)
	}
	if !strings.Contains(out, "0.1.0") {
		t.Errorf("version output %q must contain %q", out, "0.1.0")
	}
	if opts.Paths != nil || opts.ColorMode != "" {
		t.Errorf("on version, options should be zero, got %+v", opts)
	}
}

func TestParseArgs_VersionAndHelpBothSet(t *testing.T) {
	out := captureStdout(func() {
		_, _ = ParseArgs([]string{"xdu", "--help", "--version"}, "0.1.0")
	})
	if strings.Contains(out, "0.1.0") && !strings.Contains(out, "Usage:") {
		t.Error("when both help and version, should show help not version")
	}
	if !strings.Contains(out, "Usage:") {
		t.Error("should show Usage when both flags set")
	}
}

func TestParseArgs_VersionParam(t *testing.T) {
	for _, v := range []string{"0.1.0", "1.2.3", "v9.9.9-custom", "dev"} {
		out := captureStdout(func() {
			_, err := ParseArgs([]string{"xdu", "--version"}, v)
			if !errors.Is(err, ErrVersion) {
				t.Fatalf("version %q: want ErrVersion", v)
			}
		})
		if strings.TrimSpace(out) != v {
			t.Errorf("version %q: output %q want %q", v, out, v)
		}
	}
	out := captureStdout(func() {
		_, err := ParseArgs([]string{"xdu", "-v"}, "")
		if !errors.Is(err, ErrVersion) {
			t.Fatalf("empty version: want ErrVersion")
		}
	})
	if strings.TrimSpace(out) != "" {
		t.Errorf("empty version: output %q want empty", out)
	}
	for _, args := range [][]string{
		{"xdu", "-v", "-d", "-5"},
		{"xdu", "--version", "-C", "invalid"},
		{"xdu", "-v", "--depth=-10", "--color=bad"},
	} {
		out := captureStdout(func() {
			opts, err := ParseArgs(args, "9.9.9")
			if !errors.Is(err, ErrVersion) {
				t.Fatalf("args %v: want ErrVersion, got %v", args, err)
			}
			if opts.Paths != nil || opts.ColorMode != "" {
				t.Errorf(
					"args %v: on version, options should be zero, got %+v",
					args,
					opts,
				)
			}
		})
		if strings.TrimSpace(out) != "9.9.9" {
			t.Errorf("args %v: version output %q want %q", args, out, "9.9.9")
		}
	}
	out = captureStdout(func() {
		_, err := ParseArgs([]string{"xdu", "--help", "-v"}, "1.2.3")
		if !errors.Is(err, ErrHelp) {
			t.Fatalf("help+version: want ErrHelp")
		}
	})
	if !strings.Contains(out, "Usage:") {
		t.Error("help+version: want Usage")
	}
	if strings.Contains(out, "1.2.3") {
		t.Error("help+version: should not print version")
	}
}

func TestParseArgs_Errors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"unknown short", []string{"xdu", "-z"}},
		{"unknown long", []string{"xdu", "--unknown"}},
		{"unknown with value", []string{"xdu", "--unknown=value"}},
		{"invalid depth not int", []string{"xdu", "-d", "abc"}},
		{"invalid depth float", []string{"xdu", "--depth", "1.5"}},
		{"invalid depth long float", []string{"xdu", "--depth=xyz"}},
		{"missing color value at end", []string{"xdu", "-C"}},
		{"missing depth value at end", []string{"xdu", "-d"}},
		{"missing depth long value", []string{"xdu", "--depth"}},
		{"single dash unknown", []string{"xdu", "-unknown"}},
		{"-Z", []string{"xdu", "-Z"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseArgs(tc.args, "0.1.0")
			if err == nil {
				t.Fatalf("args %v: want error, got nil", tc.args)
			}
			if errors.Is(err, ErrHelp) || errors.Is(err, ErrVersion) {
				t.Fatalf(
					"args %v: error = %v, should not be ErrHelp/ErrVersion",
					tc.args,
					err,
				)
			}
		})
	}
}

func TestParseArgs_InvalidDepthErrorMessage(t *testing.T) {
	_, err := ParseArgs([]string{"xdu", "-d", "notanint"}, "0.1.0")
	if err == nil {
		t.Fatal("want error for invalid depth")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("error %q should mention invalid", err)
	}
}

func TestParseArgs_DoubleDash(t *testing.T) {
	got, err := ParseArgs(
		[]string{"xdu", "--", "-a", "--help", "/tmp"},
		"0.1.0",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantPaths := []string{"-a", "--help", "/tmp"}
	if len(got.Paths) != len(wantPaths) {
		t.Fatalf("Paths = %v, want %v", got.Paths, wantPaths)
	}
	for i, p := range wantPaths {
		if got.Paths[i] != p {
			t.Errorf("Paths[%d]=%q want %q", i, got.Paths[i], p)
		}
	}
	got2, err := ParseArgs([]string{"xdu", "-a", "--", "-t"}, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got2.ShowHidden {
		t.Error("ShowHidden should be true (parsed before --)")
	}
	if got2.Tree {
		t.Error("Tree should be false (after -- should be path, not flag)")
	}
	if len(got2.Paths) != 1 || got2.Paths[0] != "-t" {
		t.Errorf("Paths = %v, want [-t]", got2.Paths)
	}
}

func TestParseArgs_DashAsPath(t *testing.T) {
	_, err := ParseArgs([]string{"xdu", "-"}, "0.1.0")
	if err != nil {
		t.Logf("'-' as arg returned error: %v (acceptable)", err)
	} else {
		t.Logf("'-' treated as path (acceptable)")
	}
}

func TestParseArgs_RepeatedFlags(t *testing.T) {
	got, err := ParseArgs(
		[]string{"xdu", "-C", "always", "-C", "never"},
		"0.1.0",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ColorMode != "never" {
		t.Errorf("ColorMode = %q, want never (last wins)", got.ColorMode)
	}
	got2, err := ParseArgs([]string{"xdu", "-d", "1", "--depth", "5"}, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got2.MaxDepth != 5 {
		t.Errorf("MaxDepth = %d, want 5", got2.MaxDepth)
	}
	got3, err := ParseArgs([]string{"xdu", "-a", "-a", "--all"}, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got3.ShowHidden {
		t.Error("ShowHidden should be true")
	}
}

func TestParseArgs_PathWithDashPrefix(t *testing.T) {
	got, err := ParseArgs([]string{"xdu", "--", "--not-a-flag"}, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Paths) != 1 || got.Paths[0] != "--not-a-flag" {
		t.Errorf("Paths = %v, want [--not-a-flag]", got.Paths)
	}
}

func TestParseArgs_SortAndReverse(t *testing.T) {
	got, err := ParseArgs([]string{"xdu", "-s", "-r"}, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Sort || !got.Reverse {
		t.Errorf("Sort=%v Reverse=%v, want both true", got.Sort, got.Reverse)
	}
	got2, err := ParseArgs([]string{"xdu", "-r"}, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got2.Reverse {
		t.Error("Reverse should be true")
	}
	if got2.Sort {
		t.Error("Sort should be false when only -r")
	}
}

func TestParseArgs_HumanReadableWithOtherFlags(t *testing.T) {
	got, err := ParseArgs(
		[]string{"xdu", "-h", "--color=always", "-d", "3", "/path"}, "0.1.0",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.HumanReadable {
		t.Error("HumanReadable should be true")
	}
	if got.ColorMode != "always" {
		t.Errorf("ColorMode = %q, want always", got.ColorMode)
	}
	if got.MaxDepth != 3 {
		t.Errorf("MaxDepth = %d, want 3", got.MaxDepth)
	}
}

func TestParseArgs_InodesAndTotal(t *testing.T) {
	got, err := ParseArgs([]string{"xdu", "-i", "-c"}, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Inodes || !got.ShowTotal {
		t.Errorf(
			"Inodes=%v ShowTotal=%v, want both true",
			got.Inodes,
			got.ShowTotal,
		)
	}
}

func TestParseArgs_TreeAndDepth(t *testing.T) {
	got, err := ParseArgs([]string{"xdu", "-t", "-d", "0"}, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Tree {
		t.Error("Tree should be true")
	}
	if got.MaxDepth != 0 {
		t.Errorf("MaxDepth = %d, want 0", got.MaxDepth)
	}
}

func TestParseArgs_PositionalAfterFlags(t *testing.T) {
	got, err := ParseArgs([]string{"xdu", "/a", "-a"}, "0.1.0")
	if err != nil {
		t.Logf("interleaved flag after path error or not? err=%v", err)
		_ = got
		return
	}
	if got.ShowHidden {
		t.Logf(
			"flag after positional parsed as flag (ShowHidden true) - " +
				"flag package allowed interspersed",
		)
	} else {
		if len(got.Paths) != 2 || got.Paths[0] != "/a" || got.Paths[1] != "-a" {
			t.Errorf(
				"expected Paths=[/a -a] when flag after positional, got %v "+
					"(ShowHidden=%v)",
				got.Paths,
				got.ShowHidden,
			)
		}
	}
}

func TestParseArgs_EmptySlice(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("empty args panicked as expected: %v", r)
		}
	}()
	_, _ = ParseArgs([]string{}, "0.1.0")
}

func TestParseArgs_OnlyProgNameWithFlagsNoArgs(t *testing.T) {
	out := captureStdout(func() {
		_, _ = ParseArgs([]string{"prog", "--unknown"}, "0.1.0")
	})
	if strings.Contains(out, "flag") {
		t.Errorf(
			"flag error output should be discarded (io.Discard), but got %q",
			out,
		)
	}
}

func TestParseArgs_FlagErrorDoesNotSetOptions(t *testing.T) {
	opts, err := ParseArgs([]string{"xdu", "-z"}, "0.1.0")
	if err == nil {
		t.Fatal("want error")
	}
	if opts.ColorMode != "" || opts.MaxDepth != 0 {
		t.Errorf("on error, options should be zero, got %+v", opts)
	}
}

func TestParseArgs_HelpDoesNotSetColorOrDepth(t *testing.T) {
	opts, err := func() (Options, error) {
		var o Options
		var e error
		captureStdout(
			func() {
				o, e = ParseArgs(
					[]string{"xdu", "--help", "-C", "never", "-d", "5"},
					"0.1.0",
				)
			},
		)
		return o, e
	}()
	if !errors.Is(err, ErrHelp) {
		t.Fatalf("want ErrHelp, got %v", err)
	}
	if opts.ColorMode != "" {
		t.Errorf(
			"help should return zero options, ColorMode=%q",
			opts.ColorMode,
		)
	}
	if opts.MaxDepth != 0 {
		t.Errorf("help should return zero options, MaxDepth=%d", opts.MaxDepth)
	}
}

func TestParseArgs_AllShortFlagsWithPaths(t *testing.T) {
	args := []string{
		"xdu",
		"-a",
		"-c",
		"-h",
		"-i",
		"-r",
		"-s",
		"-t",
		"/my/path",
	}
	got, err := ParseArgs(args, "0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.ShowHidden || !got.ShowTotal || !got.HumanReadable || !got.Inodes ||
		!got.Reverse || !got.Sort || !got.Tree {
		t.Errorf("all bool flags should be true, got %+v", got)
	}
	if len(got.Paths) != 1 || got.Paths[0] != "/my/path" {
		t.Errorf("Paths = %v, want [/my/path]", got.Paths)
	}
}

func TestParseArgs_CaseLongFlagWithEquals(t *testing.T) {
	for _, mode := range []string{"auto", "always", "never"} {
		got, err := ParseArgs([]string{"xdu", "--color=" + mode}, "0.1.0")
		if err != nil {
			t.Fatalf("mode %q: unexpected error %v", mode, err)
		}
		if got.ColorMode != mode {
			t.Errorf(
				"mode %q: ColorMode = %q, want %q",
				mode,
				got.ColorMode,
				mode,
			)
		}
	}
	for _, mode := range []string{"AUTO", "Always", "Never", "ALWAYS"} {
		_, err := ParseArgs([]string{"xdu", "--color=" + mode}, "0.1.0")
		if err == nil {
			t.Fatalf("mode %q: want error for invalid color", mode)
		}
		if !strings.Contains(err.Error(), "invalid color mode") {
			t.Errorf(
				"mode %q: err %q should contain %q",
				mode,
				err,
				"invalid color mode",
			)
		}
	}
}

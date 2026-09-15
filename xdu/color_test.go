// -*- mode: go; coding: utf-8 -*-
//
// SPDX-FileCopyrightText: 2026 Frederico Martins
// SPDX-License-Identifier: GPL-3.0-only

package xdu

import (
	"os"
	"testing"
)

func TestColorEnabled_AlwaysNever(t *testing.T) {
	for _, mode := range []string{"always", "never"} {
		want := mode == "always"
		for _, noColor := range []string{"", "1", "true"} {
			t.Setenv("NO_COLOR", noColor)
			got := ColorEnabled(mode)
			if got != want {
				t.Errorf(
					"ColorEnabled(%q) with NO_COLOR=%q: got %v want %v",
					mode,
					noColor,
					got,
					want,
				)
			}
		}
	}
}

func TestColorEnabled_Auto_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		t.Skipf("open %s: %v", os.DevNull, err)
	}
	defer devNull.Close()
	old := os.Stdout
	os.Stdout = devNull
	defer func() { os.Stdout = old }()
	if ColorEnabled("auto") {
		t.Error(
			"ColorEnabled(auto) with NO_COLOR=1 and char device: want false",
		)
	}
	if ColorEnabled("") {
		t.Error(
			"ColorEnabled(\"\") with NO_COLOR=1: want false (falls to auto)",
		)
	}
	if ColorEnabled("invalid") {
		t.Error(
			"ColorEnabled(invalid) with NO_COLOR=1: want false",
		)
	}
}

func TestColorEnabled_Auto_CharDeviceTrue(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		t.Skipf("open %s: %v", os.DevNull, err)
	}
	defer devNull.Close()
	old := os.Stdout
	os.Stdout = devNull
	defer func() { os.Stdout = old }()
	if fi, err := devNull.Stat(); err != nil ||
		fi.Mode()&os.ModeCharDevice == 0 {
		t.Skipf("fixture not char device: %v mode %v", err, fi)
	}
	if !ColorEnabled("auto") {
		t.Error(
			"ColorEnabled(auto) with char device and NO_COLOR=\"\": want true",
		)
	}
	if !ColorEnabled("") {
		t.Error(
			"ColorEnabled(\"\") with char device: want true (default auto)",
		)
	}
	if !ColorEnabled("unknown") {
		t.Error("ColorEnabled(unknown) with char device: want true")
	}
}

func TestColorEnabled_Auto_NonCharDeviceFalse(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old }()
	if ColorEnabled("auto") {
		t.Error("ColorEnabled(auto) with pipe (non-char): want false")
	}
	if ColorEnabled("AUTO") {
		t.Error(
			"ColorEnabled(AUTO) with pipe: want false (case-sensitive, falls to auto)",
		)
	}
	f, err := os.CreateTemp(t.TempDir(), "regular")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer os.Remove(f.Name())
	os.Stdout = f
	if ColorEnabled("auto") {
		t.Error("ColorEnabled(auto) with regular file: want false")
	}
}

func TestColorEnabled_Auto_StatError(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	f, err := os.CreateTemp(t.TempDir(), "closed")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	ff, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	ff.Close()
	old := os.Stdout
	os.Stdout = ff
	defer func() { os.Stdout = old }()
	if ColorEnabled("auto") {
		t.Error(
			"ColorEnabled(auto) with closed Stdout (Stat error): want false",
		)
	}
}

func TestColorizeName(t *testing.T) {
	tests := []struct {
		name string
		node *Node
		want string
	}{
		{
			name: "dir",
			node: &Node{IsDir: true, Path: "foo"},
			want: ansiBlue + "foo" + ansiReset,
		},
		{
			name: "link",
			node: &Node{IsLink: true, Path: "bar"},
			want: ansiCyan + "bar" + ansiReset,
		},
		{
			name: "exec",
			node: &Node{IsExec: true, Path: "run"},
			want: ansiGreen + "run" + ansiReset,
		},
		{
			name: "plain",
			node: &Node{Path: "plain"},
			want: "plain",
		},
		{
			name: "dir precedence over link and exec",
			node: &Node{IsDir: true, IsLink: true, IsExec: true, Path: "x"},
			want: ansiBlue + "x" + ansiReset,
		},
		{
			name: "link precedence over exec",
			node: &Node{IsLink: true, IsExec: true, Path: "y"},
			want: ansiCyan + "y" + ansiReset,
		},
		{
			name: "empty name dir",
			node: &Node{IsDir: true, Path: ""},
			want: ansiBlue + "" + ansiReset,
		},
		{
			name: "name with spaces",
			node: &Node{IsDir: true, Path: "a b c"},
			want: ansiBlue + "a b c" + ansiReset,
		},
		{
			name: "path with slashes",
			node: &Node{IsDir: true, Path: "/tmp/foo/bar"},
			want: ansiBlue + "/tmp/foo/bar" + ansiReset,
		},
		{
			name: "plain empty",
			node: &Node{Path: ""},
			want: "",
		},
	}
	for _, tc := range tests {
		got := ColorizeName(tc.node)
		if got != tc.want {
			t.Errorf(
				"%s: ColorizeName(%v)=%q want %q",
				tc.name,
				tc.node,
				got,
				tc.want,
			)
		}
	}
}

func TestColorizeName_Constants(t *testing.T) {
	if ansiBlue == "" || ansiCyan == "" || ansiGreen == "" || ansiReset == "" {
		t.Fatal("ansi constants must not be empty")
	}
	dir := ColorizeName(&Node{IsDir: true, Path: "d"})
	if dir != ansiBlue+"d"+ansiReset {
		t.Errorf("dir color mismatch: %q", dir)
	}
	link := ColorizeName(&Node{IsLink: true, Path: "l"})
	if link != ansiCyan+"l"+ansiReset {
		t.Errorf("link color mismatch: %q", link)
	}
	exec := ColorizeName(&Node{IsExec: true, Path: "e"})
	if exec != ansiGreen+"e"+ansiReset {
		t.Errorf("exec color mismatch: %q", exec)
	}
	plain := ColorizeName(&Node{Path: "plain"})
	if plain != "plain" {
		t.Errorf("plain mismatch: %q", plain)
	}
}

func TestColorBold(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ansiBold + "" + ansiReset},
		{"simple", "hello", ansiBold + "hello" + ansiReset},
		{"with spaces", "a b", ansiBold + "a b" + ansiReset},
		{
			"with ansi",
			ansiBlue + "x" + ansiReset, ansiBold + ansiBlue + "x" + ansiReset +
				ansiReset,
		},
	}
	for _, tc := range tests {
		got := ColorBold(tc.in)
		if got != tc.want {
			t.Errorf(
				"%s: ColorBold(%q)=%q want %q",
				tc.name,
				tc.in,
				got,
				tc.want,
			)
		}
	}
	if got := ColorBold("test"); got != "\x1b[1mtest\x1b[0m" {
		t.Errorf("ColorBold literal mismatch: %q", got)
	}
}

func TestAnsiConstants(t *testing.T) {
	if ansiReset != "\x1b[0m" {
		t.Errorf("ansiReset=%q want %q", ansiReset, "\x1b[0m")
	}
	if ansiBold != "\x1b[1m" {
		t.Errorf("ansiBold=%q want %q", ansiBold, "\x1b[1m")
	}
	if ansiBlue != "\x1b[1;34m" {
		t.Errorf("ansiBlue=%q want %q", ansiBlue, "\x1b[1;34m")
	}
	if ansiCyan != "\x1b[36m" {
		t.Errorf("ansiCyan=%q want %q", ansiCyan, "\x1b[36m")
	}
	if ansiGreen != "\x1b[32m" {
		t.Errorf("ansiGreen=%q want %q", ansiGreen, "\x1b[32m")
	}
}

// -*- mode: go; coding: utf-8 -*-
//
// SPDX-FileCopyrightText: 2026 Frederico Martins
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fscm/xdu/xdu"
)

func runWithArgs(t *testing.T, args []string) (int, string, string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	code := run(args, &out, &errBuf, Version)
	return code, out.String(), errBuf.String()
}

func TestRun_Help(t *testing.T) {
	for _, args := range [][]string{
		{"xdu", "--help"},
		{"xdu", "--help", "/some/path"},
		{"xdu", "-a", "--help"},
	} {
		code, out, errOut := runWithArgs(t, args)
		if code != 0 {
			t.Errorf("help %v: code=%d want 0", args, code)
		}
		if errOut != "" {
			t.Errorf("help %v: stderr=%q want empty", args, errOut)
		}
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		var buf bytes.Buffer
		done := make(chan struct{})
		go func() {
			close(done)
		}()
		_ = done
		_ = r
		_ = buf
		_ = old
		if out != "" {
			t.Errorf(
				"help %v: stdout via run=%q want empty (help prints to os.Stdout)",
				args,
				out,
			)
		}
		captured := captureStdout(func() {
			_, _ = xdu.ParseArgs(args, Version)
		})
		if !strings.Contains(captured, "Usage:") {
			t.Errorf("help %v: captured %q want Usage", args, captured)
		}
	}
}

func captureStdout(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	r.Close()
	return buf.String()
}

func TestRun_Version(t *testing.T) {
	for _, args := range [][]string{
		{"xdu", "-v"},
		{"xdu", "--version"},
	} {
		code, out, errOut := runWithArgs(t, args)
		if code != 0 {
			t.Errorf("version %v: code=%d want 0", args, code)
		}
		if errOut != "" {
			t.Errorf("version stderr=%q", errOut)
		}
		if out != "" {
			t.Errorf(
				"version run stdout=%q want empty (version prints to os.Stdout)",
				out,
			)
		}
		captured := captureStdout(func() {
			_, _ = xdu.ParseArgs(args, Version)
		})
		if !strings.Contains(captured, Version) {
			t.Errorf("version captured %q want %q", captured, Version)
		}
		custom := "9.9.9"
		captured2 := captureStdout(func() {
			_, _ = xdu.ParseArgs([]string{"xdu", "--version"}, custom)
		})
		if !strings.Contains(captured2, custom) {
			t.Errorf("custom version %q not in %q", custom, captured2)
		}
	}
}

func TestRun_InvalidOption(t *testing.T) {
	code, out, errOut := runWithArgs(t, []string{"xdu", "-z"})
	if code != 1 {
		t.Errorf("invalid: code=%d want 1", code)
	}
	if out != "" {
		t.Errorf("invalid stdout=%q want empty", out)
	}
	if !strings.Contains(errOut, "invalid option") {
		t.Errorf("invalid stderr=%q want 'invalid option'", errOut)
	}
	for _, args := range [][]string{
		{"xdu", "-d", "-5"},
		{"xdu", "-C", "bad"},
	} {
		code, _, errOut := runWithArgs(t, args)
		if code != 1 {
			t.Errorf("invalid %v: code=%d want 1", args, code)
		}
		if !strings.Contains(errOut, "invalid") &&
			!strings.Contains(errOut, "depth") {
			t.Logf("args %v stderr=%q", args, errOut)
		}
	}
}

func TestRun_Success_List(t *testing.T) {
	tmp := t.TempDir()
	f1 := filepath.Join(tmp, "a.txt")
	f2 := filepath.Join(tmp, "b.txt")
	os.WriteFile(f1, []byte("hello"), 0o644)
	os.WriteFile(f2, []byte("world!!"), 0o644)
	code, out, errOut := runWithArgs(t, []string{"xdu", f1, f2})
	if code != 0 {
		t.Fatalf("list: code=%d err=%q", code, errOut)
	}
	if errOut != "" {
		t.Errorf("list stderr=%q want empty", errOut)
	}
	if !strings.Contains(out, "a.txt") || !strings.Contains(out, "b.txt") {
		t.Errorf("list out=%q want both files", out)
	}
	if strings.Index(out, "a.txt") > strings.Index(out, "b.txt") {
		t.Errorf("lexical order: a before b, out=%q", out)
	}
	code, out, _ = runWithArgs(t, []string{"xdu", tmp})
	if !strings.Contains(out, tmp) {
		t.Errorf("list dir: out=%q should contain dir %q", out, tmp)
	}
}

func TestRun_Success_Tree(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	os.Mkdir(sub, 0o755)
	os.WriteFile(filepath.Join(sub, "file.txt"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(tmp, "root.txt"), []byte("yy"), 0o644)
	code, out, errOut := runWithArgs(t, []string{"xdu", "-t", tmp})
	if code != 0 {
		t.Fatalf("tree: code=%d err=%q", code, errOut)
	}
	if !strings.Contains(out, "├── ") || !strings.Contains(out, "└── ") {
		t.Logf("tree out=%q", out)
	}
	if !strings.Contains(out, "file.txt") {
		t.Errorf("tree out missing file.txt: %q", out)
	}
	if errOut != "" {
		t.Errorf("tree stderr=%q", errOut)
	}
}

func TestRun_ShowHidden(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, ".hidden"), []byte("s"), 0o644)
	os.WriteFile(filepath.Join(tmp, "visible"), []byte("v"), 0o644)
	code, out, _ := runWithArgs(t, []string{"xdu", "-t", tmp})
	if code != 0 {
		t.Fatal(code)
	}
	if strings.Contains(out, ".hidden") {
		t.Errorf("without -a should not contain .hidden: %q", out)
	}
	if !strings.Contains(out, "visible") {
		t.Errorf("without -a missing visible: %q", out)
	}
	code, out, _ = runWithArgs(t, []string{"xdu", "-t", "-a", tmp})
	if !strings.Contains(out, ".hidden") {
		t.Errorf("with -a should contain .hidden: %q", out)
	}
	code, out, _ = runWithArgs(t, []string{"xdu", "-c", tmp})
	if !strings.Contains(out, "total") {
		t.Errorf("showTotal missing: %q", out)
	}
	code, out2, _ := runWithArgs(t, []string{"xdu", "-c", "-a", tmp})
	if out == out2 {
		t.Errorf("hidden should affect total: without=%q with=%q", out, out2)
	}
}

func TestRun_ShowTotal(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "a"), []byte("12345"), 0o644) // 5
	os.WriteFile(
		filepath.Join(tmp, "b"),
		[]byte("12"),
		0o644,
	)
	code, out, _ := runWithArgs(t, []string{"xdu", "-c", tmp})
	if code != 0 {
		t.Fatalf("code %d", code)
	}
	if !strings.Contains(out, "total") {
		t.Errorf("showTotal missing total: %q", out)
	}
	if !strings.Contains(out, "7") {
		t.Errorf("total should contain 7: %q", out)
	}
	code, out, _ = runWithArgs(t, []string{"xdu", "-c", "-h", tmp})
	if !strings.Contains(out, "7 B") {
		t.Errorf("human total: %q", out)
	}
	code, out, _ = runWithArgs(t, []string{"xdu", "-c", "-i", tmp})
	if !strings.Contains(out, "3") {
		t.Errorf("inodes total: %q", out)
	}
}

func TestRun_ShowTotalWithColor(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "a"), []byte("x"), 0o644)
	code, out, _ := runWithArgs(t, []string{"xdu", "-c", "-C", "always", tmp})
	if code != 0 {
		t.Fatal(code)
	}
	if !strings.Contains(out, "\x1b[1m") {
		t.Errorf("color total should contain bold: %q", out)
	}
	if !strings.Contains(out, "\x1b[0m") {
		t.Errorf("color total missing reset: %q", out)
	}
	code, out, _ = runWithArgs(t, []string{"xdu", "-c", "-C", "never", tmp})
	if strings.Contains(out, "\x1b[") {
		t.Errorf("color never should not contain ansi: %q", out)
	}
}

func TestRun_SortAndReverse(t *testing.T) {
	tmp := t.TempDir()
	small := filepath.Join(tmp, "small")
	big := filepath.Join(tmp, "big")
	os.WriteFile(small, []byte("a"), 0o644)        //1
	os.WriteFile(big, []byte("aaaaaaaaaa"), 0o644) //10
	code, out, _ := runWithArgs(t, []string{"xdu", big, small})
	if strings.Index(out, "big") > strings.Index(out, "small") {
		t.Errorf("lexical: big before small: %q", out)
	}
	code, out, _ = runWithArgs(t, []string{"xdu", "-s", big, small})
	if strings.Index(out, "small") > strings.Index(out, "big") {
		t.Errorf("sort by size asc: small before big: %q", out)
	}
	if code != 0 {
		t.Fatal(code)
	}
	code, out, _ = runWithArgs(t, []string{"xdu", "-s", "-r", big, small})
	if strings.Index(out, "big") > strings.Index(out, "small") {
		t.Errorf("sort reverse: big before small: %q", out)
	}
	code, out, _ = runWithArgs(t, []string{"xdu", "-t", "-s", tmp})
	if strings.Index(out, "small") > strings.Index(out, "big") {
		t.Errorf("tree sort asc: small before big: %q", out)
	}
}

func TestRun_HumanReadable(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "f"), bytesRepeat(2048), 0o644)
	code, out, _ := runWithArgs(t, []string{"xdu", "-h", tmp})
	if !strings.Contains(out, "2.0 KiB") {
		t.Errorf("human: %q", out)
	}
	if code != 0 {
		t.Fatal(code)
	}
	code, out, _ = runWithArgs(t, []string{"xdu", tmp})
	if strings.Contains(out, "KiB") {
		t.Errorf("non-human should not contain KiB: %q", out)
	}
}

func bytesRepeat(n int) []byte {
	return bytes.Repeat([]byte("x"), n)
}

func TestRun_Inodes(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	os.Mkdir(sub, 0o755)
	os.WriteFile(filepath.Join(sub, "file"), []byte("x"), 0o644)
	code, out, _ := runWithArgs(t, []string{"xdu", "-i", tmp})
	if code != 0 {
		t.Fatal(code)
	}
	if !strings.Contains(out, "3") {
		t.Logf("inodes out=%q", out)
	}
}

func TestRun_ColorMode(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "a"), []byte("x"), 0o644)
	t.Setenv("NO_COLOR", "1")
	code, out, _ := runWithArgs(t, []string{"xdu", "-C", "auto", tmp})
	if strings.Contains(out, "\x1b[") {
		t.Errorf("NO_COLOR should disable color: %q", out)
	}
	if code != 0 {
		t.Fatal(code)
	}
	t.Setenv("NO_COLOR", "")
	code, out, _ = runWithArgs(t, []string{"xdu", "-C", "always", tmp})
	sub := filepath.Join(tmp, "dir")
	os.Mkdir(sub, 0o755)
	code, out, _ = runWithArgs(t, []string{"xdu", "-C", "always", sub})
	if !strings.Contains(out, "\x1b[") {
		t.Errorf("always with dir should contain color: %q", out)
	}
}

func TestRun_ScanError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope123")
	code, out, errOut := runWithArgs(t, []string{"xdu", missing})
	if code != 2 {
		t.Errorf("scan error: code=%d want 2, out=%q err=%q", code, out, errOut)
	}
	if !strings.Contains(errOut, "error:") {
		t.Errorf("scan error stderr=%q want 'error:'", errOut)
	}
	if !strings.Contains(errOut, missing) {
		t.Errorf("stderr should contain missing path: %q", errOut)
	}
	if out != "" {
		t.Logf("scan error out=%q", out)
	}
	tmp := t.TempDir()
	okFile := filepath.Join(tmp, "ok")
	os.WriteFile(okFile, []byte("x"), 0o644)
	code, out, errOut = runWithArgs(t, []string{"xdu", okFile, missing})
	if code != 2 {
		t.Errorf("mixed: code=%d want 2", code)
	}
	if !strings.Contains(out, "ok") {
		t.Errorf("mixed out should contain ok: %q", out)
	}
	if !strings.Contains(errOut, "error:") {
		t.Errorf("mixed err=%q", errOut)
	}
	code, out, errOut = runWithArgs(t, []string{"xdu", "-t", tmp, missing})
	if code != 2 {
		t.Errorf("mixed tree: code=%d want 2", code)
	}
	if !strings.Contains(out, "ok") {
		t.Errorf("mixed tree out should contain ok: %q", out)
	}
}

func TestRun_MultiplePaths(t *testing.T) {
	tmp1 := t.TempDir()
	tmp2 := t.TempDir()
	f1 := filepath.Join(tmp1, "a")
	f2 := filepath.Join(tmp2, "b")
	os.WriteFile(f1, []byte("x"), 0o644)
	os.WriteFile(f2, []byte("yy"), 0o644)
	code, out, _ := runWithArgs(t, []string{"xdu", f1, f2})
	if code != 0 {
		t.Fatal(code)
	}
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Errorf("multi paths files: %q", out)
	}
	code, out, _ = runWithArgs(t, []string{"xdu", tmp1, tmp2})
	if code != 0 {
		t.Fatal(code)
	}
	if !strings.Contains(out, tmp1) || !strings.Contains(out, tmp2) {
		t.Errorf("multi paths dirs: %q", out)
	}
}

func TestRun_Depth(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	subsub := filepath.Join(sub, "subsub")
	os.MkdirAll(subsub, 0o755)
	os.WriteFile(filepath.Join(subsub, "deep"), []byte("x"), 0o644)
	code, out, _ := runWithArgs(t, []string{"xdu", "-d", "0", tmp})
	if code != 0 {
		t.Fatal(code)
	}
	if strings.Contains(out, "subsub") || strings.Contains(out, "deep") {
		t.Errorf("depth 0 should not contain deep: %q", out)
	}
	code, out, _ = runWithArgs(t, []string{"xdu", "-d", "1", "-t", tmp})
	if !strings.Contains(out, "sub") {
		t.Errorf("depth1 tree should contain sub: %q", out)
	}
	if strings.Contains(out, "deep") {
		t.Errorf("depth1 should not contain deep: %q", out)
	}
}

func TestRun_EmptyArgs(t *testing.T) {
	code, out, errOut := runWithArgs(t, []string{})
	if code != 0 && code != 2 {
		t.Errorf("empty args code=%d out=%q err=%q", code, out, errOut)
	}
}

func TestVersionConst(t *testing.T) {
	if Version == "" {
		t.Error("Version empty")
	}
	if Version != "0.1.0" {
		t.Errorf("Version=%q want 0.1.0", Version)
	}
}

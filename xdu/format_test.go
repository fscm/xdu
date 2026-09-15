// -*- mode: go; coding: utf-8 -*-
//
// SPDX-FileCopyrightText: 2026 Frederico Martins
// SPDX-License-Identifier: GPL-3.0-only

package xdu

import (
	"math"
	"strconv"
	"testing"
)

func TestFormatSize_Raw(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{512, "512"},
		{1023, "1023"},
		{1024, "1024"},
		{1536, "1536"},
		{-1, "-1"},
		{-1024, "-1024"},
		{math.MaxInt64, strconv.FormatInt(math.MaxInt64, 10)},
		{math.MinInt64, strconv.FormatInt(math.MinInt64, 10)},
	}
	for _, tc := range tests {
		got := FormatSize(tc.bytes, false)
		if got != tc.want {
			t.Errorf(
				"FormatSize(%d, human=false)=%q want %q",
				tc.bytes,
				got,
				tc.want,
			)
		}
		node := &Node{Size: tc.bytes, Inodes: 999}
		got2 := FormatUsage(node, false, false)
		if got2 != tc.want {
			t.Errorf(
				"FormatUsage(Size=%d, human=false,inodes=false)=%q want %q",
				tc.bytes,
				got2,
				tc.want,
			)
		}
	}
}

func TestFormatSize_Human_Bytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{-5, "-5 B"},
		{-1024, "-1024 B"},
	}
	for _, tc := range tests {
		got := FormatSize(tc.bytes, true)
		if got != tc.want {
			t.Errorf(
				"FormatSize(%d, human=true)=%q want %q",
				tc.bytes,
				got,
				tc.want,
			)
		}
	}
}

func TestFormatSize_Human_Units(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{1024, "1.0 KiB"},
		{1025, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{2048, "2.0 KiB"},
		{1024 * 1024, "1.0 MiB"},
		{1024*1024 + 512*1024, "1.5 MiB"},
		{1048576 + 1, "1.0 MiB"},
		{1024 * 1024 * 1024, "1.0 GiB"},
		{int64(1) << 40, "1.0 TiB"},
		{int64(1) << 50, "1.0 PiB"},
		{int64(1) << 60, "1.0 EiB"},
		{5 * (1 << 10), "5.0 KiB"},
		{10 * (1 << 20), "10.0 MiB"},
	}
	for _, tc := range tests {
		got := FormatSize(tc.bytes, true)
		if got != tc.want {
			t.Errorf(
				"FormatSize(%d, human=true)=%q want %q",
				tc.bytes,
				got,
				tc.want,
			)
		}
	}
}

func TestFormatSize_Human_LargeEiB(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{2 * (1 << 60), "2.0 EiB"},
		{math.MaxInt64, "8.0 EiB"},
		{math.MaxInt64 - 1, "8.0 EiB"},
		{(1 << 60) + (1 << 59), "1.5 EiB"},
		{int64(1536) * (1 << 50), "1.5 EiB"},
	}
	for _, tc := range tests {
		got := FormatSize(tc.bytes, true)
		if got != tc.want {
			t.Errorf(
				"FormatSize(%d, human=true)=%q want %q",
				tc.bytes,
				got,
				tc.want,
			)
		}
	}
	if got := FormatSize(7*(1<<60), true); got != "7.0 EiB" {
		t.Errorf("FormatSize(7EiB)=%q want %q", got, "7.0 EiB")
	}
}

func TestFormatSize_Human_Rounding(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{1075, "1.0 KiB"},
		{2560, "2.5 KiB"},
		{3072, "3.0 KiB"},
	}
	for _, tc := range tests {
		got := FormatSize(tc.bytes, true)
		if got != tc.want {
			t.Errorf("FormatSize(%d)=%q want %q", tc.bytes, got, tc.want)
		}
	}
}

func TestFormatUsage(t *testing.T) {
	tests := []struct {
		name   string
		node   *Node
		human  bool
		inodes bool
		want   string
	}{
		{
			"inodes true ignores human",
			&Node{Size: 9999, Inodes: 42},
			false,
			true,
			"42",
		},
		{
			"inodes true with human true",
			&Node{Size: 9999, Inodes: 7},
			true,
			true,
			"7",
		},
		{"inodes zero", &Node{Inodes: 0}, true, true, "0"},
		{"size raw", &Node{Size: 1234, Inodes: 99}, false, false, "1234"},
		{
			"size human bytes",
			&Node{Size: 512, Inodes: 99},
			true,
			false,
			"512 B",
		},
		{
			"size human KiB",
			&Node{Size: 2048, Inodes: 99},
			true,
			false,
			"2.0 KiB",
		},
		{
			"size human MiB",
			&Node{Size: 1 << 20, Inodes: 1},
			true,
			false,
			"1.0 MiB",
		},
		{"size zero human", &Node{Size: 0}, true, false, "0 B"},
		{"size zero raw", &Node{Size: 0}, false, false, "0"},
		{"negative size human", &Node{Size: -5}, true, false, "-5 B"},
	}
	for _, tc := range tests {
		got := FormatUsage(tc.node, tc.human, tc.inodes)
		if got != tc.want {
			t.Errorf(
				"%s: FormatUsage(Size=%d Inodes=%d human=%v inodes=%v)=%q want %q",
				tc.name,
				tc.node.Size,
				tc.node.Inodes,
				tc.human,
				tc.inodes,
				got,
				tc.want,
			)
		}
	}
}

func TestFormatSize_SizeUnits(t *testing.T) {
	if len(sizeUnits) != 7 {
		t.Fatalf("sizeUnits len=%d want 7", len(sizeUnits))
	}
	want := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	for i, w := range want {
		if sizeUnits[i] != w {
			t.Errorf("sizeUnits[%d]=%q want %q", i, sizeUnits[i], w)
		}
	}
}

func TestFormatUsage_Delegates(t *testing.T) {
	cases := []struct {
		size   int64
		human  bool
		inodes bool
	}{
		{0, false, false},
		{1024, true, false},
		{2048, true, false},
		{500, true, true},
		{999, false, true},
	}
	for _, c := range cases {
		node := &Node{Size: c.size, Inodes: c.size + 100}
		want := ""
		if c.inodes {
			want = strconv.FormatInt(node.Inodes, 10)
		} else {
			want = FormatSize(node.Size, c.human)
		}
		got := FormatUsage(node, c.human, c.inodes)
		if got != want {
			t.Errorf(
				"FormatUsage size=%d human=%v inodes=%v: got %q want %q",
				c.size,
				c.human,
				c.inodes,
				got,
				want,
			)
		}
	}
}

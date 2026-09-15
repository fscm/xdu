#!/usr/bin/env -S just --justfile
# -*- mode: just; coding: utf-8 -*-
#
# SPDX-FileCopyrightText: 2026 Frederico Martins
# SPDX-License-Identifier: GPL-3.0-only

set unstable # for 'which()'

PROJECT_DIR := absolute_path(justfile_directory())
PROJECT_NAME := file_name(PROJECT_DIR)
BUILD_DIR := join(PROJECT_DIR, 'build')
DIST_DIR := join(PROJECT_DIR, 'dist')
DOCS_DIR := join(PROJECT_DIR, 'docs')

export TMPDIR := join(PROJECT_DIR, 'tmp')
export GOCACHE := join(PROJECT_DIR, 'cache')
export GOTMPDIR := join(PROJECT_DIR, 'tmp')

COVERAGE_DATA := join(TMPDIR, 'coverage.dat')
COVERAGE_REPORT := join(DOCS_DIR, 'coverage-report.html')

GO := require('go')
MINGO := require('mingo')
MANDOC := require('mandoc')
TAR := require('tar')
FORMAT := require(if which('gofumpt') != '' { 'gofumpt' } else { 'gofmt' })

MKDIR := '\mkdir -p'
RM := '\rm -rf'

BUILD_ARGS := '-trimpath'
FORMAT_ARGS := '-l -w'
FORMAT_SHOW_ARGS := '-d'
MANDOC_ARGS := '-T lint'
MINGO_ARGS := '-tests' # '-v'
TEST_ARGS := '-cover -coverprofile ' + COVERAGE_DATA # '-v'

LDFLAGS := '-s -w'

[private]
default: help

# Builds the project.
build os arch:
    @{{ MKDIR }} "{{ GOTMPDIR }}"
    # Cleaning build artifacts for {{ os }}/{{ arch }}...
    @{{ RM }} "{{ join(BUILD_DIR, os, arch) }}"
    # Building artifacts for {{ os }}/{{ arch }}...
    @{{ MKDIR }} "{{ join(BUILD_DIR, os, arch) }}"
    @CGO_ENABLED=0 GOOS={{ os }} GOARCH={{ arch }} {{ GO }} build \
        {{ BUILD_ARGS }} \
        -ldflags="{{ LDFLAGS }}" \
        -o "{{ join(BUILD_DIR, os, arch, PROJECT_NAME) }}" \
        "{{ PROJECT_DIR }}"

# Builds the project for all supported systems.
build-all: \
    (build 'darwin' 'amd64') \
    (build 'darwin' 'arm64') \
    (build 'linux' 'amd64') \
    (build 'linux' 'arm64')

# Cleans all build artifacts.
[private]
build-clean:
    # Cleaning build artifacts...
    @{{ RM }} "{{ BUILD_DIR }}"

# Cleans the project builds.
clean: build-clean dist-clean

# Cleans everything.
[confirm('Are you sure you want to clean everything?')]
clean-all: clean clean-cache docs-clean

# Cleans the build and run cache artifacts.
clean-cache:
    # Cleaning caches...
    @{{ GO }} clean -cache
    @{{ RM }} "{{ GOCACHE }}"
    # Cleaning temporary files...
    @{{ RM }} "{{ GOTMPDIR }}"

# Creates distribution package.
dist os arch:
    # Cleaning distribution artifacts for {{ os }}/{{ arch }}...
    @{{ RM }} "{{ join(DIST_DIR, PROJECT_NAME) }}-{{ os }}-{{ arch }}"*
    # Building distribution package for {{ os }}-{{ arch }}...
    @{{ \
        if path_exists(join(TMPDIR, PROJECT_NAME)) != 'true' { \
            f'just build {{os}} {{arch}}' \
        } else { \
            '' \
        } \
    }}
    @{{ MKDIR }} "{{ DIST_DIR }}"
    @{{ TAR }} -c -z \
        -f "{{ join(DIST_DIR, PROJECT_NAME) }}-{{ os }}-{{ arch }}.tar.gz" \
        -C "{{ join(BUILD_DIR, os, arch) }}" "{{ PROJECT_NAME }}" \
        -C "{{ PROJECT_DIR }}" "{{ PROJECT_NAME }}.1"


# Creates distribution packages for all supported systems.
dist-all: \
    (dist 'darwin' 'amd64') \
    (dist 'darwin' 'arm64') \
    (dist 'linux' 'amd64') \
    (dist 'linux' 'arm64')

# Cleans all build artifacts.
[private]
dist-clean:
    # Cleaning distribution artifacts...
    @{{ RM }} "{{ DIST_DIR }}"

# Creates the project documentation.
docs: docs-clean
    # Generating documentation...
    @{{ MKDIR }} "{{ DOCS_DIR }}"
    @{{ GO }} doc "{{ PROJECT_DIR }}" \
        > "{{ join(DOCS_DIR, PROJECT_NAME) }}.txt"

# Cleans the documentation folder.
[private]
docs-clean:
    # Cleaning documentation...
    @{{ RM }} "{{ DOCS_DIR }}"

# Formats the code.
format:
    # Formating the code...
    @{{ FORMAT }} {{ FORMAT_ARGS }} "{{ PROJECT_DIR }}"

# Shows the required code formats.
format-show:
    # Getting required code formats...
    @-{{ FORMAT }} {{ FORMAT_SHOW_ARGS }} "{{ PROJECT_DIR }}"

# Calculates Go minimum version required.
goversion:
    # Finding minimum Go version...
    @{{ MKDIR }} "{{ GOTMPDIR }}"
    @{{ MINGO }} {{ MINGO_ARGS }} "{{ PROJECT_DIR }}"

# Shows this help message.
help:
    @just --list
    @echo ""
    @echo "Supported builds (os/arch):"
    @echo "    darwin/amd64"
    @echo "    darwin/arm64"
    @echo "    linux/amd64"
    @echo "    linux/arm64"

# Checks the project for code smells ('format' and 'vet').
lint: && format-show vet
    # Checking the code...

# Checks the man page for issues/errors.
lint-man:
    # Checking the man page...
    @{{ MANDOC }} {{ MANDOC_ARGS }} "{{ join(PROJECT_DIR, PROJECT_NAME) }}.1"

# Runs the app.
run *args:
    @{{ MKDIR }} "{{ GOTMPDIR }}"
    @-{{ GO }} run "{{ PROJECT_DIR }}" {{ args }}

# Runs the tests.
test *tests:
    @{{ MKDIR }} "{{ GOTMPDIR }}"
    # Running tests... {{ tests }}
    @-{{ GO }} test {{ TEST_ARGS }} "{{ join(PROJECT_DIR, "...") }}" \
    {{ if tests != '' { f'-run {{tests}}' } else { '' } }}

# Creates the tests coverage report (html)
test-coverage:
    # Creating tests report coverage...
    @{{ if path_exists(COVERAGE_DATA) != 'true' { 'just test' } else { '' } }}
    @{{ MKDIR }} "{{ DOCS_DIR }}"
    @{{ GO }} tool cover -html="{{ COVERAGE_DATA }}" -o "{{ COVERAGE_REPORT }}"

# Examines the code for suspicious constructs.
vet:
    @{{ MKDIR }} "{{ GOTMPDIR }}"
    # Examining the code...
    @{{ GO }} vet "{{ join(PROJECT_DIR, "...") }}"

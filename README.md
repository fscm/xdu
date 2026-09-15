# xdu

`xdu` merges the behavior of `du`, `sort`, and `tree` into one static binary.
It walks one or more paths, sums apparent sizes or inode counts for each
directory, and can show the result as a flat list or as a hierarchical tree.

## Synopsis

For each path given on the command line (or `.` when no path is given), `xdu`
builds a tree of nodes. Directories are read and walked recursively. Each node
stores the apparent size, an inode count, and flags for directory, symlink, and
executable.

A directory node aggregates its descendants. In byte mode the size of a
directory is the sum of the sizes of everything under it. The directory entry
itself does not add extra bytes. In inode mode (`-i`) a directory counts as 1
plus the count of all descendants. Symlinks are never followed. A symlink
contributes its `Lstat` size and one inode, even when it points to a directory.
Hard links are counted once per directory entry. Permission or I/O errors are
kept as warnings and do not stop the scan of other paths.

After the scan `xdu` either prints a flat list of directories (and any
non-directory paths) or renders a tree. Both outputs can be sorted by size or
inode count, limited by depth, printed with binary human-readable units,
colored by file type, and followed by a grand total. Output is deterministic
because directory entries are sorted lexically before aggregation, and totals
still include pruned or filtered descendants.

### Note (apparent sizes and symlinks)

`xdu` reports apparent sizes from `Lstat`, not allocated blocks on disk. This
matches `du --apparent-size` rather than the default `du` block count.
Directory totals are the sum of included descendants.
Symlinked directories are shown as links and are not recursed, so a symlink
loop cannot occur.

### Features

- Sums apparent sizes and inode counts for files and directories (the same
  values `du` reports with `--apparent-size` and `--inodes`).
- Shows results as a flat list or as a hierarchical tree with box-drawing
  characters.
- Sorts by size or inode count with `-s` (ascending, descending with `-r`)
  and reverses lexical order when `-r` is used without `-s`.
- Prints sizes in binary human-readable form (`B`, `KiB`, `MiB`, ...) with
  `-h`.
- Colors directories, symlinks, and executables, with `auto`, `always`, and
  `never` modes and `NO_COLOR` support.
- Limits visible depth with `-d` without changing computed totals.
- Includes or skips hidden entries (basename starting with `.`) with `-a`. An
  explicitly given hidden path is always scanned.
- Prints a grand total over all paths with `-c`.
- Keeps scan warnings on stderr and returns a partial result instead of failing
  outright.
- Zero external dependencies (standard library only, no shell-out to `du`,
  `sort`, or `tree`).
- Cross-compiled to a static binary with `CGO_ENABLED=0` for `darwin/amd64`,
  `darwin/arm64`, `linux/amd64`, and `linux/arm64`.

## Usage

`xdu` runs with no required flag. With no path it scans the current directory.

```
xdu [options] [path ...]
```

### Program Options

* `-a, --all` - Include hidden files and directories (those whose basename
  starts with `.`). Without it hidden entries and their descendants are
  skipped. A hidden path given explicitly on the command line is still scanned.
* `-c, --total` - Print a grand total after all results as `<value><TAB>total`.
  Overlapping paths are summed independently, matching `du`.
* `-C MODE, --color MODE` - Color the output. `MODE` is `auto` (default),
  `always`, or `never`. In `auto` mode color is used only when stdout is a
  terminal and `NO_COLOR` is empty. `--color=MODE` is also accepted.
* `-d DEPTH, --depth DEPTH` - Max visible depth. `0` shows only paths, `-1`
  (default) is unlimited. This affects display only, not the totals that are
  computed.
* `-h, --human-readable` - Print sizes in human-readable form using binary
  units (`B`, `KiB`, `MiB`, ...). Ignored when combined with `-i`.
* `--help` - Show a help message and exit.
* `-i, --inodes` - Show inode counts instead of bytes. Each visible entry
  counts as 1, a directory adds its descendants.
* `-r, --reverse` - Reverse sort order. With `-s` it sorts largest first
  (`desc`). Without `-s` it reverses lexical order.
* `-s, --sort` - Sort by size (or inode count with `-i`), smallest first
  (`asc`). Without `-s` the output is ordered lexically.
* `-t, --tree` - Render as a tree instead of a flat list.
* `-v, --version` - Show the program's version number and exit.

### Examples

List the current directory (like `du`):

```shell
xdu
```

List with human-readable sizes for a specific path:

```shell
xdu -h .
```

List inode counts and print a grand total:

```shell
xdu -i -c my-folder
```

Render a tree:

```shell
xdu -t my-folder
```

Render a tree that includes hidden entries:

```shell
xdu -t -a my-folder
```

Limit tree depth to one level:

```shell
xdu -t -d 1 my-folder
```

Sort by size in the tree, largest first, with human-readable sizes and a grand
total:

```shell
xdu -t -s -r -c -h my-folder
```

Control color:

```shell
xdu --color never -t my-folder
xdu --color always -t my-folder
NO_COLOR=1 xdu -t my-folder
```

Hidden file handling:

```shell
xdu my-folder        # hidden files excluded
xdu -a my-folder     # hidden files included
xdu .hidden-folder   # explicitly given hidden path is always scanned
```

## Supported Platforms

`xdu` uses only the Go standard library, so it runs anywhere Go runs. Release
binaries are statically linked and have no runtime dependencies.

| OS    | Architecture |
|-------|--------------|
| Linux | x86_64       |
| Linux | arm64        |
| macOS | x86_64       |
| macOS | arm64        |

## Build (from source)

Golang (version 1.22.0 or above) needs to be installed on your local computer.
Golang setup can be found at [go.dev](https://go.dev).

Just (version 1.46.0 or above) needs to be installed on your local computer.
Just is used to automate several steps of the development process. Just setup
can be found at [just.systems](https://just.systems).

All of the commands described below are to be executed on the root folder of
this project.

To build the `xdu` binaries use the following command:

```shell
just build-all
```

To create distribution archives of the `xdu` program use the following command:

```shell
just dist-all
```

## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request

## Versioning

This project uses [SemVer](http://semver.org/) for versioning. For the versions
available, see the [tags](https://github.com/fscm/xdu/tags) on this repository.

## Authors

* **Frederico Martins** - [fscm](https://github.com/fscm)

See also the list of [contributors](https://github.com/fscm/xdu/contributors)
who participated in this project.

## License

This project is licensed under the GPLv3 License - see the [LICENSE](LICENSE)
file for details

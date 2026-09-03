<!-- SPDX-License-Identifier: Apache-2.0 -->

# The container

Everything below is a measurement on files written by Vivado 2025.2.
Reproduce any of it with:

```sh
bazel build //hdl/corpus:all_wdb
bazel run //cmd/wdbcvt -- -dump -in "$PWD/bazel-bin/hdl/corpus/t2_flat3________/sim.wdb"
```

All integers are little endian.
Offsets are from the start of the file unless a section says otherwise.


## The shape of a file

A database is one flat file.
The `xsim.dir` tree beside it is not needed to read it; the decoder
reads nothing else.
The sections appear in this order in every corpus case:

| Section | Where | What it holds |
| :--- | :--- | :--- |
| fixed header | `0x00` to `0xc8` | magic, producer, directory pointers |
| arena table | `0xc8` to the trailer | file offsets of the arena records |
| trailer | `0x48` bytes | end time, marker offset, page size |
| directory | three 48 byte entries | `WDB.Event`, `Xilinx RTTI`, `Xilinx DBG` |
| type table | `Xilinx ISim TYPE FILE 001` | [types.md](types.md) |
| debug section | `Xilinx ISim DBG 006` | [hierarchy.md](hierarchy.md) |
| instance records | after the debug section | one per object, holds its handle |
| page directory | one `0x4c0` byte record per arena | offsets and lengths of the value pages |
| marker | 16 bytes | `[uint64 0][uint64 N]` |
| value pages | zlib streams | [values.md](values.md) |

The marker sits between the page directory and the first page in a
small simulation.
When a page filled up and was flushed before the simulation ended, that
page comes first and the marker follows it.
See [values.md](values.md).


## The fixed header

| Offset | Len | Content | Found by | Confirmed by |
| :--- | ---: | :--- | :--- | :--- |
| `0x00` | 24 | `Xilinx WAVE DATABASE 01`, NUL terminated | hex dump of any database | all 49 cases |
| `0x18` | 24 | `Xilinx Simulator`, NUL terminated | same | all 49 cases |
| `0x30` | 8 | `uint64` `0x40` | same | constant in all 49 cases |
| `0x38` | 4 | `uint32` Unix time the database was written | the noise mask, two runs of `t3_tr1` | decodes to the file's own mtime |
| `0x48` | 24 | three `uint64` file offsets of directory entries | following the values: each lands on a NUL terminated section name | all 49 cases |
| `0x98` | 12 | three `uint32` `0x30` | hex dump | constant in all 49 cases |
| `0xc0` | 4 | `uint32` `3` | hex dump | constant in all 49 cases |
| `0xc4` | 4 | `uint32` per run duration, noise | the noise mask | differs between two runs of one case |

The three `uint32` at `0x98` and the `3` at `0xc0` have not moved in any
case.
Their meaning is not known and they are left as constants.


## The arena table

From `0xc8` the header holds a table of `uint64` slots.
Slot `i` is the file offset of arena record `i` in the page directory,
or `0` when the design has no arena `i`.
Arenas are explained in [values.md](values.md): an arena is a window of
`0x800` object handles, and every object whose records are logged
belongs to one.

The table's length varies, and that is the one place the header is not
fixed.
The trailer starts where the table ends, and the first directory
pointer at `0x48` names the `WDB.Event` entry that follows the trailer,
so the table's length is `(pointer - 0x48 - 0xc8) / 8`.

| Case | Objects | Arenas | Slots |
| :--- | ---: | ---: | ---: |
| every case with 1 to 5 objects | 1 to 5 | 1 to 3 | 3 |
| `t5_sig10` | 10 | 2 | 4 |
| `t6_sig12` | 12 | 3 | 4 |
| `t6_sig20` | 20 | 4 | 6 |

*Found by* `t5_sig10` against `t2_flat3`: the trailer and every
directory entry sat 8 bytes later, and the `0x48` pointer said so.
*Confirmed by* `t6_sig05`, `t6_sig12` and `t6_sig20`, written for the
purpose, which the decoder reads with no special case.

The slot count is not the arena count: `t6_sig05` has two arenas and
three slots, `t5_sig10` two arenas and four.
It fits `max(3, ceil(objects / 4) + 1)` in every case, and that is a
guess recorded in the open questions, not a finding.

Before the table was understood, the `0xd0` word was recorded as "a file
offset, non-zero only in multi signal designs".
That reading was right as far as it went: `0xd0` is slot 1, and a second
arena appears with the second signal.


## The trailer

The trailer is `0x48` bytes and follows the arena table.
Offsets here are relative to the trailer, which starts at `0xe0` in a
three slot header.

| Offset | Len | Content | Found by | Confirmed by |
| :--- | ---: | :--- | :--- | :--- |
| `0x00` | 8 | `uint64` simulation end time in picoseconds | the correlation sweep across all cases | 49 of 49 against `end_time_ns` in `truth.json` |
| `0x08` | 4 | `uint32` `0x3e9` | hex dump | constant in all 49 cases |
| `0x0c` | 4 | `uint32` `3` | hex dump | constant in all 49 cases |
| `0x10` | 8 | `uint64` `0x800` | hex dump | constant in all 49 cases |
| `0x18` | 8 | `uint64` that grows with the object count | comparing cases | open; see below |
| `0x20` | 4 | `uint32` `0xc8` | hex dump | constant in all 49 cases |
| `0x24` | 4 | `uint32` `0` | hex dump | constant in all 49 cases |
| `0x28` | 8 | `uint64` `0` | hex dump | constant in all 49 cases |
| `0x30` | 8 | `uint64` `1` when any object is logged, `0` in `t0_nosig` | the correlation sweep | `0` for `t0_nosig` alone |
| `0x38` | 8 | `uint64` file offset of the marker, `0` in `t0_nosig` | `t5_tr1000`: the only word holding `0x1bac`, the end of the flushed page | 48 of 48 cases with a marker |
| `0x40` | 4 | `uint32` `0x2800`, the size a value page inflates to | inflating a page | every page in every case inflates to 10240 bytes |
| `0x44` | 4 | `uint32` `0x64` | hex dump | constant in all 49 cases |

The word at `0x18` is `0x11d0` for one bit with one edge, `0x1318` for
two bits, and `0x2a38` for twenty.
It grows by about `0x148` per signal object and by a few bytes per type
entry.
It is larger than the file in every case, so it is not a file offset.
It is left open.


## The directory

Each pointer at `0x48` names a 48 byte entry:

```
[24 bytes name, NUL terminated][uint64 count][uint64 offset][uint64 length]
```

| Entry | Count | Offset | Length |
| :--- | ---: | :--- | :--- |
| `WDB.Event` | 1 | `0xe0` in a three slot header, the trailer | `0x48` |
| `Xilinx RTTI` | 1 | the type table magic | to the end of the type table's offset list |
| `Xilinx DBG` | 1 | the debug section magic | the section plus its instance records |

Each entry sits at `offset + length`, directly after the section it
describes: the `WDB.Event` entry at `0x128` after the trailer, the
`Xilinx RTTI` entry after the type table's offset list, and the
`Xilinx DBG` entry after the last instance record.
So a section is written first and its entry second, and the entry's
pointer in the header is filled in when the entry is known.

The page directory follows the `Xilinx DBG` entry directly.
Arena record `i` has been at `entry + 48 + 0x4c0 * i` in every case.


## The marker

Sixteen bytes, `[uint64 0][uint64 N]`, at the offset the trailer word
`0x38` names.

`N` was first read as the object count minus one, which held for 43
cases.
`t6_var_int` broke it: two objects, one a declared process variable
with no records, and `N` is `0`.
`N` is the number of objects that have at least one record, minus one,
in all 48 cases with a marker.
The decoder checks that and refuses a file where it does not hold.

*Found by* `t2_flat3` against `t1_two_bits`: `N` was `2` and `1`.
*Corrected by* `t6_var_int`.
*Confirmed by* `t5_tr1000`, where the marker is not before the first
page but after it.


## Sizes, for orientation

A database of one bit with one edge is 3677 bytes.
Of that, the fixed header, arena table, trailer and directory are 400
bytes, the type table is a few hundred, the debug section with its
embedded source paths is about 2000, and the single value page is 98
bytes compressed.

The embedded paths dominate.
They include the path of the Vivado installation and the build machine
paths AMD compiled the standard libraries on, in the clear.
That is why every corpus case directory name is padded to one length;
see [../corpus.md](../corpus.md).


## The noise mask

Two runs of one case differ only in clocks and durations.
The union over several pairs, measured on `t4_gen_diff_two`:

| Offset | Bytes | What it holds |
| :--- | :--- | :--- |
| `0x38` | 4 | the Unix timestamp in the header |
| `0xc4` | 2 | a per run duration |
| `0x172` | 3 | varying bytes in the type table header |
| `0x23f` | 1 | a byte of the type table timestamp |
| `0x50c` | 3 | the debug section timestamp |
| `0x564` | 3 | a second copy of it |
| `0xb2f` to `0xbd8` | 4 runs of 2 to 4 bytes | the noise word of each declaration record |

The value pages are byte for byte identical between runs.
Everything outside the mask is deterministic, which is what makes a
pairwise diff between two cases meaningful.
A byte of a timestamp only shows as noise when its two values differ,
so build the mask from several pairs and take the union.
`tools/noise_mask.sh` does the two runs.

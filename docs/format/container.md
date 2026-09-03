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
| `0x00` | 24 | `Xilinx WAVE DATABASE 01`, NUL terminated | hex dump of any database | all 880 cases |
| `0x18` | 24 | `Xilinx Simulator`, NUL terminated | same | all 880 cases |
| `0x30` | 8 | `uint64` `0x40` | same | constant in all 880 cases |
| `0x38` | 4 | `uint32` Unix time the database was written | the noise mask, two runs of `t3_tr1` | decodes to the file's own mtime |
| `0x48` | 24 | three `uint64` file offsets of directory entries | following the values: each lands on a NUL terminated section name | all 880 cases |
| `0x98` | 12 | three `uint32` `0x30` | hex dump | constant in all 880 cases |
| `0xc0` | 4 | `uint32` `3` | hex dump | constant in all 880 cases |
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

| Case | Objects | Handle space | Arenas | Slots |
| :--- | ---: | ---: | ---: | ---: |
| `t0_nosig` | 0 | `0x1088` | 0 | 3 |
| `t1_bit_one_edge` | 1 | `0x11d0` | 1 | 3 |
| `t6_sig05` | 5 | `0x16f0` | 2 | 3 |
| `t7_sig07` | 7 | `0x1988` | 2 | 4 |
| `t5_sig10` | 10 | `0x1d60` | 2 | 4 |
| `t6_sig12` | 12 | `0x1ff0` | 3 | 4 |
| `t7_sig14` | 14 | `0x2280` | 3 | 5 |
| `t7_sig16` | 16 | `0x2518` | 3 | 5 |
| `t6_sig20` | 20 | `0x2a38` | 4 | 6 |
| `t7_sig24` | 24 | `0x2f60` | 4 | 6 |
| `t9_vec4096` | 2 | `0x72d8` | 6 | 15 |
| `t9_vec12000` | 1 | `0x9e50` | 7 | 20 |
| `t46_sig_1000____` | 1000 | `0x3d9e8` | 95 | 124 |
| `t46_deep_100____` | 202 | `0xedd8` | 30 | 30 |
| `t46_v_gen_70000_` | 70000 | `0x108ed3c` | 6564 | 8478 |
| `t46_gen_70000___` | 140000 | `0x23710a0` | 17776 | 18147 |

The tier 46 rows are the large end: the slot count and every count and
index in the file are whole 32 bit words, and a 49 MB file with 18147
slots, 17776 arenas in use and 140000 objects reads by the same rules
as a 5 KB one.
*Found by* `t46_gen_70000___` against `t46_sig_1000____`.
*Confirmed by* `t46_v_gen_70000_`.

A slot can be `0` in the middle of the table as well as at its end.
`t9_pkg_sig` declares a signal in a package, and that signal takes the
first handle `0x768` but is never logged, so arena 0 is never written
and slot 0 is `0` while slot 1 names the arena of the testbench's own
signal.
A wide value spans arenas, see [values.md](values.md): the 12000 byte
signal of `t9_vec12000` sits on handle `0x768` and its chunks fill
arenas 0 to 6, while slots 7 to 19 are `0`.

The slot count is `ceil(handle space / 0x800)`, where the handle space
is the trailer word at `0x18`, described below.
The reader checks that rule on every file it opens, and it holds in all
880 corpus cases.

*Found by* `t5_sig10` against `t2_flat3`: the trailer and every
directory entry sat 8 bytes later, and the `0x48` pointer said so.
An earlier guess, `max(3, ceil(objects / 4) + 1)`, fitted every case
through tier 6 and failed on `t7_sig07`, which has four slots for seven
objects.
`t7_sig14`, `t7_sig16` and `t7_sig24` then pinned the boundaries at
`0x1800`, `0x2000` and `0x2800` of handle space.
*Confirmed by* the reader's check across all 880 cases, including
`t9_vec12000`, whose one object spans 20 slots.

The slot count is not the arena count: `t6_sig05` has two arenas and
three slots, `t5_sig10` two arenas and four.
Slots past the last arena hold `0`.

The records the slots point at are not in slot order.
They sit in the page directory in the order the arenas were first
written to.
`t7_gen_for` has slots `0x14b3 0x1973 0xff3`: arena 2, which holds the
generate loop indexes and the generics, was written at elaboration,
before the signals in arenas 0 and 1 had a record.
The reader accepts any order and checks that the records are
contiguous.

*Found by* `t7_gen_for`, the first case the reader refused, with
`arena record 0 at 0x14b3, want 0xff3`.

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
| `0x00` | 8 | `uint64` simulation end time, in the time unit the DBG section names, picoseconds in every case before tier 21 | the correlation sweep across all cases | 880 of 880 against `end_time_ns` in `truth.json`; `t21_v_ts_1ns_1ns` holds 100 for 100 ns |
| `0x08` | 4 | `uint32` `0x3e9` | hex dump | constant in all 880 cases |
| `0x0c` | 4 | `uint32` number of arena table slots | recorded as the constant `3` until the tier 7 sweep over every fixed word; it is 4, 5 and 6 in the signal count cases | the reader checks it against the table length in 880 of 880 cases |
| `0x10` | 8 | `uint64` `0x800` | hex dump | constant in all 880 cases; the arena span, by its value |
| `0x18` | 8 | `uint64` handle space, the bytes of handle space the objects occupy | comparing cases | the arena table rule above, 880 of 880 cases |
| `0x20` | 4 | `uint32` `0xc8` | hex dump | constant in all 880 cases; the arena table's offset, by its value |
| `0x24` | 4 | `uint32` `0` | hex dump | constant in all 880 cases |
| `0x28` | 8 | `uint64` `0` | hex dump | constant in all 880 cases |
| `0x30` | 8 | `uint64` number of logged ranges at the marker, `0` in `t0_nosig` | read as a flag until `t9_port_rec` held `2` | 880 of 880 cases, checked against the marker |
| `0x38` | 8 | `uint64` file offset of the marker, `0` in `t0_nosig` | `t5_tr1000`: the only word holding `0x1bac`, the end of the flushed page | 140 of 140 cases with a marker |
| `0x40` | 4 | `uint32` `0x2800`, the size a value page inflates to | inflating a page | every page in every case inflates to 10240 bytes |
| `0x44` | 4 | `uint32` `0x64` | hex dump | constant in all 880 cases |

The word at `0x18` is `0x11d0` for one bit with one edge, `0x1318` for
two bits, and `0x2a38` for twenty.
It is larger than the file in every case, so it is not a file offset.
It grows by `0x148` per one bit signal in the runs of the tier 6 and
tier 7 signal count cases, and each signal's handle is `0xf0` past the
previous one, so a signal costs `0x58` beyond its handle stride.
The arena table has one slot per `0x800` of it, which is what makes it
the size of the handle space.
Tier 46 splits the `0x58`: a signal nothing drives costs `0xf8`,
`t0_bit_const` at `0x1180` against the `0x1088` of `t0_nosig`, and a
driven one `0x148`, `t1_bit_one_edge` at `0x11d0`, so the driver is
`0x50`, of which `0x30` is in the handle stride and `0x20` is not.
The sum is exact at a thousand: `t46_sig_1000____`, two driven signals
and 998 undriven ones, has `0x1088 + 2 * 0x148 + 998 * 0xf8`, which is
its `0x3d9e8`.
Where the first `0x1088` and the `0x20` per driver go is open.
Ports cost handle space too: `t8_port_in`, one signal and one connected
port that shares its handle, has `0x1288`, `0xb8` more than the one
signal of `t1_bit_one_edge`, and `t8_port_open`, two open ports and no
signal, has `0x1418`.
The connected port takes no handle and still takes handle space, so the
word counts something more than handles.

Tier 9 adds a few more prices, each against the `0x11d0` of
`t1_bit_one_edge` or the `0x1288` of `t8_port_in`:

| Case | Handle space | Cost | Of what |
| :--- | ---: | ---: | :--- |
| `t9_alias` | `0x11d0` | `0` | an alias of the signal |
| `t9_func` | `0x11d8` | `0x8` | a function with a variable |
| `t9_proc_local` | `0x11d8` | `0x8` | a procedure with a variable, inside the process |
| `t9_proc_sig` | `0x1218` | `0x48` | a procedure with a `signal` parameter |
| `t9_comp` | `0x1288` | as `t8_port_in` | a component declaration and default binding |
| `t9_port_lnk` | `0x1288` | as `t8_port_in` | a `linkage` port |
| `t9_port_slice` | `0x1280` | | an `in` port bound to one element of a 2 bit vector |
| `t9_port_slice2` | `0x1274` | | a 2 bit `in` port bound to a slice of a 4 bit vector |
| `t9_pkg_sig` | `0x1430` | | a signal in a package, on handle `0x768`, never logged |
| `t9_port_expr` | `0x1380` | `0x1b0` | an `in` port bound to the literal `'1'`, on its own handle |
| `t9_block` | `0x13d8` | `0x208` | a block with a signal and an implicit process |
| `t9_vec4096` | `0x72d8` | | two 4096 byte signals, 15 slots |
| `t9_vec12000` | `0x9e50` | | one 12000 byte signal, 20 slots |
| `t46_drv_2_next__` | `0x1428` | `0x258` | a `std_logic` with two drivers, next to a driven `std_ulogic`; as `t24_two_drivers` |
| `t46_drv_3_next__` | `0x1510` | `0x340` | the same with three drivers; as `t34_res_3drv` |

A use clause costs handle space, and the type of the signal does not.
The `bit`, `boolean`, `character`, `integer`, `real` and `time` signals
of tier 2 cost `0x340` each over `t0_nosig`, `0x1f8` more than the
`0x148` of a `std_ulogic`, and the `std_logic_vector`, `signed` and
`unsigned` of 8 elements cost `0x338`, `0x1f8` more than the `0x140`
of the `std_ulogic_vector` of `t1_vec8`.
Those nine benches are the nine with `use ieee.numeric_std.all`.
Tier 47 adds that clause to `t1_bit_one_edge` and changes nothing
else, and the handle space is the `0x13c8` of `t2_bit`; a `bit` signal
under the usual `std_logic_1164` clause, `t47_use_1164_bit`, costs the
`0x11d0` of a `std_ulogic`.
The prices, each over `t1_bit_one_edge`:

| Case | Handle space | Cost | Of what |
| :--- | ---: | ---: | :--- |
| `t47_use_none____` | `0xbcc` | `-0x604` | no `library ieee` and no use clause |
| `t47_use_lib_only` | `0xbcc` | `-0x604` | `library ieee;` and no use clause |
| `t47_use_one_name` | `0x11d0` | `0` | `use ieee.std_logic_1164.std_ulogic;`, the price of `.all` |
| `t47_use_textio__` | `0x11d0` | `0` | `use std.textio.all`, a package the file table already lists |
| `t47_use_numbit__` | `0x1358` | `0x188` | `use ieee.numeric_bit.all` |
| `t47_use_numstd__` | `0x13c8` | `0x1f8` | `use ieee.numeric_std.all` |
| `t47_use_mathrl__` | `0x15d0` | `0x400` | `use ieee.math_real.all` |
| `t47_use_pkg_emp_` | `0x1250` | `0x80` | an empty package of the design, its scope and unit |
| `t47_use_pkg_typ_` | `0x1250` | `0x80` | a package with one subtype |
| `t47_use_pkg_4arr` | `0x1250` | `0x80` | a package with four array types |
| `t47_use_pkg_fn2_` | `0x1250` | `0x80` | a package with two functions and a body |
| `t47_use_pkg_pr2_` | `0x1250` | `0x80` | a package with two procedures and a body |
| `t47_use_pkg_two_` | `0x1258` | `0x88` | a package with an `integer` and a `std_ulogic` constant, 5 bytes rounded to 8 |
| `t47_use_pkg_nul_` | `0x1270` | `0xa0` | a package with two null range vector constants, 16 each |

So `library ieee` with `use ieee.std_logic_1164.all` is `0x604` of
the `0x1088` of `t0_nosig`, a package of the design is `0x80` for its
scope and unit plus the rounded storage of its constants, and its
types and subprograms are free.
What a library package costs is not its types or subprograms by that
rule, and not its constants either: `numeric_std` declares two null
range constants, which cost `0x20` in a package of the design, and
takes `0x1f8`.
That is data without a reading.
*Found by* `t47_use_numstd__` against `t1_bit_one_edge_`.
*Confirmed by* the rest of the table.

Three trailer words describe the arena table together: `0xc8` at `0x20`
is where it starts, the word at `0x0c` is how many slots it has, and
`0x800` at `0x10` is how much handle space each slot covers.
The first and last of those are read off their values and are constant,
so that is a consistent reading and not a finding.

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

An arena record is `0x4c0` bytes:

| Offset | Len | Content |
| :--- | ---: | :--- |
| `0x000` | 8 | `uint64` file offset of a continuation record, `0` when there is none |
| `0x008` | 800 | 100 `uint64` page file offsets |
| `0x328` | 400 | 100 `uint32` compressed page lengths |
| `0x4b8` | 8 | `uint64` number of pages listed in this record |

An arena with more than 100 pages continues in another record of the
same layout, named by word 0.
The continuation is written among the pages, right after the hundredth
page, and a record with a continuation lists exactly 100 pages.
`t9_tr70000`, 70001 records over 117 pages, has its arena record at
`0x9bb` with word 0 `0x35a4d`, page 100 at `0x351fc`, the continuation
at `0x35a4d` listing 17 pages with word 0 `0`, and page 101 at
`0x35f0d`.
The last page again follows the marker, as in `t5_tr1000`.

*Found by* `t9_tr70000` against `t6_tr1300`: the reader refused a page
count of 100 with a nonzero word 0.
*Confirmed by* the reader reading the case back, 70001 records against
the truth's transition run.


## The marker: the logged ranges

At the offset the trailer word `0x38` names sits a list of 16 byte
entries, as many as the trailer word `0x30` counts.
Each entry is `[uint64 first][uint64 last]`, a closed range of object
indices into the instance records of [hierarchy.md](hierarchy.md).
Every object inside a range has at least one value record.
No object outside has any.

| Case | Objects | Entries | Why |
| :--- | ---: | :--- | :--- |
| `t1_bit_one_edge` | 1 | `[0, 0]` | one signal |
| `t2_flat3` | 3 | `[0, 2]` | three signals |
| `t6_var_int` | 2 | `[0, 0]` | a signal, then a process variable with no records |
| `t9_pkg_sig` | 2 | `[0, 0]` | a signal, then a package signal that is not logged |
| `t9_mark_gap` | 3 | `[1, 1]` | a package constant, a signal, a process variable |
| `t9_port_rec` | 3 | `[0, 0]`, `[2, 2]` | a signal, a package constant, a port |
| `t9_mark_two` | 8 | `[0, 0]`, `[2, 3]` | a signal, a package constant, two signals, four variable objects |
| `t9_var_inst3` | 13 | `[0, 3]` | four signals, then nine variable objects |

The list was first read as one record `[0][N]` with `N` the object
count minus one, which held for 43 cases.
`t6_var_int` moved the reading to the logged object count minus one,
which held for 79.
`t9_port_rec` broke that with two entries and a trailer word of `2`,
and `t9_mark_gap`, written to put an unlogged object first, showed
that the first word is an index and not a zero.
The reader marks each object in a range as logged, and refuses a file
where an object's records disagree with its range.

*Found by* `t2_flat3` against `t1_two_bits`: `N` was `2` and `1`.
*Corrected by* `t6_var_int`, then by `t9_port_rec` and `t9_mark_gap`.
*Confirmed by* `t9_mark_two` and `t9_var_inst3`, and by the reader's
check across 140 of 140 cases with a marker.
`t5_tr1000` shows the list is not before the first page but after a
page flushed mid simulation.


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

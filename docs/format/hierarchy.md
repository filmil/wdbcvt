<!-- SPDX-License-Identifier: Apache-2.0 -->

# The debug section: scopes, units, declarations and objects

The section named `Xilinx DBG` in the directory holds the design
hierarchy.
It says which scopes exist, which entity or process each scope was
elaborated from, which signals, generics and variables are declared in
each, and where in the source each of those came from.
It ends with one instance record per logged object, and that record
carries the handle that names the object's values in the pages.

Everything below is a measurement on files written by Vivado 2025.2.
Reproduce any of it with:

```sh
bazel build //hdl/corpus:all_wdb
bazel run //cmd/wdbcvt -- -dump -in "$PWD/bazel-bin/hdl/corpus/t2_flat3________/sim.wdb"
```

Offsets in this document are relative to the start of the section,
which is the byte the `Xilinx DBG` directory entry points at.


## Framing

| Offset | Size | Value |
| ---: | ---: | :--- |
| `0` | 20 | `Xilinx ISim DBG 006` and a NUL |
| `20` | 4 | uint32 timestamp, noise |
| `24` | 4 | int32 power of ten of the time unit in seconds: `-12` for the picosecond of every VHDL case and of `timescale 1ns / 1ps`, `-9`, `-10` and `-15` under a Verilog precision of `1ns`, `100ps` and `1fs`, tier 21 |
| `28` | 72 | 18 uint32 region offsets, relative to the section start |
| `100` | 16 | 4 uint32 counts: scopes, units, objects, declarations |
| `116` | 68 | 17 uint32 header words, below |
| `184` | | the regions, in the order the offsets give |

The region offsets are non decreasing.
Region `i` runs from `offset[i]` to the next larger offset.
Two equal offsets mean an empty region.
`offset[2]` is the exception: it is the end of the section proper, and
the instance records start there.

| Region | Holds | Record |
| ---: | :--- | :--- |
| 0 | scope records | 9 words |
| 1 | unit records | 9 words |
| 3 | declaration records | 11 words, padded to 8 bytes at the end |
| 4 | range records | 6 words |
| 5 to 8 | empty in every case | |
| 9 | scope name pool | NUL terminated strings, padded to 8 |
| 10 | declaration name pool | NUL terminated strings, padded to 8 |
| 11 | file name pool | NUL terminated strings, padded to 8 |
| 12 | empty in every case | |
| 13 | file table | 2 words per file |
| 14 | statement index | 2 words per file |
| 15 | statement lines | 1 word per executable statement |
| 16 | empty in every case | |
| 17 | value class entries, one per distinct class among the objects | 3 words, padded to 8 bytes at the end |

Every word is a uint32 or int32; `-1` means absent.
A name is an offset into the pool named for it.

The 17 header words are, in order:

| Word | Value | Found by |
| ---: | :--- | :--- |
| 0 | number of range records | `t1_vec8` 1, `t2_array2d` 2 |
| 1 to 4 | 0, the counts of the empty regions 5 to 8 | |
| 5 | length of the scope name pool before padding | `t2_flat3` 14 |
| 6 | length of the declaration name pool before padding | `t2_flat3` 20 |
| 7 | length of the file name pool before padding | |
| 8 | 0, the count of the empty region 12 | |
| 9 | number of files, the entries of region 13 | |
| 10 | number of files again, the entries of region 14 | |
| 11 | number of words in region 15, `0` without `-debug line` | `t2_flat3` 6; `t24_dbg_drv_only` 0 |
| 12 | 0, the count of the empty region 16 | |
| 13 | number of value class entries in region 17 | `t25_sv_two_class` 2 against `t25_sv_two_same` 1 |
| 14 | debug flags: byte 0 `1`, byte 1 `-debug drivers`, byte 2 `-debug readers` | `t24_dbg_drv_only`, `t24_dbg_readers` |
| 15 | debug flags: byte 0 `1`, byte 1 `-debug line`, byte 2 `-debug subprogram`, both under `line` alone | `t24_dbg_line`, `t24_dbg_sub_only` |
| 16 | `0x10000` | |

Words 14 and 15 are byte flags that record the `-debug` levels the
design was elaborated with; the debug level section below has the
table.
Word 13 counts the entries of region 17, which is 16 bytes long for
one entry and empty without objects; the value class section at the
end of this file has what an entry holds.
Word 16 is `0x10000` in every case.

Words 0 to 13 are one rule: word `i` counts region `i + 4`, the way the
four counts before the offsets count regions 0 to 3.
A record region is counted in records, a name pool in bytes up to and
including its last NUL, region 15 in words, and an empty region holds
0.
Regions 15 and 17 are padded to a multiple of 8 bytes after the counted
words.
Found by a sweep over the 758 corpus databases and the four external
ones, which holds in every one of them, and confirmed by the reader,
which rejects a file whose count does not fit its region.
Reproduce with `dewdb -dump`, which prints the counts as
`header words` and the region bounds as `offsets`.
The reader keeps the 17 words as `Debug.Words` and the dump prints
them as `header words`.


## Scope records

A scope is one node of the elaborated hierarchy: the root, an entity
instance, a generate iteration, or a process.
The record has 9 words:

| Word | Meaning |
| ---: | :--- |
| 0 | name, offset into the scope name pool |
| 1 | parent scope index, `-1` for the root |
| 2 | 0 |
| 3 | number of child scopes |
| 4 | index of the first child scope, `-1` if none |
| 5 | index of the first object in this scope, `-1` if none |
| 6 | file index of the scope's source, 0 for the root |
| 7 | line of the scope's source, 0 for the root |
| 8 | unit index |

The name in the pool is the leaf label, not the dotted path.
The root is named `_top`.
`t2_flat3` has `_top`, `tb` and `p`; the dump prints the path built
from the parent links, so it shows `tb.p`.

Children of one parent are contiguous, starting at word 4.
`t2_hier3` has the root at 0, `tb` at 1 with children 2 and 3, `tb.dut`
at 2 with child 4, and `tb.dut.inner` at 4 with child 5.

The root can have more than one top.
`t45_two_tops` elaborates `--top corpus.tb2 --top corpus.tb`, and
the root has two children, `tb2` at 1 and `tb` at 2, in the order of
the options, each with its process after them, and a unit each.
The handles run on across the tops as across any scopes: `tb2.t` is
`0x768` and `tb.s` `0x858`.
The default script logs the first top only: `log_wave -recursive *`
and `get_objects -r /*` reach the current scope, `/tb2`, and `tb.s`
is marked not logged, where `get_scopes /*` lists both.
`t45_two_tops_all` names `/tb2` and `/tb` in two `log_wave` calls and
both record, `tb.s` in arena 1 at key `0x58`, and the logged range
table holds `[0 1]`.
*Found by* `t45_two_tops` against `t45_log_base`.
*Confirmed by* `t45_two_tops_all`.
Two scopes elaborated from the same process share one pool string:
`p` sits at offset 20 in the `t2_hier3` pool and both `tb.p` and
`tb.dut.inner.p` name it.

The counts and indexes are whole 32 bit words, not 16 bit ones.
`t46_gen_70000___` elaborates a for generate of 70000 iterations with
a signal, a loop index and a process each, and has 140004 scopes,
140004 units, 140000 objects and 140000 declarations, with `tb` at
scope 1 holding `children 70004+1`, the last iteration's process at
scope 140003 and the last object at 139999.
`t46_v_gen_70000_` puts 70000 Verilog `reg` objects into one scope,
`tb`, since a Verilog generate block adds no scope; see the Verilog
section.
`t46_deep_100____` is a recursive entity that instantiates itself
under an if generate 100 levels deep, 306 scopes, and every path is a
chain of parent links as it is at depth 2.
The reader takes all three without a special case, and the corpus
test checks every one of the 70000 signals, indexes and registers.
*Found by* `t46_gen_70000___` against `t46_sig_1000____`.
*Confirmed by* `t46_v_gen_70000_`, `t46_deep_100____`.

Word 5 is the first object index.
The objects of a scope are contiguous in the instance record list.
`t6_proc2` has `tb` at object 0 (signals `a`, `b`), `tb.p` at object 2
(variable `v`) and `tb.q` at object 3 (variable `w`).

The line in word 7 is the line of the entity declaration for an
instance scope and the line of the `process` keyword for a process.
Found by `t2_flat3`: `tb` line 15 is `entity tb is`, `tb.p` line 23 is
`p : process`.


## Unit records

A unit is what a scope was elaborated from.
There is one unit record per scope in every case, at the same index,
even when two scopes come from the same entity.
The record has 9 words:

| Word | Meaning |
| ---: | :--- |
| 0 | entity name, offset into the scope name pool, `-1` if none |
| 1 | architecture name, same pool, `-1` if none |
| 2 | kind: `0x13` the root, `0x09` an entity, `0x0a` a package, `0x0c` a generate or a block, `0x0d` a process; for Verilog `0x00` a module, `0x03` a task, `0x04` a function, `0x05` a named block, `0x07` a process |
| 3 | number of declarations |
| 4 | 0, see below |
| 5 | file index of the architecture |
| 6 | line of the architecture |
| 7 | file index of the entity, 0 for a process |
| 8 | line of the entity, 0 for a process |

Words 5 and 6 point at the `architecture` line for an entity and at the
`process` line for a process.
Words 7 and 8 point at the `entity` line.
Found by `t2_flat3`: unit 1 has file 2 line 18 (`architecture sim of
tb`) and entity file 2 line 15 (`entity tb is`).

The declarations of a unit are the signals and generics declared in
that entity, or the variables declared in that process.
Word 4 is 0 in every case, including `t4_gen_diff_two`, where units 2
and 3 have two declarations each and a real index would read 2 for the
second.
So it is not the index of the first declaration.
The declarations of a unit are the next `count` records after the
previous unit's, in unit order.
`t6_proc2` has `a`, `b` for `tb`, then `v` for `p`, then `w` for `q`.

Within a unit the signals come first, in source order, and then the
generics, constants and variables, in source order.
`t50_ord_const1st` declares `constant i` above `signal s` in the
architecture and the file lists `s` before `i`, where `t5_tr1000`
with the constant below the signal lists the same order.
`t50_ord_proc_con` declares `constant c` above `variable v` in a
process and the file lists `c` then `v`, the source order, though the
variable kind `0x0f` is the lower number.
`t4_gen_default__` lists the generic of the entity after the signal of
the architecture, and `t22_dbg_sub_proc` lists three signals, then the
generic.
`t50_ord_two_sig_` declares `z`, `a`, `s` and lists them so, with the
handles `0x768`, `0x828`, `0x8e8` in that order, so the signal order
is neither by name nor by anything but the source.
A subprogram unit follows the same rule with its signal parameters
as the signals: `drive(constant a : ...; signal q : ...)` in
`t49_sub_vec_prm_` and `show(variable v : ...; signal q : ...)` in
`t50_sub_in_var__` both list `q` first, though `v` holds the lower
frame offset.
*Found by* `t50_ord_const1st` against `t5_tr1000_______` and
`t50_ord_proc_con` against `t6_var_int______`.
*Confirmed by* `t50_ord_two_sig_`, `t4_gen_default__`,
`t22_dbg_sub_proc`, `t49_sub_vec_prm_` and `t50_sub_in_var__`.

Names are shared with scopes through one pool.
The `tb(sim)` unit in `t2_flat3` names `tb` at offset 5 and `sim` at
offset 8, and the `tb` scope names the same offset 5.


## Declaration records

A declaration is one signal, port, generic, constant or variable, before
elaboration.
The record has 11 words, and the region is padded to a multiple of 8
bytes after the last record:

| Word | Meaning |
| ---: | :--- |
| 0 | name, offset into the declaration name pool |
| 1 | index of the region 17 entry holding the value class of the declaration's objects, see the value classes section |
| 2 | file index |
| 3 | line |
| 4 | value size, bytes for VHDL and bits for Verilog, see [values.md](values.md) |
| 5 | type index into the type table, see [types.md](types.md) |
| 6 | number of range records |
| 7 | index of the first range record, `-1` if none |
| 8 | kind, below |
| 9 | port mode, below |
| 10 | noise for a signal, 0 for a variable |

The kinds seen:

| Kind | Meaning | Found by |
| ---: | :--- | :--- |
| `0x0e` | signal, including a port | every case with a signal |
| `0x0f` | variable declared in a process | `t6_var_int`, `t6_proc2` |
| `0x12` | generic | `t4_gen_default` |
| `0x13` | constant: an architecture constant, a loop index or a generate index | `t8_gen_if`, `t5_tr1000`, `t6_tr1300`, `t7_gen_for` |
| `0x14` | parameter or variable of a subprogram, under `-debug subprogram` | `t22_dbg_subprog` |
| `0x15` | signal parameter of a subprogram, under `-debug subprogram` | `t23_sub_sig_prm` |
| `0x00` | Verilog variable: `reg`, `integer`, `real`, `time`, a SystemVerilog `logic`, `int`, struct or enum | `t11_v_bit_edge` |
| `0x01` | Verilog `parameter` | `t11_v_param` |
| `0x03` | Verilog net: a `wire` or `uwire`, and every port | `t11_v_wire`, `t11_v_port`, `t19_sv_uwire` |
| `0x04` | `wand` | `t19_v_wand` |
| `0x05` | `wor` | `t19_v_wor` |
| `0x06` | `tri` | `t19_v_tri` |
| `0x07` | `triand` | `t19_v_triand` |
| `0x08` | `trior` | `t19_v_trior` |
| `0x09` | `tri0` | `t19_v_tri0` |
| `0x0a` | `tri1` | `t19_v_tri1` |
| `0x0c` | `supply0` | `t19_v_supply0` |
| `0x0d` | `supply1` | `t19_v_supply1` |

Word 9 was recorded as the constant 5 through tier 7, where no case had
a port.
It is the port mode:

| Word 9 | Mode | Found by |
| ---: | :--- | :--- |
| 0 | `inout` | `t8_port_inout` |
| 1 | `in` | `t8_port_in`, `t8_port_open`, `t8_port_vec8` |
| 2 | `out` | `t8_port_out`, `t8_port_open` |
| 3 | `buffer` | `t8_port_buffer` |
| 4 | `linkage` | `t9_port_lnk` |
| 5 | not a port | every declaration through tier 7 |

The five modes are in the order the VHDL standard lists them.
A `linkage` port connected to a signal shares the signal's handle
like any other port, and `t9_port_lnk` is otherwise byte for byte the
layout of `t8_port_in`.

The reader exposes it as `Decl.Mode`, and the corpus test checks it
against the `port` field of each signal in `truth.json`.
A port is otherwise an ordinary signal: kind `0x0e`, declared in the
child entity's unit, and an object in the child's scope.
Where its handle comes from is in the instance record section below.

A null range declares 0 bytes.
`t24_null_range` has `signal z : std_ulogic_vector(0 downto 1)`
beside `s`, and `z` is a signal declaration of size 0 with the range
record `(0 downto 1)` kept as written, and an object with its own
handle that is marked not logged and holds no record.
The type entry is the ordinary unconstrained `STD_ULOGIC_VECTOR`.

*Found by* `t24_null_range` against `t1_bit_one_edge`.

Word 10 was called noise because it differs between two runs of the
same design.
It is 0 for both variables in `t6_proc2`, so it is not pure noise.
What it holds is open.

A range record has 6 words.
`t2_flat3` has one, for `bravo : std_ulogic_vector(7 downto 0)`:
`[7][0][0][0][-1][8]`.
The words are the left bound as a 64 bit pair, low word first, the
right bound as another, the direction, `1` for `to` and `-1` for
`downto`, and the span of the bounds plus one.
The high words are the sign: `array (-2 to 1)` of `t41_neg_arr_type`
is `[-2][-1][1][0][1][4]`, and `reg [-4:3]` of `t12_v_neg_range` is
`[-4][-1][3][0][1][8]`, so the record is the same for both languages.
The direction is in the record, not only in the type table: the two
ranges of `t2_array2d`, `(0 to 3)` and `(7 downto 0)`, hold `1` and
`-1`, and the dump prints them from that word.
The last word is not the element count: the null range `(0 downto 1)`
of `t24_null_range` holds `[0][0][1][0][-1][2]` beside a declaration
size of 0, so it is the distance between the bounds plus one, and the
reader computes the length from the bounds and the direction instead.
*Found by* `t41_neg_arr_type` against `t2_flat3` for the pairs,
`t2_array2d` for the direction, `t24_null_range` for the span.
*Confirmed by* `t43_port_unc_asc`, `(0 to 7)`; `t12_v_neg_range`;
`t11_v_mem8`, `[0:7]` and `[7:0]`.


## The file table

Region 13 has 2 words per file: an offset into the file name pool for
the path the file was compiled from, and an offset for a local path, or
`-1`.
The corpus file has only the first.
The Vivado library sources have both: a `/proj/primebuilds/...` path
that was baked in when AMD built the release, and the path in the local
installation.

Files 0 and 1 have no name in the corpus and are never referenced.
File 2 is the testbench.
The rest are the library sources the design pulled in, so a `use`
clause changes this table.
`t2_slv8` uses `ieee.numeric_std` and is 447 bytes longer than
`t1_vec8` for that reason alone: two more file entries, each with two
paths.
See [types.md](types.md) for the comparison that found this.
Tier 47 adds the use clauses one at a time to the tier 1 baseline.
`standard.vhdl`, `textio.vhdl` and `env.vhdl` are there with no clause
at all, `t47_use_none`, since `std.env.stop` ends every bench;
`library ieee;` alone adds nothing, `t47_use_lib_only`; a use clause
adds the package and its body, `numeric_std.vhdl` and
`numeric_std-body.vhdl` in `t47_use_numstd`, before `env.vhdl`, and
one naming a single item, `use ieee.std_logic_1164.std_ulogic` in
`t47_use_one_name`, adds the same two files as `.all`.
The use clause also costs handle space, by the package and not by the
type of the signal; see [container.md](container.md).
*Found by* `t47_use_none` against `t47_use_1164_bit`.
*Confirmed by* `t47_use_lib_only`, `t47_use_one_name`,
`t47_use_numstd`, `t47_use_numbit`, `t47_use_mathrl`,
`t47_use_textio`.

Region 14 has 2 words per file: an index into region 15, or `-1`, and a
count.
Region 15 is a list of source line numbers, one per executable
statement, grouped by file.
Found by `t6_proc2`: file 2 has index 0 count 9, and the nine lines are
21 to 24 and 29 to 33, which are the statements of processes `p` and
`q`.
`t2_hier3` has two files with statements: `tb.ent.vhdl` at index 0
count 2, lines 21 and 22, and `leaf.ent.vhdl` at index 2 count 3, lines
17 to 19.
Region 15 is padded with zeros to a multiple of 8 bytes.
Header word 11 is the number of lines before padding.

Regions 14 and 15 are the `line` debugging ability.
`t22_dbg_wave`, elaborated with `-debug wave` in place of the default
`-debug typical`, has every region 14 entry at `-1 0` and region 15
empty, and nothing else of the section changes but the 32 bytes the
eight statement lines took.
*Found by* `t22_dbg_wave` against `t22_base`.


## Instance records

After `offset[2]` come the objects, 56 bytes each, as many as the
third count says.
An object is a declaration elaborated in a scope: one signal, generic or
variable that the simulator gave storage.

| Offset | Size | Meaning |
| ---: | ---: | :--- |
| `0` | 8 | handle |
| `8` | 8 | second handle for a signal, 0 for the others |
| `16` | 4 | scope index |
| `20` | 4 | offset into the value at the handle, 0 unless the object is a port bound to a slice: bytes for VHDL, bits for Verilog |
| `24` | 4 | 0 |
| `28` | 4 | storage class: 0 signal, 1 port on a language boundary, 2 generic, constant, variable or loop index, 3 scalar subprogram local, 4 composite subprogram local, 6 signal parameter |
| `32` | 8 | declaration index |
| `40` | 4 | position of a Verilog port in its module's port list, from 0; 0 for every VHDL object and every non port |
| `44` | 4 | not written: `-1`, 0, or a value that differs between runs; see below |
| `48` | 8 | 0 |

The word at `16` was read as a `uint64` scope index through tier 8,
where the upper half was always 0.
`t9_port_slice` binds a `std_ulogic` port to `x(0)` of a
`std_ulogic_vector(1 downto 0)`, and the port's record holds scope 2
and `1` in the upper half: the byte offset of `x(0)`, the right and
therefore second element, in the signal's value.
`t9_port_slice2` binds a 2 bit port to `x(2 downto 1)` of a 4 bit
vector and holds 1 again, the offset of `x(2)`, and `t9_port_sliceto`
binds `x(0)` of a `0 to 1` vector and holds 0.
So the port's value is `size` bytes of the signal's value from that
offset, and the reader decodes it from the signal's records.
A Verilog port bound to a slice of a net holds the offset in bits
from bit 0 of the net: `wb_mem_adr[12:2]` in `//hdl/serv:sim` gives
offset 2, `o_dbus_dat[26:24]` gives 24, and `v[39:34]` of a 40 bit
wire in `t37_v_port_pair1` gives 34, into the second word pair.
An output port bound to a slice holds the same word: `.o(v[1])` in
`t63_pdr_port_bit` gives 1, `.o(v[7:4])` in `t63_pdr_port_slc` 4 and
`.o(v[63:32])` in `t63_pdr_port_hi_` 32, each on the net's handle.
See [values.md](values.md).

*Found by* `//hdl/serv:sim` against `t9_port_slice`.
*Confirmed by* `t37_v_port_slc__`, `t37_v_port_bit__`,
`t37_v_port_pair1`, `t37_v_port_span_`, `t63_pdr_port_bit`,
`t63_pdr_port_slc` and `t63_pdr_port_hi_`.

The word at `40` was recorded as 0 through tier 47, when every case
with a Verilog port had read as 0 or 1.
A sweep of the word over the corpus found 1, 2 and 3 on the ports of
the tier 36 children and 0 to 29 on the ports of `//hdl/serv:sim`,
where `servant_sim` holds 0 for `wb_clk`, 1 for `wb_rst`, 2 for
`pc_adr`, 3 for `pc_vld` and 4 for `q`, the order of its port list.
Tier 48 fixes what the number counts.
`t48_v_port_pos4_` instantiates a child with ports `a`, `b`, `c`, `d`
and holds 0, 1, 2, 3 on them.
`t48_v_port_rev__` connects them by name in reverse order, and
`t48_v_port_posit` by position, and both hold 0, 1, 2, 3, so the
connection does not count.
`t48_v_port_nansi` declares the same ports in a non ANSI header,
`module child(a, b, c, d);` followed by `output d; output c; input b;
input a;`, and its objects come in declaration order, `d`, `c`, `b`,
`a`, holding 3, 2, 1, 0: the number is the place in the port list,
and the object order is the declaration order.
`t48_v_port_open_` leaves `d` unconnected and `d` still holds 3.
Every VHDL port holds 0: the 557 objects of `//hdl/potato:sim` and
every port case of the VHDL tiers.
The word is written on the first instance of a unit only.
Every port of every later instance of the same unit holds 0:
`t64_ord_three___` instantiates `child(input i, output o)` three
times and holds 0, 1 on `u0` and 0, 0 on `u1` and `u2`;
`t64_ord_two_pos4` holds 0, 1, 2, 3 on `u0` and four zeros on `u1`;
and the two instances of a generate loop, `t64_ord_gen_kids`, hold
0, 1 under `g[0]` and 0, 0 under `g[1]`.
The first instance of a second unit holds its own: `child2 u1` after
`child u0` in `t64_ord_two_mods` holds 0, 1.
So the number can place a port only where the unit is instantiated
once.
The reader keeps it as `Position` and the dump prints it after a port.

*Found by* the sweep over the corpus, then `t48_v_port_nansi` against
`t48_v_port_pos4_`.
*Confirmed by* `t48_v_port_rev__`, `t48_v_port_posit`,
`t48_v_port_open_`, `t36_v_hier_and__`, `t13_v_hier3_net_`,
`t11_v_port______` and `//hdl/serv:sim`.
The first instance rule was *found by* `t64_ord_two_kids` against
`t63_pdr_port_bit`, `tb.u1.o` at 0, and *confirmed by*
`t64_ord_two_nets`, `t64_ord_two_same`, `t64_ord_three___`,
`t64_ord_two_pos4`, `t64_ord_two_mods` and `t64_ord_gen_kids`, against
`t64_ord_pos_expr` and `t64_ord_pos_bit3`, one instance each, at 1.

The word at `44` is not written by the producer, on the evidence of
the same sweep.
Across the 746 databases in the tree it is 0 on every object of 235,
one of `0x7ffc` to `0x7fff` on every object of 352, all of them
Verilog, `-1` on the objects of a few VHDL cases, and an arbitrary
32 bit value elsewhere, with different objects of one file holding
different values: 0 on the loop indexes and `0x60b54a78` on the
signals of `t7_gen_for______`.
The `0x7ffc` to `0x7fff` range is the shape of the upper half of a
stack address on the producing host, and the arbitrary values the
shape of heap addresses, so the reading is memory the writer copied
without setting.
The reader masks it.

The word at `28` was read as 0 for a signal and 2 for the rest through
tier 48, the two values every case before the subprogram and mixed
language tiers holds.
The same sweep found 1, 3, 4 and 6 as well, and tier 49 fixes what
each stands for.

| Value | Objects | Cases |
| ---: | :--- | :--- |
| 0 | a signal, net or Verilog variable, ports included | every case |
| 1 | a port on a language boundary, whichever side it is on | `t21_mix_v_in_vh_`, `t21_mix_vh_in_v_`, `t49_mix_2port___`, `t49_mix_deep____` |
| 2 | a generic, constant, parameter, process variable or loop index; the objects with no second handle | `t7_gen_for______`, `t6_proc2________`, `t49_sub_var_prm_` |
| 3 | a subprogram parameter or variable of a scalar or access type, `constant` and `variable` class alike, `in` or `inout` | `t23_sub_sizes___`, `t49_sub_var_prm_`, `t50_sub_in_var__`, `t50_sub_acc_loc_` |
| 4 | a subprogram parameter or variable of an array, string or record type, whatever its class | `t23_sub_sizes___`, `t49_sub_rec_loc_`, `t49_sub_int_arr_`, `t49_sub_vec_prm_`, `t50_sub_var_vec_`, `t50_sub_var_rec_`, `t50_sub_str_loc_`, `t50_sub_func_prm` |
| 6 | a signal parameter of a subprogram, in or out, scalar, vector or record | `t23_sub_sig_prm_`, `t49_sub_sig_in__`, `t49_sub_sig_vec_`, `t50_sub_sig_rec_` |

The boundary port holds 1 on the Verilog side and on the VHDL side:
`t49_mix_deep____` puts a VHDL leaf under a Verilog child under a VHDL
testbench, and the two ports of the child and the two of the leaf all
hold 1.
The scalar and composite locals differ by the type and not by the
class or mode: `t49_sub_var_prm_` gives an `inout` `variable`
parameter 3, and `t49_sub_vec_prm_` a `constant` vector parameter 4.
Tier 50 holds that against the class and mode: an `inout` `variable`
vector parameter and an `inout` `variable` record parameter are 4 in
`t50_sub_var_vec_` and `t50_sub_var_rec_`, an `in` `variable` scalar
parameter is 3 in `t50_sub_in_var__`, a local of an access type is 3
in `t50_sub_acc_loc_`, a local `string(1 to 4)` is 4 in
`t50_sub_str_loc_`, and a function's vector parameter and scalar local
are 4 and 3 in `t50_sub_func_prm` as a procedure's are.
A signal parameter of a record type is 6 in `t50_sub_sig_rec_`.
So 3 is a value the frame holds in place, a scalar or a pointer, 4 a
value the frame holds through a descriptor, and 6 a reference to a
signal, and none of the three depends on the class or the mode.
5 has not been seen, in the 48 cases built with `-debug subprogram`
through tier 55.
Tier 55 hunts for it among the declarations of a subprogram that are
neither parameters nor plain variables, and finds each of them either
a local of class 3 or 4 or absent from the file, see the constant,
alias, file and protected paragraphs below.
The reader keeps the word as `Storage` and the dump prints it when it
is not 0; `Generic` stays the test for 2.

*Found by* the sweep over the corpus, then `t49_sub_rec_loc_` and
`t49_sub_int_arr_` against `t23_sub_sizes___` for the composite
reading of 4.
*Confirmed by* `t49_sub_var_prm_`, `t49_sub_vec_prm_`,
`t49_sub_sig_in__`, `t49_sub_sig_vec_`, `t49_mix_2port___` and
`t49_mix_deep____`, the eight `t50_sub_` cases, and the `storage`
field of `truth.json` on the tier 21, 22 and 23 cases named above.

A subprogram parameter of an unconstrained type is not in the file.
`t49_sub_str_prm_` declares `constant name : in string` beside a
signal parameter, and the file holds the signal parameter's
declaration and object and nothing for `name`: two declarations and
two objects, the same handle space as `t23_sub_sig_prm_` with its
scalar constant parameter.
`t50_sub_ivec_prm` does the same with `constant v : in integer_vector`
and holds the signal parameter alone, so the absence is about the
unconstrained bound and not about `string`.
The absent parameter still takes room in the frame: the signal
parameter after it is on `0xe8` in both cases, 24 bytes past the
procedure base `0xd0` it is on when it comes first, the size of a
vector descriptor, see the frame section below.

*Found by* `t49_sub_str_prm_` against `t23_sub_sig_prm_`.
*Confirmed by* `t50_sub_ivec_prm`.

A `file` parameter is absent the same way and takes 8 bytes of the
frame: `procedure put(file f : int_file; signal q : out std_ulogic)`
in `t51_sub_file_prm` has `q` alone, on `0xd8`.
The file object of the architecture, `file fo : int_file open
write_mode is "t51.bin"`, is a declaration of the variable kind
`0x0f` with size 0 and the file type, and an object with one handle,
storage class 2 and no record, as an access variable is.
*Found by* `t51_sub_file_prm` against `t23_sub_sig_prm_`.

A constant declared in a subprogram is a local of the same kind `0x14`
as a variable, with no trace of being a constant: `k` of
`t55_sub_con_loc_` is `local k : integer` of class 3, and the array
constant `a` of `t55_sub_con_arr_` is class 4, as a variable of each
type is.
It takes more of the frame than a variable does.
`t55_sub_loop____` declares `c : std_ulogic` and `variable v : integer`
on `0x40` and `0x44`; `t55_sub_con_loc_` puts `constant k : integer`
between them and `v` moves to `0x58`, 20 bytes past `k`, whether or
not the initialiser of `v` reads `k` (`t55_sub_con_nori`), and a
second constant in `t55_sub_2con____` is on `0x58` with `v` on
`0x6c`.
A real constant in `t55_sub_con_real` is on `0x48` with `v` on `0x60`,
24 bytes past it.
So a scalar constant takes its size plus 16 bytes of the frame, where
a variable of the same type takes its size alone: `u` and `v` of
`t55_sub_var_init` are on `0x44` and `0x48`.
The array constant of `t55_sub_con_arr_` is on `0x48` with `v` on
`0x60`, the 24 byte descriptor an array variable takes.
None of them moves the handle space, which stays `0x11d0` from
`t55_sub_loop____` through `t55_sub_con_real`.
*Found by* `t55_sub_con_loc_` against `t55_sub_loop____`.
*Confirmed by* `t55_sub_con_nori`, `t55_sub_2con____`,
`t55_sub_con_real`, `t55_sub_var_init` and `t55_sub_con_arr_`.

An alias, a file object and a variable of a protected type declared in
a subprogram are all absent from the file: no declaration, no object,
the same handle space `0x11e8` as `t51_sub_loop_idx`, whose procedure
has the signal parameter and the integer local alone.
The frame shows what each takes.
The integer local after the signal parameter is on `0x110` in
`t51_sub_loop_idx`.
It stays on `0x110` in `t55_sub_alias___`, where `alias b : integer is
v` follows it, so an alias takes nothing.
It is on `0x138` in `t55_sub_file_loc`, where `file fl : int_file`
precedes it, so a file local takes 40 bytes where the file parameter
of `t51_sub_file_prm` took 8.
It is on `0x11c`, `0x12c` and `0x13c` in `t55_sub_prot_loc`,
`t55_sub_prot_2__` and `t55_sub_prot_3__`, where one, two and three
variables of a protected type precede it, so the first takes 12 bytes
and each further one 16.
The 12 is not a multiple of the 8 that a pointer takes, and what the
frame holds for a protected variable is open.
A `for` loop index in a function is absent as it is in a procedure:
`t55_sub_loop____` has `c` and `v` on `0x40` and `0x44` with the loop
between them.
*Found by* `t55_sub_alias___`, `t55_sub_file_loc` and
`t55_sub_prot_loc` against `t51_sub_loop_idx`.
*Confirmed by* `t55_sub_prot_2__`, `t55_sub_prot_3__` and
`t55_sub_loop____` against `t50_sub_func_prm`.

The static values of a subprogram take handle space, though its
locals do not.
Tier 56 holds a function with a `std_ulogic` parameter and an integer
local, `t56_typ_none____` at `0x11d0`, and adds one thing at a time:

| Added | Handle space | Case |
| :--- | ---: | :--- |
| an array type of four integers, used by nothing | `+0` | `t56_typ_arr_unus` |
| a local of it, initialised by `(others => 0)` | `+0x10` | `t56_typ_arr_loc_` |
| a local of it, not initialised | `+0x10` | `t56_typ_arr_noin` |
| a local of it, initialised by `(others => n)` from a parameter | `+0` | `t56_sub_arr_dyni` |
| two locals of it, or of two such types | `+0x20` | `t56_typ_arr_2loc`, `t56_typ_arr_2typ` |
| a local of an array of eight integers | `+0x20` | `t56_typ_arr8_loc` |
| the uninitialised local, assigned `(others => 2)` in the body | `+0x20` | `t56_typ_arr_lit_` |
| the uninitialised local, assigned `(others => v)` in the body | `+0x10` | `t56_typ_arr_dyn_` |
| a `std_ulogic_vector(3 downto 0)` local, initialised or of a named subtype | `+4` | `t56_typ_vec4_loc`, `t56_typ_vec4_sub` |
| two of them | `+8` | `t56_typ_vec4_2lc` |
| one, not initialised, assigned `(others => '0')` in the body | `+8` | `t56_typ_vec_noin` |
| an `integer range 0 to 7` local, or an enumeration local | `+0` | `t56_typ_int_rng_`, `t56_typ_enum_loc` |
| a record local of 8 bytes, initialised by a literal or not | `+0xc` | `t56_typ_rec_loc_`, `t56_typ_rec_noin`, `t56_sub_rec_1int`, `t56_sub_rec_2int`, `t56_typ_rec_arr_` |
| a record local of 16 bytes | `+0x14` | `t56_sub_rec_3int`, `t56_sub_rec_4int`, `t56_sub_rec_2rl_` |
| a record local initialised by `(a => c, n => 1)` from the parameter | `+0` | `t56_typ_rec_prm_` |

So a type costs nothing, and a scalar local nothing, but every static
value of a composite type costs its bytes: the initial value of a
composite local when it is a literal or the default, and every
aggregate or string literal in the body, or in the call, which is the
4 that `"0001"` costs `t50_sub_func_prm` and `"abcd"` costs
`t50_sub_str_loc_`.
An initial value computed at the call costs nothing, and a local that
gets one has no default to store.
An array's bytes are its elements, a record's its declared size plus
4, and the declared size of a record is a multiple of 8, so a record
of one integer declares 8 bytes as a record of two does.
A process variable is not affected: `t56_prc_vec_noin` and
`t56_prc_arr_noin` drop the initialiser of the tier 52 variables and
keep their handle space and strides, `0x14` and `0x20`.
*Found by* `t56_typ_arr_loc_` against `t56_typ_arr_unus`, then
`t56_typ_arr_lit_` against `t56_typ_arr_noin` for the literal, and
`t56_sub_arr_dyni` and `t56_typ_rec_prm_` for the computed initial
value.
*Confirmed by* the rest of tier 56, `t49_sub_int_arr_`,
`t49_sub_rec_loc_` at `0x11d0` with its record initialised from the
parameter, `t50_sub_func_prm` and `t50_sub_str_loc_`.

The handle is the number a value record in a page carries, split as
`handle >> 11` for the arena and `handle & 0x7ff` for the key.
See [values.md](values.md).

Signals get handles from one counter.
The first signal is `0x768` in every case.
The second handle is the first plus the value size rounded up to a
multiple of 8, and the next signal's handle is the second handle plus
`0xe8` when the signal has one driver.
So the stride between one byte signals is `0xf0`, and between the two
16 byte records of `t2_record_two` it is `0xf8`.

| Case | Size | Handle | Second | Next |
| :--- | ---: | ---: | ---: | ---: |
| `t6_sig20` | 1 | `0x768` | `0x770` | `0x858` |
| `t2_flat3` | 1, 8, 4 | `0x768`, `0x858`, `0x948` | `+8` each | |
| `t2_record_two` | 16 | `0x768` | `0x778` | `0x860` |
| `t2_record_nested` | 24 | `0x768` | `0x780` | |
| `t2_array2d` | 32 | `0x768` | `0x788` | |

That is consistent with a handle being an address: a block of the
value's size, then `0xe8` bytes of something per signal.
Nothing in the file depends on the reading, and the decoder does not
use it.

Tier 46 splits the `0xe8` into `0xb8` for the signal and `0x30` for
its driver.
`t46_sig_1000____` declares a thousand `std_ulogic` signals and drives
two of them, `s0` and `s999`.
The driven `s0` at `0x768` is followed by `s1` at `0x858`, `0xf0` on,
and every undriven signal after it is `0xc0` on, which is `0xb8` plus
its 8 bytes of storage, the stride an open port has.
`t46_gen_70000___` drives every one of its 70000 signals from a
process of its own, and every stride is `0xf0`.
A resolved `std_logic` with more drivers costs more: `0x140` for two,
`t46_drv_2_next__`, and `0x178` for three, `t46_drv_3_next__`, which
is `0xc0` plus `0x80` and `0xc0` plus `0xb8`.
Two points fit `0x30` per driver plus `0x10` plus 8 per driver from
the second driver on, and that is a guess in
[format.md](../format.md).
The reader does not use any of it.

| Object | Handle stride | Found by |
| :--- | :--- | :--- |
| signal, one driver | `0xe8` plus the size rounded up to 8 | the table above |
| signal, no driver | `0xb8` plus the size rounded up to 8 | `t46_sig_1000____` |
| `std_logic`, two and three drivers | `0x140` and `0x178` | `t46_drv_2_next__`, `t46_drv_3_next__` |
| open port | `0xb8` plus the size rounded up to 8 | `t8_port_open3`, `t8_port_vec8`, `t8_port_vec16` |
| connected port | none, it shares the signal's handle | `t8_port_in`, `t8_port_chain` |

Generics, constants, variables and loop indexes get their handles after
the last signal of the whole design, whatever scope declares them.
`t46_gen_70000___` interleaves a signal and a loop index in every
generate scope, and the objects list them in that order, yet the 70000
signals run from `0x768` at `0xf0` and the 70000 indexes from
`0x10066a8` at `0x118`, `0x640` past the end of the last signal's
stride.
`t46_deep_100____` does the same with 101 signals at `0xf0` and 101
generics at `0x140`, again `0x640` past the last signal.
The 109 cases of the corpus with both kinds of object all put every
signal before every other object.
Within the second region the stride from one object to the next is
the value's size, the next object starting on a multiple of its own
size, and an array or a string takes 16 bytes more than its elements.
Tier 52 declares `a : T` and then `b : integer` for eleven types, as
two process variables, as two architecture constants and as two
generics, and reads the distance from `a` to `b`:

| Type of `a` | Stride | Cases |
| :--- | ---: | :--- |
| `std_ulogic`, `boolean`, `integer` | 4 | `t52_var_sul_____`, `t52_var_bool____`, `t52_var_int_____`, `t52_con_int_____`, `t52_gen_int_____` |
| `real`, `time` | 8 | `t52_var_real____`, `t52_var_time____`, `t52_con_real____` |
| a record of a `std_ulogic` and an `integer` | 8 | `t52_var_rec_____` |
| `std_ulogic_vector(3 downto 0)`, `string(1 to 4)` | `0x14` | `t52_var_vec4____`, `t52_var_str4____` |
| `std_ulogic_vector(7 downto 0)` | `0x18` | `t52_var_vec8____`, `t52_con_vec8____`, `t52_gen_vec8____` |
| an array of four integers | `0x20` | `t52_var_arr4____` |

The one byte types read 4 because `b` is an integer and starts on a
multiple of 4.
`t9_gen_types` has a `string(1 to 3)` 1 past a `boolean`, a four
element vector 19 past the string and a `real` 20 past the vector, on
a multiple of 8, which is the same rule with one byte neighbours.
A constant and a generic have the strides of a variable, and the
handle space grows with the stride: `0x11d8` with two integers, then
`0x11e0`, `0x11e8`, `0x11f0` and `0x11f8` for the strides 8, `0x14`,
`0x18` and `0x20`, which is the two sizes summed and rounded up to a
multiple of 8.

The larger strides between like objects of different scopes are the
cost of the scopes between them.
Tier 52 instantiates a `child` with a generic `k` twice, as `d0` and
`d1`, and elaborates a two iteration generate with the index `i`, and
varies the body of the child and of the iteration in step:

| Body | `k` stride | `i` stride | Cases |
| :--- | ---: | ---: | :--- |
| nothing | `0x30` | no index at all | `t52_inst2_empty_`, `t52_gi2_empty___` |
| a process | `0xc0` | `0xc0` | `t52_inst2_proc__`, `t52_gi2_proc____` |
| an undriven signal | `0x68` | `0x68` | `t52_inst2_sig___`, `t52_gi2_sig_____` |
| a signal driven by a process | `0x118` | `0x118` | `t52_inst2_sigprc`, `t52_gi2_sigprc__` |

So an instance and a generate iteration cost the same: `0x28` for the
scope and 8 for its integer, tier 53 below, a process `0x90` more, a
signal `0x38` more in this region beside its own `0xc0` in the first,
and the signal's driver `0x20` more beside its `0x30` in the first.
That is the `0xf8` per signal and `0x50` per driver that tier 46 read
off the handle space and could only half account for in the first
region.
Within a scope's block the signals come before the data objects and
the process and the driver after them: `k` of `d0` sits at `0x1070`
with the undriven signal, which is the `0xeb8` of the empty child
plus the two `0xc0` of the first region plus one `0x38`, and at
`0x10d0` with the driven signal, the same past two `0xf0`.
The handle space of the four instance cases, `0x1288`, `0x13a8`,
`0x1478` and `0x1638`, grows by twice `0x90`, twice `0xc0` plus
`0x38`, and twice `0xf0` plus `0x38` plus `0x20` plus `0x90` from the
empty child, the sum of both regions.
The generate cases have `0x18` to `0x20` less handle space than the
instance cases of the same body, `0x1390`, `0x1458` and `0x1620`,
which the strides do not show, and that difference is open.

The strides of the earlier cases fit the same costs.
`t7_gen_for` puts its three loop indexes at `0x1040`, `0x1070` and
`0x10a0`, `0x30` apart, a bare iteration each, and then the three
generics of the children at `0x1130`, `0x1248` and `0x1360`, `0x118`
apart, a child with a signal driven by a process each; the child
instances follow the iterations because the scopes are laid out
breadth first.
`t46_gen_70000___` has its indexes `0x118` apart, and each iteration
holds a signal driven by a process.
`t46_deep_100____` has its generics `0x140` apart, `0x28` more than
the signal, the process and the driver of each level, which is read
as the if generate scope holding the next level, on that one point.
`t6_proc2` has the integer variables of its two processes 4 apart.

Tier 53 varies the body of the child one item at a time on the two
instance shape and reads the cost of each:

| Body | `k` stride | Cases |
| :--- | ---: | :--- |
| nothing, one child; three children | `0x30` | `t53_inst1_empty_`, `t53_inst3_empty_` |
| a second generic; an architecture constant | `0x30`, the second object 4 past `k` | `t53_inst2_2gen__`, `t53_inst2_const_` |
| two processes | `0x150` | `t53_inst2_2proc_` |
| a process with a variable | `0xc0`, the variable 4 past `k` | `t53_inst2_var___` |
| two undriven signals | `0xa0` | `t53_inst2_2sig__` |
| a signal driven by a concurrent assignment | `0x118` | `t53_inst2_conc__` |
| a `std_logic` driven by two processes | `0x1c8` | `t53_inst2_2drv__` |
| an input port connected to `s`; left open | `0x68` | `t53_inst2_port__`, `t53_inst2_portop` |
| an empty grandchild with a generic | `0x60` | `t53_inst2_nest__` |
| the child under an if generate; under a block | `0x30`, and `0x50` before the first `k` | `t53_ifgen_inst__`, `t53_blk_inst____` |

So a scope's block is `0x28`, plus its data objects laid out by the
size rule and rounded up to a multiple of 8, plus `0x38` per signal
or port, connected or open, plus `0x90` per process, a concurrent
assignment being one, plus `0x20` per driver.
The block holds the signals first, then the data objects, then the
rest; a process variable sits among the data objects of the scope
holding the process, `t53_inst2_var___` and `t6_proc2`.
The generate and block scopes of a unit are laid out depth first
after the unit's own block, and the instances the unit holds follow,
each with the scopes and instances of its own unit the same way:
`t53_ifgen_inst__` has `g0`, `g1`, `d0`, `d1`; `t53_inst2_nest__`
has `d0`, `d0.e`, `d1`, `d1.e`; `t7_gen_for` has `g(0)`, `g(1)`,
`g(2)`, `g`, then the three `dut`; and `t8_gen_nest` has `g(0)`,
`g(0).h(0)`, `g(0).h(1)`, `g(0).h`, `g(1)`, its three, then `g`,
which ends at `0x12c8`, and the first `dut` puts its signal there and
its `k` at `0x1300`, then the other three `dut` `0x118` apart.

The region starts `0x580` past the end of the signals, with the
block of the root unit, and the signals themselves start at `0x738`,
each handle `0x30` into its block.
Tier 54 probes the start with a variable `a` of `tb.p` and takes the
libraries, the signal and the `std.env.stop` away in turn:

| Case | Packages in the file | `s` | `a` | Handle space |
| :--- | :--- | ---: | ---: | ---: |
| `t54_none_noenv__` | `standard` | none | `0x738` | `0x8cc` |
| `t54_noenv_sig___` | `standard` | `0x768` | `0x860` | `0xa14` |
| `t54_none_nosig__` | `standard`, `textio`, `env` | none | `0x810` | `0xa8c` |
| `t54_lib_none_var` | `standard`, `textio`, `env` | `0x768` | `0x938` | `0xbd4` |
| `t54_1164_noenv__` | `standard`, `textio`, `std_logic_1164` | `0x768` | `0xda0` | `0x1128` |
| `t54_nosig_var___` | those and `env` | none | `0xcb8` | `0x1090` |
| `t54_lib_1164_bit` | those and `env`, `s` a `bit` | `0x768` | `0xde0` | `0x11d8` |
| `t54_lib_numstd_v` | those and `numeric_std` | `0x768` | `0xed8` | `0x13d0` |
| `t54_lib_mathrl_v` | those and `math_real` | `0x768` | `0x10e8` | `0x15d8` |
| `t54_pkg_con_var_` | those and `pk` with one constant | `0x768` | `0xe10` | `0x1258` |

With `standard` alone and no signal the variable sits at `0x738`,
and with one `bit` signal, whose handle is `0x768`, at `0x860`: `0xf0`
for the signal's block and `0x38` for the signal in `tb`'s block, so
the signal's block runs from `0x738` and its handle is `0x30` into
it, and the root unit's block follows the signals at once.
`std.env`, which `std.env.stop` brings in with `std.textio`, moves
the variable by `0xd8` in both shapes; `std_logic_1164` with
`textio` by `0x540`; `env` on top of those by `0x40` more, so
`textio` is `0x98` of the `0xd8`; `numeric_std` by `0xf8` more and
`math_real` by `0x308` more.
Those are the blocks of the packages, between the signals and the
root unit, and the rest of what each package costs the handle space,
tier 47, lies past the second region: `0x78` for `textio`, `0x70`
for `env`, `0x15c` for `std_logic_1164`, `0x100` for `numeric_std`
and `0xf8` for `math_real`.
A package of the design has the block a scope has, `0x28` plus its
constants rounded up to 8: `pk` with one integer constant moves the
variable by `0x30`, its constant sits at `0xd40`, `0x98` before the
root unit's block, and the two constants of `t54_pkg_2con_var` at
`0xd40` and `0xd44` move it by `0x30` as well.
The package is in the file when a use clause names it or a name
refers into it, `t54_pkg_use_var_` and `t54_pkg_con_var_` alike, and
absent when neither does, `t54_pkg_unused__`, with the handle space
of a bench without it.
The type of the signal changes nothing, `t54_lib_1164_bit` against
`t52_var_int_____`, and neither does a signal's presence for the
packages' blocks: `t54_nosig_var___` puts the variable at `0x738`
plus `0x580`.

So with the usual `std_logic_1164` and `std.env.stop`, `tb` of tier
52 has `s` with its driver and `p`, `0x38` plus `0x20` plus `0x90`
plus `0x28`, a data object of `tb` or of `tb.p` sits at `0x828` plus
`0x580` plus `0x38`, `0xde0`, and the children start `0x110` past
`0xda8`, at `0xeb8`, whether there are one, two or three of them.
`t46_deep_100____` and `t46_gen_70000___` fit the same: `tb` with
`p` alone costs `0xb8`, the first generic or index is `0x38` past
that, `0x640` past the last signal's handle plus `0xf0`, and each
level of the chain costs `0x28` for its if generate, `0x30` for its
scope and generic, and `0x38`, `0x20` and `0x90` for its signal,
driver and process, `0x140`.
The handle space is 8 more than the strides account for when a scope
holds no data object: `0x58` for the two wrappers of
`t53_ifgen_inst__`, `0x30` for the `g` of `t52_gi2_empty___` against
the `0x11d0` the two integer cases imply, and that 8 is open.

The first region has its own driver costs, and they are less regular:
a concurrent assignment costs `0x50` where a process driver costs
`0x30`, `c` `0x110` apart in `t53_inst2_conc__` for `0xf0`, and two
process drivers cost `0x80`, `0x140` in `t53_inst2_2drv__` as in
`t46_drv_2_next__`.
None of it is needed: the handle is read from the record.

The decoder exposes the fifth word as `Object.Generic`, because it is 2
for exactly the objects that are not signals.

Generics of other types get handles the same way.
`t9_gen_types` has `kb : boolean`, `ks : string(1 to 3)`,
`kv : std_ulogic_vector(3 downto 0)` and `kr : real` at `0xe98`,
`0xe99`, `0xeac` and `0xec0`: 1, 19 and 20 apart for values of 1, 3, 4
and 8 bytes, which is the size plus 16 for the string and the vector,
the second region rule above.
Each is a declaration of kind `0x12` with the type's size in word 4,
and the corpus test checks name, type, size and recorded value against
`truth.json`.

A port bound to a literal, `a => '1'` in `t9_port_expr`, gets its own
handle `0x858` like an open port, in a second arena, at `0x1380` of
handle space against the `0x1288` of `t8_port_in`.
A component declaration with default binding, `t9_comp`, an alias of a
signal, `t9_alias`, a function with a variable, `t9_func`, and a
procedure with a variable declared in the process, `t9_proc_local`,
add no unit, scope, declaration or object.
The subprograms cost 8 bytes of handle space each, a procedure with a
`signal` parameter costs `0x48`, and none of that space shows up as
an object; see [container.md](container.md).


## Generics and entity naming

`t4_gen_same_two` and `t4_gen_diff_two` instantiate `child` twice, with
the same generic value and with different values.
The comparison shows how generics change the section:

| | `same_two` | `diff_two` |
| :--- | ---: | ---: |
| scopes | 7 | 7 |
| units | 5 | 7 |
| declarations | 2 | 4 |
| objects | 4 | 4 |

Two instances with equal generics share one `child(sim)` unit and one
set of declarations.
Two instances with different generics get a unit record each, still
named `child(sim)`, and their declarations are repeated.
The unit name never carries the generic value.
The scope names are the instance labels, `dut` and `dut2`, in both.
So the file does not name a parameterized entity differently.
A reader that wants to tell the two apart has to read the generic's own
value record.

A generic is an object: it has an instance record with the fifth word
set to 2, a declaration of kind `0x12`, and one value record at time 0.
The default and the explicit value produce the same layout, found by
`t4_gen_default` against `t4_gen_explicit`.

A Verilog module does not repeat its unit for a second parameter set.
`t21_v_param_diff` instantiates `child` with `K` 7 and with `K` 9 and
has one `child()` unit, two declarations and four objects, the same
counts as `t21_v_param_same` with 7 and 7.
The two `K` objects hold their own values.
So the VHDL repetition of `t4_gen_diff_two` is a VHDL property, not a
rule of the section.
*Found by* `t21_v_param_diff` against `t21_v_param_same`.

The repetition is not about generics as such: it follows every
declaration that elaborates differently.
An unconstrained port, `a : in std_ulogic_vector` in `t43_port_uncons`,
is declared with the bounds and the size of its actual, `(7 downto 0)`
and 8 bytes for an eight bit signal, `(0 to 7)` for an ascending one
in `t43_port_unc_asc`, and the file is otherwise that of a constrained
port.
Two instances bound to an eight bit and a four bit signal,
`t43_port_unc_two`, get two `child(sim)` units and two sets of
declarations, `a` with `(7 downto 0)` and 8 bytes under one and
`(3 downto 0)` and 4 bytes under the other, where two instances bound
to two eight bit signals, `t43_port_unc_sam`, share one unit and one
set: 7 units and 6 declarations against 5 and 4, with the same 7
scopes and 6 objects.
An unconstrained `out` port is the same, `t43_port_unc_out`, and so
is a port of a record with an unconstrained field, `t43_port_unc_rec`,
whose declaration carries the actual's `(7 downto 0)` and 16 bytes
while the record type, from a package, keeps the field unconstrained.
So a unit is one elaboration of an entity, and a reader that keys
units by entity and architecture name has to allow several.
*Found by* `t43_port_unc_two` against `t43_port_unc_sam`.
*Confirmed by* `t43_port_uncons` against `t8_port_vec8`, and
`t43_port_unc_asc`, `t43_port_unc_out` and `t43_port_unc_rec` against
`truth.json`.

The declaration of an array generic carries the range of the value it
ends up with, not the range of its subtype.
A generic declared as the unconstrained `std_ulogic_vector` with the
default `x"A"`, `t40_gen_uncons`, is declared with 4 bytes and the
range `(0 to 3)`, the ascending range a literal takes; the same generic
declared `std_ulogic_vector(3 downto 0)`, `t40_gen_cons`, is declared
with `(3 downto 0)`, and the two files differ nowhere else.
A string generic overridden on the command line takes the length of
the value given: `--generic_top ks=hello` over `ks : string := "abc"`,
`t40_gen_str_top`, declares `ks` as 5 bytes with `(1 to 5)` where the
default gives 3 bytes with `(1 to 3)`, and the one record holds
`hello`.
The handle of the next generic moves from `0xdf4` to `0xdf8` with it.
`//hdl/potato:sim` shows both at once: `RESET_ADDRESS`, an
unconstrained `std_logic_vector := x"00000200"` at the top, is
declared `(0 to 31)`, and the constrained `(31 downto 0)` generics of
the same name in `pp_core`, `pp_fetch` and `pp_decode` keep their
range; `IMEM_FILENAME` and `DMEM_FILENAME`, set by `-generic_top` to
paths of 19 and 44 characters over defaults of 17, are declared
`(1 to 19)` and `(1 to 44)`.
A reader that sizes a string generic from the source is wrong as soon
as the command line sets it.
*Found by* `//hdl/potato:sim` against `t9_gen_types`, then
`t40_gen_uncons` against `t40_gen_cons` and `t40_gen_str_top`
against `t40_gen_uncons`.
*Confirmed by* the corpus test's range check on the three tier 40
cases.

A VHDL 2008 type generic is not a declaration and not an object.
The type mapped to it enters the type table under the formal's name:
`generic (type data_t; init : data_t; next_v : data_t)` mapped to
`integer` in `t42_gen_type` gives one entry, `integer "data_t"
-2147483648 to 2147483647`, and no `INTEGER` entry at all; mapped to
`std_ulogic` in `t42_gen_type_enu` it gives `enum "data_t"` with the
nine literals of `STD_ULOGIC` and no `STD_ULOGIC` entry.
The signal and the two value generics of the child are declared with
that entry, 4 bytes for the integer and 1 for the enumeration, and
their records hold the values, `05 00 00 00` for `init => 5`.
So the name of a type in the table is the name the design used at the
declaration, and a reader that expects `INTEGER` for an integer has
to look at the kind instead.
*Found by* `t42_gen_type` against `t4_gen_explicit`, whose generic `k`
is declared with an `INTEGER` entry.
*Confirmed by* `t42_gen_type_enu`, and the corpus test's type check on
both.


## Mixed language

A VHDL testbench over a Verilog child, `t21_mix_v_in_vh`, and a Verilog
testbench over a VHDL child, `t21_mix_vh_in_v`, put both languages in
one file.
Each unit keeps its own kind: `entity tb(sim)` beside `module child()`,
and `process` beside `vprocess`.
Each declaration keeps its own language: the Verilog port is a `net`
of `logic` sized in bits, the VHDL port a `signal` of `STD_ULOGIC`
sized in bytes, and the type table holds an origin `2` entry beside an
origin `5` entry.
The child's source is in the file table, and the Verilog file numbers
land after the VHDL library files.

The port at the language boundary is not on the signal's handle.
`t9_comp` joins a VHDL signal and the VHDL port it drives on one
handle; both mixed cases give the port a handle of its own, and the
port's records are its own history.
A VHDL port under a Verilog testbench holds `U` at time 0 before the
driven `0`, the VHDL default of `STD_ULOGIC`, where the same port under
a VHDL testbench, `t9_comp`, holds the driven value only.
A Verilog port under a VHDL testbench holds `X` then `0`, as a Verilog
net with a driver does.
*Found by* `t21_mix_vh_in_v` against `t11_v_port` and `t9_comp`.
*Confirmed by* `t21_mix_v_in_vh` against `t9_comp`.

Tier 49 adds an output at the boundary and a second boundary.
`t49_mix_2port___` gives the Verilog child an `assign b = a` output
into the VHDL signal `y`, and `y` holds `U`, `X`, `0`, `1`: the VHDL
default, then the `X` the assign produced from the port's `X`, then
the driven values.
`t49_mix_deep____` puts a VHDL leaf under the Verilog child, so a
VHDL signal crosses into Verilog and back.
The leaf's input port holds `0` then `1` and no `U`, where the VHDL
port of `t21_mix_vh_in_v` held `U` first: the Verilog net above it
carries the VHDL `0` from elaboration, where the `reg` of
`t21_mix_vh_in_v` gets its `0` from an `initial` at time 0.
The leaf's output port holds `U`, `0`, `1`, and `y` above it the same,
without the `X` of `t49_mix_2port___`, because no Verilog expression
sits between them.
Every boundary port has a handle of its own on both levels: `tb.x` at
`0x768`, `tb.dut.a` at `0x960` and `tb.dut.u.a` at `0xb28`.

*Found by* `t49_mix_deep____` against `t49_mix_2port___`.


## For generate

`t7_gen_for` elaborates `g: for i in 0 to 2 generate` with one
instance of `child` per iteration, `generic map (k => i)`.

```
[0]  _top
[1]  tb
[2]  tb.\g(0)\          unit 2, generate, 1 declaration: i
[3]  tb.\g(1)\          unit 3, generate, 1 declaration: i
[4]  tb.\g(2)\          unit 4, generate, 1 declaration: i
[5]  tb.g               unit 5, generate, no declarations, no children
[6]  tb.p               unit 6, process
[7]  tb.\g(0)\.dut      unit 7, child(sim), 2 declarations: s, k
[8]  tb.\g(1)\.dut      unit 8, child(sim)
[9]  tb.\g(2)\.dut      unit 9, child(sim)
[10] tb.\g(0)\.dut.p    unit 10, process
[11] tb.\g(1)\.dut.p    unit 11, process
[12] tb.\g(2)\.dut.p    unit 12, process
```

Each iteration is a scope whose leaf name is the label with the index
in parentheses, spelled as a VHDL extended identifier: `\g(0)\`, with
the backslashes in the name pool.
A fourth scope named plainly `g` sits beside them with no children,
no declarations and no objects.
All four have unit kind `0x0c`, which no other case produces, and no
entity or architecture name.
Their file and line point at the `generate` statement.

Each iteration scope declares the loop index `i`, kind `0x13`, and
that object gets one record at time 0 holding the iteration's value,
0, 1 and 2.
That is unlike a process loop index, which records 0; see
[values.md](values.md).

The three `child` instances have different generics, and as in
`t4_gen_diff_two` that costs a `child(sim)` unit each with its own two
declarations, so the section holds 13 scopes, 13 units, 9 declarations
and 9 objects.
The generic `k` records 0, 1 and 2.

A reader that matches paths from another tool has to strip the
backslashes, or add them.
The decoder's test compares `tb.g(0).dut.s` from `truth.json` against
the pool's `tb.\g(0)\.dut.s` that way.

*Found by* `t7_gen_for`, the first generate case, which failed on both
the arena record order and the path spelling before the reader was
taught both.

A for generate whose body declares nothing and holds no statement
elaborates to the plain label scope alone.
`t52_gi2_empty___` has `g: for i in 0 to 1 generate end generate;`
and the file holds `tb.g` with no children, one generate unit with no
declarations, and no index object at all, where `t52_gi2_proc____`,
whose iterations each hold a process, has `\g(0)\` and `\g(1)\`
beside `g` with an index each.

*Found by* `t52_gi2_empty___` against `t52_gi2_proc____`.


## Nested for generate

`t8_gen_nest` nests `h: for j in 0 to 1 generate` inside
`g: for i in 0 to 1 generate`, with one `child` per inner iteration.
The shape repeats at every level.

```
[1]  tb
[2]  tb.\g(0)\                   generate, 1 declaration: i
[3]  tb.\g(1)\                   generate, 1 declaration: i
[4]  tb.g                        generate, empty
[5]  tb.p                        process
[6]  tb.\g(0)\.\h(0)\            generate, 1 declaration: j
[7]  tb.\g(0)\.\h(1)\            generate, 1 declaration: j
[8]  tb.\g(0)\.h                 generate, empty
[9]  tb.\g(1)\.\h(0)\            generate, 1 declaration: j
[10] tb.\g(1)\.\h(1)\            generate, 1 declaration: j
[11] tb.\g(1)\.h                 generate, empty
[12] tb.\g(0)\.\h(0)\.dut        child(sim), 2 declarations: s, k
...
```

Every outer iteration gets its own empty `h` beside its two `\h(j)\`
scopes, and one empty `g` sits beside the `\g(i)\` scopes.
The four instances have four different `k`, so there are four
`child(sim)` units: 20 scopes, 20 units, 14 declarations and 14
objects.

*Found by* `t8_gen_nest` against `t7_gen_for`.


## If generate

`t8_gen_if` has `g: if with_dut generate` and
`h: if not with_dut generate`, with `with_dut` an architecture
constant `true`, and one `child` in each branch.

```
[1]  tb              entity, 1 declaration: with_dut
[2]  tb.g            generate, no declarations
[3]  tb.h            generate, no declarations, no children
[4]  tb.p            process
[5]  tb.g.dut        child(sim), 2 declarations: s, k
[6]  tb.g.dut.p      process
```

An `if` generate is one scope per label, named plainly, unit kind
`0x0c` like a `for` generate.
The false branch is still there, as an empty scope, so a reader cannot
tell which condition held from the scope list alone; it has to look for
children.
There is no index and no extended identifier.

The constant `with_dut` is a declaration of kind `0x13` in the entity
unit, an object in `tb`, and gets one record at time 0 holding its
value, like a generic.
That is what made `0x13` a constant kind rather than a loop index kind.

*Found by* `t8_gen_if` against `t7_gen_for` and `t4_gen_explicit`.


## Case generate

`t24_case_gen` has `g: case k generate` over an architecture constant
`k : integer := 1`, with the alternatives `when zero: 0 =>` declaring
a signal `a` and `when one: others =>` declaring a signal `b`, each
with one concurrent assignment.

```
[1]  tb              entity, 1 declaration: k
[2]  tb.g            generate, 1 declaration: b
[3]  tb.p            process
[4]  tb.g.line__26   process, the assignment of the taken alternative
```

A `case` generate is one scope named by the generate label, unit kind
`0x0c` like the other generates, holding the declarations of the
alternative that was taken.
The alternative label `one` is not a scope and is not in the name,
and the alternative not taken leaves nothing, where an `if` generate
keeps its false branch as an empty scope.
`k` is a constant of `tb` with one record holding `1`, as `with_dut`
of `t8_gen_if`.

*Found by* `t24_case_gen` against `t8_gen_if`.


## Blocks

A `block` statement is a scope of unit kind `0x0c`, the kind of a
generate, under the architecture that holds it.
`t9_block` has `b: block` with a signal `y` and a concurrent
assignment, and the file holds scopes `tb`, `tb.b`, `tb.p` and
`tb.b.line__20`, with `y` an object of `tb.b` on its own handle.
A block and a generate are not told apart by the unit kind.
The unit's file and line point at the `block` line.

*Found by* `t9_block` against `t7_gen_for`.


## Packages

A package that declares an object is a scope of its own.
`t9_port_rec` uses a package `pair_pkg` with a record type and a
constant `zero` of that type, and the file holds:

| Scope | Parent | Unit kind | First object |
| :--- | ---: | ---: | ---: |
| `_top` | `-1` | `0x13` | none |
| `tb` | 0 | `0x09` | 0 |
| `pair_pkg` | 0 | `0x0a` | 1 |
| `tb.dut` | 1 | `0x09` | 2 |
| `tb.p` | 1 | `0x0d` | none |

The package is a child of the root next to `tb`, not under it, and its
unit has no entity file or line.
Its constant is a declaration of kind `0x13`, the constant kind, and an
object with handle `0xd40` and no records.
`t9_mark_two` and `t9_mark_gap` have the same for an integer constant,
and `t9_pkg_sig` has a package signal as an object with the first
handle `0x768` and no records; see [values.md](values.md).
A package with only a type in it gets the scope too, with no object:
`trio_pkg` of `t34_pmap_field`, one record type, and `pp` of
`t42_pkg_subtype`, one subtype.
This was recorded until tier 42 as "no scope, as in `t2_record`", but
`t2_record` declares its type in the architecture and has no package.

A SystemVerilog package takes the same place: `p` of `t13_sv_pkg`,
imported into `tb` for a typedef and a parameter, is the second child
of the root beside `tb`, with a unit of kind `0x08` and the parameter
`W` as its object.
The typedef is a type entry and not an object, as in VHDL.
Its scope and object cost handle space: `t13_sv_pkg` has `0xafc`
where `t12_sv_typedef`, the same typedef inside `tb`, has `0x91c`.

A SystemVerilog package enters the file when a declaration uses one of
its types, and its parameters come along whether they are used or not.
`t25_sv_pkg_tdef`, a package with only `typedef logic [7:0] byte_t`
and a `byte_t s` in `tb`, has the unit and the scope `p` with no
object.
`t25_sv_pkg_prm`, a package with only `parameter int W = 8` used in
the cast `W'(8'ha5)`, and `t25_sv_pkg_unusd`, the same package
imported and not used, have no unit, no scope and no object for `p`,
and `W` is nowhere in the file.
`t13_sv_pkg` uses `W` in the same cast, and has the object `p.W`
because `byte_t` of the same package types its `s`.
The cast still costs handle space: `t25_sv_pkg_prm` has `0xac4` and
`t25_sv_pkg_unusd` `0x9cc`, `0xf8` apart, the stride of a dynamic
object; see the Verilog section below.

*Found by* `t9_port_rec` against `t2_record`, where the extra scope
sat between `tb` and `tb.dut` in the scope list.
*Confirmed by* `t13_sv_pkg` against `t12_sv_typedef`, and
`t25_sv_pkg_prm` against `t25_sv_pkg_tdef` for the condition.

A VHDL 2008 generic package instance is a package like any other, and
the instance is invisible.
`package gp8 is new work.gp generic map (n => 8)` in `t42_gen_pkg`,
with `subtype word_t is std_ulogic_vector(n - 1 downto 0)` in `gp`,
gives a scope named `gp`, not `gp8`, whose unit points at the file and
line of `package gp is`, and a constrained `word_t` entry
`(7 downto 0)`.
The generic `n` is nowhere: no declaration, no object, no type entry.
`t42_pkg_subtype`, a plain package `pp` with the same subtype written
out, differs from it in the scope name and the file paths and in
nothing else.
Two instances, `gp8` and `gp4` in `t42_gen_pkg_two`, give two scopes
both named `gp`, two package units both at the line of `package gp
is`, and two entries both named `word_t`, one `(7 downto 0)` and one
`(3 downto 0)`.
A reader that keys packages by name sees one; the scope index and the
type index tell them apart, and the declaration of a signal points at
the right entry.
A constant of the generic, `constant width : natural := n` in
`t42_gen_pkg_cons`, is an object under `gp` like any package constant,
not logged and with no record, so the value of `n` is not in the file
even then.
A package instantiated inside an architecture does not elaborate:
`xelab` stops with `The "Vhdl 2008 Package Instantiation Declaration
in Architecture Body" is not supported yet for simulation`, so the
corpus has no such case.
*Found by* `t42_gen_pkg` against `t42_pkg_subtype`, the same subtype
in a plain package.
*Confirmed by* `t42_gen_pkg_two` and `t42_gen_pkg_cons`.


## Implicit processes

A concurrent signal assignment such as `q <= a;` in an architecture is
an implicit process.
It gets a process scope named `line__NN`, where `NN` is the source line
of the statement, with a process unit pointing at the same line.
`t8_port_open` has `tb.dut.line__18` and `t8_port_vec8` has the same,
both from one concurrent assignment in `child`.

*Found by* `t8_port_open`, whose child has no labelled process.

A concurrent assignment from a signal attribute is an implicit
process too, and `'delayed` makes two of them.
`t24_att_delayed` has `d <= s'delayed(2 ns);` at line 18, and the
file has two scopes named `tb.line__18`, units 2 and 4, each a
process unit pointing at line 18, beside `tb.p`.
`t24_att_stable`, `t24_att_quiet` and `t24_att_transact` have
`b <= s'stable(1 ns)`, `q <= s'quiet(1 ns)` and `t <= s'transaction`
at the same line and one `tb.line__18` each.
None of the four has a declaration or an object for the implicit
signal the attribute denotes: the declarations are `s` and the
target, and the target records the attribute's value.
`d` changes at 52 ns for the change of `s` at 50 ns; `b` and `q` go
false at 50 ns and true at 51 ns; `t` toggles at 50 and 70 ns, where
the second assignment of `s` repeats its value and leaves no record
on `s`.
An external name has the same shape: the assignment
`a <= << signal .tb.dut.s : std_ulogic >>;` at line 19 of
`t24_ext_name` is the scope `tb.line__19`.

*Found by* `t24_att_delayed` against `t24_att_stable`, two scopes
against one for the same statement shape.
*Confirmed by* `t24_att_quiet`, `t24_att_transact` and
`t24_ext_name`.


## Variables

A variable declared in a process is a declaration of kind `0x0f` in the
process's unit, an object in the process's scope, and nothing else.
No arena is allocated for it and no value record is ever written.
Found by `t6_var_int`, whose variable changes twice and leaves no
trace, and `t6_proc2`, which has two.

A `for` loop index in a process is a declaration of kind `0x13`, the
constant kind.
It gets one record at time 0, holding 0, whatever the loop's first
value is.
Found by `t5_tr1000` and `t6_tr1300`, whose loops run `0 to 999` and
`0 to 1299`.
A `for` loop index in a subprogram is not in the file, even under
`-debug subprogram`: `t51_sub_loop_idx` loops `for i in 0 to 3` in a
procedure and the procedure's declarations are its signal parameter
and its variable `n` alone.
Found by `t51_sub_loop_idx` against `t23_sub_sig_prm_`.
A `for generate` index and an architecture constant are the same kind
and record their value instead; see above.

Both kinds count toward the marker only when they have a record.
See the marker section of [container.md](container.md).

The objects of a process variable multiply with the instances of its
entity.
`t9_var_inst3` instantiates `child` three times, and `child` has one
process with one variable `v`.
The variable is one declaration in the process's unit, unit 4, shared
by the three process scopes `tb.d0.p`, `tb.d1.p` and `tb.d2.p`.
Each of those scopes lists three objects for `v`, on the three handles
`0x11c0`, `0x12d8` and `0x13f0`, `0x118` apart, so the file holds nine
objects for three variables.
`t9_mark_two` does the same with two instances: four objects, on
`0x1100` and `0x1218`, two per scope.
`t8_gen_nest`, whose four instances take different generics and so get
four units and four declarations each, lists one object per variable.
So the multiplication happens when instances share a unit, and which
of the shared handles belongs to which scope cannot be told from the
file.
It does not matter to the reader, because a variable has no records,
and the corpus test counts objects by distinct path.

*Found by* `t9_mark_two` against `t6_var_int`, written to put a
second logged range after an unlogged object, and showing eight objects
where six were expected.
*Confirmed by* `t9_var_inst3`.

A shared variable declared in the architecture is the same kind
`0x0f`, in the entity's unit and the entity's scope, with one handle
and no record: `sv` of `t23_shared_int`, an integer assigned once from
the process, sits on `0xde0` as the process variable `v` of
`t6_var_int` does, and only the scope it is listed under differs.
A shared variable of a protected type is the exception: `c` of
`t23_protected` is kind `0x0f` too, but its object has both handles,
as a signal does, and sits at `0x858`, the next signal handle after
`s`, rather than in the variable handles.
It is not logged and has no record either.
Its type is a record entry, see [types.md](types.md).

A file object is a variable of kind `0x0f` as well, in the unit and
scope of the architecture that declares it, on the variable handle
`0xde0` with no second handle and no record, and declares 0 bytes:
`f` of `t23_file_text`, `t23_file_int` and `t23_file_sul`.
So the `textio` files that `-debug all` showed in tier 22 are the
ordinary shape of a file object, not something the debug level adds.
A variable of an access type is a process variable of 48 bytes with no
record: `p` of `t23_access` and `t23_access_vec`.

*Found by* `t23_shared_int` against `t6_var_int`, `t23_protected`
against `t23_shared_int`, and `t23_file_text` against `t22_dbg_all`.
*Confirmed by* `t23_file_int`, `t23_file_sul`, `t23_access` and
`t23_access_vec`.


## Subprograms

A VHDL function or procedure leaves nothing in the file under the
default `-debug typical`: `t22_base` declares a function `inc` with a
parameter and a variable and has no scope, unit or object for either.
Under `-debug subprogram`, given beside `typical`, the subprogram
becomes a unit and a scope under the entity that declares it, and its
parameters and variables become declarations and objects.
A function is a unit of kind `0x11` and a procedure one of kind
`0x12`, each with the file and line of its declaration and the two
entity words at 0, as a process.
The scope is named after the subprogram: `tb.inc`, `tb.bump`.
Each parameter and each variable is a declaration of kind `0x14`, a
kind no other object has, and a parameter carries its mode in word 9
as a port does: `x` of `inc` is `port in`, `a` and `d` of `bump` are
`port inout` and `port in`, and the variables `v` and `w` have no
mode.
The objects have both handles, as a signal does, and are never logged.
Their handles are small and step by the size of the value: `x` and `v`
of `inc` are `0x40` and `0x44`, and `a`, `d` and `w` of `bump` are
`0xd0`, `0xd4` and `0xd8`, so each subprogram numbers its locals from
its own base, and the number looks like an offset into the
subprogram's frame rather than a place in the handle space.
They do not move the handle space: `t22_dbg_subprog` has `0x1468` as
`t22_base` does.
A composite local and a composite literal do, by the bytes of the
value, see the static values paragraph below.
`-debug subprogram` on its own, `-debug line` and `-debug off` produce
no database at all: xsim refuses `log_wave` with "compiled without
trace information", and asks for `all` or `typical`.

*Found by* `t22_dbg_subprog` against `t22_base`.
*Confirmed by* `t22_dbg_sub_proc`, which adds the procedure, and
`t22_dbg_all`.

The frame offsets follow the alignment of each local.
`t23_sub_sizes` declares a function with parameters `c : std_ulogic`
and `n : integer` and variables `r : real`,
`w : std_ulogic_vector(7 downto 0)` and `m : integer`, and their
handles are `0x40`, `0x44`, `0x48`, `0x50` and `0x68`: the 1 byte
`c` is followed by `n` at the next multiple of 4, `r` at the next
multiple of 8, and `w` after the 8 bytes of `r`.
`m` follows `w` at `0x68`, 24 bytes later, and stays at `0x68` when
`w` is 16 elements in `t23_sub_vec16` and 32 elements in
`t23_sub_vec32`, while the declaration record's size grows from 8 to
16 to 32.
So a vector local occupies 24 bytes of the frame whatever its length,
which is the size of a descriptor rather than of the elements.
*Found by* `t23_sub_sizes` against `t22_dbg_subprog`.
*Confirmed by* `t23_sub_vec16` and `t23_sub_vec32`.

A signal parameter is a declaration of kind `0x15`, its own kind,
with its mode in word 9: `q` of `procedure drive(signal q : out
std_ulogic; constant v : in std_ulogic)` in `t23_sub_sig_prm` is
`0x15` `port out`, and `v` beside it is the ordinary `0x14` `port
in`.
`q` is on `0xd0`, the procedure base of `t22_dbg_sub_proc`, and `v`
on `0x110`, 64 bytes later, so a signal parameter takes 64 bytes of
the frame where a `std_ulogic` value takes one.
The object has both handles and no record, and the signal `s` passed
to `q` keeps its own handle and records.
*Found by* `t23_sub_sig_prm` against `t22_dbg_sub_proc`.

The 64 bytes of a signal parameter start on a multiple of 8:
`procedure show(variable v : in integer; signal q : out std_ulogic)`
in `t50_sub_in_var__` puts `v` on `0xd0` and `q` on `0xd8`, and
`t50_sub_acc_loc_` and `t50_sub_str_loc_` put a local after a signal
parameter on `0x110`.
A vector parameter is the 24 byte descriptor a vector local is: `a`
of `function low(a : std_ulogic_vector(3 downto 0))` in
`t50_sub_func_prm` is on the function base `0x40` and the scalar local
after it on `0x58`.
An `inout` `variable` parameter of a vector or record type is on the
procedure base `0xd0` in `t50_sub_var_vec_` and `t50_sub_var_rec_`,
and nothing follows it to show its size.
*Found by* `t50_sub_in_var__` against `t23_sub_sig_prm_`, and
`t50_sub_func_prm` against `t49_sub_vec_prm_`.

A procedure declared in a package is a scope under the package:
`t51_sub_pkg_proc` calls `work.pk.drive(s, '1')` and the file has the
package scope `pk` beside `tb` under the root, with `pk.drive` as its
child, a procedure unit at the line of the body's declaration, and the
parameters on `0xd0` and `0x110` as in `t23_sub_sig_prm_`.
The scopes are listed breadth first, `tb`, `pk`, `tb.p`, `pk.drive`.
*Found by* `t51_sub_pkg_proc` against `t23_sub_sig_prm_`.

A procedure declared inside a process, the shape of `t9_proc_local`,
gets two scopes: `t23_sub_in_proc` declares `flip` in process `p` and
the file has `tb.flip` as a child of `tb`, listed before `tb.p`, and
`tb.p.flip` as the child of `tb.p`.
Both point at the same unit, the one procedure unit of kind `0x12`,
and each lists an object for the local `r` on the same handle `0xd0`,
so the file holds three objects for two declarations.
*Found by* `t23_sub_in_proc` against `t22_dbg_sub_proc`.

A function declared inside a function gets the same two scopes.
`t55_sub_nested__` declares `g` inside `f`, and the file has `tb.g`
as a child of `tb`, listed before `tb.p`, and `tb.f.g` as the child
of `tb.f`, both on the one function unit of `g`, and each with its
own objects for the parameter `n` and the local `w` on `0x40` and
`0x44`, the function base, so the nested function numbers its frame
from the base as a top level function does.
So the file holds six objects for the four declarations of `f` and
`g`, beside the signal.
*Found by* `t55_sub_nested__` against `t50_sub_func_prm`.

The methods of a protected type are subprogram scopes too, under
`-debug subprogram`, and they come in two copies as a nested
subprogram does.
`t55_sub_prot_loc` declares `type counter_t is protected` with a
procedure `bump` and an impure function `get` in the architecture,
and a local `ct : counter_t` in a procedure, and the file has units
for `bump` and `get`, at the lines of the protected body, with no
declarations, and the scopes `tb.bump` and `tb.get` twice each, all
four children of `tb`, the first pair before `tb.drive` and `tb.p`
and the second pair after them.
The variable `n` of the protected body is in the file nowhere, and
neither is `ct`.
The scopes do not move the handle space, which is the `0x11e8` of
`t51_sub_loop_idx`.
The type alone brings nothing: `t55_sub_prot_typ` declares the type
and no variable of it and has neither the units nor the scopes.
A shared variable of the type brings the same four scopes, whether
its methods are called from a procedure in `t55_prot_shared_` or from
the process in `t55_prot_arch_pr`, and with two processes in
`t55_prot_arch_2p` the second pair is still under `tb`, after
`tb.p2`.
The shared variable is the object of tier 23, on the signal handle
`0x858` or `0x870`, and costs `0x100` of handle space against
`t55_sub_prot_typ`.

When the protected type is declared in a package the second pair
moves.
`t55_prot_pkg____` declares the type in `pk` and the shared variable
in the architecture, and the file has `pk.bump` and `pk.get` under
`pk`, and `tb.p.bump` and `tb.p.get` under the process `tb.p`, the
only process, on the same two units.
With two processes `p` and `p2` the second pair is under `tb.p2`, the
last one declared, in `t55_prot_pkg_2p_` where `p2` calls the methods
first and in `t55_prot_pkg_2pl` where it calls them last, and
`t55_prot_pkg_prc`, whose process calls the methods itself, has them
under `tb.p` as `t55_prot_pkg____` does, whose procedure calls them.
So the second pair sits under the last process of the architecture
that declares the variable, whatever calls the methods and when.
A process scope can therefore have children, and the reader's scope
tree allows it.
`t55_prot_pkg_sv_` declares the shared variable in the package as well
and reaches it through two package subprograms named `bump` and `get`
like the methods.
The file has `pk.bump` and `pk.get` twice under `pk`, first the
methods' units and then the package subprograms' units, and
`tb.p.bump` and `tb.p.get` under the process on the methods' units
alone, so the package subprograms get one scope as `pk.drive` of
`t51_sub_pkg_proc` does, and the package shared variable is absent as
a package constant is not.
*Found by* `t55_sub_prot_loc` against `t51_sub_loop_idx`, and
`t55_prot_pkg____` against `t55_prot_shared_`.
*Confirmed by* `t55_sub_prot_typ`, `t55_sub_prot_2__`,
`t55_prot_arch_pr`, `t55_prot_arch_2p`, `t55_prot_pkg_prc`,
`t55_prot_pkg_2p_`, `t55_prot_pkg_2pl` and `t55_prot_pkg_sv_`.


## Two architectures of one entity

An entity with two architectures gets one unit per architecture that
is instantiated, and none for the others.
`t23_arch_b` instantiates `entity work.child(b)` of a `child` with
architectures `a` and `b`, and the file has the one unit `child(b)`,
with the unit's line at the `architecture b` line and its entity line
at the entity, plus the one process unit of that architecture.
`t23_arch_both` instantiates both as `da` and `db`, and the file has
`child(a)` and `child(b)` as separate units, each with its own
declaration of `s` and its own process unit, so the two instances
share nothing, as two instances with different generics do in
`t8_gen_nest`.
The architecture name in the unit record is therefore the one the
instance chose, and `t4_gen_explicit`, whose child has one
architecture `sim`, is the degenerate case.

*Found by* `t23_arch_b` against `t4_gen_explicit`.
*Confirmed by* `t23_arch_both`.

A configuration specification chooses the architecture the same way.
`t24_config_spec` declares `component child`, binds it with
`for dut : child use entity work.child(a);` and instantiates the
component, and the file has the unit `child(a)`, where `t23_arch_b`
instantiates `entity work.child(b)` of the same child and has
`child(b)`.
The scope `tb.dut`, its process and its object `s` are the same in
both, so the component and the binding leave no trace beyond the
architecture name.

*Found by* `t24_config_spec` against `t23_arch_b`.


## Library packages under `-debug all`

`-debug all` adds `xlibs`, visibility into the precompiled libraries,
and the file grows from 6783 to 11814 bytes on the same source.
The packages the design uses become children of the root beside `tb`,
each a unit of kind `0x0a` as a user package is: `env`, `standard`,
`textio` and `std_logic_1164`, pointing at the `package` line of the
library source.
Their constants and variables are declarations in the package's unit
and objects in its scope, with one handle and no records, as a user
package constant is: `env.DIR_SEPARATOR`, `textio.INPUT` and
`textio.OUTPUT`, and the twelve tables and `NBSP` of
`std_logic_1164`.
`standard` has a unit and a scope with no declaration and no object.
The function `resolved` of `std_logic_1164` is a scope of kind `0x11`
under its package with its variable `result` as a local, and the type
table grows from 3 to 19 entries with the types of those constants:
`character` with all 256 literals, `STRING`, the `TEXT` file type and
the nine element tables of the package.
The value records and the handles of the design's own objects do not
move.

*Found by* `t22_dbg_all` against `t22_base`.

`xlibs` alone brings the packages.
`t24_dbg_xlibs` runs the two driver design of `t24_two_drivers` under
`-debug wave -debug xlibs`, and the file has the same four packages
as root children with the same 15 constants and variables, the file
growing from 5229 bytes under `-debug wave -debug drivers` to 9772.
It has no `resolved` scope and no `result` local, so those come from
a level that `all` adds beyond `xlibs`.

*Found by* `t24_dbg_xlibs` against `t24_dbg_drv_only`.


## Elaboration options that change nothing

`--O0`, `--mt off` and `-debug drivers` over `-debug typical`, which
already includes it, produce a file byte identical to the default
outside the noise mask.
`--generic_top k=9` changes the value record of `k` from 7 to 9 and
the record of `n`, which is computed from it, and nothing else: the
same unit, the same declaration and the same handles as the default of
7.
A Verilog top elaborated with `-generic_top` is another matter: the
top scope and its unit of `//hdl/serv:sim`, run with `-generic_top
memfile=...`, are named

```
tb(memfile="external/+http_archive+serv/sw/hello_uart.hex")
```

the parameter and its value spelled into the name, and the VCD's
`$scope module` line carries the same name.
Every object path below starts with it, and `TestVCD` strips the
parenthesis when it decides what the VCD may leave out.

*Found by* `t22_o0`, `t22_mt_off` and `t22_gen_top`, each against
`t22_base`, and `t24_dbg_drivers` against `t24_two_drivers`; the
Verilog top name by `//hdl/serv:sim` against `t22_gen_top`.


## Debug levels

`xelab -debug` takes `line`, `wave`, `drivers`, `readers`, `xlibs`
and `subprogram`, with `typical` for `line`, `wave` and `drivers`
and `all` for everything but `subprogram`, according to its help.
The corpus runs under `typical` unless a case says otherwise.
Tier 24 elaborates the two driver design of `t24_two_drivers` under
one level at a time, and the levels show in three places: header
words 14 and 15, the statement regions 14 and 15, and the packages
of `xlibs` above.

| Case | `-debug` | Word 14 | Word 15 | Word 11 |
| :--- | :--- | ---: | ---: | ---: |
| `t24_dbg_drv_only` | `wave drivers` | `0x101` | `0x1` | 0 |
| `t24_dbg_line` | `wave line` | `0x1` | `0x10101` | 9 |
| `t24_dbg_sub_only` | `wave subprogram` | `0x1` | `0x10101` | 9 |
| `t24_dbg_xlibs` | `wave xlibs` | `0x1` | `0x1` | 0 |
| `t24_two_drivers` | `typical` | `0x101` | `0x101` | 9 |
| `t24_dbg_drivers` | `typical drivers` | `0x101` | `0x101` | 9 |
| `t24_dbg_readers` | `typical readers` | `0x10101` | `0x101` | 9 |

`t22_dbg_wave` under `wave` alone has `0x1` in both words, and
`t22_dbg_all` and the `typical subprogram` cases of tiers 22 and 23
have `0x101` and `0x10101`.
So byte 0 of each word is always `1`; byte 1 of word 14 is
`drivers` and byte 2 is `readers`; byte 1 of word 15 is `line` and
byte 2 is `subprogram`, which `line` alone also sets and `typical`
does not, though `typical` includes `line`.
Why `line` alone sets byte 2 of word 15 is open.
`xlibs` sets no byte.
The statement index and lines of regions 14 and 15 exist under
`line`: word 11 counts 9 statement lines in every case that has it
and 0 in the others, where regions 15 and 16 start at the same
offset.
`readers` over `typical` changes the one byte and nothing else
outside the noise mask.
Handles, records and the type table are the same under every level
but `xlibs`.

*Found by* `t24_dbg_readers` against `t24_two_drivers`, one byte at
file offset `0x303`, and `t24_dbg_line` against `t24_dbg_drv_only`
for the two words and region 15 moving together.
*Confirmed by* `t24_dbg_sub_only`, `t24_dbg_xlibs`, `t24_dbg_drivers`
and the `header words` line of the dump of every case.


## Verilog modules, processes and nets

Tier 11 repeats the hierarchy cases in Verilog.
The debug section has the same regions and record shapes, and the
differences are in the kind words, the scope names and the handle
strides.
Every claim below reproduces with `wdbcvt -dump`.

**Units.**
A module is a unit of kind `0x00`, with the module name at word 0 and
no architecture name.
Words 5 to 8 all point at the `module` line.
A process is a unit of kind `0x07` with words 7 and 8 zero, as a VHDL
process, and a named block, `begin : blk`, is a unit of kind `0x05`.
The root is `0x13`, as for VHDL.

A `task` is a unit of kind `0x03` and a `function` one of kind `0x04`,
each with a scope named after it under the module, `tb.inc` at the
`task` or `function` line, and the arguments and locals of the
subprogram are declarations of that unit and objects of that scope.
`t12_v_task` declares `v`, an `input` argument with port mode `1` and
`(7 downto 0)`, then the local `tmp`; both hold `X` records at time 0
and their values at the call.
`t12_v_func` declares the return variable `inc` first, then `v` and
`tmp`, and the call writes `v`, `tmp` and then `inc`.
The subprogram scope has no unit name and no process scope of its own.
A SystemVerilog function called from a process is laid out the same
way: `function int f()` in `t29_sv_fn_noc` is a function unit with
the scope `tb.f` holding the object `tb.f.f`, which records `0` at
time 0 and `3` at the call.
A function called only from an initializer leaves nothing: `logic s =
f()` in `t26_sv_logic_fn`, with `f` returning `1'b0`, has the units,
scopes, declarations and handle space of `t11_sv_logic`, and `s`
records `0` once, so the call was folded at elaboration.

An `automatic` subprogram keeps its unit and scope and loses its
arguments and locals: `task automatic inc(input logic [7:0] v)` with
a local `tmp` in `t51_sv_task_auto` has the `tb.inc` scope with no
first object, the task unit with no declarations, and the handle space
`0xbb4` for the `0xc14` of the static task of `t51_sv_task_stat`.
A `ref` argument, `t51_sv_task_ref_`, and an `automatic` function,
`t51_sv_func_auto`, are the same.
A `static` local of an `automatic` task is back in the file:
`t51_sv_task_stvr` lists `tb.inc.tmp` and nothing for the argument.
The `output` and `inout` arguments of a static task are listed with
their modes, `t51_sv_task_out_` and `t51_sv_task_inou`, and every
argument holds 0 in the word at `40` of its instance record, whatever
its place in the argument list, where a module port holds its
position.
*Found by* `t51_sv_task_auto` against `t51_sv_task_stat`.
*Confirmed by* `t51_sv_task_ref_`, `t51_sv_func_auto`,
`t51_sv_task_stvr`, `t51_sv_task_out_` and `t51_sv_task_inou`.

A cast in an initializer leaves a hidden variable in the module scope:
`int s = int'(1.5)` in `t28_sv_int_cast` declares
`xilinx_isim_temp_0_ln5castingOp`, an `int` at file 0 line 0, as the
module's first declaration and first object, at `0x768` before `s` at
`0x828`, and adds a second process unit and scope at the line of the
declaration, `tb.Initial5_1`.
The name carries the line of the cast, `ln6` in `t28_sv_enum_cast`,
whose hidden variable is a `state_t`, and the variable has the type
of the cast, an unnamed 8 bit vector for `8'(0)` in `t28_sv_v8_szcast`.
It is logged, holds one record at time 0 with the cast's value, and
costs `0x190` of handle space over `t11_sv_int`, `0xaac` for `0x91c`.
Each cast in an initializer leaves one, numbered through the module:
`int s = int'(1.5)` and `int t = int'(2.5)` in `t29_sv_cast_two`
declare `xilinx_isim_temp_0_ln5castingOp` before `s` and
`xilinx_isim_temp_1_ln6castingOp` before `t`, so the objects sit at
`0x768`, `0x828`, `0x8e8` and `0x9a8`, and the two initializers share
one implicit process, `tb.Initial5_1`.
The count starts again in a child module: `t29_sv_cast_sub` has
`tb.u.xilinx_isim_temp_0_ln5castingOp` before `tb.u.s`.
The variable takes the type of the cast, and its value class with it:
an 8 bit vector of class 1 for `signed'(8'h05)` in `t29_sv_cast_sgn`
and a `real` of class 0 for `real'(3)` in `t29_sv_cast_real`, each
in a module of `0xaac` handle space like `t28_sv_int_cast`.
Two casts in one initializer leave nothing: `int'(1.5) + int'(2.5)` in
`t29_sv_cast_same` has the declarations, units and `0x91c` of
`t11_sv_int`, and `s` records `5` once, so the sum was folded.
No other place for a cast leaves a variable, though most cost handle
space.
`s = int'(2.5)` in a process, `t29_sv_cast_proc`, has the objects of
`t29_sv_incr` and `0xb0c` for its `0x91c`; `return int'(2.5)` in a
function, `t29_sv_cast_fn`, has those of `t29_sv_fn_noc` and `0xde4`
for `0xbf4`, the same `0x1f0`.
`always_comb w = 8'(s + 1)` in `t29_sv_cast_alwc` costs `0xf8` over
`t29_sv_alwc_noc`, `0xba4` for `0xaac`.
`assign w = 8'(s + 1)` in `t29_sv_cast_asgn` costs `0x198` over
`t29_sv_asgn_noc`, `0xc74` for `0xadc`, and adds a process unit and
scope: the assignment has the scopes `tb.NetRegassign7_1` and
`tb.NetRegassign7_2` at its line where the control has
`tb.NetRegassign7_1` alone.
`parameter K = int'(1.5)` in `t29_sv_cast_prm` costs nothing over
`t28_sv_prm_expr`, `0x924` in both, and `K` is the unnamed 32 bit
vector of the untyped parameter, holding `2`.
A streaming operator, `{<<{8'h05}}` in `t29_sv_stream`, `$bits` in
`t29_sv_bits` and `s++` in `t29_sv_incr` leave nothing and cost
nothing.

*Found by* `t28_sv_int_cast` against `t27_sv_int_real`.
*Confirmed by* `t28_sv_enum_cast`, `t28_sv_v8_szcast`, and the tier
29 cases, `t29_sv_cast_proc` against `t29_sv_incr`, `t29_sv_cast_fn`
against `t29_sv_fn_noc`, `t29_sv_cast_asgn` against `t29_sv_asgn_noc`,
`t29_sv_cast_alwc` against `t29_sv_alwc_noc` and `t29_sv_cast_prm`
against `t28_sv_prm_expr`.

A named block with a declaration of its own, `initial begin : blk reg
t;` in `t13_v_blk_var`, is a block unit with one declaration, the
next after the module's one, and the block scope `tb.blk` holds the
object `tb.blk.t` at `0x828` after `tb.s` at `0x768`.
The block sits at the `initial` line beside the process scope
`tb.Initial9_0` of the same line.
A loop index declared in the loop is a block of the same kind, named
`Block<line>_<n>` after the line of the loop: `for (int i = 0; i < 3;
i++)` in `t29_sv_for_int` adds the block unit and the scope
`tb.Block7_1` with the one object `tb.Block7_1.i`, an `int` declared
at line 7, before the process scope `tb.Initial7_0`, and `foreach
(a[i])` in `t29_sv_foreach` adds `tb.Block8_1` with `tb.Block8_1.i`
the same way.
The index is logged: the `foreach` index records `0`, `1`, `2` and `3`,
while the `for` index records `0` at time 0 and `3` at the end of the
loop, and nothing between.
An index declared in the module, `integer i` in `t29_sv_for_modi`, is
an object of the module scope and records every value, `X`, `0`, `1`,
`2`, `3`.

*Found by* `t29_sv_for_int` against `t29_sv_incr`, a block unit,
a scope and an object for the loop.
*Confirmed by* `t29_sv_foreach` and `t29_sv_for_modi`.

A SystemVerilog interface is a unit of kind `0x01`, named after the
interface, with its signals as declarations: `bus_if` of
`t13_sv_iface` declares `d`.
A SystemVerilog package is a unit of kind `0x08`, named after the
package, with its parameters as declarations: `p` of `t13_sv_pkg`
declares `W`.
Both point their four file and line words at the `interface` or
`package` line, as a module does.
A modport is a unit of kind `0x02`, named `bus_if.slave` after the
interface and the modport, with one declaration per modport signal
carrying the port mode: `modport slave(input d)` of `t15_sv_iface_mp`
declares `d` as `port in` at the `modport` line.
Its two entity words are 0, as for a process.
The kinds seen so far are:

| Kind | Unit |
| ---: | :--- |
| `0x00` | module |
| `0x01` | interface |
| `0x02` | modport |
| `0x03` | task |
| `0x04` | function |
| `0x05` | named block |
| `0x07` | process |
| `0x08` | SystemVerilog package |
| `0x0a` | VHDL package |
| `0x11` | VHDL function |
| `0x12` | VHDL procedure |
| `0x13` | root |

*Found by* `t11_v_bit_edge` against `t1_bit_one_edge`.
*Confirmed by* `t11_v_always`, which adds the block, `t12_v_task`
and `t12_v_func`, which add the two subprogram kinds, `t13_v_blk_var`,
which puts a declaration in a block, `t13_sv_iface` and
`t13_sv_pkg`, which add the interface and package kinds, and
`t15_sv_iface_mp`, which adds the modport.

**Process scopes.**
Every `initial`, every `always` and every continuous assignment is a
process scope under its module, named after its kind, its source line
and a counter:

| Statement | Scope name | Found by |
| :--- | :--- | :--- |
| `initial` at line 15 | `Initial15_0` | `t11_v_bit_edge` |
| `always` at line 10 | `Always10_0` | `t11_v_always` |
| `assign` at line 10 | `NetRegassign10_2` | `t11_v_wire` |
| `always ... begin : blk` at line 12 | `blk` and `Always12_1`, both at line 12 | `t11_v_always` |
| `always_ff` at line 13 | `Always13_1` | `t13_sv_alwaysff` |
| `always_comb` at line 14 | `Always14_2` | `t13_sv_alwaysff` |
| `fork` branch at line 13 | `Forked13_1` | `t24_sv_fork` |
| `and (w, s, 1'b1)` at line 11 | `Forked11_1` | `t62_str_and_____` |
| `pullup (w)` at line 11 | `Forked11_1` | `t62_str_pullup__` |

The counter after the underscore is per design.
`always_ff` and `always_comb` are `Always` scopes with nothing to tell
them from `always`.
`t12_v_proc_order` gives a child and its parent one `initial` block,
one initializer and one `assign` each, and numbers them
`child.Initial12_0`, `child.Initial7_1`, `tb.Initial14_2`,
`tb.Initial7_3`, `tb.NetRegassign10_4`, `child.NetRegassign10_5`.
So the procedural processes of a design are numbered module by module
in post order, children before their parent, and within a module in
source order with the implicit `Initial` last; the continuous
assignments follow all of them, parent first.
That is the order `t11_v_port` and `t11_v_hier1` showed in part: in
`t11_v_port` `tb` has the procedural processes and the child has the
assignment, and in `t11_v_hier1` the child has the processes.
The counter runs over elaborated instances and the name is fixed at
the first: `t11_v_gen_for` has `tb.Initial14_4`, because each of the
two child instances took two numbers, and both instances carry
`Initial9_0` and `Initial7_1`.

*Found by* `t12_v_proc_order` against `t11_v_port` and `t11_v_hier1`.
*Confirmed by* `t11_v_gen_for`.

A `fork ... join` gives each branch a process scope of its own.
`t24_sv_fork` forks two statements at lines 13 and 14 inside the
`initial` block at line 10, and the file has `tb.Initial10_0`,
`tb.Forked13_1` and `tb.Forked14_2`, each a `vprocess` unit at its
line, numbered on the same counter as the block that holds them.
A clocking block leaves nothing.
`t24_sv_clocking` has `clocking cb @(posedge clk)` with `input s` at
line 10, and the file has `tb.Always9_0` and `tb.Initial14_1` and no
scope, unit, declaration or object for `cb` or its input.

A `final` block is an `Always` scope: `final begin ... end` at line 11
of `t66_prc_final___` is `tb.Always11_0`, the scope an `always` block
at that line would leave, and `always_latch`, `t66_prc_latch___`, is
one too, as `always_ff` and `always_comb` are.
An assertion leaves nothing of its own.
An immediate assertion inside an `always`, `t66_prc_ass_imm_`, has
that block's `tb.Always11_0` and no more; a concurrent
`assert property (@(posedge c) 1'b1);`, `t66_prc_ass_conc`, adds no
scope, unit or declaration to the clock's `Always` and the `Initial`,
and a named `sequence` and `property`, `t66_prc_prop____`, add none
either.
They cost handle space: `0x9b4` for a `final` block and `0x9bc` for
the immediate assertion against the `0x91c` of `t11_sv_logic____`,
and `0xf04` for the named property against the `0xd04` of the bare
concurrent one.
A `covergroup` is the exception.
`covergroup cg @(posedge s); coverpoint s; endgroup` with `cg c1 =
new;`, `t66_prc_covgrp__`, leaves nine scopes under `tb`:
`tb.xlnx_isim_covergroup_cg::new`, `::update` and
`::xlnx_isim_covergroup_sample`, three `function` units at the
`covergroup` line, two `Block11_1` and `Block11_3` scopes under
`::new`, and `tb.Forked11_0` and `tb.Forked11_2` beside them, all on
the module's process counter, which the `initial` blocks then
continue at 4 and 5.
So the writer elaborates a covergroup into generated subprograms and
processes, and their names are the only ones in the corpus with a
`::` in them.

A `program` is a module unit and its instance an ordinary scope:
`prog p();` of `t66_prc_program_` is `tb.p` with `tb.p.Initial20_0`
under it, numbered before the testbench's own `Initial13_1` by the
post order rule above.
The simulation ends when the program's `initial` block ends, at 10 ns
here, so the testbench's write at 50 ns never happens.
A `bind` puts the bound instance under the target scope at the line
of the `bind` statement: `bind tb watcher b(.i(s));` at line 22 of
`t66_prc_bind____` is `tb.b` at line 22 with the port object
`tb.b.i`, on its own handle because the actual is a variable, by the
tier 37 rule.
A `specify` block leaves no scope, no unit and no declaration, and a
path delay in it delays the records: `(i => o) = 1` in the child of
`t66_prc_specify_` moves `o` and the net it drives to 1 ns and 51 ns
where `t66_prc_kid_____`, the same child without the block, writes at
0 and 50 ns, and it costs `0x120` of handle space, `0xe8c` for
`0xd6c`.
A path delay of 0, `t66_prc_spec_0__`, records at 0 and 50 ns and
costs nothing, so the cost is the delay and not the block.

*Found by* `t24_sv_fork` against `t11_v_always`, and
`t24_sv_clocking` against `t11_sv_logic`; `t66_prc_final___` and
`t66_prc_covgrp__` against `t11_sv_logic____`, one `Always` scope and
nine scopes with generated names.
*Confirmed by* `t66_prc_latch___`, `t66_prc_ass_imm_`,
`t66_prc_ass_conc`, `t66_prc_prop____`, `t66_prc_program_`,
`t66_prc_bind____`, `t66_prc_specify_`, `t66_prc_kid_____` and
`t66_prc_spec_0__`.

A gate primitive, a switch primitive or a pull source is a `Forked`
scope too, one per instance, at the line of the instantiation and on
the same counter.
`t62_str_and_____` has `and (w, s, 1'b1);` at line 11 and the file
has `tb.Forked11_1`, a `vprocess` unit at line 11, beside
`tb.Initial13_0`, where `t62_str_wire____` has `tb.NetRegassign11_1`
for `assign w = s;` at the same line.
`t62_str_and_2___` instantiates two gates in one statement and has
`tb.Forked11_1` and `tb.Forked11_2`.
The instance name is not used: `bufif1 g1 (w, 1'b1, s);` of
`t62_str_bufif_n_` leaves the same `tb.Forked11_1` as the unnamed
gate of `t62_str_bufif___`.
A pull source is one instance whatever its width: `pullup (w);` of
`t62_str_pullup__` and `pullup p [3:0] (v);` of `t62_str_vec_pu__`
each leave one `Forked` scope.
A gate delay, `and #3` of `t62_str_gate_dly`, adds nothing to the
gate's scope, and a drive strength on an `assign` leaves its
`NetRegassign` scope as it is; the net's declaration is the same net
kind word with the same type in every case of the tier.
The handle space moves with them, all against the `0xadc` of
`t62_str_wire____`: a pullup alone `0xadc`, an `and` gate `0xc6c`, a
`bufif1` or `nmos` `0xc84`, the delayed gate `0xcbc`, a second plain
`assign` `0xbd4`, a second driver with a strength on either `0xbdc`,
a pullup beside the driver `0xbcc`, a second driver of a 4 bit net
`0xc4c` and four pullups beside it `0xc34`.
So a strength costs 8 once, a gate delay `0x50`, and the second driver
of a 4 bit net `0x78` more than of a scalar; none of it is read.
`tran`, `tranif1` and `trireg` do not elaborate in this version:
`xelab` reports `Primitive "tran" is not supported` and
`Trireg is not supported`.

*Found by* `t62_str_and_____` against `t62_str_wire____`, the scope
name.
*Confirmed by* `t62_str_and_2___`, `t62_str_bufif___`,
`t62_str_bufif_n_`, `t62_str_nmos____`, `t62_str_gate_dly`,
`t62_str_pullup__`, `t62_str_pulldn__`, `t62_str_vec_pu__`, and the
strength cases `t62_str_weak____`, `t62_str_strong__`,
`t62_str_mixed___`, `t62_str_supply__` and `t62_str_wand____`.

**Implicit initial scopes.**
A module whose variables have initializers, `reg s = 1'b0;`, gets one
extra process scope, `Initial<line>_<n>`, at the line of the first
such declaration.
`t11_v_bit_edge` has `tb.Initial13_1` for the `reg` at line 13 beside
`tb.Initial15_0` for the `initial` block at line 15.
`t11_v_two_w64` initializes two variables at lines 7 and 8 and has one
scope, `tb.Initial7_1`.
Every `.v` initializer produces it: `reg`, `reg [7:0]`, `integer`,
`real` and `time`.
In SystemVerilog only an `enum` and a `string` initializer do;
`logic`, `bit`, `int`, `byte`, `longint`, a vector and a struct with an
initializer do not.
The scope adds `0x98` bytes of handle space: `t11_v_bit_edge` has
`0x9b4` where `t11_sv_logic`, the same design with `logic` instead of
`reg`, has `0x91c`.
It also changes the records at time 0; see [values.md](values.md).
Without an initializer the scope is absent: `t12_v_noinit`, a `reg`
first written at 50 ns, has `0x91c` of handle space, as
`t11_sv_logic` does, and `t12_sv_enum_noin`, an enum without an
initializer, has `0x934` where `t11_sv_enum` has `0x9cc`.
The same holds for a `string`, whose variable is otherwise absent from
the file: `t68_str_noinit__` has `0xa14` of handle space where
`t68_str_lit4____`, the same design with `string str = "ZQXJ";`, has
`0xaac`, the `0x98` of the scope.
An unpacked array of strings takes the scope too, and 8 bytes more,
`0xab4` in `t68_str_arr_____`.
A `reg` declared inside a generate loop gets one implicit scope per
iteration: `t12_v_gen_reg` has `tb.Initial11_0` and `tb.Initial11_1`
for two iterations, beside `tb.Initial10_3` for the module's own
initializer and `tb.Initial17_2` for its `initial` block.

*Found by* `t11_v_bit_edge` against `t1_bit_one_edge`, four scopes
where three were expected.
*Confirmed by* `t11_sv_logic` against `t11_v_bit_edge`, and by
`t11_sv_enum`, `t11_sv_enum4` and `t11_sv_str`, the SystemVerilog
cases that have the scope.

**Instances and generates.**
An instance is a scope named after the instance, `tb.dut`, with the
child module's unit, as an entity instance is.
A generate loop does not get a scope of its own: `t11_v_gen_for` has
`tb.g[0].dut` and `tb.g[1].dut` directly under `tb`, where
`t7_gen_for` has `\g(0)\` scopes of kind `0x0c`.
The `genvar` is not a declaration and not an object.
The two instances share one unit, and their process scopes carry the
same names, `Initial9_0` and `Initial7_1` under each.

A variable declared in a generate block belongs to the module, under
an escaped name: `reg r` in `for (i = 0; i < 2; i = i + 1) begin : g`
of `t12_v_gen_reg` is declared in `tb` as `\g[0].r ` and `\g[1].r `,
with the backslash and the trailing space of a Verilog escaped
identifier, and the objects are `tb.\g[0].r ` and `tb.\g[1].r `.
There is no `g[0]` scope.
An `assign` in a generate block is a `NetRegassign` scope of the
module on the module's process counter, with nothing of the block in
its name: the four `assign v[i] = s;` of `t64_ord_gen4____` are
`tb.NetRegassign11_1` to `tb.NetRegassign11_4`, and the loop counting
down in `t64_ord_gen_rev_` leaves the same four names.
An instance in the block keeps the block's name, `tb.g[0].u` and
`tb.g[1].u` of `t64_ord_gen_kids`, with its own `NetRegassign20_1`
under each.
An `if` generate does the same without the index: `reg r` in
`if (1) begin : g` of `t13_v_gen_if_reg` is `tb.\g.r `, with one
implicit `Initial9_1` for its initializer, and the handle space is
`0x9b4`, that of `t11_v_bit_edge` with a plain `reg`.

An interface instance is a scope named after the instance with the
interface's unit: `bus_if b();` in `t13_sv_iface` is `tb.b`, with the
object `tb.b.d`.
The child that takes the interface as a port, `child dut(b)` with
`module child(bus_if p)`, gets a second scope of the same unit under
its own, `tb.dut.p`, whose object `tb.dut.p.d` shares the handle
`0x768` of `tb.b.d`.
So an interface port is a scope that stands for the instance, as a
port of a net stands for the net, and the records are written once.
`t15_sv_iface_vec` adds `logic [7:0] v` to the interface, and the
vector is a second declaration of the interface unit and a second
object in each of the two scopes, `tb.b.v` and `tb.dut.p.v`, both at
`0x828`, so the sharing holds per object.
A modport adds a scope: `child dut(b.slave)` in `t15_sv_iface_mp`
gives the instance a child `tb.b.slave` of the modport unit, and the
child's port `tb.dut.p` is a second scope of that same modport unit,
not of the interface.
Both hold `d` as an object at `0x768`, the handle of `tb.b.d`, with
port mode `in` from the modport's declaration.
The child's `always_comb` is `tb.dut.Always8_0` and the counter runs
on as for any child.

*Found by* `t11_v_gen_for` against `t7_gen_for`; `t64_ord_gen4____`
against `t63_pdr_two_bits`, four `NetRegassign` scopes under `tb` and
no `g[i]`.
*Confirmed by* `t12_v_gen_reg`, the loop with a variable and no
instance, `t13_v_gen_if_reg`, the `if` generate, `t13_sv_iface`,
the interface, and `t15_sv_iface_vec` and `t15_sv_iface_mp`, the
vector and the modport.

**Declarations.**
A `reg` or any other variable is kind `0x00`, a `wire` is kind `0x03`,
and a port is kind `0x03` with the port mode in word 9, `1` for
`input`, `2` for `output` and `0` for `inout`, whether or not the
port has a net type.
The other net kinds have kinds of their own, `0x04` to `0x0d` in the
order of the kinds table above, which is the order the Verilog
standard lists them in with `tri0` and `tri1` after `trior`.
`0x0b` is where `trireg` would fall; xsim refuses one with
`[XSIM 43-4096] Trireg is not supported`, so it is unseen.
A `uwire` is kind `0x03` like a `wire`.
The VCD writer names each of these nets by the same keyword in its
`$var` line, and `TestVCD` holds the declaration kind to it.
A net's kind changes nothing else: `t19_v_tri` differs from
`t11_v_wire` by that one word.

*Found by* `t19_v_wand` against `t11_v_wire`, kind `0x04` against
`0x03`.
*Confirmed by* the other eight tier 19 net cases, each of one kind,
and the `declared` keyword in their truths, which `TestCorpus` checks
against the kind.
A `parameter` is kind `0x01`, with the size 32, the unnamed vector type
and a range `(31 downto 0)`.
Word 4 is the value size in bits, not bytes: 1 for a `reg`, 8 for
`reg [7:0]`, 32 for `integer`, `real` and `int`, 64 for `time`, the
sum of the element sizes for a memory, and for a struct the rule in
[values.md](values.md).

*Found by* `t11_v_wire` against `t11_v_bit_edge`, and `t11_v_port`
against `t8_port_in`.
*Confirmed by* `t13_v_inout` for the `inout` mode.

**Objects and handles.**
Nets are given handles before variables.
In `t11_v_wire` the wire `w` holds `0x768` and the `reg` `s` holds
`0x828`, and in `t11_v_port` the wire `y` holds `0x768`, the input
port `a` holds `0x828` and the `reg` `x` holds `0x8e8`.
An output port shares the handle of the wire on its net, `0x768` for
`b` and `y`, as a VHDL port does.
An input port connected to a `reg` does not share the `reg`'s handle;
`a` and `x` are distinct objects with their own records.
Such a port takes its handle in the order of the connections written
in the instantiation, not the order of the port list:
`t48_v_port_pos4_` connects `.a(a), .b(b)` and holds `0x8e8` on `a`
and `0x9a8` on `b`, and `t48_v_port_rev__` connects `.d(d), .c(c),
.b(b), .a(a)` and holds `0x8e8` on `b` and `0x9a8` on `a`.
The output ports `c` and `d` share the parent's wires either way.
An input port connected to a `wire` does share it: in
`t12_v_port_wire`, where the parent drives `wire x` by an `assign`,
`x` and `tb.dut.a` hold `0x768` together, and `y` and `tb.dut.b` hold
`0x828`.
An `output reg` port does not share the parent's wire: in
`t12_v_port_reg` the wire `y` holds `0x768`, the input port `a`
`0x828`, the `reg` `x` `0x8e8` and the output `reg` `b` `0x9a8`.
So sharing happens between a port and a net, never between a port
and a variable, and the shared handle is the net's.
Nets come first in the whole design, in pre order: `t12_v_proc_order`
has the child's wire `u` at `0x768`, `tb`'s wire `w` at `0x828`, then
the child's `reg` `t` at `0x8e8` and `tb`'s `reg` `s` at `0x9a8`.
`t13_v_hier3_net` runs a net through `tb`, `mid` and `leaf`, each
with a wire and a `reg`, and holds `tb.w0` at `0x768`, `tb.y` at
`0x828`, `mid.w1` at `0x8e8` and `leaf.w2` at `0x9a8`, then `tb.r0`
at `0xa68`, `mid.r1` at `0xb28` and `leaf.r2` at `0xbe8`.
The input port of `mid` shares `w0`, the input port of `leaf` shares
`w1`, and the output ports of both share `y`, so a port shares the
handle of the net it is connected to in the parent, however many
levels the net runs through, and the wire that is a port's own
declaration in the parent is the one that holds the handle.
An `inout` port shares as an input port does: `tb.dut.w` of
`t13_v_inout` holds `0x768` with `tb.w`.
The second handle of a variable is the handle plus the record size,
8 bytes per 32 bits of value, and the next object's handle is the
second handle plus `0xb8`.
`t11_v_two_w64` has `p`, 64 bits, at `0x768` and `q` at `0x830`, which
is `0x768 + 16 + 0xb8`.
So Verilog objects are `0xb8` plus the record size apart, the stride
of a VHDL open port, where VHDL signals are `0xe8` plus the rounded
size apart, `0x30` of which is the VHDL driver; see above.
A Verilog `reg` pays nothing for its `initial` or `always` writer:
`t46_v_gen_70000_` has 70000 registers at `0x768` and every stride is
`0xc0`, whichever of them the `initial` block writes.
A wire does pay for its continuous assignments, and not linearly:
the object after a wire with no `assign` or one is `0xc0` on,
`t19_v_wire_nodrv` and `t11_v_wire`, after one with two and a child
port reading it `0xe8`, `t19_v_2drv_port`, with three `0xf0`,
`t19_v_wire_3drv`, and with four `0xf0` again, `t46_v_wire_4asg_`.
No rule fits those points.
The handle space grows by more than that per object: `t11_v_two_w64`,
which widens the variable to 64 bits and adds a second one bit
variable, has `0x100` more than `t11_v_bit_edge`, where the strides
account for `0xc8`.
VHDL signals do the same; see the open questions in
[../format.md](../format.md).
A `parameter` is an object with a handle and no second handle, the
shape of a VHDL generic, in a second arena: `0x8c0` in `t11_v_param`.
The parameters of one module sit at consecutive handles, each 8 bytes
per 32 bits of value: the five of `t12_v_params`, `K`, `P`, `Q`, `R`
and `L`, at `0x8c0`, `0x8c8`, `0x8d0`, `0x8d8`, `0x8e0`, and in
`t12_v_param64` the 64 bit `W` at `0x8c0` and the 8 bit `P` at
`0x8d0`.
A `localparam` is a parameter like any other, in source order.
A package parameter is an object of the package scope with the same
shape, `p.W` of `t13_sv_pkg` at `0x890` in arena 1 with no second
handle, and it is marked not logged and has no record; see
[values.md](values.md).
A string parameter is a parameter of 40 bits, `P` of `t13_v_str_param`
at `0x8c0`, whose record is two pairs.

*Found by* `t11_v_two_w64` against `t11_v_vec8`, and `t11_v_port`
against `t8_port_in`.
*Confirmed by* `t12_v_port_wire`, `t12_v_port_vec8`, `t12_v_port_reg`
and `t12_v_proc_order` for the sharing and the order, `t13_v_hier3_net`
and `t13_v_inout` for three levels and the third port mode, and
`t12_v_params`, `t12_v_param64`, `t13_sv_pkg` and `t13_v_str_param`
for the parameters.

**What is missing.**
A `string` variable, `t11_sv_str`, has its implicit `initial` scope
and nothing else: no declaration, no object, no type entry, and a
logged range count of 0.
Its handle space is `0x9b4`, the same as `t11_v_bit_edge`, so the
string appears to be given handle space and left out of the instance
list.
The dynamic objects of SystemVerilog go the same way.
A queue `int q[$]`, a dynamic array `int d[]`, an associative array
`int a[string]` and a class handle `c h` of a class with one `int`
member, each written once from the `initial` block, leave no type
entry, no declaration and no object: `t24_sv_queue`,
`t24_sv_dynarr`, `t24_sv_assoc` and `t24_sv_class` hold the one
declaration `s` of `t11_sv_logic`.
Their handle space is `0xa14` in all four, `0xf8` over the `0x91c`
of `t11_sv_logic`, so each is given the same handle space whatever
it holds, and the class member `x` is not there either.

A `parameter string P = "hello"` goes the same way: `t26_sv_str_prm`
has no object for `P` and 8 bytes of handle space over `t11_sv_logic`,
where the untyped `parameter P = "hello"` of `t27_sv_str_untyp` is an
object with a 40 bit record, as in the `.v` `t13_v_str_param`.
An `event e` leaves no declaration and no object either, and is given
`0x2c0` of handle space: the `logic s` declared after it in
`t26_sv_event` has the handle `0xa28` where `t11_sv_logic` has
`0x768`, and the handle space is `0xc14` for `0x91c`.

*Found by* `t24_sv_queue` against `t11_sv_logic`.
*Confirmed by* `t24_sv_dynarr`, `t24_sv_assoc` and `t24_sv_class`,
`t26_sv_str_prm` against `t27_sv_str_untyp`, and `t26_sv_event`
against `t11_sv_logic`.
The clock toggle due at the `$finish` time in `t11_v_always` is not
recorded, so the last record of `clk` is at 75 ns in a 100 ns run.

**What -debug all brings back.**
Elaborated with `xelab -debug all`, tier 60, the same objects are
declared.
A `string`, `t60_dbg_str_____`, and a class handle, `t60_dbg_class___`,
get a declaration of 32 bits and value class 0, and an object at the
usual second handle `0x828`, which is logged.
A queue, a dynamic array and an associative array,
`t60_dbg_queue___`, `t60_dbg_dynarr__`, `t60_dbg_assoc___` and
`t60_dbg_assoc_i_`, get a declaration of 32 bits and value class 3
with one range `(0 to 0)`, the range of a one element array, and an
object at `0x828` that is not logged: the logged ranges name `s` only
and the second arena is never written.
The declaration names the file and line of the variable, as any
declaration does.
The handle space is what typical gives: `0xa14` for the containers and
`0xaac` for the string, `0x98` over `t60_dbg_vec_____` as
`t11_sv_str` is `0x98` over `t11_sv_logic`.
A second handle of the class, `h2` of `t60_dbg_class_2h`, takes the
next slot `0x8e8`, and the file holds one class entry.
The ordinary objects of `t60_dbg_vec_____`, `t60_dbg_int_____`,
`t60_dbg_real____`, `t60_dbg_struct__` and `t60_dbg_mem_____` are
declared as under typical.
The records of the logged objects are in [values.md](values.md).

*Found by* `t60_dbg_str_____` and `t60_dbg_queue___` against
`t60_dbg_none____`, whose only object is `s`.
*Confirmed by* the other cases named above.

## Value classes

Region 17 holds one entry of three words per distinct value class
among the objects of the file, in the order the classes first appear
in the object list, and header word 13 counts the entries.
The first word of an entry is the class code and the other two are 0
in every case.
The region is padded to 8 bytes at the end: 16 bytes for one entry,
24 for two, 40 for three, and empty for none.
The reader keeps the entries as `Debug.Classes`, checks the region
length against word 13, and the dump prints them as `value classes`.

Every VHDL object is class 0, and every VHDL case with objects holds
the one entry `[0 0 0]`.
A Verilog or SystemVerilog object is classed by its type and by the
form of its initializer.
The integral types `shortint`, `int`, `integer` and `longint`, with or
without `unsigned`, are class 3 whatever the initializer, and `time`
is class 4 whatever the initializer.
`real` and `shortreal` are class 0 whatever the initializer.
Everything else, the packed types, is classed by the initializer
alone.
The table names each class's objects, and the first case that showed
each form:

| Class | Objects | Found by |
| ---: | :--- | :--- |
| 0 | every VHDL object; a `real`, a `realtime`, the hidden `real` of `real'(3)`; a packed type with no initializer, `logic`, `bit`, `byte`, `bit [7:0]`, `logic [7:0]`, a net; every packed type in a `.v` file, whose initializer runs as an implicit process; a packed type from a real literal, a real expression `5 + 1.5`, a time literal or `$time`, whose initializer runs as one; an enum from a literal, an unpacked struct or array, a packed struct from a `'{}` pattern; a `real` or `realtime` parameter, typed or untyped, or an untyped parameter from a real or time expression, `5.0 * 2`, `10ns * 2` | `t11_v_bit_edge`, `t29_sv_cast_real`, `t12_sv_noinit`, `t27_sv_byte_noin`, `t28_sv_v8_real`, `t30_sv_v8_realx`, `t28_sv_v64_time`, `t28_sv_v64_stime`, `t11_sv_enum`, `t28_sv_enum_pkd`, `t28_sv_pstr_pat`, `t28_sv_uarr_pat`, `t25_sv_real_lit`, `t25_v_prm_real`, `t28_sv_prm_realu`, `t28_sv_rtime_prm`, `t30_sv_prm_realx`, `t30_sv_ptm_expr` |
| 1 | a packed type from a sized literal, `1'b0`, `1'bx`, `8'h00`, `8'sh05`, `-8'sd1`, `-8'd1`, `32'd5`, `64'h0`, or an unsized based literal, `'d5`, `'sd5`; from a fill literal `'0`, `'1`, `'x`, `'z`; from a concatenation, a replication, a conditional, a function call, `$signed(8'h05)`, `$unsigned(5)`, a comparison `(1 < 2)`, a part select of a parameter or a parameter holding a sized literal; from an expression with a sized operand, `4'd5 + 4'd1`, `8'd5 + 1`, `1'b1 ? 8'd5 : 0`; a parameter of a packed type, `parameter bit`, `parameter logic [7:0]`, `parameter [7:0]`, or an untyped parameter from a sized literal of any width, `8'sd5`, `-8'd1`, `40'h1`, a based literal `'d5`, a sized expression `4'd5 + 4'd1`, a comparison, or a string literal into a typed parameter; a packed struct or union from `'0` or a sized literal; the hidden variable of `signed'(8'h05)` | `t11_sv_logic`, `t25_sv_logic_x`, `t30_sv_v8_negsz`, `t30_sv_v8_ubase`, `t30_sv_v8_sbase`, `t27_sv_v8_xfill`, `t30_sv_v8_1fill`, `t26_sv_v8_cat`, `t26_sv_v8_rep`, `t26_sv_logic_cnd`, `t26_sv_logic_fn`, `t28_sv_v8_signed`, `t30_sv_v8_uns`, `t30_sv_v8_cmp`, `t28_sv_v8_bitsel`, `t30_sv_v8_szexp`, `t30_sv_v8_mixed`, `t30_sv_v8_cnd`, `t26_sv_logic_prm`, `t26_sv_bit_prm`, `t12_v_param64`, `t30_sv_prm_szsgn`, `t30_sv_prm_neg8`, `t30_sv_prm_wide`, `t30_sv_prm_ubase`, `t28_sv_prm_szexp`, `t30_sv_prm_cmp`, `t28_sv_prm_lstr`, `t24_sv_union`, `t28_sv_pstr_szd`, `t29_sv_cast_sgn` |
| 3 | every `shortint`, `int`, `integer`, `longint`, `int unsigned`, `longint unsigned` and `integer unsigned`, with a sized, unsized, negative, real, string or time literal, `'x`, a parameter, a cast or no initializer; a signed packed type from an unsized literal, `logic signed [7:0] s = -1`, `logic signed [7:0] s = 5`, `byte s = 0`, `byte s = -1`; an untyped, `integer`, `int`, `int unsigned` or enum parameter or localparam, in a module or a package, from a literal, `2 * 3`, `$clog2(8)`, `-1`, `1 << 40`, `1'b1 ? 3 : 4` or an enum literal; the hidden variable of a cast, `int'(1.5)`, `state_t'(1)`, `8'(0)`; a loop index, `int i` or `integer i`; the return variable of `function int f` | `t11_sv_int`, `t27_sv_int_uns`, `t26_sv_int_szd5`, `t27_sv_int_str`, `t27_sv_int_real`, `t26_sv_integer_x`, `t26_sv_sgn8_neg`, `t27_sv_sgn8_pos`, `t11_sv_byte`, `t11_v_param`, `t13_sv_pkg`, `t28_sv_prm_expr`, `t28_sv_prm_clog`, `t28_sv_prm_neg`, `t30_sv_prm_shft`, `t30_sv_prm_cnd`, `t28_sv_prm_int_u`, `t28_sv_prm_enum`, `t28_sv_int_cast`, `t28_sv_v8_szcast`, `t29_sv_for_int`, `t29_sv_for_modi`, `t29_sv_fn_noc` |
| 4 | every `time`, from `0`, `64'h0`, `10ns` or no initializer; a `time` parameter from `10ns`; an untyped parameter from a time literal, `10ns`, `10ps`, `1us`, `1s`, `10.5ns`, under any timescale, declared before or after the variable, one or two of them; an unsigned packed type from an unsized literal or expression, `logic s = 0`, `logic s = 1`, `logic [7:0] s = $signed(5)`, `bit s = 0`, `logic [7:0] s = 0`, `logic [7:0] s = -1`, `bit [7:0] s = 0`, `bit [7:0] s = -1`, `logic [31:0] s = 0`, `logic [63:0] s = 0`, `byte unsigned s = 0`, `logic [7:0] s = 2 ** 3`, `logic [7:0] s = K` with `parameter K = -1`, a packed struct from `0` | `t11_v_time`, `t27_sv_time_szd`, `t28_sv_prm_tmtyp`, `t28_sv_prm_time`, `t30_sv_ptm_10ps`, `t30_sv_ptm_1us`, `t30_sv_ptm_1s`, `t30_sv_ptm_frac`, `t30_sv_ptm_ps_ts`, `t30_sv_ptm_late`, `t30_sv_ptm_two`, `t25_sv_logic_int`, `t30_sv_v8_sgnu`, `t26_sv_logic_1`, `t25_sv_bit_unsz`, `t27_sv_v8_neg`, `t28_sv_bit8_neg`, `t26_sv_v32_int`, `t27_sv_byte_uns`, `t28_sv_v8_pow`, `t28_sv_v8_prmneg`, `t28_sv_pstr_int` |
| 6 | a packed type or an untyped parameter from a string literal or a concatenation of string literals: `logic [7:0] s = "a"`, `logic [7:0] s = "ab"`, `logic [15:0] s = "a"`, `logic [15:0] s = {"a", "b"}`, `parameter P = "hello"`, `parameter K = {"a", "b"}` | `t26_sv_v8_str`, `t30_sv_v8_str2`, `t28_sv_v16_str`, `t30_sv_v16_strc`, `t13_v_str_param`, `t27_sv_str_untyp`, `t30_sv_prm_strc` |

The classes 2 and 5 have not been seen; the tier 30 sweep over
thirteen more initializer forms and nine more parameter forms, with a
based literal, a comparison, a string concatenation, a real
expression and a shift past 32 bits among them, produced only 0, 1,
3, 4 and 6, and the tier 69 sweep over thirteen forms the earlier
tiers had not declared produced only 0, 1, 3 and 4:

| Form | Class | Case |
| :--- | ---: | :--- |
| a `specparam` in a `specify` block | 3 | `t69_vcl_specprm_` |
| a `supply0` or `supply1` net | 0 | `t69_vcl_supply0_`, `t69_vcl_supply1_` |
| `const logic [7:0] k = 8'd5` | 1 | `t69_vcl_const_v_` |
| `const int k = 5` | 3 | `t69_vcl_const_i_` |
| a parameter set by `defparam` | 3 | `t69_vcl_defparam` |
| a parameter overridden with `-generic_top` | 3 | `t69_vcl_gtop_prm` |
| a variable of a `parameter type` | 4 | `t69_vcl_typeprm_` |
| a parameter from `$bits` | 3 | `t69_vcl_bits_prm` |
| `bit b = 1.5` | 0 | `t69_vcl_bit_real` |
| an enumeration from `e_t'('x)` | 0 | `t69_vcl_enum_xcs` |
| a net with an initializer, `wire w = 1'b1` | 0 | `t69_vcl_wire_ini` |

So a `const` variable takes the class of its initializer and nothing
else, an override from a `defparam` or from the command line leaves a
parameter looking like any other, and a `specparam` is a parameter
declaration with an object and a record.
The net is the one that separates two forms which look alike in the
source: `wire w = 1'b1` is class 0 where `logic w = 1'b1` is class 1,
because a net's initializer is a continuous assignment and not an
initial value.
A `chandle` leaves no declaration at all under typical,
`t69_vcl_chandle_`, as a `string` and a queue do.
Two forms have no case: xsim rejects `trireg (large) t;` with
`ERROR: [XSIM 43-4096] Trireg is not supported` and `let five = 5;`
with `ERROR: [XSIM 43-3980] The SystemVerilog feature "Let" is not
supported yet for simulation`.

Word 1 of a declaration record is the index of the entry that holds
the class of the declaration's objects.
`t12_v_params` declares `s`, `K`, `P`, `Q`, `R`, `L` and holds
`[0 0 0] [3 0 0] [1 0 0]`; the six records index the entries as `0`,
`1`, `2`, `1`, `0`, `1`, so `s` the `reg` and `R` the `real` parameter
are class 0, `K`, `Q` and `L` are class 3, and `P` the `[7:0]`
parameter is class 1.
Swapping two declarations swaps the word: `t31_sv_w1_i5`, `logic s`
before `int i = 5`, holds `[1 0 0] [3 0 0]` with `s` at `0` and `i` at
`1`, and `t31_sv_w1_swap`, `i` before `s`, holds `[3 0 0] [1 0 0]`
with `i` at `0` and `s` at `1`.
`t25_sv_two_class`, `logic s = 1'b0` beside `int i = 5`, holds
`[1 0 0] [3 0 0]`, and `t25_sv_two_same`, the same `logic` beside
`logic t = 1'b1`, holds `[1 0 0]` with both declarations at `0`.
The word does not follow the value or the writes: `int i` at `0`,
`1`, `5` or `165`, never written, or written after its own delay,
indexes `1` in `t31_sv_w1_i0`, `t31_sv_w1_i1`, `t31_sv_w1_i5`,
`t31_sv_w1_i165`, `t31_sv_w1_nowrt` and `t31_sv_w1_own50`, and the
`int` alone in `t31_sv_w1_s5` indexes `0` of `[3 0 0]`.
The word was recorded as `0` until tier 31: every VHDL file has one
entry, and the first declaration of every file indexes entry `0`.
The reader keeps the word as `Decl.Class`, checks it against the entry
count, and `File.ValueClass` returns the code; the dump prints
`class N` on each declaration line.
A net after a `logic` with a sized initializer adds class 0 after
class 1: `t25_sv_wire`, `t19_sv_uwire`, `t25_sv_net_init`, and so does
the uninitialized output of `always_comb`, `always_latch` and
`always_ff`, `t25_sv_alw_comb`, `t25_sv_alw_latch`, `t13_sv_alwaysff`.
The same declaration classes differently in the two languages: `reg
[7:0] s = 8'h00` in `t25_v_vec8_sz` is class 0 and `logic [7:0] s =
8'h00` in `t25_sv_vec8_sz` is class 1, while `integer s = 32'h0` in
`t25_v_int_sized` and `int s = 32'h0` in `t25_sv_int_sized` are both
class 3.
Signedness moves an unsized literal between 3 and 4 for a packed
type, `logic signed [7:0] s = 5` against `logic [7:0] s = -1`, and
`byte` against `byte unsigned`, but not for an integral type, where
`int unsigned` stays 3; a sized literal is 1 on either.
The signedness of the expression moves it too: `logic [7:0] s =
$signed(5)` is 4 and `$unsigned(5)` is 1, `t30_sv_v8_sgnu` against
`t30_sv_v8_uns`.
A based literal without a size counts as sized: `'d5` and `'sd5` are
1, `t30_sv_v8_ubase` and `t30_sv_v8_sbase`, and so is an expression
with one sized operand, `8'd5 + 1` in `t30_sv_v8_mixed`, a
conditional with a sized arm, and a comparison.
A real expression is 0 like a real literal, `5 + 1.5` in
`t30_sv_v8_realx`.
The class is not the value's representation in the pages: the records
of `t27_sv_byte_uns` are those of `t11_sv_byte` outside the value.
The class follows the target, not the expression: `logic [7:0] s = K`
with `parameter K = -1` in `t28_sv_v8_prmneg` holds `[4 0 0] [3 0 0]`,
`s` unsigned and `K` signed, where `s = -1` alone is 4 and `K` alone
3.
A string literal is 6 into a variable and 1 into a typed parameter,
`t28_sv_v16_str` against `t28_sv_prm_lstr`.
A cast in an initializer takes its class on the hidden variable it
leaves, and the variable it initializes runs as a process and is 0:
`t28_sv_int_cast` holds `[3 0 0]` for both, and `t28_sv_v8_szcast`
`[3 0 0] [0 0 0]`, the hidden variable first; see the cast paragraph
above.
The hidden variable takes the class of the cast: 1 for `signed'(8'h05)`
in `t29_sv_cast_sgn`, `[1 0 0] [0 0 0]`, and 0 for `real'(3)` in
`t29_sv_cast_real`, `[0 0 0]` alone.
What the codes stand for is open; the table reads as the kind of
constant the elaborator holds for the initial value, none, a bit
string, a signed integer, an unsigned integer or a string, with the
integral types and `time` given their kind whatever the initializer,
and that is a guess.

*Found by* `t25_sv_two_class` against `t25_sv_two_same`, where word
13 went from 1 to 2 with the second object's type, and region 17 from
16 to 24 bytes.
*Confirmed by* the region length check in 1090 of 1090 cases, and the
tier 25 to 30 sweeps over the initializer forms.
The word 1 index was *found by* `t31_sv_w1_swap` against
`t31_sv_w1_i5`, where swapping the declarations swapped the words,
and *confirmed by* the reader's range check in 1090 of 1090 cases and
by `t12_v_params`, where the six words select the entry the table
above gives each declaration.

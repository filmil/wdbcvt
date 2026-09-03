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
| 1 to 4 | 0 | |
| 5 | length of the scope name pool before padding | `t2_flat3` 14 |
| 6 | length of the declaration name pool before padding | `t2_flat3` 20 |
| 7 | length of the file name pool before padding | |
| 8 | 0 | |
| 9 | number of files | |
| 10 | number of files again | |
| 11 | number of words in region 15, `0` without `-debug line` | `t2_flat3` 6; `t24_dbg_drv_only` 0 |
| 12 | 0 | |
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
Two scopes elaborated from the same process share one pool string:
`p` sits at offset 20 in the `t2_hier3` pool and both `tb.p` and
`tb.dut.inner.p` name it.

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
`[7][0][0][0][-1][8]`, which reads as left bound, right bound, two
zeros, `-1`, and length.
`t2_array2d` has two.
The direction of the range is not in the record; it is in the type
table.
The record's own meaning beyond the bounds and length is open.


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
| `28` | 4 | 0 for a signal, 2 when the second handle is 0 |
| `32` | 8 | declaration index |
| `40` | 4 | 0 |
| `44` | 4 | `-1`, 0, or a value that differs between runs; open |
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
See [values.md](values.md).

*Found by* `//hdl/serv:sim` against `t9_port_slice`.
*Confirmed by* `t37_v_port_slc__`, `t37_v_port_bit__`,
`t37_v_port_pair1` and `t37_v_port_span_`.

The handle is the number a value record in a page carries, split as
`handle >> 11` for the arena and `handle & 0x7ff` for the key.
See [values.md](values.md).

Signals get handles from one counter.
The first signal is `0x768` in every case.
The second handle is the first plus the value size rounded up to a
multiple of 8, and the next signal's handle is the second handle plus
`0xe8`.
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

A port connected to a signal has no handle of its own.
Its object carries the handle of the signal it is connected to, and so
does every port further down a chain.
In `t8_port_in`, `tb.x` and `tb.dut.a` both hold `0x768`, and in
`t8_port_chain` the signal `tb.x`, the port `tb.dut.a` and the port
`tb.dut.u.b` below it all hold `0x768`.
The mode does not matter: `t8_port_out`, `t8_port_inout` and
`t8_port_buffer` share the handle the same way.
So a net is one handle, and the objects on it are names for it.
The reader reports each object separately and the value records under
each of them, since a record is looked up by handle.

A port left `open` gets its own handle, and it costs `0xb8` plus the
value size rounded up to 8, where a signal costs `0xe8` plus the same.
`t8_port_open3` has one signal `x` at `0x768`, then three open one bit
ports at `0x858`, `0x918` and `0x9d8`, `0xc0` apart, then a signal `s`
of the child at `0xa98`.
`t8_port_vec8` and `t8_port_vec16` have one open vector port at `0x768`
and the child's signal `s` at `0x828` and `0x830`, so the port's stride
is `0xb8` plus 8 and `0xb8` plus 16.

| Object | Handle stride | Found by |
| :--- | :--- | :--- |
| signal | `0xe8` plus the size rounded up to 8 | the table above |
| open port | `0xb8` plus the size rounded up to 8 | `t8_port_open3`, `t8_port_vec8`, `t8_port_vec16` |
| connected port | none, it shares the signal's handle | `t8_port_in`, `t8_port_chain` |

Generics, constants and variables get handles from somewhere else:
`0xe98` for the lone generic of `t4_gen_default`, `0xf88` and `0x10a0`
for the two of `t4_gen_diff_two`, `0xdf8` for the loop index of
`t5_tr1000`, `0xde0` for the variable of `t6_var_int`, `0xf08` and
`0xf0c` for the two variables of `t6_proc2`, which are 4 bytes apart,
and `0xda8` for the architecture constant of `t8_gen_if`.
There is no pattern yet that predicts them, and none is needed: the
handle is read from the record.

The decoder exposes the fifth word as `Object.Generic`, because it is 2
for exactly the objects that are not signals.

Generics of other types get handles the same way.
`t9_gen_types` has `kb : boolean`, `ks : string(1 to 3)`,
`kv : std_ulogic_vector(3 downto 0)` and `kr : real` at `0xe98`,
`0xe99`, `0xeac` and `0xec0`: 1, 19 and 20 apart for values of 1, 3, 4
and 8 bytes, so the stride is not the size.
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
A package with only a type in it, as in `t2_record`, gets no scope.

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

A procedure declared inside a process, the shape of `t9_proc_local`,
gets two scopes: `t23_sub_in_proc` declares `flip` in process `p` and
the file has `tb.flip` as a child of `tb`, listed before `tb.p`, and
`tb.p.flip` as the child of `tb.p`.
Both point at the same unit, the one procedure unit of kind `0x12`,
and each lists an object for the local `r` on the same handle `0xd0`,
so the file holds three objects for two declarations.
*Found by* `t23_sub_in_proc` against `t22_dbg_sub_proc`.


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

*Found by* `t24_sv_fork` against `t11_v_always`, and
`t24_sv_clocking` against `t11_sv_logic`.

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

*Found by* `t11_v_gen_for` against `t7_gen_for`.
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
size apart.
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
3, 4 and 6.

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
*Confirmed by* the region length check in 675 of 675 cases, and the
tier 25 to 30 sweeps over the initializer forms.
The word 1 index was *found by* `t31_sv_w1_swap` against
`t31_sv_w1_i5`, where swapping the declarations swapped the words,
and *confirmed by* the reader's range check in 675 of 675 cases and
by `t12_v_params`, where the six words select the entry the table
above gives each declaration.

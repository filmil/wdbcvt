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
| `24` | 4 | int32 `-12` in every case |
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
| 17 | 16 zero bytes when there are objects, else empty | |

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
| 11 | number of words in region 15 | `t2_flat3` 6 |
| 12 | 0 | |
| 13 | 1 when there are objects, 0 in `t0_nosig` | |
| 14 | `0x101` | |
| 15 | `0x101` | |
| 16 | `0x10000` | |

The meaning of the last three is open.
They are the same in all 252 cases.


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
| 1 | 0 |
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
| `0x00` | Verilog variable: `reg`, `integer`, `real`, `time`, a SystemVerilog `logic`, `int`, struct or enum | `t11_v_bit_edge` |
| `0x01` | Verilog `parameter` | `t11_v_param` |
| `0x03` | Verilog net: a `wire`, and every port | `t11_v_wire`, `t11_v_port` |

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
| `20` | 4 | byte offset into the value at the handle, 0 unless the object is a port bound to a slice |
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
See [values.md](values.md).

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

*Found by* `t9_port_rec` against `t2_record`, where the extra scope
sat between `tb` and `tb.dut` in the scope list.
*Confirmed by* `t13_sv_pkg` against `t12_sv_typedef`.


## Implicit processes

A concurrent signal assignment such as `q <= a;` in an architecture is
an implicit process.
It gets a process scope named `line__NN`, where `NN` is the source line
of the statement, with a process unit pointing at the same line.
`t8_port_open` has `tb.dut.line__18` and `t8_port_vec8` has the same,
both from one concurrent assignment in `child`.

*Found by* `t8_port_open`, whose child has no labelled process.


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

A named block with a declaration of its own, `initial begin : blk reg
t;` in `t13_v_blk_var`, is a block unit with one declaration, the
next after the module's one, and the block scope `tb.blk` holds the
object `tb.blk.t` at `0x828` after `tb.s` at `0x768`.
The block sits at the `initial` line beside the process scope
`tb.Initial9_0` of the same line.

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
The clock toggle due at the `$finish` time in `t11_v_always` is not
recorded, so the last record of `clk` is at 75 ns in a 100 ns run.

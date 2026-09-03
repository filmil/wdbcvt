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
They are the same in all 79 cases.


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
| 2 | kind: `0x13` the root, `0x09` an entity, `0x0c` a generate, `0x0d` a process |
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
| 4 | value size in bytes, see [values.md](values.md) |
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

Word 9 was recorded as the constant 5 through tier 7, where no case had
a port.
It is the port mode:

| Word 9 | Mode | Found by |
| ---: | :--- | :--- |
| 0 | `inout` | `t8_port_inout` |
| 1 | `in` | `t8_port_in`, `t8_port_open`, `t8_port_vec8` |
| 2 | `out` | `t8_port_out`, `t8_port_open` |
| 3 | `buffer` | `t8_port_buffer` |
| 5 | not a port | every declaration through tier 7 |

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
| `16` | 8 | scope index |
| `24` | 4 | 0 |
| `28` | 4 | 0 for a signal, 2 when the second handle is 0 |
| `32` | 8 | declaration index |
| `40` | 16 | 0 |

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

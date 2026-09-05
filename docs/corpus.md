<!-- SPDX-License-Identifier: Apache-2.0 -->

# The corpus, and how to read it


## The point

One waveform database teaches almost nothing.
A byte at offset 400 could be a signal count, a width, a name length, a
block size, or padding, and a single file cannot tell those apart.

Two databases that differ in exactly one property can.
If the only difference between two designs is that the second has one
more signal, then the bytes that differ between their databases are the
bytes that encode "how many signals" and "one signal's record".
That is the whole method.
Complexity rises along the ladder, but each rung changes exactly one
thing from the rung below it.


## Rule 1: measure the noise before measuring anything else

A file format usually embeds things that change between two runs of the
same simulation: a creation timestamp, a host name, an absolute path, a
tool build id, a random id.
Those bytes differ every run, and they differ for reasons that have
nothing to do with the design.

So the first experiment is not a comparison between two designs.
It is the same design, simulated twice:

```sh
bazel run //tools:noise_mask -- //hdl/corpus/t1_bit_one_edge_:sim /tmp/mask
```

Note what the script had to do to be correct.
An earlier version passed a changing `--action_env` value, on the theory
that it would change the action key. It does not, for this rule: two
builds with different values both reported `1 total action` with nothing
executed, and produced identical files. That looked like a perfectly
deterministic format, and was really one cached file compared with
itself. The script now runs `bazel clean` and disables the disk cache,
which forces a genuine re-simulation.

Every offset it reports is noise.
Pass the two files it keeps to `wdbdiff` as `-mask-a` and `-mask-b` for
every later comparison:

```sh
bazel run //cmd/wdbdiff -- \
    -mask-a /tmp/mask/run1.wdb -mask-b /tmp/mask/run2.wdb \
    -a bazel-bin/hdl/corpus/t1_bit_one_edge/sim.wdb \
    -b bazel-bin/hdl/corpus/t1_two_bits/sim.wdb
```

Bazel is the obstacle in this one experiment rather than the help.
A second build of an unchanged target returns the cached file, so the
naive "build it twice" compares a file with itself and reports perfect
determinism.
The script passes a different `WDB_NONCE` through `--action_env`, which
changes the action key and makes the simulation run again.
Nothing reads the variable, so it does not change what is simulated.

Do not reach for `bazel clean --expunge` to force the re-run.
It deletes the output base, and the hermetic Vivado repository lives
there.

Skipping this step produces confident, wrong conclusions: a field is
declared to be "the signal count" because it changed, when it was the
run timestamp all along.

If the two files are byte for byte identical, say so explicitly.
That is a strong and useful property, not an absence of a result.


## Rule 2: every case directory name is the same length

Vivado embeds the absolute path of the source file in the database, once
per file. So a case in a directory whose name is one character longer
produces a database one byte longer, whatever the design does.

That is a confound with teeth, because it is invisible and it is the
right order of magnitude to be mistaken for a real result. It already
produced one wrong measurement here. `t2_unsigned8` and `t2_signed8`
differ by exactly two bytes, which was read as evidence that type names
are stored one byte per character. The two directory names also differed
by two characters, so the measurement said nothing. Padding every case
name to the same length and repeating it gave the same two bytes, and
only then was the conclusion earned.

Every directory under `hdl/corpus` is therefore padded with trailing
underscores to a common length. The padding is ugly on purpose: it is
easier to keep than to remember.

To check the invariant:

```sh
ls -d hdl/corpus/*/ | sed 's|.*/\([^/]*\)/|\1|' | awk '{print length($0)}' | sort -u
```

That must print exactly one number.


## Rule 3: hold everything constant except the axis

Every case in the corpus keeps these the same, so they cannot become
confounds:

* the top level entity is always named `tb`,
* the Vivado library name is always `corpus`, except that a mixed
  language case compiles its child into a library named `child`,
  because `rules_vivado` compiles one language per library,
* the simulation ends through `std.env.stop`, at 100 ns unless the case
  is about a longer run,
* the source files use the same header, the same libraries and the same
  VHDL standard.

Two confounds were found the hard way and are worth naming:

* A `use` clause adds entries to the file table inside the database,
  with two absolute paths each.
  `t2_slv8` is 447 bytes longer than `t1_vec8` for `use
  ieee.numeric_std.all` alone.
  A pair that is meant to measure a type must use the same clauses.
* A `for` loop adds an object, the loop index, and that object gets a
  record at time 0.
  A pair that is meant to measure transitions must declare the loop in
  `truth.json` as a variable of kind `loop`, or count it.

What is left different between a pair is the one axis under test.
The Bazel target name differs too, and the output file is named after
it. Whether that name reaches inside the file is itself one of the first
questions to answer.


## Rule 4: every case declares its own ground truth

Each case directory holds a `truth.json` stating what the simulation
actually did: the signals, their widths and types, every transition with
its time and value, the end time, and any generics and variables the
design declares.

The decoder is then developed against declared truth rather than against
somebody's reading of a hex dump.
A test says "parse this `.wdb` and reproduce this `truth.json`", and it
either does or it does not.
`sim.vcd`, written by the same simulation run, is the independent check
that the truth file is itself correct.

The fields grew with the ladder, and `//pkg/wdb:corpus_test.go` reads
them all.
A signal or a variable can say `"logged": false`, and the test then
expects an object with no records and outside every logged range.
A generic can carry `name`, `type`, `value` and `width`, and the test
checks the declaration and the recorded value against them; a `scope`
puts it outside `tb`, as the package parameter of `t13_sv_pkg` is, and
`"logged": false` expects no record of it.
`transition_runs` lists a signal, a start, a step, a count and the
values the run cycles through, and the test expands it into
transitions.
`final_per_time` names the signals whose writes within one time step
the file cannot order, and the test compares the last value of each
time step for them instead of the sequence; see the tier 12 notes.
A signal can carry `records`, the number of records its object holds
with repeats of one value included, for the cases that count the `X`
records of a net at time 0; see the tier 16 notes.
A signal can say `"unsigned": true` for an integral type declared
`unsigned`, which the database does not record, and the test reads the
decoded value back as an unsigned number of `width` bits before the
comparison; see the tier 27 notes.
An `absent` list names declarations the source has and the file does
not, `parameter string P` or an `event`, and the test checks that no
object carries the name; see the tier 25 and 26 notes.
A generic of an untyped time parameter gives `value` as the reader
spells the `float64`, `1e+09` for `1s`; the `stored` field that once
named the storage went away when the reader learned to decode it, in
tier 30.


## The ladder

Tier 0 establishes the floor: what a database costs when it holds
almost nothing.

| Case | What it is |
| :--- | :--- |
| `t0_nosig` | No signals at all. The container skeleton, and nothing else. |
| `t0_bit_const` | One bit that never changes. Adds one signal, no transitions. |

Tier 1 changes one property of the baseline `t1_bit_one_edge`, which is
one bit with one transition.

| Case | Axis it moves | Question it answers |
| :--- | :--- | :--- |
| `t1_bit_one_edge` | baseline | |
| `t1_bit_two_edges` | transition count 1 to 2 | Size and shape of one transition record. |
| `t1_two_bits` | signal count 1 to 2 | Size and shape of one signal record, and where the count lives. |
| `t1_bit_long_name` | name length 1 to 40 | Are names inline or in a table? Length prefixed or terminated? |
| `t1_vec8` | width 1 to 8 | Where the width lives, and whether a vector value is packed. |
| `t1_nine_state` | value alphabet | How many bits per value, and the code for each of U X 0 1 Z W L H and dash. |
| `t1_hier1` | hierarchy depth 0 to 1 | How a scope is opened and closed. |

Tier 2 moves the type and the structure, holding the one signal and the
one transition of the baseline fixed.

| Case | Axis |
| :--- | :--- |
| `t2_bit`, `t2_boolean`, `t2_character`, `t2_integer`, `t2_real`, `t2_time` | the predefined scalar types from `std.standard` |
| `t2_slv8`, `t2_unsigned8`, `t2_signed8` | the IEEE vector types, and resolved against unresolved |
| `t2_enum` | a user-defined enumeration, whose literal names appear nowhere else |
| `t2_record` | a record of three fields of different shapes |
| `t2_array2d` | an array of four vectors |
| `t2_hier3` | hierarchy depth 1 to 3 |

One limit of the answer key shows up at this tier and is worth stating.
GHDL's VCD writer emits nothing for `character`, `time`, enumerations,
records and arrays, because VCD cannot represent them.
For those six cases the independent simulator check does not apply, and
the `truth.json` plus the VCD Vivado itself writes are the only guards.
See [provenance.md](provenance.md).

Tier 3 moves the value change data, holding one signal of one type
fixed.

| Case | Axis |
| :--- | :--- |
| `t3_tr1`, `t3_tr2`, `t3_tr4`, `t3_tr8`, `t3_tr16` | transition count, 1 to 16 |
| `t3_late` | the same single transition at 1000 ns rather than 10 ns |
| `t3_valz` | the same single transition to `'Z'` rather than `'1'` |

The last two are the useful shape: they are the **same size** as the
baseline and differ from it in exactly one thing, so a masked diff
points at the field rather than at everything downstream of an insertion.
Prefer that shape when designing a case, because a case that changes the
file length moves every offset after the change and buries the answer.

Tier 4 moves generics, to find out whether a generic reaches the file
and whether it changes how an entity is named.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t4_gen_default` | one instance of `child`, generic at its default | a generic is an object with one record at time 0 |
| `t4_gen_explicit` | the same, generic set in the instantiation | no difference in the file |
| `t4_gen_same_two` | two instances, equal generics | one `child(sim)` unit, one set of declarations |
| `t4_gen_diff_two` | two instances, different generics | two `child(sim)` units, declarations repeated, names unchanged |

Tier 5 moves alignment, the object count and the page size.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t5_int_arr` | an array of four integers | 16 bytes, elements back to back |
| `t5_arr_rec` | an array of three two field records | records stride 8 inside an array |
| `t5_rec_real` | a record with a `real` field | a real aligns to 8 |
| `t5_rec_sub5` | a nested record whose inner record is 5 bytes | a record field aligns to 8 |
| `t5_sig10` | ten one bit signals | the arena table grows; the trailer moves; `handle >> 11` |
| `t5_tr1000` | 1000 transitions | pages overflow at 600 one byte records; the marker moves |

Tier 6 moves the same axes further, to turn a guess into a rule.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t6_sig05`, `t6_sig12`, `t6_sig20` | 5, 12 and 20 signals | arena table of 3, 4 and 6 slots; four arenas |
| `t6_tr1300` | 1300 transitions | three pages, marker after the second |
| `t6_var_int` | a process variable | an object with no records |
| `t6_proc2` | two processes with variables | statement lines per file; objects grouped by scope |

Tiers 4 to 6 were generated by a script from a table of cases rather
than written by hand, because the cases differ from the baseline in one
declared line each.
The `truth.json` of each is still written from the design, not from the
database.

Tier 7 asks the questions tiers 5 and 6 left open, one case per
question.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t7_sig07`, `t7_sig14`, `t7_sig16`, `t7_sig24` | 7, 14, 16 and 24 signals | slots are `ceil(trailer +0x18 / 0x800)`; the earlier guess was wrong |
| `t7_int700`, `t7_wide700` | 700 changes of a 4 byte and an 8 byte value | pages hold 510 and 425 records; the limit is 10240 bytes |
| `t7_delta` | two assignments separated by `wait for 0 ns` | two records at the same time, in order |
| `t7_rec_in2`, `t7_rec_in16` | an inner record of scalars only, 2 and 16 bytes | no extra range triple at all |
| `t7_rec_vfirst` | the inner vector before the scalar | the `(0, 8, 1)` follows its field |
| `t7_rec_bitv`, `t7_rec_intv` | the inner scalar is a `bit`, an `integer` | the triple is the scalar's own range |
| `t7_rec_in2v` | two inner vectors | one triple per inner field |
| `t7_gen_for` | three instances under a `for generate` | generate scopes `\g(0)\`; arena records in write order; records at one time unsorted |

`t7_gen_for` is the one case that broke the reader rather than a guess:
it refused the file until it accepted arena records in any order, and
its test failed until the path spelling was normalised.
The five record cases were written two at a time as each answer raised
the next question, and the table lists them in that order.

Tier 8 takes the questions tier 7 left in the open list, and the
first axis no earlier tier had touched: entity ports.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t8_delta3` | three assignments across three deltas | one record per delta |
| `t8_delta_same`, `t8_same` | assignments of the value already held | no record without a change |
| `t8_ps` | waits of 1 ps, 998 ps and 1500 fs | picosecond unit; femtoseconds truncated |
| `t8_rec_realv` | a `real` beside a vector in the inner record | the real contributes no triple |
| `t8_port_in` | an `in` port on the child | declaration word 9 is the port mode; a connected port shares the signal's handle |
| `t8_port_out`, `t8_port_inout`, `t8_port_buffer` | the other modes | modes 2, 0 and 3; the handle is shared whatever the mode |
| `t8_port_open` | both ports left open | an open port owns a handle; the `line__NN` scope of a concurrent assignment |
| `t8_port_open3` | three open ports beside a signal | an open port's stride is `0xc0` |
| `t8_port_vec8`, `t8_port_vec16` | an open 8 and 16 bit port | the stride is `0xb8` plus the rounded size |
| `t8_port_chain` | a port connected through two levels | all objects on the net share the handle, one time 0 record each |
| `t8_gen_if` | an `if` generate with a constant condition | plain branch scopes, an empty false branch, kind `0x13` is a constant |
| `t8_gen_nest` | a `for` generate inside a `for` generate | the shape repeats per level |

Two port cases had to be redesigned before their truth was right.
An `inout` or `buffer` port with no driver inside the child and an
unresolved `std_ulogic` type stays `'U'`, since the testbench and the
port are two sources on one unresolved signal.
The `inout` case uses `std_logic` and drives `'Z'` from inside, and the
`buffer` case is driven only from inside.
And `t8_ps` first claimed a change at 1001 ps for a wait of 1500 fs,
which the simulator cut to 1 ps, so the truth was wrong and the file
was right; the xsim VCD settled it.

Tier 9 has three strands: the two cases tier 8 listed as not written
yet, the language constructs no tier had touched, and a size sweep
after the first wide value came back in pieces.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t9_port_lnk` | a `linkage` port | port mode 4 |
| `t9_gen_types` | generics of `boolean`, `string`, vector and `real` | each records in its type's encoding |
| `t9_port_slice`, `t9_port_slice2`, `t9_port_sliceto` | a port bound to an element or slice of a signal | instance word `+20` is a byte offset from the left element |
| `t9_port_expr` | a port bound to `'1'` | the port owns a handle, as an open port |
| `t9_port_rec` | a record port with a package constant as the default | a package scope of kind `0x0a`; an unlogged object; the marker is a list of ranges |
| `t9_mark_gap`, `t9_mark_two` | an unlogged object first, then two logged ranges | the ranges hold indices; variable objects multiply per instance |
| `t9_var_inst3` | three instances of an entity with a process variable | nine objects for three variables |
| `t9_pkg_sig` | a signal in a package | takes the first handle and is not logged; arena 0 unwritten |
| `t9_block` | a `block` with a signal | unit kind `0x0c`, as a generate |
| `t9_comp`, `t9_alias`, `t9_func`, `t9_proc_local` | a component, an alias, a function, a procedure | nothing in the file but handle space |
| `t9_proc_sig` | a procedure with a `signal` parameter | the change is recorded twice |
| `t9_vec200` to `t9_vec12000`, 18 sizes | the value size | values over 257 bytes are chunked; the chunk size table |
| `t9_int73` | 73 integers, the 292 bytes of `t9_vec292` | chunks split bytes, not elements |
| `t9_tr70000` | 70000 transitions | arena records continue past 100 pages |

The `t9_vec*` sizes were chosen after the first three: 200, 261 and
257 stayed whole, 292 split, and the rest walk the powers of two, the
multiples of 146 and a few odd sizes between to find the rule.
The rule was not found; the table is the result.
`t9_tr70000` is the one case whose truth is not a value list but a
`transition_runs` entry, a start, a step and a count, because 70001
values do not belong in a JSON file.

Tier 10 is a sweep along one axis, the value size, to turn the chunk
table of tier 9 into a rule.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t10_vec258` to `t10_vec272` | sizes between 257, whole, and 292, split | still whole |
| `t10_vec273` to `t10_vec276` | one byte at a time | 274 whole, 275 two chunks, 276 four |
| `t10_vec280` to `t10_vec500` | sizes in the four chunk band | four chunks throughout |
| `t10_vec573` to `t10_vec576` | one byte at a time | the step to six is at 575 |
| `t10_vec872`, `t10_vec874` | one byte either side | the step to eight is at 874 |
| `t10_vec1022` to `t10_vec1200` | sizes around eight chunks of 146 | 146 is not a limit; the step to ten is between 1169 and 1200 |
| `t10_vec11970`, `t10_vec11980` | just below 12000 | 82 chunks both, so the steps are 299 apart, not 300 |
| `t10_vec20000`, `t10_vec30000` | wider than a page | 134 and 202 chunks of 149 and 148 |
| `t10_real40` | 40 reals, the 320 bytes of `t10_vec320` | the same four chunks of 80 |

The first twenty five cases were written from the tier 9 table, and
the rule `2 * ceil((size + 24) / 299)` was fitted to their counts.
The last twelve were written to sit one byte either side of the
boundaries the rule predicts, and every one fell where it should.
The reader now enforces the rule's addresses on every wide value, so
a file that chunks differently is refused rather than read.

Tier 11 is the ladder again in Verilog and SystemVerilog, to find out
whether the source language reaches the database.
It does, in the type table's origin word, in the unit and declaration
kinds, in the scope names and in the shape of every value record.
The `.v` cases are `t11_v_*` and the `.sv` cases `t11_sv_*`, and each
holds one transition at 50 ns and a `$finish` at 100 ns unless the
axis says otherwise.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t11_v_bit_edge` | `t1_bit_one_edge` in Verilog | origin `5`; unit kinds `0x00`, `0x07`; the implicit `Initial` scope; word pair records; the `X` record |
| `t11_sv_logic`, `t11_sv_bit` | `logic` and `bit` in SystemVerilog | no implicit scope for a `.sv` initializer; `bit` is two state with variant `1` |
| `t11_v_vec8`, `t11_v_vec8_asc`, `t11_v_signed8` | an 8 bit vector, ascending, signed | one shared unnamed vector entry; direction and signedness change nothing in the record |
| `t11_v_vec33`, `t11_v_vec100`, `t11_v_vec64x` | 33 and 100 bits; a bit set to `x` | one pair per 32 bits; `X` is both words; a partial record |
| `t11_v_integer`, `t11_v_time`, `t11_v_real` | the Verilog scalar types | named vector entries with bounds; origin `0xd`; a real is one pair |
| `t11_sv_int`, `t11_sv_byte`, `t11_sv_longint` | the SystemVerilog integral types | `bit` based, 32, 8 and 64 bits |
| `t11_v_two_w64` | two variables, 64 and 1 bits | the handle stride is `0xb8` plus the record size |
| `t11_v_wire`, `t11_v_port` | a `wire` with `assign`; ports on a child | kind `0x03`; the `NetRegassign` scope; nets first; the output port shares the wire's handle |
| `t11_v_hier1`, `t11_v_param`, `t11_v_gen_for` | a child module, a `parameter`, a generate loop | instance scopes; kind `0x01` parameter objects; `g[0].dut` with no generate scope |
| `t11_v_always` | `always` blocks and a named block | `Always` scopes, unit kind `0x05`; the toggle at `$finish` is unrecorded |
| `t11_v_mem4`, `t11_v_mem4_desc`, `t11_v_mem8` | memories of four and eight bytes, descending | layout `2`; contiguous bits, leftmost element at the top; one record per element write |
| `t11_v_mem4w4`, `t11_v_mem4w5`, `t11_v_mem3w5` | element widths under a byte | no padding between elements |
| `t11_v_mem2w9` to `t11_v_mem2w64`, six widths | element widths across the 32 bit boundary | elements straddle pairs |
| `t11_sv_struct`, `t11_sv_ustruct`, `t11_sv_pstruct40` | a struct packed and unpacked; a 41 bit packed struct | layout `3` and `2`; the `0x07` alias; packed is a vector |
| `t11_sv_struct3`, `t11_sv_struct40`, `t11_sv_struct_r` | three fields; a 40 bit field; a `real` field | one slot of pairs per field, last field lowest |
| `t11_sv_arr2d` | `logic [1:0][3:0]` | an array of the vector entry; two declaration ranges |
| `t11_sv_enum`, `t11_sv_enum4` | an enum over `int`; over `logic [3:0]` with values | kind `0x04`; the base type and bounds; the `XXXX` record |
| `t11_sv_str` | a `string` | not in the database at all, beyond its implicit scope |

The `truth.json` of a Verilog case names the declared keyword under
`declared` and leaves `type` empty for an unnamed vector, memory or
struct, because that is what the type table holds.
Every `.v` case with an initializer lists the `X` record before its
value, and the memory cases list one state per element write.

Tier 12 perturbs tier 11 where its findings left gaps: wide Verilog
values, the process counter, ports on wires, subprograms, parameters,
generate blocks with variables, and SystemVerilog typedefs and
unpacked arrays.
Each case is one axis against a tier 11 case, and the transition and
`$finish` times are as in tier 11.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t12_v_vec1088`, `t12_v_vec1089` | 272 and 280 bytes of record | the chunk threshold is 275 for Verilog too; chunks split inside a pair |
| `t12_v_vec2272`, `t12_v_vec2304` | one byte under a chunk step, and on it | 4 chunks of 142, 6 of 96 |
| `t12_v_vec4800`, `t12_v_vec4800x` | ten chunks; one bit of it set later | a pair write into a chunked value is an 8 byte record |
| `t12_v_vec12000` | 3000 bytes over three arenas | the chunks of `t9_vec3000` |
| `t12_v_mem40w32`, `t12_v_mem40_t0` | forty element writes at forty times, and at one | the split rest ambiguity; the cross arena order loss |
| `t12_v_vec8_z` | `8'bz0z1xx01` | `Z` is the `b` word alone |
| `t12_v_neg_range` | `reg [-4:3]` | signed bounds, the record unchanged |
| `t12_v_noinit`, `t12_sv_noinit`, `t12_sv_enum_noin` | no initializer | no implicit scope; one `X` record, or the first literal for an enum |
| `t12_sv_shortreal` | `shortreal` | the `real` entry and record |
| `t12_sv_typedef` | `typedef logic [7:0]` | the alias carries the bounds |
| `t12_sv_unp2d` | `logic [3:0] m [0:1][0:2]` | one array entry per unpacked dimension |
| `t12_v_params`, `t12_v_param64` | five parameter declarations; a 64 bit one | parameter types; 16 bits for a `real`; consecutive handles by record size |
| `t12_v_task`, `t12_v_func` | a `task` and a `function` | unit kinds `0x03` and `0x04`; arguments and locals are objects; the return variable first |
| `t12_v_gen_reg` | a `reg` in a generate loop | the escaped name `\g[0].r `; one implicit scope per iteration |
| `t12_v_proc_order` | a child and a parent with three processes each | the process counter order; nets before variables in pre order |
| `t12_v_port_wire`, `t12_v_port_vec8`, `t12_v_port_reg` | an input port on a wire, 8 bits wide, an `output reg` | an input port shares a wire's handle and not a reg's; three `X` records on the shared net |

`t12_v_mem40_t0` is the one case whose `truth.json` cannot list its
transitions in order, because the file does not hold the order; it
sets `final_per_time` and the test compares the value at the end of
each time step.

Tier 13 takes the cases listed as not written after tier 12, and adds
the ones the tier 12 findings asked for: a script that logs a package,
the SystemVerilog constructs that had not been seen, arrays of `real`
and of structs, a string parameter, three levels of nets, and long
runs that spill a page.
`t13_pkg_log_all` is the one case with a script of its own; the
`tcl` attribute of `wdb_case` replaces the default `xsim.tcl` with a
file of the case directory, which gets the same `{{VCD_FILE}}` and
`{{TOP}}` substitutions.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t13_pkg_log_all` | `log_wave -recursive /sig_pkg` added to the script of `t9_pkg_sig` | the package signal records like a signal of `tb`; the logged range count grows |
| `t13_sv_iface` | an interface instance passed to a child | unit kind `0x01`; the interface port is a scope sharing the instance's handles |
| `t13_sv_pkg` | a package with a typedef and a parameter | unit kind `0x08` beside `tb`; the parameter is an object with no record |
| `t13_sv_alwaysff` | `always_ff` and `always_comb` | `Always` scopes |
| `t13_v_inout` | an `inout` port driven `Z` then `1` | port mode `0`; the port shares the wire; `Z` as one record |
| `t13_sv_tdef_ua` | a typedef of an unpacked array | the alias carries both ranges |
| `t13_sv_real_arr` | `real r [0:1]` | one pair per element, last lowest; an unchanged value writes nothing |
| `t13_sv_struct_ar` | an unpacked array of a packed struct | one contiguous value, element 0 at the top |
| `t13_v_str_param` | `parameter P = "hello"` | a 40 bit vector, `h` at the top |
| `t13_v_same_t` | three writes at one time | three records in write order |
| `t13_v_hier3_net` | a net through three modules | nets pre order over three levels; ports share upward |
| `t13_v_gen_if_reg` | a `reg` in an `if` generate | `\g.r `; one implicit scope |
| `t13_v_blk_var` | a `reg` declared in a named `initial` block | the block unit holds the declaration; the block scope holds the object |
| `t13_v_tr2000` | two thousand toggles | five Verilog pages; no `X` record at time 0 |
| `t13_v_tr420`, `t13_v_tr430`, `t13_v_tr430_2` | 421 records, 430, and 430 beside a second arena | the `X` record goes with the arena that spills into a second page |
| `t13_tr430` | the 430 ns clock in VHDL | one page of 431 records; the toggle at `std.env.stop` is recorded |

The long runs list their toggles as a `transition_runs` entry in
`truth.json`, a start time, a step and a count, rather than two
thousand transitions.

Every case is also held to its own `sim.vcd` by `TestVCD`; see
[format/vcd.md](format/vcd.md).

Tier 14 asks the one question tier 13 left about the page that
spills: whether the missing `X` record is dropped when the page is
written out, or written into a page still in memory at the close.
A `reg d` without an initialiser shares the clock's arena, so that the
arena spills with a second key in it, and is written at times chosen
to fall in one page or the other.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t14_v_page_d` | `d` beside a clock that stays in one page | `X` records for all, in handle order, then the initial values |
| `t14_v_spill_d` | `d` first written in the second page of the spilling arena | `d` keeps its `X`; the clock loses its own |
| `t14_v_spill_d0`, `t14_v_spill_d2` | `d` written in the first page, and in both | the same |
| `t14_v_spill_dfst` | `d` declared before the clock | the same; the first record of the arena is not what goes |
| `t14_v_page_dd` | two writes of `d` across `#0` in a page that stays | two records at one time |
| `t14_v_spill_dd` | the two writes in the first page of the spilling arena | one record, the last |
| `t14_v_spill_dd2` | the two writes in the second page | two records |

So a page written out during the run keeps the last record per key and
time, and the `X` record was lost that way, sharing time 0 with the
initial value.
The last page of an arena, written at the close, keeps every delta.

Tier 15 takes the rest of that list.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t15_sv_pkg_log` | `log_wave -recursive /p` added to the script of `t13_sv_pkg` | the package parameter is logged, one record at time 0 |
| `t15_sv_iface_vec` | `logic [7:0] v` added to the interface | one more declaration and one more shared object per scope |
| `t15_sv_iface_mp` | the child takes `bus_if.slave` | unit kind `0x02`, named `bus_if.slave`; a modport scope under the instance; the port scope is of the modport unit; port mode `in` |

Tier 16 asks what writes the extra `X` record of a net at time 0,
by adding readers to the wire of `t11_v_wire`.
These truths carry a `records` count per signal, and `TestCorpus`
holds the object to it.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t16_v_wire_rd1` | a second wire assigned from the first | the read wire holds one `X` more; the reading wire does not |
| `t16_v_wire_rd2` | two wires assigned from the first | still one more |
| `t16_v_wire_rdp` | an `always` process reading the wire | one more |
| `t16_v_wire_rdi` | an input port that nothing inside reads | none more |

Tier 17 asks whether a Verilog write of the value already held
leaves a record, as tier 8 asked for VHDL, and whether a nonblocking
write differs from a blocking one.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t17_v_reg_same` | `s = 1'b0` to a `reg` holding 0 | no record |
| `t17_v_net_same` | a wire held at 0 by `assign w = s & 1'b0` while `s` toggles | no record |
| `t17_v_mem_same` | `m[2] = 8'h00` to an element holding 0 | no record |
| `t17_v_reg_nb` | `s <= 1'b1` | the same records as `s = 1'b1` |
| `t17_v_nb_swap` | `a <= b; b <= a;` | one record each, swapped |

Tier 18 asks what the `dims` word of an array type entry counts, since
every array type so far had one index dimension, and whether a VHDL
signal with a reader gets the extra time 0 record a Verilog net does.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t18_arr_2dim` | `array (0 to 1, 0 to 2) of std_ulogic` | `dims` `2`, two index words, row major value |
| `t18_arr_3dim` | `array (0 to 1, 0 to 1, 0 to 2) of std_ulogic` | `dims` `3`, three index words |
| `t18_sig_read` | `y <= s` beside `t1_bit_one_edge` | no extra record on `s` |

Tier 19 follows the two dimensional array with vectors as elements and
with an array over it, and walks the Verilog net kinds, first to see
whether the declaration kind word tells them apart and then to count
the time 0 records against drivers rather than readers.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t19_arr_2d_vec` | `array (0 to 1, 0 to 2) of std_ulogic_vector(3 downto 0)` | `dims` `2`, three triples, row major value |
| `t19_arr_of_2dim` | `array (0 to 1) of mat_t` | `dims` `1`, one triple |
| `t19_v_tri` | `tri w` | kind `0x06`, nothing else |
| `t19_v_wand` | `wand w` with two drivers | kind `0x04`; two `X` records |
| `t19_v_wor` | `wor w` with two drivers | kind `0x05`; two `X` records |
| `t19_v_supply1` | `supply1 w` | kind `0x0d`; `X` then `1` |
| `t19_v_wand_rd` | the `wand` read by an assignment | two `X` records |
| `t19_v_wire_3drv` | three drivers of one wire | two `X` records |
| `t19_v_wire_nodrv` | a wire with no driver | one `Z` record |
| `t19_v_2drv_port` | two drivers and an input port | three `X` records |
| `t19_v_nodrv_rd` | the undriven wire read | one `Z` record; the reader `X`, `Z` |
| `t19_sv_uwire` | `uwire w` | kind `0x03` |
| `t19_v_triand` | `triand w` with two drivers | kind `0x07` |
| `t19_v_trior` | `trior w` with two drivers | kind `0x08` |
| `t19_v_tri0` | `tri0 w` | kind `0x09`; `X` then `0` |
| `t19_v_tri1` | `tri1 w` | kind `0x0a`; `X` then `1` |
| `t19_v_supply0` | `supply0 w` | kind `0x0c`; `X` then `0` |

A `trireg` case was written and dropped: xsim refuses it with
`[XSIM 43-4096] Trireg is not supported`.

Tier 20 gives user enumerations the literals of the predefined types,
to tell whether the class word goes with the name or the literals,
crosses the 256 literal boundary to find the value size, leaves a two
dimensional array type unconstrained, and moves a `real` parameter's
value and keyword.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t20_enum_bitlike` | `type mybit_t is ('0', '1')` | class `2`, as `BIT` |
| `t20_enum_ul_like` | the nine `STD_ULOGIC` literals under another name | class `3`, as `STD_ULOGIC` |
| `t20_enum_chars` | `type sym_t is ('a', 'b', 'c')` | class `4`, as `CHARACTER` |
| `t20_enum_mixed` | `type mix_t is (alpha, 'b', gamma)` | class `4` |
| `t20_enum_one` | `type one_t is (only)` | class `5`; one record only |
| `t20_enum_two_id` | `type flag_t is (no, yes)` | class `5`, as `BOOLEAN` |
| `t20_enum_300` | 300 literals | last word `4`; a 4 byte value |
| `t20_enum_256` | 256 literals | last word `1`; a 1 byte value |
| `t20_enum_257` | 257 literals | last word `4` |
| `t20_enum_300_arr` | an array of two of the 300 literal type | 8 bytes |
| `t20_enum_300_rec` | a record of `std_ulogic` and the 300 literal type | 8 bytes, the wide field at 4 |
| `t20_arr_2d_uncon` | `array (natural range <>, natural range <>)` constrained at the signal | two `(0, 0, -2)` triples |
| `t20_rec_2dim` | a two dimensional array as a record field | two triples on the field |
| `t20_v_realp_big` | `parameter real R = 123456.789` | 16 bits |
| `t20_v_realp_lp` | `localparam real R = 1.5` | 16 bits |

Tier 21 turns to what every earlier case held constant.
It moves the Verilog time scale, since every case so far ran at 1 ps
and the DBG word `-12` had never moved; puts two parameter sets on one
Verilog module, as tier 4 did for VHDL generics; crosses the language
boundary in both directions; and spells values the earlier cases only
spelled one way: a negative integer, a negative real, an integer
subtype, a user physical type and a `bit_vector`.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t21_int_neg` | `s <= -165` | `5b ff ff ff` |
| `t21_int_sub` | `subtype small_t is integer range 0 to 7` | an entry `small_t` `0 to 7`; 4 bytes |
| `t21_int_newtype` | `type small_t is range 0 to 7` | byte identical to the subtype |
| `t21_real_neg` | `s <= -1.5` | `0xbff8000000000000` |
| `t21_phys_user` | `units um; mm = 1000 um; m = 1000 mm` | `um=1 mm=1000 m=1000000`; `3 mm` as 3000 |
| `t21_bitvec8` | `bit_vector(7 downto 0)` | `BIT_VECTOR` over `BIT`, 8 bytes |
| `t21_v_param_same` | two instances, `K` 7 and 7 | one unit, two `K` objects |
| `t21_v_param_diff` | two instances, `K` 7 and 9 | still one unit |
| `t21_v_ts_1ns_1ns` | `timescale 1ns / 1ns` | the DBG word `-9`; times in ns |
| `t21_v_ts_1ps_1ps` | `timescale 1ps / 1ps` | `-12` |
| `t21_v_ts_10ns` | `timescale 10ns / 1ns`, `#5` | `-9`; 50 |
| `t21_v_ts_1ns_100` | `timescale 1ns / 100ps`, `#50.55` | `-10`; 506, end 1001 |
| `t21_v_ts_1ps_1fs` | `timescale 1ps / 1fs`, `#50.5` | `-15`; 50500 |
| `t21_v_ts_none` | no `timescale` | `-12` |
| `t21_mix_v_in_vh` | a Verilog child under a VHDL testbench | both languages' units and declarations; the port on its own handle |
| `t21_mix_vh_in_v` | a VHDL child under a Verilog testbench | the VHDL port holds `U` then `0` |
| `t21_mix_ts_1ns` | the same under `timescale 1ns / 1ns` | `-12`; the VHDL precision wins |

The time scale cases changed the reader's interface: the times it
returns are in the file's own unit, `File.TimeUnit`, and the truth of a
case may add `time_ps` and `time_fs` to `time_ns`.

Tier 22 holds the source still and moves the elaboration.
Every case but one compiles the same testbench, with a generic on
`tb`, a `TIME` signal, a `1500 fs` wait and a function with a
parameter and a variable, and passes different options to xelab
through the `xelab_args` attribute of `wdb_case`.
The default of every other case is `-debug typical`.
Three options produce no database at all and are not cases: `-debug
line`, `-debug off` and `-debug subprogram` on its own make xsim refuse
`log_wave` with "compiled without trace information".

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t22_base` | the defaults | the baseline; `tb.k` a generic object; the function absent |
| `t22_dbg_wave` | `-debug wave` | DBG regions 14 and 15 empty |
| `t22_dbg_subprog` | `-debug typical -debug subprogram` | the function as a unit of kind `0x11`; locals of kind `0x14` |
| `t22_dbg_sub_proc` | the same with a procedure beside the function | a unit of kind `0x12`; `inout` and `in` modes |
| `t22_dbg_all` | `-debug all` | four library packages as root children; the `TEXT` file type |
| `t22_vh_fs` | `--timeprecision_vhdl 1fs` | `-15`; `TIME` in femtoseconds |
| `t22_vh_ns` | `--timeprecision_vhdl 1ns` | `-9`; the wait rounds to nothing |
| `t22_o0` | `--O0` | nothing |
| `t22_mt_off` | `--mt off` | nothing |
| `t22_gen_top` | `--generic_top k=9` | the values of `k` and `n` |

Tier 23 declares what no earlier VHDL case had declared: file
objects, a shared variable, a protected type, an access type,
subprogram locals of every size, a signal parameter, a procedure
inside a process, two architectures of one entity, and the 2008
predefined vectors.
The subprogram cases pass `-debug typical -debug subprogram` as tier
22 did; the rest use the defaults.
The access case crashed the reader on an unseen type kind.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t23_file_text` | `file f : text` in the architecture | the `TEXT` kind under the default debug level; `f` a 0 byte variable with one handle |
| `t23_file_int` | a file of integer | the words `8` and `40` again |
| `t23_file_sul` | a file of std_ulogic | the words `8` and `40` a third time |
| `t23_shared_int` | a shared integer variable | kind `0x0f` in scope `tb`, on the variable handle, no record |
| `t23_protected` | a protected type with one variable and one method | a record entry; the object with both handles in the signal handles |
| `t23_access` | `type int_ptr is access integer` | kind `0x8`, the words `8` and `48`, a 48 byte variable |
| `t23_access_vec` | an access to an unconstrained array | the same words |
| `t23_sub_sizes` | locals of 1, 4, 8, 8 and 4 bytes | frame offsets aligned to each size; the vector 24 bytes |
| `t23_sub_vec16` | the vector local at 16 elements | the next offset unchanged |
| `t23_sub_vec32` | the vector local at 32 elements | the next offset unchanged; a descriptor |
| `t23_sub_sig_prm` | `signal q : out std_ulogic` parameter | kind `0x15` with the mode; a 64 byte slot |
| `t23_sub_in_proc` | a procedure declared in the process | two scopes, `tb.flip` and `tb.p.flip`, one unit |
| `t23_arch_b` | `entity work.child(b)` of a child with two architectures | the unit `child(b)` only |
| `t23_arch_both` | both architectures instantiated | two units, nothing shared |
| `t23_int_vector` | `integer_vector(0 to 3)` | an unconstrained entry over `INTEGER`; the bounds in the declaration |
| `t23_real_vector` | `real_vector(0 to 3)` | 32 bytes |
| `t23_time_vector` | `time_vector(0 to 3)` | 32 bytes; the `TIME` entry |
| `t23_bool_vector` | `boolean_vector(0 to 3)` | 4 bytes |

Tier 24 goes after what the format does with things that have no
object of their own: signal attributes, a null range, two drivers of
one signal, an external name, a case generate, a configuration
specification, and the dynamic types of SystemVerilog; and it
elaborates one design under each `-debug` level in turn.
The debug cases share the two driver design and pass
`-debug wave -debug <level>`, or `-debug typical -debug <level>`.
The union crashed the reader on an unseen layout word.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t24_att_delayed` | `d <= s'delayed(2 ns)` | two implicit processes at one line; no object for the attribute |
| `t24_att_stable` | `s'stable(1 ns)` | one implicit process |
| `t24_att_quiet` | `s'quiet(1 ns)` | the same |
| `t24_att_transact` | `s'transaction` | the same; a repeated value leaves no record on `s` |
| `t24_null_range` | `std_ulogic_vector(0 downto 1)` | a 0 byte declaration; an object marked not logged |
| `t24_two_drivers` | two processes driving a `std_logic` | the initial value once per driver |
| `t24_dbg_drivers` | `typical` and `drivers` | nothing |
| `t24_dbg_readers` | `typical` and `readers` | header word 14 byte 2 |
| `t24_dbg_line` | `wave` and `line` | word 15 bytes 1 and 2; the statement regions |
| `t24_dbg_drv_only` | `wave` and `drivers` | word 14 byte 1; empty statement regions |
| `t24_dbg_sub_only` | `wave` and `subprogram` | word 15 bytes 1 and 2, like `line` |
| `t24_dbg_xlibs` | `wave` and `xlibs` | the four packages; no flag byte |
| `t24_case_gen` | `case k generate` | one scope, the taken alternative's declarations |
| `t24_ext_name` | `<< signal .tb.dut.s : std_ulogic >>` | an implicit process; the change recorded twice |
| `t24_config_spec` | a component and a configuration specification | the unit `child(a)` |
| `t24_sv_queue` | `int q[$]` | nothing but handle space |
| `t24_sv_dynarr` | `int d[]` | the same |
| `t24_sv_assoc` | `int a[string]` | the same |
| `t24_sv_class` | a class with one member | the same |
| `t24_sv_union` | `union packed` | layout `6`, the width of one field |
| `t24_sv_fork` | `fork ... join` | a `vprocess` scope per branch |
| `t24_sv_clocking` | `clocking cb @(posedge clk)` | nothing |

Tier 25 goes after DBG header word 13, which tier 24 left open, by
varying what was different in the cases where it was not `1`:
Verilog parameters in every form, `always_ff`, `always_comb` and
`always_latch` alone, a `wire` in a `.sv` file, and packages holding
one thing each.
The count followed the types and initializers of the objects, so the
second half of the tier varies the initializer of one object at a
time.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t25_v_prm_dflt` | `#(.K(5))` over a default of 5 | nothing but the value |
| `t25_v_prm_none` | no override | the same |
| `t25_v_prm_two` | two parameters overridden | two objects |
| `t25_v_prm_lp` | `localparam L = K + 1` | an object with the computed value |
| `t25_v_prm_lp_ind` | `localparam L = 3` | the same |
| `t25_v_defparam` | `defparam dut.K = 7` | nothing but the value |
| `t25_v_prm_tb` | a parameter of the root module | `tb.K` |
| `t25_sv_alw_ff` | `always_ff` alone | an `Always` scope; word 13 `1` |
| `t25_sv_alw_comb` | `always_comb` alone | word 13 `2`: the uninitialized output |
| `t25_sv_alw_latch` | `always_latch` alone | the same |
| `t25_sv_always` | `always` in a `.sv` file | word 13 `1` |
| `t25_sv_wire` | `wire` in a `.sv` file | word 13 `2`: the net |
| `t25_sv_pkg_tdef` | a package with a typedef only | the unit and scope, no object |
| `t25_sv_pkg_prm` | a package with a parameter only, used in a cast | no package at all |
| `t25_sv_pkg_unusd` | the package imported and not used | the same, `0xf8` less handle space |
| `t25_sv_logic_int` | `logic s = 0` | class 4 |
| `t25_sv_vec8_int` | `logic [7:0] s = 0` | class 4 |
| `t25_sv_vec8_sz` | `logic [7:0] s = 8'h00` | class 1 |
| `t25_sv_int_sized` | `int s = 32'h0` | class 3 |
| `t25_sv_int_noini` | `int s` | class 3 |
| `t25_sv_two_class` | `logic s = 1'b0` beside `int i = 5` | two entries, word 13 `2` |
| `t25_sv_two_same` | two `logic` with sized literals | one entry |
| `t25_sv_real_lit` | `real s = 1.5` | class 0 |
| `t25_sv_time_lit` | `time s = 10ns` | class 4; an implicit process, `X` then `10` |
| `t25_sv_time_noin` | `time s` | class 4; one `X` record |
| `t25_v_reg_int` | `reg s = 0` in `.v` | class 0 |
| `t25_v_vec8_sz` | `reg [7:0] s = 8'h00` in `.v` | class 0, where `.sv` gives 1 |
| `t25_v_int_sized` | `integer s = 32'h0` in `.v` | class 3 |
| `t25_v_int_noinit` | `integer s` in `.v` | class 3 |
| `t25_v_prm_real` | a `real` parameter beside an `integer` | `[3 0 0] [0 0 0]` |
| `t25_sv_net_init` | `wire w = s` | class 0 for the net |
| `t25_sv_bit_unsz` | `bit s = 0` | class 4 |
| `t25_sv_logic_one` | `logic s = '1` | class 1 |
| `t25_sv_v64_unsz` | `logic [63:0] s = 0` | class 4 |
| `t25_sv_byte_szd` | `byte s = 8'h05` | class 1 |
| `t25_sv_logic_exp` | `logic s = 1'b0 \| 1'b0` | class 1 |
| `t25_sv_logic_x` | `logic s = 1'bx` | class 1 |

Tier 26 varies the initializer forms tier 25 did not: a parameter
reference, a function call, unsized hex, negative literals, sized
literals narrower and wider than the target, `shortint`, two state
vectors, signed vectors, a string literal into a vector, a
concatenation, a replication, a conditional, an `event`, and the
SystemVerilog parameter types in the root module.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t26_sv_logic_prm` | `logic s = K` | class 1, as `K` |
| `t26_sv_int_prm` | `int s = K` | class 3 |
| `t26_sv_logic_fn` | `logic s = f()` | class 1; no unit or scope for `f` |
| `t26_sv_v8_unshex` | `logic [7:0] s = 'h00` | class 1 |
| `t26_sv_logic_1` | `logic s = 1` | class 4 |
| `t26_sv_int_neg` | `int s = -1` | class 3 |
| `t26_sv_int_szd5` | `int s = 5'd3` | class 3 |
| `t26_sv_int_szd64` | `int s = 64'h0` | class 3 |
| `t26_sv_int_unhex` | `int s = 'h0` | class 3 |
| `t26_sv_shortint` | `shortint s = 0` | class 3; the `shortint` entry |
| `t26_sv_lng_szd` | `longint s = 64'h0` | class 3 |
| `t26_sv_bit8_szd` | `bit [7:0] s = 8'h00` | class 1 |
| `t26_sv_bit8_int` | `bit [7:0] s = 0` | class 4 |
| `t26_sv_v32_int` | `logic [31:0] s = 0` | class 4 |
| `t26_sv_v32_szd` | `logic [31:0] s = 32'h0` | class 1 |
| `t26_sv_sgn8_neg` | `logic signed [7:0] s = -1` | class 3 |
| `t26_sv_sgn8_szd` | `logic signed [7:0] s = 8'h00` | class 1 |
| `t26_sv_real_int` | `real s = 1` | class 0 |
| `t26_sv_v8_str` | `logic [7:0] s = "a"` | class 6 |
| `t26_sv_v8_cat` | `{4'h0, 4'h0}` | class 1 |
| `t26_sv_v8_rep` | `{8{1'b0}}` | class 1 |
| `t26_sv_logic_cnd` | `(1 > 0) ? 1'b0 : 1'b1` | class 1 |
| `t26_sv_integer_x` | `integer s = 'x` | class 3 |
| `t26_sv_byte_neg` | `byte s = -1` | class 3 |
| `t26_sv_str_prm` | `parameter string P` | no object |
| `t26_sv_bit_prm` | `parameter bit B = 1'b1` | class 1 |
| `t26_sv_lp_int` | `localparam int L = 3` | class 3 |
| `t26_sv_real_prm` | `parameter real R = 1.5` in `.sv` | class 0 |
| `t26_sv_v8_prm` | `parameter logic [7:0] P = 8'h5a` | class 1 |
| `t26_sv_event` | `event e` | no object; `0x2c0` of handle space |

Tier 27 varies signedness, after tier 26 left the codes 3 and 4
looking like signed and unsigned integer initial values: `unsigned`
on the integral types, negative and positive unsized literals into
unsigned and signed vectors, signed sized literals, `time` from a
sized literal and from `0`, string, real and time literals into an
`int`, fill literals, and the packed types without an initializer.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t27_sv_int_uns` | `int unsigned s = 0` | class 3; `unsigned` not recorded |
| `t27_sv_int_unsni` | `int unsigned s` | class 3 |
| `t27_sv_byte_uns` | `byte unsigned s = 0` | class 4; reads back signed |
| `t27_sv_lng_uns` | `longint unsigned s = 0` | class 3 |
| `t27_sv_intg_uns` | `integer unsigned s = 0` | class 3 |
| `t27_v_intg_uns` | the same in `.v` | class 3 |
| `t27_sv_v8_neg` | `logic [7:0] s = -1` | class 4 |
| `t27_sv_sgn8_pos` | `logic signed [7:0] s = 5` | class 3 |
| `t27_sv_v8_ssized` | `logic [7:0] s = 8'sh05` | class 1 |
| `t27_sv_sgn8_szdn` | `logic signed [7:0] s = -8'sd1` | class 1 |
| `t27_v_sgn8_neg` | `reg signed [7:0] s = -1` in `.v` | class 0 |
| `t27_sv_time_szd` | `time s = 64'h0` | class 4; one record |
| `t27_sv_time_uns` | `time s = 0` in `.sv` | class 4; one record |
| `t27_sv_int_str` | `int s = "a"` | class 3; `97` |
| `t27_sv_int_real` | `int s = 1.5` | class 3; `2` |
| `t27_sv_int_time` | `int s = 10ns` | class 3; `0` then `10` |
| `t27_sv_real_szd` | `real s = 8'h05` | class 0 |
| `t27_sv_v8_xfill` | `logic [7:0] s = 'x` | class 1 |
| `t27_sv_v8_zfill` | `'z` | class 1 |
| `t27_sv_v8_0fill` | `'0` | class 1 |
| `t27_sv_v8_uns32` | `logic [7:0] s = 32'd5` | class 1 |
| `t27_sv_bit_noini` | `bit s` | class 0 |
| `t27_sv_byte_noin` | `byte s` | class 0 |
| `t27_sv_bit8_noin` | `bit [7:0] s` | class 0 |
| `t27_sv_v8_noini` | `logic [7:0] s` | class 0 |
| `t27_sv_str_untyp` | `parameter P = "hello"` in `.sv` | an object of class 6 |

Tier 28 hunts the value classes 2 and 5 with the initializers the
earlier tiers did not try: a real and a time literal into a vector,
enum literals and casts, assignment patterns into packed and unpacked
targets, parameter expressions, `$clog2`, a time parameter typed and
untyped, an enum parameter, `$time`, `$signed`, casts, a string into
a wider vector and into a typed parameter, a negative parameter into
an unsigned vector, and `realtime` after the untyped time parameter
came back as a vector holding a `float64`.
Neither class turned up.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t28_sv_v8_real` | `logic [7:0] s = 1.5` | class 0; `2` once |
| `t28_sv_v64_time` | `logic [63:0] s = 10ns` | class 0; an implicit process |
| `t28_sv_enum_pkd` | an enum over `logic [1:0]` from `RUN` | class 0; `XX` then `RUN` |
| `t28_sv_enum_cast` | `state_t'(1)` | a hidden variable `xilinx_isim_temp_0_ln6castingOp` |
| `t28_sv_pstr_pat` | a packed struct from `'{a: 1'b0, b: 4'b0000}` | class 0 |
| `t28_sv_pstr_szd` | a packed struct from `5'b00000` | class 1 |
| `t28_sv_pstr_int` | a packed struct from `0` | class 4 |
| `t28_sv_uarr_pat` | `logic [1:0] s [0:1] = '{2'b00, 2'b01}` | class 0 |
| `t28_sv_uarr_dflt` | `'{default: 2'b01}` | class 0 |
| `t28_sv_prm_expr` | `parameter K = 2 * 3` | 32 bits, class 3 |
| `t28_sv_prm_szexp` | `parameter K = 4'd5 + 4'd1` | 4 bits, class 1 |
| `t28_sv_prm_clog` | `parameter K = $clog2(8)` | 32 bits, class 3 |
| `t28_sv_prm_neg` | `parameter K = -1` | 32 bits, class 3 |
| `t28_sv_prm_time` | `parameter T = 10ns` | a 64 bit vector, class 4, a `float64` record |
| `t28_sv_prm_tmtyp` | `parameter time T = 10ns` | `time`, class 4 |
| `t28_sv_prm_realu` | `parameter K = 1.5` | `real`, 32 bits, class 0 |
| `t28_sv_prm_int_u` | `parameter int unsigned K = 5` | `int`, class 3 |
| `t28_sv_prm_enum` | `parameter state_t S = RUN` | the alias, class 3 |
| `t28_sv_v64_stime` | `logic [63:0] s = $time` | class 0; an implicit process |
| `t28_sv_v8_signed` | `logic [7:0] s = $signed(8'h05)` | class 1 |
| `t28_sv_int_cast` | `int s = int'(1.5)` | a hidden variable, class 3 |
| `t28_sv_v8_szcast` | `logic [7:0] s = 8'(0)` | a hidden variable of class 3, `s` class 0 |
| `t28_sv_v16_str` | `logic [15:0] s = "a"` | class 6 |
| `t28_sv_prm_lstr` | `parameter logic [7:0] P = "a"` | class 1 |
| `t28_sv_v8_prmneg` | `logic [7:0] s = K`, `parameter K = -1` | `s` class 4, `K` class 3 |
| `t28_sv_bit8_neg` | `bit [7:0] s = -1` | class 4 |
| `t28_sv_v8_bitsel` | `logic [7:0] s = K[7:0]` | class 1 |
| `t28_sv_v8_pow` | `logic [7:0] s = 2 ** 3` | class 4 |
| `t28_sv_real_time` | `real s = 10ns` | class 0; `0` then `10` |
| `t28_sv_str_pat` | an unpacked struct from `'{1'b0, 4'b0000}` | class 0 |
| `t28_sv_rtime_var` | `realtime s = 10ns` | the `realtime` entry; `0` then `10` |
| `t28_sv_rtime_prm` | `parameter realtime T = 10ns` | `realtime`, 16 bits, class 0 |
| `t28_sv_rtime_noi` | `realtime s` | `0` once |

Tier 29 asks where else the elaborator leaves a hidden object, after
tier 28 found the cast variable in an initializer: a cast in a
process statement, two casts in two initializers and in one, a cast
in a continuous assignment, in a function, in `always_comb`, in a
parameter and in a child module, a signing cast and a cast to `real`,
a streaming operator, `$bits`, an increment, a `for` loop with an
`int` index, the same loop with a module level `integer`, and a
`foreach`.
Four cases without the cast are the controls for the handle space.
A `case inside` statement was planned and dropped: `xsim` rejects it
as an unsupported construct.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t29_sv_cast_proc` | `s = int'(2.5)` in a process | no hidden variable; `0x1f0` more handle space |
| `t29_sv_cast_two` | two initializers with casts | `temp_0_ln5`, `temp_1_ln6`; one implicit process |
| `t29_sv_cast_same` | `int'(1.5) + int'(2.5)` in one initializer | nothing; `5` once |
| `t29_sv_cast_asgn` | `assign w = 8'(s + 1)` | two `NetRegassign` scopes; no hidden variable |
| `t29_sv_asgn_noc` | `assign w = s + 1` | one `NetRegassign` scope |
| `t29_sv_cast_fn` | `return int'(2.5)` in a function | no hidden variable; `0x1f0` more handle space |
| `t29_sv_fn_noc` | `return 3` in a function | the object `tb.f.f` |
| `t29_sv_cast_alwc` | `always_comb w = 8'(s + 1)` | no hidden variable; `0xf8` more handle space |
| `t29_sv_alwc_noc` | `always_comb w = s + 1` | the control |
| `t29_sv_cast_prm` | `parameter K = int'(1.5)` | a 32 bit vector, `2`; no cost |
| `t29_sv_cast_sub` | `int'(1.5)` in a child module | the hidden variable in `tb.u`, numbered 0 |
| `t29_sv_cast_sgn` | `logic [7:0] s = signed'(8'h05)` | a hidden 8 bit variable of class 1 |
| `t29_sv_cast_real` | `real s = real'(3)` | a hidden `real` |
| `t29_sv_stream` | `logic [7:0] s = {<<{8'h05}}` | nothing; `10100000` |
| `t29_sv_bits` | `int s = $bits(s)` | nothing; `32` |
| `t29_sv_incr` | `s++` | nothing |
| `t29_sv_for_int` | `for (int i = 0; i < 3; i++)` | `tb.Block7_1.i`, `0` then `3` |
| `t29_sv_for_modi` | `integer i` in the module | `i` records every value |
| `t29_sv_foreach` | `foreach (a[i])` | `tb.Block8_1.i` records every value |

Tier 30 goes after two things tier 28 left open: the value classes
2 and 5, which no case had produced, and the 16 byte record of the
untyped time parameter.
Thirteen `logic [7:0]` initializers and nine untyped parameters try
the forms not yet seen: a based literal without a size, signed and
unsigned, a negated sized literal, an expression with one sized
operand, a conditional with a sized arm, a comparison, a fill, a real
expression, `$signed` and `$unsigned` of an unsized literal, a string
wider than the target, a concatenation of strings, a 40 bit literal
and a shift past 32 bits.
Nine untyped time parameters vary the literal, the timescale, the
position and the count.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t30_sv_v8_ubase` | `'d5` | class 1 |
| `t30_sv_v8_sbase` | `'sd5` | class 1 |
| `t30_sv_v8_negsz` | `-8'd1` | class 1 |
| `t30_sv_v8_mixed` | `8'd5 + 1` | class 1 |
| `t30_sv_v8_szexp` | `4'd5 + 4'd1` | class 1 |
| `t30_sv_v8_uns` | `$unsigned(5)` | class 1 |
| `t30_sv_v8_sgnu` | `$signed(5)` | class 4 |
| `t30_sv_v8_realx` | `5 + 1.5` | class 0 |
| `t30_sv_v8_cnd` | `1'b1 ? 8'd5 : 0` | class 1 |
| `t30_sv_v8_1fill` | `'1` | class 1 |
| `t30_sv_v8_cmp` | `(1 < 2)` | class 1 |
| `t30_sv_v8_str2` | `"ab"` into 8 bits | class 6; `01100010` |
| `t30_sv_v16_strc` | `{"a", "b"}` into 16 bits | class 6 |
| `t30_sv_prm_ubase` | `parameter K = 'd5` | class 1, 32 bits |
| `t30_sv_prm_szsgn` | `parameter K = 8'sd5` | class 1, 8 bits |
| `t30_sv_prm_wide` | `parameter K = 40'h1` | class 1, 40 bits, 16 bytes; 8 more of handle space |
| `t30_sv_prm_shft` | `parameter K = 1 << 40` | class 3, 32 bits, `0` |
| `t30_sv_prm_realx` | `parameter K = 5.0 * 2` | `real`, class 0 |
| `t30_sv_prm_cnd` | `parameter K = 1'b1 ? 3 : 4` | class 3 |
| `t30_sv_prm_cmp` | `parameter K = (1 < 2)` | class 1, a `logic` of 1 bit |
| `t30_sv_prm_strc` | `parameter K = {"a", "b"}` | class 6, 16 bits, `62 61` |
| `t30_sv_prm_neg8` | `parameter K = -8'd1` | class 1, 8 bits, `ff` |
| `t30_sv_ptm_20ns` | `parameter T = 20ns` | `float64` `20` |
| `t30_sv_ptm_10ps` | `parameter T = 10ps` | `0.01`; the fraction is kept |
| `t30_sv_ptm_1us` | `parameter T = 1us` | `1000` |
| `t30_sv_ptm_1s` | `parameter T = 1s` | `1e9` |
| `t30_sv_ptm_frac` | `parameter T = 10.5ns` | `10.5` |
| `t30_sv_ptm_expr` | `parameter T = 10ns * 2` | the `real` entry, 32 bits, class 0 |
| `t30_sv_ptm_late` | the parameter after the variable | nothing |
| `t30_sv_ptm_ps_ts` | `timescale 1ps / 1ps` | `10000`; the unit of the file |
| `t30_sv_ptm_two` | two time parameters | the second half of a record is the next parameter's value |

Classes 2 and 5 did not appear.
The time parameter's record resolved: the `float64` is the first eight
bytes, and the reader decodes it, so the `stored` field of
`truth.json` went away.

Tier 31 is one word.
Declaration word 1 had been recorded as `0` since tier 2, and every
VHDL case agreed, because a VHDL file has one value class entry.
A sweep of the word over every case showed `1` and `2` on some
SystemVerilog declarations.
Ten cases move one `int i = 5` beside a `logic s` through the values,
the writes, the order and the language, with a `logic [7:0]` and a
`.v` file for contrast.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t31_sv_w1_i5` | `int i = 5` after `logic s` | word 1 `0` on `s`, `1` on `i`; `[1 0 0] [3 0 0]` |
| `t31_sv_w1_swap` | `i` before `s` | `0` on `i`, `1` on `s`; `[3 0 0] [1 0 0]` |
| `t31_sv_w1_i0`, `t31_sv_w1_i1`, `t31_sv_w1_i165` | the value of `i` | the same words |
| `t31_sv_w1_nowrt` | `i` never written | the same words |
| `t31_sv_w1_own50` | `i` written after its own delay | the same words; the end time moves to 110 ns |
| `t31_sv_w1_s5` | the `int` alone | `0`; `[3 0 0]` |
| `t31_sv_w1_v8_5` | `logic [7:0] i = 5` | `1`; `[1 0 0] [4 0 0]` |
| `t31_v_w1_int5` | the pair in a `.v` file | `1`; `[0 0 0] [3 0 0]` |

Word 1 is the index of the value class entry.
The reader keeps it, checks it, and the dump prints the class beside
each declaration.

Tier 32 came from the first design that was not written for the
ladder.
`//hdl/counter:sim`, the counter this repository started with, bundles
`clk`, `reset` and `enable` into one record and drives them one field
at a time, and the reader that reproduced 549 cases refused it: the
record had records at three addresses where the chunk rule predicts
one.
Every corpus case through tier 31 assigned VHDL signals whole, so the
rule that a VHDL record is the whole value had never been tested.
Twenty eight cases move one assignment through the parts of a record,
a vector and an array, the number of parts, their adjacency, the
delta, the driver and the width.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t32_rec_whole___` | the whole record | 8 bytes at the handle |
| `t32_rec_field___` | `r.b` alone | 1 byte at `+1` |
| `t32_rec_conc____` | `r.b` from a concurrent assignment | the same record |
| `t32_rec_two_adj_` | `r.b`, `r.c` in one delta | one record of 2 bytes at `+1` |
| `t32_rec_two_gap_` | `r.a`, `r.c` in one delta | two records, `+0` and `+2`, at one time |
| `t32_rec_delta___` | `r.a`, then `r.b` after `wait for 0 ns` | two records at one time |
| `t32_rec_wthenf__`, `t32_rec_fthenw__` | the whole record and a field in one delta | one whole record with the result |
| `t32_rec_vecfield` | a slice of a vector field | 4 bytes at `+5` |
| `t32_rec_intfld__`, `t32_rec_intlast_` | an integer field, and the field behind it | 4 bytes at `+0`; 1 byte at `+4` |
| `t32_vec_slice___`, `t32_vec_elem____`, `t32_vec_to_slice` | a slice, an element, a `to` slice | `+4`, `+5`, `+0` |
| `t32_vec_two_slc_`, `t32_vec_adj_slc_` | two slices apart, two slices touching | two records in source order; one of 6 bytes |
| `t32_vec_slc_conc` | a concurrent slice assignment | the same record as the process |
| `t32_vec_slc_over` | the whole vector then a slice | one whole record |
| `t32_arr_elem____`, `t32_arr_row_____`, `t32_arr_row_bit_`, `t32_arr2d_elem__` | array parts | `+4`, `+8`, `+13`, `+3` |
| `t32_wide_slice__`, `t32_wide_top____` | 300 of 600 bytes | four chunks of 75 from the write's address |
| `t32_wide_small__` | 4 of 600 bytes | one record at `+596` |
| `t32_wide_field__`, `t32_wide_tail___`, `t32_wide_tail_a_` | a 300 byte field before and behind a `std_ulogic` | chunks at `+0` and `+1`; 1 byte at `+0` |

A VHDL record is one write: the bytes one delta changed in one driver,
merged where they touch, chunked when 275 bytes or more.
The reader overlays writes, and the counter design decodes.
Its VCD is now part of `TestVCD`, the first check against a database
the corpus did not produce.

Tier 33 asks the Verilog question tier 32 left: every Verilog partial
write seen was one word pair.
Nine cases write 2400 bits of a 4800 bit reg at once, as the high
half, the low half, a slice off the pair grid, four bits, 1088 and
1089 bits either side of the threshold, the same in a `.sv` file, a
memory row, and a struct field.

| Case | Axis | Found |
| :--- | :--- | :--- |
| `t33_v_wsl_hi____` | `s[4799:2400]` | six chunks of 100 at `+600` |
| `t33_v_wsl_lo____` | `s[2399:0]` | six chunks at the handle, split at `0x800` |
| `t33_v_wsl_mid___` | `s[2415:16]` | 608 bytes, pairs 0 to 75, in six chunks |
| `t33_v_wsl_4b____` | `s[3:0]` | one pair |
| `t33_v_wsl_272___`, `t33_v_wsl_280___` | 1088 and 1089 bits from pair 34 | one record of 272; four of 70 |
| `t33_sv_wsl_hi___` | the high half in a `.sv` file | the same six chunks |
| `t33_v_mem_row___` | `m[1]` of four 2400 bit rows | six chunks at `+1200` |
| `t33_sv_st_wide__` | a 2400 bit struct field | six chunks at `+8` |

A Verilog partial write is whole pairs, and is chunked from its own
address by the rule of the VHDL one.
The two readers, one per language, became one: a whole write is the
first unused record at each chunk address of the value, and any
other record starts a partial write, chunked or alone.
The memory rows of `t33_v_mem_row___` also showed a split chunk's
rest sitting behind the other writes of its time in the next arena,
so the reader chains chunks by address, not by position.

Tier 34 moves a partial write out of the one process tier 32 used:
into a child, through a port bound whole, to a field and to a slice;
into a second process and a `for generate`; and onto a resolved type
with two or three drivers.

| Case | Moves | Records |
| :--- | :--- | :--- |
| `t34_port_fld_out` | `p.b <= '1'` in the child of `p : out trio_t` | 1 byte at `+1` on the shared handle |
| `t34_pmap_field__` | `a : out std_ulogic` bound to `r.b` | offset 1; 1 byte at `+1` |
| `t34_pmap_slice__` | `v(1 downto 0)` written in the child of `v => x(3 downto 0)` | 2 bytes at `+6` |
| `t34_two_prc_adj_` | `r.b` from `p`, `r.c` from `q` | two records, `q`'s first |
| `t34_two_prc_rev_` | `r.c` from `p`, `r.b` from `q` | two records, `q`'s first |
| `t34_gen_elems___` | `v(i) <= s` under `for i in 0 to 2 generate` | three records, `g(2)` first |
| `t34_res_two_fld_` | `t34_two_prc_adj_` on `std_logic` fields | the same two records |
| `t34_res_same_fld` | both processes on `r.b`, `'Z'` at time 0 | the whole record and `Z` at `+1` at time 0 |
| `t34_res_two_drv0` | `t24_two_drivers` without the time 0 assignment | `Z` once at time 0 |
| `t34_res_txn_zero` | one driver assigning `'Z'` at time 0 | `Z` once |
| `t34_res_3drv____` | a third driver assigning `'Z'` at 80 ns | `Z` twice, `1`, `X`, `X` |

A write through a port sits at the port's offset plus the part's, a
write is per driver and the processes' records come in the order the
simulator ran the processes, which `//hdl/uart:sim` showed is not the
source order, and the `t24_two_drivers` reading of one record per
driver was wrong: a resolved signal with several drivers records each
transaction, and the second `Z` was the assignment at time 0.
The reader needed no change.

Tier 35 puts SystemVerilog unpacked structs into arrays and arrays
into unpacked structs, the typedef of an unpacked struct array that
tier 34 left unwritten among them, and writes one element, one field
of one element, or one element of a field.

| Case | Shape | Records |
| :--- | :--- | :--- |
| `t35_sv_ust_arr__` | `rec_t m [0:1]`, `m[1]` written whole | 128 bits; 16 bytes at pair 0 |
| `t35_sv_ust_tdef_` | `typedef rec_t arr_t [0:1]` | the same, under a second alias |
| `t35_sv_pst_tdef_` | `typedef s_t arr_t [0:1]` over a packed struct | one 10 bit word, as `t13_sv_struct_ar` |
| `t35_sv_ust_fld__` | `m[1].b` written alone | 8 bytes at pair 0 |
| `t35_sv_st_uarr__` | `struct { logic a; logic [3:0] v [0:1]; }`, `s.v[1]` | 64 bits; pair 0 rewritten whole |
| `t35_sv_ust_nest_` | `struct { logic a; rec_t r; }`, `s.r.b` | 96 bits; 8 bytes at pair 0 |

An unpacked struct element takes a slot of its own pairs with the last
element lowest, a nested unpacked struct flattens into the outer
slots, and an unpacked array field packs into its own pairs.
The reader's Verilog decoder gained the slot per element for an
unpacked struct element, which it only had for `real`.

`//hdl/uart:sim` is the realistic design after the counter: a UART
transmitter looped back into a receiver whose characters fill a FIFO,
under one `uart_loop` entity with two generics, driven by a bench that
sends six characters, takes them out of the queue and counts
mismatches in `errors`.
The database holds 59 objects over 24 scopes, 17 types, 8 arenas and
17 pages, with enumerated states, ranged integers, a memory of
vectors, nested records and fields driven through three levels of
ports.
The reader reads every value, `TestVCD` agrees with Vivado's VCD, and
`errors` reads back as 0.
The design is not a corpus case: it has no `truth.json`, and the VCD
and the bench's own check stand in for it.

`//hdl/serv:sim` is the design from outside this repository: **SERV**
1.4.0, the bit serial RISC-V core, as the `servant_sim` system with a
RAM, a timer and a GPIO, running the `hello_uart` program from its own
`sw` directory.
The plan above named `rules_fusesoc` as the way to build it.
The bench uses a pinned `http_archive` of the SERV release instead,
with a build file in `third_party/serv` that lists the sources by
hand and one patch that moves a declaration above its first use, which
`xvlog` requires.
The archive is smaller than a FuseSoC tool chain, its hash pins the
sources, and the bench controls the elaboration: the top gets the
program through `-generic_top memfile=...`, and the firmware ends the
run itself when it writes the halt address.
The bench is `hdl/serv/tb.v`, adapted from SERV's `servant_tb.v` with
its UART decoder and its timeout.
The database holds 968 objects over 281 scopes, 23 module instances
under the top, 3.3 ms of run, and the reader reads every value and
`TestVCD` agrees with Vivado's VCD.
It is not a corpus case: it has no `truth.json`, and the VCD stands in
for it.

SERV taught three things the ladder had not, each then pinned by a
tier of its own: a port bound to a slice of a Verilog net, whose
object offset counts bits, tier 37; the writes of the value held by a
clocked nonblocking assignment and by a shared net, tier 36; and the
name of a Verilog top elaborated with `-generic_top`,
`tb(memfile="...")`.
It also caught two names `go-vcd-parser` does not read, `bufreg` and
the top's, so `TestVCD` now hides every name before parsing.

`//hdl/potato:sim` is the VHDL counterpart: **Potato** 0.3, a RV32I
processor in VHDL by Kristian Klomsten Skordal, as its own
`tb_processor` bench with an instruction memory, a data memory and
`textio` loading of both from hex files.
The bench is the release's own; the sources come from a pinned
`http_archive` with a build file in `third_party/potato` listing them
in the order `xvhdl` needs, and one patch that renames a conversion
function which VHDL-2008 makes a homograph of the one in
`std_logic_1164`.
The release's programs need a RISC-V tool chain the runner does not
have, so `hdl/potato/imem.hex` is a nine instruction program assembled
by hand: it stores a word, loads it back, counts it down and writes
`mtohost`, on which the bench prints `Success!` and stops.
The memory file names reach the top through `-generic_top`, the way
SERV's does.
The database holds 557 objects in 144 scopes, 136 units and 24 types,
among them a `file` and an `access` type and a record with a vector
field, and two 32768 byte memories; the reader reads every value and
`TestVCD` agrees with Vivado's VCD.
It is not a corpus case either.

Potato taught two things, each then pinned by a tier: the rest of a
chunked value is chunked again when it reaches 276 bytes, tier 39,
which the reader had refused as records at the wrong address; and an
array generic is declared with the range of its value, not of its
subtype, tier 40, so a string generic set on the command line has the
length of what was given.

`//hdl/picorv32:sim` is **PicoRV32** at commit `a473fc8`, the size
optimised RISC-V core by Claire Xen, under the project's own
`testbench_ez.v`, which holds a six instruction program in an
`initial` block and so needs no firmware image and no tool chain.
Nothing is written here but the build file in
`third_party/picorv32`; the core has no release to pin, so the archive
is pinned to a commit.
The database holds 280 objects over 62 scopes, 40589 values and 11 us
of run, and the reader read every value and agreed with the VCD on the
first attempt.
It taught nothing new, which is what a design of a language the corpus
already covers is expected to do.

`//hdl/neorv32:sim` is **NEORV32** 1.11.7, the RISC-V processor in
VHDL by Stephan Nolting, under the project's own `neorv32_tb`, booting
the instruction memory image the release ships, as a dual core system
with caches, an external bus and every peripheral the testbench turns
on.
The testbench has no stop of its own, because GHDL's `--stop-time`
ends it, so `hdl/neorv32/tb.ent.vhdl` instantiates it and calls
`std.env.stop` after 200 us; that is the only file written here.
The database holds 5696 objects over 4832 scopes, 3025 units, 198
types and 1395 arenas, 18875466 values, and a handle space of
`0x2b931c`, an order of magnitude past SERV.

NEORV32 taught one thing, and the reader was wrong on 39 of its 5696
objects before it: the records of an object belong to the signal at
its handle, so the signal's size sets the chunk boundaries, not the
object's.
`bus_req_i` of the DMA is a 88 byte record port at offset 1408 of
`iodev_req`, an array of 1760 bytes, and the array's initial write is
chunked at 146 bytes, so a chunk boundary falls inside the port's own
bytes: the port sees the ends of two chunks and no record that covers
it.
The reader took the chunk map from the object and refused the file
with `a first write of 146 bytes at 0x9422, which does not cover it`.
It now takes the chunk map from the largest object on the handle, and
checks that the first write's records together cover the object
rather than that one of them does.
Every design and all 1097 cases pass unchanged.

NEORV32 also caught a limitation of `go-vcd-parser`: a VCD identifier
code may be any printable ASCII, and a design this size gets codes
such as `#0`, which the parser's lexer reads as a timestamp, and `R0`,
which it reads as a real.
The fix is https://github.com/filmil/go-vcd-parser/pull/23, released
in v0.3.0, which is the version this repository takes; it carried the
patch in `third_party/go_vcd_parser` until then.

`//hdl/ibex:sim` is **Ibex** at commit `34b0705`, the lowRISC RISC-V
core, as the `simple_system` example the project ships: the core with
a bus, a memory, a timer and a control register block, in
SystemVerilog throughout.
It is the design that reads that language at scale, where the corpus
knows it only in miniature.
Two files are written here: `hdl/ibex/tb.sv`, which instantiates the
system and hands it the memory image, and `hdl/ibex/prog.vmem`, a
twelve instruction program assembled by hand because building anything
for RISC-V needs a tool chain the runner does not have.
The program sums 7 twelve times, stores the answer, reads it back and
writes to the simulation control register, which ends the run at
156 ns.
Ibex starts at the boot address plus `0x80`, so the image begins at
word `0x20`; the run before that was reading a memory the image never
wrote and the core's own assertions said so.
One patch comments out the `export "DPI-C"` lines: xsim generates C++
glue for an export and compiles it with the C compiler, which fails on
`extern "C"`.
The database holds 3287 objects over 1481 scopes and 150 types, and
73835 values.

Ibex found a defect in the checker rather than in the reader, which is
the first time that has happened.
A SystemVerilog enumeration declares its width in the entry's own
range, after the literals, and not in the base type, which for
`enum logic [1:0]` is a plain `logic`.
The reader had this right, because `bitsOf` passes the entry's ranges
down.
`TestVCD` did not: it spelled an enumeration with the width of the
base type, so an enumeration field of a packed struct contributed one
bit instead of two or four and every field after it moved.
The corpus had `t11_sv_enum4`, an enumeration of four bits, and the
error was invisible there because a VCD strips leading zeros and the
enumeration was not inside anything.
Tier 67 pins it.

Not written yet: a `string` value, which has no object to hold one,
see `t11_sv_str`.

Realistic designs come last, not first.
A FIFO or a UART is where the reader gets confirmed, not where it gets
discovered.
SERV came after the ladder, not instead of it: a database from SERV
differs from the baseline in hundreds of ways at once, and each of the
three things it taught was worked out in a tier of minimal pairs.

**Tier 36: writes of the value held.**
Verilog only.
SERV holds 2965811 records that repeat the value before them, and the
tier 17 rule, that a write of the value held records nothing, covers
none of them.
The tier takes the clocked nonblocking assignment and the shared net
apart, one axis per case, with the record count of every signal pinned
as `records` in `truth.json`.
The `truth.json` transitions list changes, since `TestCorpus` collapses
a repeated value; `records` is what pins the repeats.

| Case | Differs from | In | Records of the signal |
| :--- | :--- | :--- | :--- |
| `t36_v_nb_clk_lit` | `t17_v_reg_same` | `s <= 1'b0` at every posedge | 3: `X`, `0`, `0` at 25 ns |
| `t36_v_bl_clk_lit` | `t36_v_nb_clk_lit` | blocking | 2 |
| `t36_v_nb_clk_exp` | `t36_v_nb_clk_lit` | `s <= a & b`, operands held | 3 |
| `t36_v_init_expr_` | `t36_v_nb_clk_exp` | `s = a & b` from `initial` | 2 |
| `t36_v_net_mux___` | `t17_v_net_same` | `w = sel ? a : b`, `a = b` | 2 |
| `t36_v_net_and___` | `t17_v_net_same` | `w = a & b`, `b = 0` held | 2 |
| `t36_v_comb_and__` | `t36_v_net_and___` | `always @(a or b) s = a & b` | 2 |
| `t36_v_nb_clk_tog` | `t36_v_nb_clk_exp` | `a` toggling at 30 and 60 ns | 4: `0` at 25 and 75 ns |
| `t36_v_nb_clk_chg` | `t36_v_nb_clk_tog` | `s <= a`, `a` set at 30 ns | 4 |
| `t36_v_nb_clk_x__` | `t36_v_nb_clk_lit` | no initializer | 2: `X`, `0` at 25 ns |
| `t36_v_nb_two_lit` | `t17_v_reg_same` | `s <= 1'b0` twice from `initial` | 2 |
| `t36_v_nb_evt_lit` | `t36_v_nb_clk_lit` | `always @(a) s <= 1'b0` | 2 |
| `t36_v_nb_ini_evt` | `t36_v_nb_clk_lit` | two `@(posedge clk)` in `initial` | 2 |
| `t36_v_nb_dly_lit` | `t36_v_nb_clk_lit` | `always #25 s <= 1'b0` | 2 |
| `t36_v_nb_clk_150` | `t36_v_nb_clk_tog` | three edges, one toggle | 4: nothing at 125 ns |
| `t36_v_nb_clk_two` | `t36_v_nb_clk_tog` | `t <= 1'b0` in the same block | `t` 4, as `s` |
| `t36_v_nb_clk_net` | `t36_v_nb_clk_tog` | `assign w = s \| b` from the flop | `w` 2 |
| `t36_v_net_and_b_` | `t36_v_net_and___` | `b` drops to 0 at 20 ns | 2 |
| `t36_v_net_flop__` | `t36_v_net_and_b_` | `a` a flop | 2 |
| `t36_v_net_cnt___` | `t36_v_net_flop__` | `w = (c[4:2] == 3'b111)` | 2 |
| `t36_v_net_mux_w_` | `t36_v_net_mux___` | `a` and `b` written to 1 at 10 ns | 3 |
| `t36_v_net_mux_v6` | `t36_v_net_mux_w_` | 6 bit vectors | 3 |
| `t36_v_hier_mux__` | `t36_v_net_mux_w_` | the mux in a child, ports | 4: one `X` per object on the handle, then the changes |
| `t36_v_hier_int__` | `t36_v_hier_mux__` | `wire i` inside, port from `i` | `i` 7 |
| `t36_v_hier_and__` | `t36_v_net_and_b_` | the and in a child | 3 |
| `t36_v_hier_regs_` | `t36_v_hier_int__` | portless child | `i` 3 |
| `t36_v_net_wires_` | `t36_v_net_mux_w_` | operands wires of the regs | 3 |
| `t36_v_hier_i_and` | `t36_v_hier_int__` | `i = a & b` | `i` 5 |
| `t36_v_hier_i_or_` | `t36_v_hier_int__` | `i = a \| b` | `i` 6 |
| `t36_v_hier_i_sel` | `t36_v_hier_int__` | `sel` a reg of the child | `i` 7 |
| `t36_v_hier_i_ab_` | `t36_v_hier_int__` | `a`, `b` regs of the child | `i` 7 |
| `t36_v_net_copy__` | `t36_v_net_mux_w_` | `wire c = w` | `w` 7 |
| `t36_v_hier_i_noc` | `t36_v_hier_int__` | nothing reads `i` | `i` 3 |
| `t36_v_net_rd_not` | `t36_v_net_copy__` | `wire c = ~w` | `w` 7 |
| `t36_v_net_rd_cat` | `t36_v_net_copy__` | `wire [1:0] c = {w, sel}` | `w` 7 |
| `t36_v_net_rd_alw` | `t36_v_net_copy__` | `always @(w) c = w` | `w` 7 |
| `t36_v_bl_clk_rd_` | `t36_v_bl_clk_lit` | `wire c = s` | 2 |
| `t36_v_nb_clk_rd_` | `t36_v_nb_clk_lit` | `wire c = s` | 3 |
| `t36_v_comb_rd___` | `t36_v_comb_and__` | `wire c = s` | 2 |
| `t36_v_hier_p_nba` | `t36_v_nb_clk_lit` | `s` bound to a child input port | the port `X`, `X`, `0` |
| `t36_v_net_2drv__` | `t36_v_net_and___` | `assign w = a & b` twice | 5 |

The two rules are in [format/values.md](format/values.md): a clocked
nonblocking assignment records on its first run and after an event on
an operand of the block, and a net with two or more drivers and
readers records every evaluation of a driver.
The reader changed nothing for this tier.
`TestVCD` changed: the VCD keeps a few of these writes and drops the
rest, so the test compares the changes on both sides, see
[format/vcd.md](format/vcd.md).

**Tier 37: a port bound to a slice of a Verilog net.**
`t9_port_slice` bound a VHDL port to a slice and found the object
offset in bytes.
SERV binds `i_wb_adr` to `wb_mem_adr[12:2]` and the reader, counting
bytes, read the wrong bits.
The tier moves the slice through a 40 bit net.

| Case | Differs from | In | Offset |
| :--- | :--- | :--- | :--- |
| `t37_v_port_slc__` | `t9_port_slice` | Verilog, `v[5:2]` of 8 bits | 2, in bits |
| `t37_v_port_bit__` | `t37_v_port_slc__` | `v[3]` | 3 |
| `t37_v_port_pair1` | `t37_v_port_slc__` | `v[39:34]` of 40 bits | 34, in pair 1 |
| `t37_v_port_span_` | `t37_v_port_pair1` | `v[35:28]` | 28, across pairs 0 and 1 |
| `t37_v_port_reg__` | `t37_v_port_slc__` | the actual a `reg` | none: a handle of its own, 5 records |

The reader's `Changes` gained the bit offset for a Verilog port: it
reads the pairs the bits fall in and shifts them down.

**Tier 38: a memory loaded from a file.**
Verilog only.
The RAM of SERV is loaded by `$readmemh` and spans nine arenas, and
neither had a corpus case.
The bench names the file by its path from the workspace root, and the
new `data` attribute of `wdb_case` makes it a run time input of the
simulation.

| Case | Differs from | In | Records |
| :--- | :--- | :--- | :--- |
| `t38_v_mem4w32___` | `t11_v_mem2w32` | four words written one per statement at time 0 | 6 |
| `t38_v_rmh_4w____` | `t38_v_mem4w32___` | `$readmemh` from a four line file | 6, the same |
| `t38_v_rmb_4w____` | `t38_v_rmh_4w____` | `$readmemb` | 6, the same |
| `t38_v_rmh_2of4__` | `t38_v_rmh_4w____` | a two line file | 4 |
| `t38_v_rmh_at2___` | `t38_v_rmh_2of4__` | the file starts with `@2` | 4, `m[2]` and `m[3]` |
| `t38_v_rmh_rng___` | `t38_v_rmh_4w____` | `$readmemh(f, m, 1, 2)` | 4, `m[1]` and `m[2]` |
| `t38_v_rmh_desc__` | `t38_v_rmh_4w____` | `m [3:0]` | 6, `m[0]` first |
| `t38_v_rmh_twice_` | `t38_v_rmh_4w____` | the file loaded twice | 6: the second load records nothing |
| `t38_v_mem512____` | `t38_v_rmh_4w____` | 512 words, three arenas | 6; 28 chunks for the `X` write |

A load from a file is one element write per line, and the reader
changed nothing for the tier.

**Tier 39: the rest of a chunked value.**
VHDL.
The instruction memory of Potato, 32768 bytes, is 220 chunks of 148
by the chunk rule, and the reader found the last chunk of 356 bytes as
four records of 89.
No corpus value had a rest over 275 bytes: the rest is at most `n - 1`
bytes over the chunk, and `n` passes 148 only above 20000 bytes.
The tier is `std_ulogic_vector` signals sized to put the rest at the
threshold and around it, in the `t10_vec*` bench, plus one array of
4096 bytes whose loop writes a 30720 byte part.

| Case | Differs from | Rest | Rest's records |
| :--- | :--- | ---: | :--- |
| `t39_vec30022____` | `t10_vec30000` | 274 | one |
| `t39_vec30023____` | `t39_vec30022____` | 275 | one |
| `t39_vec20120____` | `t39_vec30023____` | 275 | one |
| `t39_vec20121____` | `t39_vec20120____` | 276 | four of 69 |
| `t39_vec20125____` | `t39_vec20121____` | 280 | four of 70 |
| `t39_vec20561____` | `t39_vec20125____` | 285 | 71, 71, 71, 72 |
| `t39_vec22347____` | `t39_vec22348____` | 295 | 73, 73, 73, 76 |
| `t39_vec22199____` | `t39_vec22348____` | 296 | four of 74, chunks of 147 |
| `t39_vec22348____` | `t39_vec22349____` | 296 | four of 74, chunks of 148 |
| `t39_vec22349____` | `t39_vec22348____` | 297 | 74, 74, 74, 75 |
| `t39_vec22647____` | `t39_vec22349____` | 299 | 74, 74, 74, 77 |
| `t39_vec22791____` | `t39_vec22647____` | 300 | four of 75 |
| `t39_vec32768____` | `t39_vec22791____` | 356 | four of 89, as Potato |
| `t39_mem4096_____` | `t39_vec32768____` | 175 | one; a 30720 byte part in 206 chunks |

The threshold for a rest is 276 where the threshold for a value is
275, and a rest of the rest below 276 stays whole.
The reader's `chunkLens` recurses on the rest and the corpus test
reads every case back.

**Tier 40: the range of a generic.**
VHDL.
Potato declares `RESET_ADDRESS` as an unconstrained
`std_logic_vector` with a literal default, and the file gives it
`(0 to 31)`; its file name generics, set by `-generic_top`, are
declared with the length of the path given.
The tier is one bench with a string generic and a vector generic, run
three ways.

| Case | Differs from | Axis | Declared |
| :--- | :--- | :--- | :--- |
| `t40_gen_uncons__` | `t9_gen_types` | generics of the top, `kv` unconstrained | `ks (1 to 3)`, `kv (0 to 3)` |
| `t40_gen_cons____` | `t40_gen_uncons__` | `kv : std_ulogic_vector(3 downto 0)` | `kv (3 downto 0)` |
| `t40_gen_str_top_` | `t40_gen_uncons__` | `--generic_top ks=hello` | `ks (1 to 5)`, 5 bytes, `hello` |

A vector generic set on the command line is not in the tier: `xelab`
reads `kv=01010101` as an integer literal, and a quoted value reaches
it with the quotes stripped.
`truth.json` gained a `range` field for a generic, and the corpus test
checks it.

**Tier 41: bounds below zero, and the IEEE fixed and float packages.**
VHDL.
No corpus case had a VHDL index range with a negative bound, and none
used `ieee.fixed_pkg` or `ieee.float_pkg`, whose types are declared
with them.
`std_ulogic_vector` is indexed by `NATURAL` and cannot take one, so
the tier declares `vec_t`, an `array (integer range <>) of
std_ulogic`, and constrains it at the signal.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t41_uvec________` | `t1_vec8` | `array (natural range <>) of std_ulogic` in the architecture | an entry of the `STD_ULOGIC_VECTOR` shape |
| `t41_neg_vec_____` | `t41_uvec________` | `vec_t(3 downto -4)` over `integer range <>` | the bound signed; `INTEGER` as the index |
| `t41_neg_asc_____` | `t41_neg_vec_____` | `vec_t(-4 to 3)` | `(-4 to 3)` |
| `t41_arr_subtype_` | `t41_neg_vec_____` | `subtype byte_t is vec_t(3 downto -4)` | an entry named `byte_t` holding the range; no `vec_t` |
| `t41_neg_arr_type` | `t5_int_arr` | `array (-2 to 1) of integer` | `(-2 to 1)` in the type entry |
| `t41_neg_int_sub_` | `t21_int_sub` | `integer range -8 to 7` | `-8 to 7` in the entry |
| `t41_sfixed______` | `t41_neg_vec_____` | `sfixed(3 downto -4)` | an entry `sfixed`, lower case, of the `vec_t` shape |
| `t41_ufixed______` | `t41_sfixed______` | `ufixed(3 downto -4)` | the same as `ufixed` |
| `t41_float32_____` | `t41_sfixed______` | `float32` | a constrained entry `(8 downto -23)`, 32 bytes |

The `range` field of a signal in `truth.json`, written since tier 11
and never read, is checked by the corpus test from this tier on, and
the 32 cases that had one pass.

**Tier 42: VHDL 2008 composites and generics.**
VHDL.
No case had a record field declared without bounds, an array of an
unconstrained element, a generic package or a type generic.
The record cases follow `t2_record2`, the array cases `t2_array2d`,
and the generic cases `t4_gen_explicit` and `t1_vec8`.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t42_rec_uncons__` | `t2_record2` | `bravo : std_ulogic_vector`, the signal `bundle_t(bravo(7 downto 0))` | the field triple `(0, 0, -2)`; the bounds in the declaration only; the reader misread the value |
| `t42_rec_subtype_` | `t42_rec_uncons__` | `subtype b8_t is bundle_t(bravo(7 downto 0))` | the entry renamed, the field still unconstrained |
| `t42_rec_two_cons` | `t42_rec_uncons__` | a second signal constrained `(3 downto 0)` | one entry, two declarations |
| `t42_rec_two_unc_` | `t42_rec_uncons__` | both fields unconstrained | `(3 downto 0) (7 downto 0)` on the declaration |
| `t42_rec_mix_unc_` | `t42_rec_two_unc_` | the first field constrained in the record | the same declaration list |
| `t42_rec_unc_nest` | `t42_rec_uncons__` | the unconstrained field in an inner record | `(0, 0, -2)` on the outer field |
| `t42_rec_unc_arr_` | `t42_rec_uncons__` | `array (0 to 1) of bundle_t`, `arr_t(open)(bravo(3 downto 0))` | the array's own index `(0, 0, -2)` |
| `t42_rec_unc_2dim` | `t20_rec_2dim` | an unconstrained two dimensional field | two `(0, 0, -2)` triples |
| `t42_arr_unc_elem` | `t2_array2d` | `array (0 to 1) of std_ulogic_vector` | both triples `(0, 0, -2)` |
| `t42_arr_unc_both` | `t42_arr_unc_elem` | `array (natural range <>) of std_ulogic_vector` | the index word only |
| `t42_gen_pkg_____` | `t1_vec8` | `package gp8 is new work.gp generic map (n => 8)` | a scope `gp`; no `n` |
| `t42_pkg_subtype_` | `t42_gen_pkg_____` | the subtype in a plain package | the name and paths only |
| `t42_gen_pkg_two_` | `t42_gen_pkg_____` | a second instance `gp4` | two scopes `gp` |
| `t42_gen_pkg_cons` | `t42_gen_pkg_____` | `constant width : natural := n` | an unlogged object `gp.width` |
| `t42_gen_type____` | `t4_gen_explicit` | `generic (type data_t; ...)` mapped to `integer` | the entry named `data_t` |
| `t42_gen_type_enu` | `t42_gen_type____` | `data_t => std_ulogic` | `enum "data_t"` |

A package instantiated inside an architecture was planned as a case
and dropped: `xelab` reports `The "Vhdl 2008 Package Instantiation
Declaration in Architecture Body" is not supported yet for simulation`.

The record cases changed the reader: `Decode` now takes the bounds of
a record's fields from the declaration's range list, and the field
triples are a fallback.

**Tier 43: unconstrained ports.**
VHDL.
Every port in the corpus had its bounds in the entity.
A port declared `std_ulogic_vector` takes them from the actual at each
instance, which asks where the bounds go and whether two instances
with different actuals repeat the unit the way different generics do.
The child reads `a(a'right)` into a signal so that the port is used.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t43_port_uncons_` | `t8_port_vec8` | `a : in std_ulogic_vector` bound to eight bits | the actual's bounds and size on the declaration |
| `t43_port_unc_two` | `t43_port_uncons_` | a second instance bound to four bits | the unit and declarations repeated |
| `t43_port_unc_sam` | `t43_port_unc_two` | the second instance bound to eight bits too | one unit, one set of declarations |
| `t43_port_unc_asc` | `t43_port_uncons_` | the actual `(0 to 7)` | `(0 to 7)` |
| `t43_port_unc_out` | `t43_port_uncons_` | `a : out std_ulogic_vector` | the same |
| `t43_port_unc_rec` | `t43_port_uncons_` | a port of a record with an unconstrained field | `(7 downto 0)` on the port; the field unconstrained in the type |

**Tier 44: times past 32 bits, and strings.**
VHDL and Verilog.
Every record time and page bound is read as 8 bytes, but the largest
time in the corpus was 70010000 ps, so the tier puts changes at 5 ms
and 5 s to see whether the high half is used.
A `string` had entered the type table only behind a `text` file, so
the tier declares a string signal and a string variable.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t44_time_5ms____` | `t1_bit_one_edge` | an edge at `5 ms` | 5000000000 in the record, the page bound and the end time |
| `t44_time_5s_____` | `t44_time_5ms____` | an edge at `5 sec` | 5000000000000 |
| `t44_time_late___` | `t44_time_5ms____` | edges at 1 ns and 5 ms | `t0` 0, `t1` 5000001000 |
| `t44_v_time_5ms__` | `t11_v_bit_edge` | Verilog `#5000000` under `1ns / 1ps` | the same as VHDL |
| `t44_str_sig_____` | `t2_character` | `s : string(1 to 5)` | `STRING` over `character` by `POSITIVE`; `(1 to 5)`, 5 bytes |
| `t44_str_sig_3to7` | `t44_str_sig_____` | `string(3 to 7)` | `(3 to 7)` as written |
| `t44_str_var_____` | `t6_var_int` | a `string(1 to 5)` variable | a 5 byte declaration, no record |

**Tier 45: what the script logs, and when.**
VHDL.
Every case so far logged everything under the top from time 0 in one
`run -all`.
The tier starts logging late, logs twice, logs one signal or one
scope, splits the run, and elaborates two top entities.
The cases with a script of their own carry it as `xsim.tcl`, and the
late ones name the log time as `log_ns` in `truth.json`, because the
VCD backdates the first value to `#0`.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t45_log_base____` | `t1_bit_two_edges` | three edges under the default script | the baseline of the tier |
| `t45_log_late____` | `t45_log_base____` | `log_wave` after `run 10 ns` | the first record at 10 ns holding the value held; `t0` 10000 |
| `t45_run_steps___` | `t45_log_base____` | the run in three `run` commands | no difference |
| `t45_log_twice___` | `t45_log_base____` | `log_wave` again at 10 ns | one record of the value held |
| `t45_log_one_____` | `t45_log_base____` | `log_wave /tb/s` beside an unlogged `u` | `u` listed and marked not logged |
| `t45_log_dut_____` | `t45_log_one_____` | `log_wave -recursive /tb/dut` | `tb.s` not logged, `tb.dut.c` as usual |
| `t45_log_dut_late` | `t45_log_dut_____` | the child log after `run 10 ns` | the child's first record at 10 ns |
| `t45_two_tops____` | `t45_log_base____` | two `--top` options | two root children in option order; only the first logged by `*` |
| `t45_two_tops_all` | `t45_two_tops____` | `log_wave -recursive /tb2` and `/tb` | both tops record |

**Tier 46: scale, and the cost of a driver.**
VHDL and Verilog.
Every case so far had fewer than 2000 objects and 300 scopes.
The tier pushes the counts past 65535 in both languages, nests an
entity in itself 100 levels deep, and separates the price of a signal
from the price of its drivers.
The large cases list their signals in `truth.json` as one entry with a
`count` and a `%d` in the names, which the corpus test expands.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t46_sig_1000____` | `t45_log_base____` | 1000 `std_ulogic` signals, two driven | undriven signals `0xc0` apart, driven `0xf0`; handle space `0x1088 + 2 * 0x148 + 998 * 0xf8` |
| `t46_gen_70000___` | `t46_sig_1000____` | a for generate of 70000 iterations | 140004 scopes, 140000 objects and 18147 slots in whole 32 bit words; indexes after the last signal |
| `t46_v_gen_70000_` | `t46_gen_70000___` | the same in Verilog | 70000 registers in one scope, `0xc0` apart, no stride for the writer |
| `t46_deep_100____` | `t8_gen_if_______` | a recursive entity, 100 levels | 306 scopes, paths of 101 names; generics after the last signal |
| `t46_drv_2_next__` | `t24_two_drivers_` | a driven signal after a two driver `std_logic` | the next handle `0x140` on |
| `t46_drv_3_next__` | `t46_drv_2_next__` | a third driver | `0x178` on |
| `t46_v_wire_4asg_` | `t19_v_wire_3drv_` | a fourth `assign` on a wire | `0xf0` on, as with three |

**Tier 47: what a use clause costs.**
VHDL.
The nine tier 2 cases with `use ieee.numeric_std.all` had `0x1f8` more
handle space than the rest, which tier 46 recorded as a type cost.
The tier adds use clauses to the tier 1 baseline one at a time, and
packages of the design with nothing, types, subprograms and constants
in them.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t47_use_numstd__` | `t1_bit_one_edge_` | `use ieee.numeric_std.all`, unused | `0x1f8`: the clause, not the type |
| `t47_use_numbit__` | `t47_use_numstd__` | `numeric_bit` | `0x188` |
| `t47_use_mathrl__` | `t47_use_numstd__` | `math_real` | `0x400` |
| `t47_use_textio__` | `t47_use_numstd__` | `std.textio` | nothing; already in the file table |
| `t47_use_none____` | `t1_bit_one_edge_` | no library or use clause, a `bit` signal | `0x604` less; only `standard`, `textio` and `env` in the file table |
| `t47_use_1164_bit` | `t47_use_none____` | the usual clauses back | the `0x11d0` of a `std_ulogic` |
| `t47_use_lib_only` | `t47_use_none____` | `library ieee;` alone | nothing |
| `t47_use_one_name` | `t47_use_lib_only` | `use ieee.std_logic_1164.std_ulogic;` | the price and files of `.all` |
| `t47_use_pkg_emp_` | `t47_use_pkg_typ_` | an empty package of the design | `0x80`, the scope and unit |
| `t47_use_pkg_typ_` | `t47_use_pkg_two_` | a package with one subtype | `0x80` |
| `t47_use_pkg_4arr` | `t47_use_pkg_typ_` | four array types | `0x80` |
| `t47_use_pkg_fn2_` | `t47_use_pkg_typ_` | two functions with a body | `0x80` |
| `t47_use_pkg_pr2_` | `t47_use_pkg_fn2_` | two procedures with a body | `0x80` |
| `t47_use_pkg_two_` | `t47_use_numstd__` | an `integer` and a `std_ulogic` constant | `0x88`; two unlogged objects |
| `t47_use_pkg_nul_` | `t47_use_pkg_two_` | two null range constants | `0xa0`; size 0 declarations with `(0 downto 1)` |

**Tier 48: the port position word.**
Verilog.
A sweep of the two open words of the instance record over the corpus
and `//hdl/serv:sim` found the word at `40` at 0 to 29 on the ports
of `serv`, in the order of each module's port list, and the word at
`44` holding values with the shape of addresses.
The tier separates the port list from the connection and from the
declaration order.
The `position` field of `truth.json` checks it.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t48_v_port_pos4_` | `t36_v_hier_and__` | four ANSI ports connected by name in order | 0, 1, 2, 3 |
| `t48_v_port_rev__` | `t48_v_port_pos4_` | connected by name in reverse order | 0, 1, 2, 3; the input port handles swap |
| `t48_v_port_posit` | `t48_v_port_pos4_` | connected by position | 0, 1, 2, 3 |
| `t48_v_port_nansi` | `t48_v_port_pos4_` | non ANSI header, declarations reversed | objects `d`, `c`, `b`, `a` with 3, 2, 1, 0 |
| `t48_v_port_open_` | `t48_v_port_pos4_` | output `d` unconnected | 3 still; `d` gets its own handle |

**Tier 49: the storage class word.**
VHDL and mixed language.
The sweep of tier 48 also found the word at `28` of the instance
record at 1, 3, 4 and 6 on the mixed language and subprogram cases,
beside the 0 and 2 the reader tested for.
The tier varies the type, class and mode of subprogram objects under
`-debug subprogram`, and puts an output and a second language
boundary into the mixed design.
The `storage` field of `truth.json` checks the word.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t49_sub_rec_loc_` | `t23_sub_sizes___` | a record local | 4 |
| `t49_sub_int_arr_` | `t23_sub_sizes___` | an integer array local | 4 |
| `t49_sub_str_prm_` | `t23_sub_sig_prm_` | an unconstrained `string` parameter | absent from the file |
| `t49_sub_var_prm_` | `t23_sub_sig_prm_` | an `inout` `variable` parameter | 3 |
| `t49_sub_sig_in__` | `t23_sub_sig_prm_` | an `in` signal parameter | 6 |
| `t49_sub_vec_prm_` | `t49_sub_str_prm_` | a constrained vector `constant` parameter | 4 |
| `t49_sub_sig_vec_` | `t23_sub_sig_prm_` | a vector signal parameter | 6 |
| `t49_mix_2port___` | `t21_mix_v_in_vh_` | a Verilog child with an output | 1 on both ports; `U`, `X`, `0` on the driven VHDL signal |
| `t49_mix_deep____` | `t49_mix_2port___` | a VHDL leaf under the Verilog child | 1 on the leaf's ports; no `U` on its input |


**The header count sweep.**
No new case.
The 17 DBG header words of every database, the 758 corpus cases and
the four external designs, were held against the lengths of the 18
regions.
Word `i` for 0 to 13 counts region `i + 4`, in records, in bytes to the
last NUL of a name pool, in words for region 15, and as 0 for the empty
regions 5 to 8, 12 and 16.
The reader now rejects a file whose counts do not fit, and a unit test
raises word 0 of a corpus database by one to see the rejection.
The [hierarchy](format/hierarchy.md) file has the table.

**Tier 50: storage classes by class and mode, and declaration order.**
VHDL.
The tier holds the storage class against the class and the mode of a
subprogram object, looking for the unseen 5, and measures the frame
offsets around a signal parameter and an absent parameter.
Three cases then move a constant above a signal, a constant above a
process variable, and put two signals out of name order, to fix what
orders the declarations of a unit.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t50_sub_var_vec_` | `t49_sub_var_prm_` | an `inout` `variable` vector parameter | 4 |
| `t50_sub_var_rec_` | `t50_sub_var_vec_` | an `inout` `variable` record parameter | 4 |
| `t50_sub_in_var__` | `t49_sub_var_prm_` | an `in` `variable` scalar parameter beside a signal parameter | 3; the signal parameter on `0xd8`, listed first |
| `t50_sub_acc_loc_` | `t23_sub_sig_prm_` | an access type local | 3, on `0x110` after the signal parameter |
| `t50_sub_str_loc_` | `t49_sub_int_arr_` | a `string(1 to 4)` local | 4, on `0x110` |
| `t50_sub_sig_rec_` | `t49_sub_sig_vec_` | a record signal parameter | 6 |
| `t50_sub_ivec_prm` | `t49_sub_str_prm_` | an unconstrained `integer_vector` parameter | absent; 24 bytes of frame |
| `t50_sub_func_prm` | `t49_sub_vec_prm_` | a vector parameter of a function | 4 on `0x40`; the scalar local on `0x58` |
| `t50_ord_const1st` | `t5_tr1000_______` | a constant declared above the signal | the signal listed first |
| `t50_ord_proc_con` | `t6_var_int______` | a constant above a variable in a process | source order; the constant has a record at 0 |
| `t50_ord_two_sig_` | `t5_tr1000_______` | signals `z`, `a`, `s` | source order, and the handles follow it |

5 was not seen.
Signals first, then the data objects, each in source order, is the
rule, and a subprogram's signal parameters count as signals.

**Tier 51: subprogram objects of SystemVerilog, and three VHDL shapes.**
SystemVerilog and VHDL.
The SystemVerilog half varies the lifetime and the argument modes of a
task, against a static task with the shape of `t12_v_task`.
The VHDL half puts a loop, a file parameter and a package around a
procedure.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t51_sv_task_stat` | `t12_v_task______` | the static task in a `.sv` file | the same as in `.v` |
| `t51_sv_task_auto` | `t51_sv_task_stat` | `task automatic` | no argument or local in the file |
| `t51_sv_task_ref_` | `t51_sv_task_auto` | a `ref` argument | nothing |
| `t51_sv_task_out_` | `t51_sv_task_stat` | an `output` argument | its mode; 0 in the word at `40` |
| `t51_sv_task_inou` | `t51_sv_task_out_` | an `inout` argument | its mode; `X`, then the value in and out |
| `t51_sv_func_auto` | `t51_sv_task_auto` | `function automatic` | nothing |
| `t51_sv_task_stvr` | `t51_sv_task_auto` | a `static` local in the automatic task | the local listed |
| `t51_sub_loop_idx` | `t23_sub_sig_prm_` | a `for` loop in the procedure | no index |
| `t51_sub_file_prm` | `t23_sub_sig_prm_` | a `file` parameter | absent, 8 bytes of frame; the file object a size 0 variable |
| `t51_sub_pkg_proc` | `t23_sub_sig_prm_` | the procedure in a package | `pk.drive` under `pk` |

**Tier 52: strides of the second handle region, and scope costs.**
VHDL.
Fifteen cases declare `a : T` and then `b : integer`, as process
variables, as architecture constants or as generics, and read the
distance from `a` to `b`.
Eight cases put two instances of a `child` with a generic `k`, or two
iterations of a generate with the index `i`, under `tb`, and vary the
body of the child and of the iteration in step.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t52_var_int_____` | `t50_ord_proc_con` | a second integer variable | `b` 4 past `a` |
| `t52_var_sul_____`, `t52_var_bool____` | `t52_var_int_____` | `a` one byte | 4, `b` on a multiple of 4 |
| `t52_var_real____`, `t52_var_time____` | `t52_var_int_____` | `a` eight bytes | 8 |
| `t52_var_rec_____` | `t52_var_int_____` | `a` a record of a `std_ulogic` and an `integer` | 8 |
| `t52_var_vec4____`, `t52_var_str4____` | `t52_var_int_____` | `a` four elements | `0x14`, 16 more than the elements |
| `t52_var_vec8____` | `t52_var_int_____` | `a` eight elements | `0x18` |
| `t52_var_arr4____` | `t52_var_int_____` | `a` four integers | `0x20` |
| `t52_con_int_____`, `t52_con_real____`, `t52_con_vec8____` | the `t52_var_` case of the type | constants | the same strides |
| `t52_gen_int_____`, `t52_gen_vec8____` | the `t52_var_` case of the type | generics | the same strides |
| `t52_inst2_empty_` | `t4_gen_diff_two_` | two empty children with a generic each | `k` `0x30` apart |
| `t52_inst2_proc__` | `t52_inst2_empty_` | a process in the child | `0xc0` |
| `t52_inst2_sig___` | `t52_inst2_empty_` | an undriven signal in the child | `0x68` |
| `t52_inst2_sigprc` | `t52_inst2_sig___` | the process driving the signal | `0x118`, the `t7_gen_for` stride |
| `t52_gi2_empty___` | `t7_gen_for______` | the iteration empty | no iteration scope and no index |
| `t52_gi2_proc____` | `t52_inst2_proc__` | an iteration for the instance | `0xc0` |
| `t52_gi2_sig_____` | `t52_inst2_sig___` | an iteration for the instance | `0x68` |
| `t52_gi2_sigprc__` | `t52_inst2_sigprc` | an iteration for the instance | `0x118` |

**Tier 53: the cost of each item of a scope in the second region.**
VHDL.
The two instance shape of tier 52 with one item of the child's body
varied per case, and three cases that change the number of children
or wrap each child in an if generate or a block.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t53_inst1_empty_` | `t52_inst2_empty_` | one child | the first `k` at `0xeb8` still |
| `t53_inst3_empty_` | `t52_inst2_empty_` | three children | `0x30` each from `0xeb8` |
| `t53_inst2_2gen__` | `t52_inst2_empty_` | a second generic | 4 past `k`, the stride `0x30` still |
| `t53_inst2_const_` | `t52_inst2_empty_` | an architecture constant | the same |
| `t53_inst2_2proc_` | `t52_inst2_proc__` | a second process | `0x150`, `0x90` each |
| `t53_inst2_var___` | `t52_inst2_proc__` | a variable in the process | 4 past `k`, the stride `0xc0` still |
| `t53_inst2_2sig__` | `t52_inst2_sig___` | a second undriven signal | `0xa0`, `0x38` each |
| `t53_inst2_conc__` | `t52_inst2_sigprc` | the driver a concurrent assignment | `0x118` still; `0x110` in the first region |
| `t53_inst2_2drv__` | `t52_inst2_sigprc` | a `std_logic` with two driving processes | `0x1c8`; `0x140` in the first region |
| `t53_inst2_port__` | `t52_inst2_empty_` | an input port connected to `s` | `0x68`, as a signal, sharing `0x768` |
| `t53_inst2_portop` | `t53_inst2_port__` | the port left open | `0x68` |
| `t53_inst2_nest__` | `t52_inst2_empty_` | an empty grandchild with a generic | `d0`, `d0.e`, `d1`, `d1.e` at `0x30` each |
| `t53_ifgen_inst__` | `t52_inst2_empty_` | each child under an if generate | `0x28` per wrapper, both before the children |
| `t53_blk_inst____` | `t52_inst2_empty_` | each child under a block | the same |

**Tier 54: where the second region starts.**
VHDL.
A variable of `tb.p` probes the start of the second handle region
while the libraries, the signal and the `std.env.stop` come and go,
and a package of the design is used, referenced, or neither.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t54_none_noenv__` | `t54_none_nosig__` | no library, no signal, no `std.env.stop` | the variable at `0x738`; `standard` alone |
| `t54_noenv_sig___` | `t54_none_noenv__` | a `bit` signal | `s` at `0x768`, the variable at `0x860` |
| `t54_none_nosig__` | `t54_nosig_var___` | no library clause | the variable at `0x810`; `textio` and `env` with `standard` |
| `t54_lib_none_var` | `t52_var_int_____` | no library clause, a `bit` signal | the variable at `0x938` |
| `t54_lib_1164_bit` | `t54_lib_none_var` | `std_logic_1164` | the variable at `0xde0`, as with a `std_ulogic` |
| `t54_1164_noenv__` | `t52_var_int_____` | no `std.env.stop` | the variable at `0xda0`; the end time 50 ns |
| `t54_nosig_var___` | `t52_var_int_____` | no signal | the variable at `0xcb8` |
| `t54_lib_numstd_v` | `t52_var_int_____` | `numeric_std` as well | `0xf8` on |
| `t54_lib_mathrl_v` | `t52_var_int_____` | `math_real` as well | `0x308` on |
| `t54_pkg_con_var_` | `t52_var_int_____` | a package with a constant, read by name | `0x30` on; the constant at `0xd40` |
| `t54_pkg_2con_var` | `t54_pkg_con_var_` | a second constant | `0x30` on still; `0xd44` |
| `t54_pkg_use_var_` | `t54_pkg_con_var_` | a use clause and no reference | the same |
| `t54_pkg_unused__` | `t54_pkg_con_var_` | neither | the package absent |

**Tier 55: the declarations of a subprogram beyond its variables.**
VHDL, under `-debug subprogram`.
A hunt for storage class 5: a constant, a loop, an alias, a file, a
protected variable and a nested function in a subprogram, then the
method scopes of a protected type wherever the type and its variable
are declared.
None of them is class 5.
Two shapes xelab refuses: a VHDL-2008 subprogram instantiation,
`function f3 is new f generic map (k => 3)`, elaborates with "not
supported yet for simulation", and a method called on a package's
shared variable by its selected name, `work.pk.ct.bump`, ends
compilation with SIGABRT, so `t55_prot_pkg_sv_` reaches the variable
through package subprograms instead.

| Case | Differs from | Axis | Found |
| :--- | :--- | :--- | :--- |
| `t55_sub_loop____` | `t50_sub_func_prm` | a `for` loop in a function | no index; `c` and `v` on `0x40`, `0x44` |
| `t55_sub_con_loc_` | `t55_sub_loop____` | a constant between them | a local of class 3; `v` on `0x58` |
| `t55_sub_con_nori` | `t55_sub_con_loc_` | the initialiser of `v` not reading it | `0x58` still |
| `t55_sub_2con____` | `t55_sub_con_nori` | two constants | `0x44`, `0x58`, `v` on `0x6c` |
| `t55_sub_con_real` | `t55_sub_con_nori` | a real constant | `0x48`, `v` on `0x60` |
| `t55_sub_var_init` | `t55_sub_con_loc_` | a variable in its place | `0x44`, `v` on `0x48` |
| `t55_sub_con_arr_` | `t55_sub_con_loc_` | an array constant | class 4; `0x48`, `v` on `0x60` |
| `t55_sub_alias___` | `t51_sub_loop_idx` | an alias of the local | absent; `v` on `0x110` |
| `t55_sub_file_loc` | `t51_sub_loop_idx` | a file local before `v` | absent; `v` on `0x138` |
| `t55_sub_prot_loc` | `t51_sub_loop_idx` | a protected local before `v` | absent; `v` on `0x11c`; the methods as scopes, twice |
| `t55_sub_prot_2__` | `t55_sub_prot_loc` | two of them | `v` on `0x12c` |
| `t55_sub_prot_3__` | `t55_sub_prot_2__` | three | `v` on `0x13c` |
| `t55_sub_prot_typ` | `t55_sub_prot_loc` | the type and no variable of it | no method scopes |
| `t55_sub_nested__` | `t50_sub_func_prm` | a function inside the function | `tb.g` and `tb.f.g`, one unit, `0x40` in each |
| `t55_prot_shared_` | `t55_sub_prot_typ` | a shared variable of the type | the four scopes; `0x100` of handle space |
| `t55_prot_arch_pr` | `t55_prot_shared_` | the process calling the methods | the same scopes |
| `t55_prot_arch_2p` | `t55_prot_arch_pr` | two processes | the second pair under `tb` after `tb.p2` |
| `t55_prot_pkg____` | `t55_prot_shared_` | the type in a package | `pk.bump`, `pk.get`; `tb.p.bump`, `tb.p.get` |
| `t55_prot_pkg_prc` | `t55_prot_pkg____` | the process calling | the same |
| `t55_prot_pkg_2p_` | `t55_prot_pkg_prc` | two processes, `p2` calling first | the second pair under `tb.p2` |
| `t55_prot_pkg_2pl` | `t55_prot_pkg_2p_` | `p2` calling last | under `tb.p2` still |
| `t55_prot_pkg_sv_` | `t55_prot_pkg_prc` | the variable in the package, behind package subprograms | `pk.bump`, `pk.get` twice under `pk`; the variable absent |

**Tier 56: the types and composite values of a subprogram.**
VHDL, under `-debug subprogram`.
A function with a `std_ulogic` parameter and an integer local,
`t56_typ_none____`, gains one type declaration or one composite local at
a time, to find what each costs the handle space and the frame.
A type costs nothing, a composite local costs the bytes of its static
initial value, and each aggregate or string literal in the body costs
its bytes again.
Three process variable cases without initialisers check that the tier
52 rule does not depend on the initialiser.

| Case | Adds | Differs from |
| :--- | :--- | :--- |
| `t56_typ_none____` | the baseline function `f(c)` with local `v` | `t55_sub_loop____` |
| `t56_typ_arr_unus` | `type arr_t is array (0 to 3) of integer`, used by nothing | `t56_typ_none____` |
| `t56_typ_arr_loc_` | a local `a : arr_t := (others => 0)` | `t56_typ_arr_unus` |
| `t56_typ_arr_2loc` | two locals of `arr_t` | `t56_typ_arr_loc_` |
| `t56_typ_arr_2typ` | two array types, a local of each | `t56_typ_arr_2loc` |
| `t56_typ_arr8_loc` | a local of an array of eight integers | `t56_typ_arr_loc_` |
| `t56_typ_vec4_loc` | a local `w : std_ulogic_vector(3 downto 0) := "0000"` | `t56_typ_none____` |
| `t56_typ_vec4_2lc` | two such locals | `t56_typ_vec4_loc` |
| `t56_typ_vec4_sub` | the local of a named subtype of the vector | `t56_typ_vec4_loc` |
| `t56_typ_int_rng_` | a local `n : integer range 0 to 7 := 0` | `t56_typ_none____` |
| `t56_typ_enum_loc` | an enumeration type and a local of it | `t56_typ_none____` |
| `t56_typ_rec_loc_` | a record type of a `std_ulogic` and an integer, a local of it with a literal | `t56_typ_none____` |
| `t56_typ_rec_arr_` | a record local with a vector field | `t56_typ_rec_loc_` |
| `t56_typ_arr_prc_` | the array type in the architecture, a process variable of it | `t56_typ_arr_unus` |
| `t56_typ_arr_noin` | the array local without an initialiser | `t56_typ_arr_loc_` |
| `t56_typ_arr_dyn_` | the uninitialised local assigned `(others => v)` in the body | `t56_typ_arr_noin` |
| `t56_typ_arr_lit_` | the uninitialised local assigned `(others => 2)` in the body | `t56_typ_arr_noin` |
| `t56_typ_vec_noin` | the vector local without an initialiser, assigned `(others => '0')` | `t56_typ_vec4_loc` |
| `t56_typ_rec_noin` | the record local without an initialiser | `t56_typ_rec_loc_` |
| `t56_typ_rec_prm_` | the record local initialised by `(a => c, n => 1)` from the parameter | `t56_typ_rec_loc_` |
| `t56_prc_vec_init` | a process with two vector variables, initialised | `t52_var_vec4____` |
| `t56_prc_vec_noin` | the same without initialisers | `t56_prc_vec_init` |
| `t56_prc_arr_noin` | a process with two array variables, not initialised | `t56_typ_arr_prc_` |
| `t56_sub_arr_dyni` | `f(c, n)` with a local `a : arr_t := (others => n)` | `t56_typ_arr_loc_` |
| `t56_sub_rec_2int` | a record local of two integers | `t56_typ_rec_noin` |
| `t56_sub_rec_1int` | a record local of one integer | `t56_sub_rec_2int` |
| `t56_sub_rec_3int` | a record local of three integers | `t56_sub_rec_2int` |
| `t56_sub_rec_4int` | a record local of four integers | `t56_sub_rec_3int` |
| `t56_sub_rec_2rl_` | a record local of two reals | `t56_sub_rec_2int` |

**Tier 57: what `log_wave` can name.**
VHDL, under `-debug typical` unless noted.
One design holds a scalar, a vector, a record, a constant and a shared
variable in the top, a `for generate` with a signal, and a process with
a variable and a `for` loop, and every case runs it under a script that
names one object or scope with `log_wave`, and nothing else.
The truth marks everything the script did not reach `logged: false`.
Two observations of xsim shaped the scripts: `log_vcd {/tb/v[3]}`
writes the whole vector to the VCD while `log_wave {/tb/v[3]}` logs
nothing, so `t57_log_bit_____` writes no VCD at all, and a Tcl brace
list cannot end in the closing backslash of `\g(1)\`, so
`t57_log_gen_it__` quotes the path with doubled backslashes instead.

| Case | `log_wave` | Differs from |
| :--- | :--- | :--- |
| `t57_log_all_____` | `-recursive *` | `t7_gen_for______` |
| `t57_log_none____` | none | `t57_log_all_____` |
| `t57_log_var_____` | `/tb/p/w`, a process variable | `t57_log_all_____` |
| `t57_log_var_all_` | the same under `-debug all` | `t57_log_all_____` |
| `t57_log_shv_____` | `/tb/sv`, a shared variable | `t57_log_all_____` |
| `t57_log_con_____` | `/tb/c`, a constant | `t57_log_all_____` |
| `t57_log_loop____` | `/tb/p/k`, a loop index | `t57_log_all_____` |
| `t57_log_slice___` | `{/tb/v[2:1]}`, a slice | `t57_log_all_____` |
| `t57_log_bit_____` | `{/tb/v[3]}`, one bit | `t57_log_all_____` |
| `t57_log_rec_fld_` | `/tb/r.n`, a record field | `t57_log_all_____` |
| `t57_log_rec_____` | `/tb/r`, the record | `t57_log_all_____` |
| `t57_log_gen_sig_` | `{/tb/\g(1)\/gs}`, a signal of one iteration | `t57_log_all_____` |
| `t57_log_gen_idx_` | `{/tb/\g(1)\/i}`, the index of one iteration | `t57_log_all_____` |
| `t57_log_gen_it__` | `"/tb/\\g(1)\\"`, one iteration | `t57_log_all_____` |
| `t57_log_gen_____` | `/tb/g`, the generate statement | `t57_log_all_____` |
| `t57_log_proc____` | `/tb/p`, the process | `t57_log_all_____` |
| `t57_log_top_____` | `/tb`, the top without `-recursive` | `t57_log_all_____` |

**Tier 58: what `log_wave` can name, in SystemVerilog.**
SystemVerilog, under `-debug typical`.
The tier 57 question over a module with a `logic`, a vector, a
memory, a packed struct, an `int`, a `real`, a `parameter`, a
`localparam`, a generate block with a wire, a named block with an
`int`, and a static task with an argument and a local.
The generate block's wire has no path `log_wave` accepts, so
`t58_sv_log_gen_w` hands it the result of `get_objects -regexp`.

| Case | `log_wave` | Differs from |
| :--- | :--- | :--- |
| `t58_sv_log_all__` | `-recursive *` | `t57_log_all_____` |
| `t58_sv_log_none_` | none | `t58_sv_log_all__` |
| `t58_sv_log_bit__` | `{/tb/v[3]}`, one bit | `t58_sv_log_all__` |
| `t58_sv_log_slc__` | `{/tb/v[2:1]}`, a slice | `t58_sv_log_all__` |
| `t58_sv_log_mem_e` | `{/tb/m[1]}`, a memory element | `t58_sv_log_all__` |
| `t58_sv_log_mem__` | `/tb/m`, the memory | `t58_sv_log_all__` |
| `t58_sv_log_st_fl` | `/tb/st.a`, a struct field | `t58_sv_log_all__` |
| `t58_sv_log_st___` | `/tb/st`, the struct | `t58_sv_log_all__` |
| `t58_sv_log_int__` | `/tb/i`, the module's `int` | `t58_sv_log_all__` |
| `t58_sv_log_real_` | `/tb/r`, the module's `real` | `t58_sv_log_all__` |
| `t58_sv_log_prm__` | `/tb/P`, the parameter | `t58_sv_log_all__` |
| `t58_sv_log_lprm_` | `/tb/L`, the localparam | `t58_sv_log_all__` |
| `t58_sv_log_blkv_` | `/tb/blk/bv`, the named block's variable | `t58_sv_log_all__` |
| `t58_sv_log_blk__` | `/tb/blk`, the named block | `t58_sv_log_all__` |
| `t58_sv_log_tsk_l` | `/tb/inc/tmp`, the task's local | `t58_sv_log_all__` |
| `t58_sv_log_tsk_a` | `/tb/inc/x`, the task's argument | `t58_sv_log_all__` |
| `t58_sv_log_tsk__` | `/tb/inc`, the task | `t58_sv_log_all__` |
| `t58_sv_log_gen_w` | `[get_objects -regexp {/tb/.*gb\[1\].*}]`, the generate wire | `t58_sv_log_all__` |
| `t58_sv_log_gen__` | `{/tb/gb[1]}`, the generate block by path | `t58_sv_log_all__` |
| `t58_sv_log_top__` | `/tb`, the module without `-recursive` | `t58_sv_log_all__` |

**Tier 59: forced and deposited values.**
VHDL under a custom script, and SystemVerilog.
A `std_ulogic` driven `'1'` at 10 ns and `'0'` at 20 ns beside a
4 bit vector driven `"0101"` and `"1010"` at the same times, and the
script forces or deposits a value on one of them with `add_force`,
`remove_forces` and `set_value`.
The SystemVerilog cases put a `force` and a `release` in a second
`initial` block of a module with the same driver.
The truth lists the changes and pins the count of records, repeats
of the value held included, through `records`.

| Case | The force | Differs from |
| :--- | :--- | :--- |
| `t59_frc_none____` | none | `t3_late_________` |
| `t59_frc_s_const_` | `add_force /tb/s 1` before the run | `t59_frc_none____` |
| `t59_frc_s_cancel` | the same with `-cancel_after 5ns` | `t59_frc_none____` |
| `t59_frc_s_pat___` | `add_force /tb/s {0 0ns} {1 2ns} -repeat_every 4ns` | `t59_frc_none____` |
| `t59_frc_v_const_` | `add_force /tb/v 1111` | `t59_frc_none____` |
| `t59_frc_v_bit___` | `add_force {/tb/v[3]} 1` | `t59_frc_none____` |
| `t59_frc_mid_____` | `run 15 ns`, then `add_force /tb/s 0` | `t59_frc_none____` |
| `t59_frc_mid_same` | `run 15 ns`, then `add_force /tb/s 1`, the value held | `t59_frc_none____` |
| `t59_frc_release_` | `add_force /tb/s 0`, `run 15 ns`, `remove_forces /tb/s` | `t59_frc_none____` |
| `t59_frc_rel_same` | `add_force /tb/s 1`, `run 15 ns`, `remove_forces /tb/s` | `t59_frc_none____` |
| `t59_frc_twice___` | `add_force /tb/s 1`, `run 15 ns`, `add_force /tb/s 0` | `t59_frc_none____` |
| `t59_frc_deposit_` | `set_value /tb/s 1` before the run | `t59_frc_none____` |
| `t59_frc_dep_mid_` | `run 15 ns`, then `set_value /tb/s 0` | `t59_frc_none____` |
| `t59_frc_dep_same` | `set_value /tb/s 0`, the value held, before the run | `t59_frc_none____` |
| `t59_frc_sv_none_` | none, SystemVerilog | `t59_frc_none____` |
| `t59_frc_sv_force` | `force s = 1'b1` at 5 ns, `release s` at 15 ns | `t59_frc_sv_none_` |
| `t59_frc_sv_frc_0` | `force s = 1'b0` at 5 ns, `release s` at 15 ns | `t59_frc_sv_force` |
| `t59_frc_sv_long_` | `force s = 1'b1` at 5 ns, `release s` at 25 ns | `t59_frc_sv_force` |
| `t59_frc_sv_norel` | `force s = 1'b1` at 5 ns, no release | `t59_frc_sv_long_` |
| `t59_frc_sv_relon` | `release s` at 15 ns, no force | `t59_frc_sv_none_` |
| `t59_frc_sv_tcl__` | `add_force /tb/s 0` before the run, SystemVerilog | `t59_frc_sv_none_` |

The value after `remove_forces` in `t59_frc_rel_same` is `0`, though
the driver assigned `'1'` at 10 ns; the VCD xsim writes agrees, and
the truth records what the file holds.

**Tier 60: SystemVerilog under -debug all.**
A `logic s` written at 50 ns beside one more variable, elaborated with
`xelab_args = ["-debug", "all"]` where every SystemVerilog case before
ran under xsim's typical level, to see what the flag adds to the type
table, the debug section and the records.
The variable is written once from the `initial` block after the write
to `s`.
The truth lists the placeholder record of a string or class handle as
a 32 bit signal with one transition of zeros, pins a double record
through `records`, and marks a container `"logged": false`.

| Case | The variable | Differs from |
| :--- | :--- | :--- |
| `t60_dbg_none____` | none | `t11_sv_bit______` |
| `t60_dbg_vec_____` | `logic [3:0] v` | `t60_dbg_none____` |
| `t60_dbg_int_____` | `int i` | `t60_dbg_none____` |
| `t60_dbg_real____` | `real r` | `t60_dbg_none____` |
| `t60_dbg_str_____` | `string str` | `t60_dbg_none____` |
| `t60_dbg_str_log_` | the string, with `log_wave /tb/str` after `log_wave -recursive *` | `t60_dbg_str_____` |
| `t60_dbg_queue___` | `int q[$]` | `t60_dbg_int_____` |
| `t60_dbg_q_log___` | the queue, with `log_wave /tb/q` | `t60_dbg_queue___` |
| `t60_dbg_dynarr__` | `int d[]` | `t60_dbg_int_____` |
| `t60_dbg_assoc___` | `int a[string]` | `t60_dbg_int_____` |
| `t60_dbg_assoc_i_` | `int a[int]` | `t60_dbg_assoc___` |
| `t60_dbg_class___` | `c_t h` of `class c_t; int f = 1; endclass` | `t60_dbg_int_____` |
| `t60_dbg_class_2_` | the class with a second field `logic [3:0] g` | `t60_dbg_class___` |
| `t60_dbg_class_d_` | `c_t extends b_t`, one `int` field each | `t60_dbg_class___` |
| `t60_dbg_class_2h` | two handles `h` and `h2` of the class | `t60_dbg_class___` |
| `t60_dbg_class_n_` | `c_t h = new` at declaration | `t60_dbg_class___` |
| `t60_dbg_struct__` | `struct packed { logic a; logic [2:0] b; }` | `t60_dbg_vec_____` |
| `t60_dbg_mem_____` | `logic [3:0] m [0:1]` | `t60_dbg_vec_____` |

A case with `set_value /tb/str cd` after `run 25 ns` was tried and
dropped: the command ends the batch script without a message, and the
database closes at 25 ns.
The generators of tiers 57 to 60 are in `tools/corpus/`.

**Tier 61: the numbering under -debug all.**
The tier 60 shape again, elaborated with `-debug all`, with the type
declarations arranged to show how the numbers that replace the `-99`
trailer of an array, sit in the id word of a class and follow the
element of a container are assigned; see "The numbering" in
[format/types.md](format/types.md).
Each case moves one thing: the order of two declarations, the kind of
one field, the element of one queue.
The truth holds a class or string handle as in tier 60 and marks a
container `"logged": false`.

| Case | The variable | Differs from |
|---|---|---|
| `t61_num_cls_3f__` | `c_t` with `int`, `logic [3:0]` and `real` fields | `t60_dbg_class_2_`: `real` takes no number |
| `t61_num_cls_rev_` | the three fields in reverse order | `t61_num_cls_3f__`: the numbers swap with the order |
| `t61_num_cls_byte` | `int` and `byte` fields | `t60_dbg_class___`: both hold `1` |
| `t61_num_cls_byti` | `byte` then `int` | `t61_num_cls_byte`: the same |
| `t61_num_cls_long` | `int` and `longint` | `t61_num_cls_byte`: `longint` shares too |
| `t61_num_cls_2int` | two `int` fields | `t60_dbg_class___`: one entry, one number |
| `t61_num_cls_2vec` | `logic [3:0]` and `logic [7:0]` fields | `t60_dbg_class_2_`: one entry `logic` with number `1` for both |
| `t61_num_cls_ibv_` | `int`, `byte`, `logic [3:0]` | `t61_num_cls_byte`: the vector, declared last, holds `1`; `int` and `byte` hold `2` |
| `t61_num_cls_str_` | `int` and `string` fields | `t60_dbg_class___`: `string` takes no number |
| `t61_num_cls_q___` | `int` and `int q[$]` fields | `t60_dbg_class___`: `int` `1`, the queue `2` |
| `t61_num_cls_cls_` | `b_t { int g }`, `c_t { b_t hb }` | `t60_dbg_class_d_`: a handle field's class follows like a parent, `c_t` `0`, `b_t` `1`, `int` `2` |
| `t61_num_two_cls_` | `a_t ha` and `b_t hb`, both held | `t60_dbg_class_2h`: `a_t` `0`, its `int` `1`, `b_t` `2` |
| `t61_num_cls_int_` | `c_t h` beside `int i` written `7` then `9` | `t60_dbg_class___`: `i` declared, logged and recorded as under typical |
| `t61_num_q_cls___` | `c_t q[$]` | `t60_dbg_queue___`: `c_t` `0`, `int` `1`, the queue `2`; the declaration takes class `0` |
| `t61_num_q_q_____` | `int q[$][$]` | `t60_dbg_queue___`: `int` `0`, inner `1`, outer `2`; two `(0 to 0)` ranges |
| `t61_num_q_str___` | `string q[$]` | `t60_dbg_queue___`: the queue holds `0` |
| `t61_num_q_vec___` | `logic [3:0] q[$]` | `t60_dbg_queue___`: the vector `0`, the queue `1` |
| `t61_num_q_byte__` | `byte q[$]` | `t60_dbg_queue___`: `byte` `0`, the queue `1` |
| `t61_num_a_then_q` | `int a[string]` then `int q[$]` | `t60_dbg_assoc___`: `int` `0`, the assoc `2`, the queue `3` |
| `t61_num_q_then_a` | `int q[$]` then `int a[string]` | `t61_num_a_then_q`: the queue `1`, the assoc `3` |
| `t61_num_ai_thn_q` | `int a[int]` then `int q[$]` | `t61_num_a_then_q`: the assoc `3`, the queue `4` |

**Tier 62: net strengths, pull sources and gate primitives.**
A `logic s` written at 50 ns beside one net, under typical, to see
whether a drive strength, a pull source, a switch or a gate primitive
leaves anything in the declaration, the hierarchy or the records.
Nothing of the strength does; a gate or pull is a `Forked` scope; a
net with two or more drivers records bit by bit.
Each net's raw record count is pinned as `records`.
`tran`, `tranif1` and `trireg` were tried and do not elaborate in
this version.

| Case | The net | Differs from |
|---|---|---|
| `t62_str_none____` | none | `t11_sv_logic____`: the tier's base |
| `t62_str_wire____` | `wire w; assign w = s;` | `t62_str_none____`: `NetRegassign11_1`; `X`, `0`, `1` |
| `t62_str_tri_____` | `tri` | `t62_str_wire____`: kind `0x06` |
| `t62_str_uwire___` | `uwire` | `t62_str_wire____`: kind `0x03` |
| `t62_str_pullup__` | `wire w; pullup (w);` | `t62_str_wire____`: `Forked11_1`; `X`, `1`; the same handle space |
| `t62_str_pulldn__` | `pulldown (w)` | `t62_str_pullup__`: `X`, `0` |
| `t62_str_pu_drv__` | a pullup under `assign w = s ? 1'bz : 1'b0;` | `t62_str_pullup__`: `X`, `0`, `0`, then `1` |
| `t62_str_weak____` | `assign (weak0, weak1) w = 1'b1;` under the same driver | `t62_str_pu_drv__`: the same records |
| `t62_str_strong__` | a weak literal `0` and a strong `s` | `t62_str_wire____`: `X`, `0`, `0`, `1`; 8 more of handle space |
| `t62_str_equal___` | the same two drivers without strengths | `t62_str_strong__`: `X` at 50 ns; one record fewer at time 0 |
| `t62_str_mixed___` | `(strong0, weak1) w = s` against a pulled `1` | `t62_str_strong__`: `0` then `1` |
| `t62_str_supply__` | a supply literal `0` against `s` | `t62_str_strong__`: `0` throughout |
| `t62_str_wand____` | `wand` with a weak `0` and a strong `s` | `t62_str_strong__`: `1` at 50 ns, as the VCD |
| `t62_str_bufif___` | `bufif1 (w, 1'b1, s);` | `t62_str_wire____`: `Forked11_1`; `X`, `Z`, `Z`, `1` |
| `t62_str_bufif_n_` | the gate named `g1` | `t62_str_bufif___`: nothing |
| `t62_str_and_____` | `and (w, s, 1'b1);` | `t62_str_bufif___`: `X`, `0`, `1` |
| `t62_str_and_2___` | two `and` gates in one statement, two nets | `t62_str_and_____`: `Forked11_1` and `Forked11_2` |
| `t62_str_nmos____` | `nmos (w, 1'b1, s);` | `t62_str_bufif___`: the same |
| `t62_str_vec_pu__` | `wire [3:0] v` under `pullup p [3:0] (v);` and a driver of `zz01` | `t62_str_pu_drv__`: one `Forked` scope; 9 records at time 0 and 4 at 50 ns, per bit |
| `t62_str_vec_1drv` | `wire [3:0] v` with one driver | `t62_str_wire____`: `XXXX`, `0000`, `1101` |
| `t62_str_vec_2drv` | a second literal driver `z1zz` | `t62_str_vec_1drv`: one record per bit per write; `0X00` then `Z101` |
| `t62_str_gate_dly` | `and #3 (w, s, 1'b1);` | `t62_str_and_____`: `0` at 3 ns, `1` at 53 ns |

**Tier 63: partial drivers on a net.**
A `logic s` written at 50 ns beside one net, under typical, to see
whether a driver of a bit, a slice or a port bound to part of a net
records the whole net or the part.
The driver writes the pairs its bits fall in, whole; the first record
holds `X` on the driven bits and `Z` on the rest; an output port on
part of a net shares the handle with the bit offset.
Each net's raw record count is pinned as `records`.

| Case | Axis | Differs from, and what it showed |
| :--- | :--- | :--- |
| `t63_pdr_bit0____` | `wire [3:0] v; assign v[0] = s;` | `t62_str_vec_1drv`: `ZZZX`, `ZZZ0`, `ZZZ1` |
| `t63_pdr_bit3____` | `assign v[3] = s;` | `t63_pdr_bit0____`: `XZZZ`, `0ZZZ`, `1ZZZ` |
| `t63_pdr_two_bits` | `v[0] = s` and `v[3] = ~s` | `t63_pdr_bit0____`: five records, one per driver write |
| `t63_pdr_slice___` | `wire [7:0] v; assign v[3:0] = {4{s}};` | `t63_pdr_bit0____`: `ZZZZXXXX`, `ZZZZ0000`, `ZZZZ1111` |
| `t63_pdr_w64_bit0` | `wire [63:0] v; assign v[0] = s;` | `t63_pdr_bit0____`: a 16 byte first record, 8 byte writes at the handle |
| `t63_pdr_w64_bit6` | `assign v[63] = s;` | `t63_pdr_w64_bit0`: the writes at the handle plus 8 |
| `t63_pdr_w64_hi__` | `assign v[63:32] = {32{s}};` | `t63_pdr_w64_bit0`: the same as bit 63 |
| `t63_pdr_w64_all_` | `assign v = {64{s}};` | `t63_pdr_w64_hi__`: three whole records; 16 more of handle space |
| `t63_pdr_2400_bit` | `wire [2399:0] v; assign v[0] = s;` | `t63_pdr_w64_bit0`: the first record chunked, the writes 8 bytes |
| `t63_pdr_2400_hi_` | `assign v[2399:2000] = {400{s}};` | `t63_pdr_2400_bit`: 104 byte writes at byte 96 of chunk 4 |
| `t63_pdr_2400_all` | `assign v = {2400{s}};` | `t63_pdr_2400_hi_`: three chunked records |
| `t63_pdr_concat__` | `wire a, b; assign {a, b} = {s, ~s};` | `t62_str_wire____`: two handles, three records each |
| `t63_pdr_port_bit` | `child u(.i(s), .o(v[1]));` on 4 bits | `t63_pdr_bit0____`: `tb.u.o` on the net's handle, offset 1; `ZZXZ` twice |
| `t63_pdr_port_slc` | `.o(v[7:4])` on 8 bits | `t63_pdr_port_bit`: offset 4 |
| `t63_pdr_port_hi_` | `.o(v[63:32])` on 64 bits | `t63_pdr_port_slc`: offset 32; the writes at the handle plus 8 |

**Tier 64: several partial drivers on one net.**
A `logic s` written at 50 ns beside one net with two or more partial
drivers, under typical, to see the order and the place of the records
the drivers write, and what a second instance of a child leaves.
Each driver writes its own pair record; the order within a time is the
scheduler's; the port position word is written on the first instance
of a unit only; an `assign` in a generate block gets no block scope.
Each net's raw record count is pinned as `records`.

| Case | Axis | Differs from, and what it showed |
| :--- | :--- | :--- |
| `t64_ord_src_rev_` | `assign v[3] = ~s; assign v[0] = s;` | `t63_pdr_two_bits`: `1ZZX` then `1ZZ0`, the source order |
| `t64_ord_gen4____` | `for (i = 0; i < 4; ...) assign v[i] = s;` | `t63_pdr_two_bits`: `tb.NetRegassign11_1` to `_4`; nine records |
| `t64_ord_gen_rev_` | the loop counting down | `t64_ord_gen4____`: the same scopes; bit 3 first |
| `t64_ord_w64_two_` | `v[0] = s; v[63] = ~s;` on 64 bits | `t63_pdr_two_bits`: records at the handle and at the handle plus 8 |
| `t64_ord_2400_two` | `v[0] = s; v[2399] = ~s;` on 2400 bits | `t64_ord_w64_two_`: the second at byte 92 of chunk 5 |
| `t64_ord_unp_elem` | `wire [3:0] v [0:1]; assign v[1][2] = s;` | `t63_pdr_bit0____`: `(ZZZZ, ZXZZ)`, `(ZZZZ, Z0ZZ)`, `(ZZZZ, Z1ZZ)` |
| `t64_ord_unp_whol` | `assign v[1] = {4{s}};` | `t64_ord_unp_elem`: `(ZZZZ, XXXX)` and on |
| `t64_ord_two_kids` | `child u0(.i(s), .o(v[1])); child u1(.i(~s), .o(v[3]));` | `t63_pdr_port_bit`: `tb.u1.o` position 0; seven records |
| `t64_ord_two_same` | both inputs on `s` | `t64_ord_two_kids`: `u1` first at 50 ns |
| `t64_ord_gen_kids` | the two children from a generate loop, on `v[i * 3]` | `t64_ord_two_same`: `tb.g[0].u`, `tb.g[1].u`; `g[1].u.o` position 0 |
| `t64_ord_pos_expr` | one child on a scalar net from `~s` | `t64_ord_two_kids`: position 1 |
| `t64_ord_pos_bit3` | one child on `v[3]` from `s` | `t64_ord_two_kids`: position 1 |
| `t64_ord_two_nets` | two children on two scalar nets | `t64_ord_two_kids`: `tb.u1.o` position 0 |
| `t64_ord_three___` | three children on three nets | `t64_ord_two_nets`: `u1` and `u2` at 0 |
| `t64_ord_two_pos4` | two children of a four port module | `t64_ord_two_nets`: 0, 1, 2, 3 then four zeros |
| `t64_ord_two_mods` | a child of a second module after the first | `t64_ord_two_nets`: position 1 on its `o` |
| `t64_ord_inout___` | `assign v[1] = s; bidi u(.io(v[3]));` | `t63_pdr_port_bit`: offset 3, mode 0; five records |
| `t64_ord_self____` | `assign v[0] = s; assign v[1] = v[0];` | `t63_pdr_two_bits`: `ZZX0` twice; six records |
| `t64_ord_chain___` | `assign w[1] = v[0]` on a second net | `t64_ord_self____`: three records on each |

**Tier 65: times past 32 bits, across a page and in another unit.**
Three cases that put the tier 44 reading of the 8 byte times where it
had not been: a page whose records cross 2^32 of the unit, a write
past 2^32 nanoseconds, and an end time of 1 s.

| Case | Axis | Differs from, and what it showed |
| :--- | :--- | :--- |
| `t65_tim_1s______` | a write at 999 ms, an end at 1 s | `t44_time_5s_____`: 1000000000000 ps read back |
| `t65_tim_cross___` | 3000 writes every 1 ns from 4.293 ms | `t44_v_time_5ms__`: eight pages, the crossing inside page 3 |
| `t65_tim_ns_5s___` | a write at 4.5 s under `1ns / 1ns` | `t44_v_time_5ms__`: 4500000000 units of nanoseconds |

**Tier 66: the SystemVerilog constructs the corpus had not seen.**
A `logic s` written at 50 ns beside the construct, under typical, to
see what scope it leaves and what it declares.
A `final` block and an `always_latch` are `Always` scopes; assertions
and `specify` blocks leave nothing; a `covergroup` leaves nine
generated scopes; a `program` ends the run; a `bind` places the
instance at the `bind` line.
`checker` was tried and does not elaborate in this version.

| Case | Axis | Differs from, and what it showed |
| :--- | :--- | :--- |
| `t66_prc_final___` | `final begin ... end` | `t11_sv_logic____`: `tb.Always11_0` |
| `t66_prc_latch___` | `always_latch if (s) q <= 1'b1;` | `t11_sv_logic____`: `tb.Always12_0`, `q` recording `X` then `1` |
| `t66_prc_ass_imm_` | an immediate assertion in an `always` | `t11_sv_logic____`: the `Always` scope alone |
| `t66_prc_ass_conc` | `assert property (@(posedge c) 1'b1);` | `t66_prc_ass_imm_`: no scope for the assertion |
| `t66_prc_prop____` | a named `sequence` and `property` | `t66_prc_ass_conc`: no scope; `0x200` more of handle space |
| `t66_prc_task____` | a task called from an `initial` | `t11_sv_logic____`: the `task` unit and `tb.t` |
| `t66_prc_func____` | a function called from an `assign` | `t66_prc_task____`: the `function` unit and `tb.f` |
| `t66_prc_program_` | `prog p();` beside the module | `t11_sv_logic____`: `tb.p`, `tb.p.Initial20_0`; the run ends at 10 ns |
| `t66_prc_bind____` | `bind tb watcher b(.i(s));` | `t66_prc_func____`: `tb.b` at line 22, its port on its own handle |
| `t66_prc_specify_` | a child with `(i => o) = 1` | `t66_prc_func____`: records at 1 ns and 51 ns |
| `t66_prc_kid_____` | the same child without the block | `t66_prc_specify_`: records at 0 and 50 ns; `0x120` less |
| `t66_prc_spec_0__` | the path delay set to 0 | `t66_prc_specify_`: 0 and 50 ns, and the handle space of the plain child |
| `t66_prc_covgrp__` | `covergroup cg @(posedge s)` with one coverpoint | `t66_prc_ass_imm_`: nine scopes, three `xlnx_isim_covergroup_cg::` functions |

**Tier 67: an enumeration inside a packed struct.**
A packed struct of a logic, an enumeration and another logic, so that
the neighbours move if the enumeration's width is wrong.
The width is the range the enumeration entry carries after its
literals; the base type of `enum logic [1:0]` is a plain `logic` and
says nothing.

| Case | Axis | Differs from, and what it showed |
| :--- | :--- | :--- |
| `t67_esz_pk_2bit_` | `enum logic [1:0]` between two logics | `t11_sv_struct___`: 4 bits, and the VCD reads `1011` where a one bit enumeration would give `111` |
| `t67_esz_pk_4bit_` | the same over four bits | `t67_esz_pk_2bit_`: 6 bits, `100011` |
| `t67_esz_pk_int__` | the same over `int` | `t67_esz_pk_2bit_`: 34 bits, the width on the base type instead |

**Tier 68: where the value of a SystemVerilog string goes.**
Every case carries characters that occur nowhere else in a database,
`ZQXJ` and `WPMK`, so that the whole file can be searched for them:

```
bazel run //tools/pagegrep -- -pat ZQXJ \
    "$PWD/bazel-bin/hdl/corpus/t68_str_lit4____/sim.wdb"
```

The search reads the bytes as they lie and every record of every
inflated page, and finds nothing in any of the string cases.
`t68_str_byte____` is the control: the same characters in an unpacked
array of `byte` are found at once, so the answer is about strings and
not about the search.

| Case | Axis | Differs from, and what it showed |
| :--- | :--- | :--- |
| `t68_str_lit4____` | `string str = "ZQXJ";` under typical | `t11_sv_logic____`: no declaration, no object, no record, and the implicit initializer process |
| `t68_str_lit40___` | forty characters instead of four | `t68_str_lit4____`: nothing moves, the same 2619 bytes |
| `t68_str_noinit__` | the string without an initializer | `t68_str_lit4____`: `0xa14`, the `0x98` of the process |
| `t68_str_arr_____` | `string a [0:1]` under typical | `t68_str_lit4____`: absent as the scalar is, 8 bytes more of handle space |
| `t68_str_log_____` | `log_wave /tb/str` under typical | `t68_str_lit4____`: nothing, and the warning `No matching HDL object or HDL scope found` |
| `t68_str_dbg_____` | the four character string under `-debug all` | `t68_str_lit4____`: the declaration, the object and the eight zero bytes of tier 60 |
| `t68_str_dbg40___` | forty characters under `-debug all` | `t68_str_dbg_____`: the same file and the same record |
| `t68_str_dbg_arr_` | the array under `-debug all` | `t68_str_dbg_____`: one 64 bit object with one 16 byte record of zeros |
| `t68_str_byte____` | four `byte` holding the same characters | `t68_str_lit4____`: the characters, in one record, in reverse element order |

**Tier 69: the value class of the forms nothing had declared.**
Region 17 classes an object by its type and the form of its
initializer, and two of the codes, 2 and 5, have never appeared.
This tier declares thirteen forms the earlier sweeps did not have, one
per case, and reads the class of each.
None of them is a 2 or a 5, which is what the tier was looking for;
what it found instead is a handful of declarations nothing had
recorded.

| Case | Axis | Differs from, and what it showed |
| :--- | :--- | :--- |
| `t69_vcl_specprm_` | a `specparam` in a `specify` block | `t11_sv_logic____`: a parameter declaration, class 3, with an object and a record |
| `t69_vcl_supply0_` | a `supply0` net | `t11_sv_logic____`: the keyword is the declaration kind, class 0, X then `0` at time 0 |
| `t69_vcl_supply1_` | a `supply1` net | `t69_vcl_supply0_`: the same with `1` |
| `t69_vcl_const_v_` | `const logic [7:0] k = 8'd5` | `t11_sv_logic____`: class 1, the initializer's, as without `const` |
| `t69_vcl_const_i_` | `const int k = 5` | `t69_vcl_const_v_`: class 3, the `int`'s |
| `t69_vcl_defparam` | a child parameter set by `defparam` | `t11_sv_logic____`: an ordinary class 3 parameter |
| `t69_vcl_gtop_prm` | a parameter overridden with `-generic_top` | `t11_sv_logic____`: the same, and the top scope is renamed `tb(P=5)` |
| `t69_vcl_typeprm_` | a variable of a `parameter type` | `t11_sv_logic____`: the parameter's name `T` is the variable's type name |
| `t69_vcl_bits_prm` | a parameter from `$bits` | `t11_sv_logic____`: class 3, like any integer expression |
| `t69_vcl_bit_real` | `bit b = 1.5` | `t11_sv_logic____`: class 0, the real initializer's |
| `t69_vcl_enum_xcs` | an enumeration from `e_t'('x)` | `t11_sv_logic____`: class 0, and the cast's hidden variable is class 1 |
| `t69_vcl_wire_ini` | `wire w = 1'b1` | `t11_sv_logic____`: class 0, where the same literal on a variable is class 1 |
| `t69_vcl_chandle_` | a `chandle` | `t11_sv_logic____`: nothing at all, as a `string` leaves nothing |

Two forms have no case because xsim rejects them:
`trireg (large) t;` is `ERROR: [XSIM 43-4096] Trireg is not supported`
and `let five = 5;` is `ERROR: [XSIM 43-3980] The SystemVerilog
feature "Let" is not supported yet for simulation`.

**Tier 70: what an associative array's spare numbers are.**
Under `-debug all` an associative array of `int` keyed by `string`
takes the numbers 0 and 2, and one keyed by `int` takes 0 and 3, so
one or two numbers go to something the file does not write, tier 61.
Neither of those cases can say what: their key is either a `string`,
which takes no number, or the element's own type.
Every case here gives the key a type of its own.

| Case | Axis | Differs from, and what it showed |
| :--- | :--- | :--- |
| `t70_num_a_v_str_` | `logic [3:0] a[string]` | `t60_dbg_assoc___`: the numbers do not move with the element |
| `t70_num_a_b_str_` | `byte a[string]` | `t60_dbg_assoc___`: nor with a narrower one |
| `t70_num_a_i_byte` | `int a[byte]` | `t60_dbg_assoc_i_`: the `byte` key holds the number 1 in its own entry |
| `t70_num_a_b_int_` | `byte a[int]` | `t70_num_a_b_str_`: so does an `int` key that is not the element |
| `t70_num_a_e_key_` | `int a[e_t]` | `t60_dbg_assoc_i_`: an enumeration key spends none, and its typedef spends two before the element |
| `t70_num_a_2dim__` | `int a[string][int]` | `t60_dbg_assoc___`: the rule repeats per dimension, 3 then 5 |
| `t70_num_a_in_cls` | a class with an `int a[string]` field | `t60_dbg_class___`: the count starts at the class, which takes 0 |
| `t70_num_d_then_q` | `int d[]` then `int q[$]` | `t61_num_a_then_q`: a dynamic array and a queue leave nothing over |

What is left after that is one number per associative array and
nothing else; that it belongs to an iterator is still a guess, and
question 24 says so.

**Tier 71: the width a real parameter declares.**
A `real` parameter declares 16 bits where a `real` variable declares
32, and both record one `float64`.
Tier 20 ruled out the value and the `localparam` keyword; this tier
puts the type keyword in every other place it fits.

| Case | Axis | Differs from, and what it showed |
| :--- | :--- | :--- |
| `t71_rlw_sreal_p_` | `parameter shortreal R = 1.5` | `t25_v_prm_real__`: 16, as `real` does |
| `t71_rlw_rtime_p_` | `parameter realtime R = 1.5` | `t71_rlw_sreal_p_`: 16, on the `realtime` entry |
| `t71_rlw_untyped_` | `parameter R = 1.5` | `t71_rlw_sreal_p_`: 32, so the keyword carries the 16 |
| `t71_rlw_specprm_` | `specparam d = 1.5` | `t71_rlw_sreal_p_`: 32, as an untyped parameter |
| `t71_rlw_pkg_prm_` | the parameter in a package | `t71_rlw_sreal_p_`: 16, and no record, as a package parameter has none |
| `t71_rlw_kid_prm_` | the parameter in a child module | `t71_rlw_sreal_p_`: 16 |
| `t71_rlw_arr_prm_` | `parameter real R [0:1]` | `t71_rlw_sreal_p_`: 64, so 32 an element, on the same type entry |
| `t71_rlw_vhdl_gen` | a VHDL `generic r : real` | `t71_rlw_sreal_p_`: 8 bytes, the float itself |

The array is the one that settles it: two of the same parameters
declare 32 each, so the 16 is not the type's and not the value's but
the scalar parameter form's, and why that form declares half is what
question 14 now asks.

**Tier 72: what each -debug mode writes.**
Tier 24 read the flag bytes of header words 14 and 15 off a design
with two drivers and nothing else, which left byte 2 of word 15
looking as though `line` set it by accident.
This tier repeats every mode on one design whose function has an
input and a local, so that what each mode writes is visible beside
what it flags.

| Case | `-debug` | Words 14, 15 | Differs from, and what it showed |
| :--- | :--- | :--- | :--- |
| `t72_dbg_typical_` | `typical` | `0x101`, `0x101` | `t11_sv_logic____`: the function's scope, none of its objects |
| `t72_dbg_wave____` | `wave` | `0x1`, `0x1` | `t72_dbg_typical_`: not even the scope |
| `t72_dbg_line____` | `wave line` | `0x1`, `0x10101` | `t72_dbg_wave____`: the scope and its three objects |
| `t72_dbg_subprog_` | `wave subprogram` | `0x1`, `0x10101` | `t72_dbg_line____`: the same file, to the byte |
| `t72_dbg_all_____` | `all` | `0x101`, `0x10101` | `t72_dbg_typical_`: drivers and the subprogram together |
| `t72_dbg_drivers_` | `wave drivers` | `0x101`, `0x1` | `t72_dbg_wave____`: one byte, and no scope |
| `t72_dbg_readers_` | `wave readers` | `0x10001`, `0x1` | `t72_dbg_drivers_`: byte 2 alone, so the bytes are independent |

So `line` and `subprogram` are one mode under two names, byte 1 of
word 15 is a subprogram's scope and byte 2 is its own declarations,
and question 21 is answered.
A narrow mode on its own writes no database: `-debug line` without
`wave` ends the run with `ERROR: [Simulator 45-10] The current
simulation was compiled without trace information`, which is why
every case pairs it with `wave`.

## Record which comparison produced which finding

A finding is only as good as the comparison behind it, and a comparison
that looked sound can turn out not to be. That already happened here:
the two byte difference between `t2_unsigned8` and `t2_signed8` was
recorded as evidence about type names when it was really the directory
names, and the only way to catch that was to know which pair the number
came from.

So the findings table in [format.md](format.md) has a **Found by**
column and a **Confirmed by** column, and every row fills both. Naming
the pair costs a few words and makes a wrong finding recoverable instead
of merely wrong.

Three shapes of comparison are in use, and they find different things:

* **A pair of cases differing in one axis.** Finds anything whose size or
  bytes move with that axis. Blind to fields that are the same in both.
* **Two runs of one case**, the noise mask. Finds the clocks, and
  nothing else, which is what makes it a mask.
* **The correlation sweep across all cases**, described in
  [format.md](format.md). Finds fields that are correct everywhere, and
  which a pairwise diff therefore never shows.

Reach for the third when a field is expected to exist but no pair moves
it.


## Working order

1. Run the noise experiment. Produce the mask.
2. Diff Tier 0 against Tier 1 baseline. Find the signal record.
3. Walk Tier 1 one axis at a time. Each answer is a row in the findings
   table in [format.md](format.md), with the command that
   reproduces it.
4. Only once the reader reproduces every `truth.json` in Tiers 0 and 1
   is it worth writing anything larger.
5. The reader now reproduces all 1097 cases through tier 72, and
   matches the VCD of every one of them, and of `//hdl/counter:sim`,
   `//hdl/uart:sim`, `//hdl/serv:sim` and `//hdl/potato:sim`, where
   the VCD holds anything.
   The next cases are the ones listed as not written yet.

A writer comes after a reader that works.
Round-tripping a database the reader understands is the test that the
writer is right.

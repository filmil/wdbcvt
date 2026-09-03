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
tools/noise_mask.sh //hdl/corpus/t1_bit_one_edge:sim /tmp/mask
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

Not written yet: a `string` value, which has no object to hold one,
see `t11_sv_str`; and a design from outside this repository.

Realistic designs come last, not first.
A FIFO or a UART is where the reader gets confirmed, not where it gets
discovered.

The intended end point is a real processor rather than a toy: **SERV**,
the bit serial RISC-V core, built through `rules_fusesoc`, which is
already a module in the registry. That gives a design nobody wrote for
this experiment, with a hierarchy, a register file, and a trace long
enough to cross whatever block boundaries the format has. It is the
test that the reader works on something it was not tuned against.

It comes after the ladder, not instead of it. A database from SERV
differs from the baseline in hundreds of ways at once, so on its own it
would teach nothing about which byte means what.


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
5. The reader now reproduces all 603 cases through tier 35, and
   matches the VCD of every one of them, and of `//hdl/counter:sim`
   and `//hdl/uart:sim`, where the VCD holds anything.
   The next cases are the ones listed as not written yet.

A writer comes after a reader that works.
Round-tripping a database the reader understands is the test that the
writer is right.

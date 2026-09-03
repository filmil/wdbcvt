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
* the Vivado library name is always `corpus`,
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

The `t9_vec*` sizes were chosen after the first three: 200, 256 and
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

Not written yet: a `string` value, which has no object to hold one,
see `t11_sv_str`; a typedef of an unpacked struct array; and a
realistic design.

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
5. The reader now reproduces all 256 cases through tier 16, and
   matches the VCD of every one of them where the VCD holds anything.
   The next cases are the ones listed as not written yet.

A writer comes after a reader that works.
Round-tripping a database the reader understands is the test that the
writer is right.

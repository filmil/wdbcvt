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

Not written yet, in order: a `linkage` port, to fill mode 4; a
generic of a type other than integer; and then the same ladder in
Verilog and SystemVerilog, to find out whether the source language
reaches the database at all.

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
5. The reader now reproduces all 79 cases through tier 8.
   The next cases are the ones listed as not written yet.

A writer comes after a reader that works.
Round-tripping a database the reader understands is the test that the
writer is right.

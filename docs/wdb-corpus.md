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
bazel build //hdl/corpus/t1_bit_one_edge:sim
cp bazel-bin/hdl/corpus/t1_bit_one_edge/sim.wdb /tmp/a.wdb
bazel clean --expunge_async     # or touch a source and rebuild
bazel build //hdl/corpus/t1_bit_one_edge:sim
cp bazel-bin/hdl/corpus/t1_bit_one_edge/sim.wdb /tmp/b.wdb
cmp -l /tmp/a.wdb /tmp/b.wdb
```

Every offset that differs is noise.
Record those offsets as a mask.
Apply the mask to every later comparison.

Skipping this step produces confident, wrong conclusions: a field is
declared to be "the signal count" because it changed, when it was the
run timestamp all along.

If the two files are byte for byte identical, say so explicitly.
That is a strong and useful property, not an absence of a result.


## Rule 2: hold everything constant except the axis

Every case in the corpus keeps these the same, so they cannot become
confounds:

* the top level entity is always named `tb`,
* the Vivado library name is always `corpus`,
* the simulation always ends at 100 ns through `std.env.stop`,
* the source files use the same header, the same libraries and the same
  VHDL standard.

What is left different between a pair is the one axis under test.
The Bazel target name differs too, and the output file is named after
it. Whether that name reaches inside the file is itself one of the first
questions to answer.


## Rule 3: every case declares its own ground truth

Each case directory holds a `truth.json` stating what the simulation
actually did: the signals, their widths and types, every transition with
its time and value, and the end time.

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

Tier 2 and beyond are not written yet. In order: wider vectors, 64 bits
and past it; integer, boolean, real, time and user enumerations; records
and multidimensional arrays; deeper and wider hierarchy, including
for-generate; transition counts large enough to cross a block boundary
and trigger compression; different timescales; simulation times large
enough to widen a timestamp; and then the same ladder in Verilog and
SystemVerilog, to find out whether the source language reaches the
database at all.

Realistic designs come last, not first.
A FIFO or a UART is where the reader gets confirmed, not where it gets
discovered.


## Working order

1. Run the noise experiment. Produce the mask.
2. Diff Tier 0 against Tier 1 baseline. Find the signal record.
3. Walk Tier 1 one axis at a time. Each answer is a row in the findings
   table in [wdb-format.md](wdb-format.md), with the command that
   reproduces it.
4. Only once the reader reproduces every `truth.json` in Tiers 0 and 1
   is it worth writing anything larger.

A writer comes after a reader that works.
Round-tripping a database the reader understands is the test that the
writer is right.

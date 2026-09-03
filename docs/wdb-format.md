<!-- SPDX-License-Identifier: Apache-2.0 -->

# Decoding the xsim waveform database (`.wdb`)


## What this document is

AMD does not document the `.wdb` container that `xsim` writes.
This file records what has actually been measured about it, and how.
Everything here is either a measurement or a statement marked as a guess.
Nothing gets promoted from guess to fact without a reproduction.


## Where the sample comes from

`//hdl/counter:sim` runs a small VHDL testbench under `xsim` and writes
its waveform database.

```sh
bazel build //hdl/counter:sim
ls -l bazel-bin/hdl/counter/sim.wdb
```

The same target also writes `sim.vcd`.
That pairing is the whole reason this design is the sample: VCD is a
documented text format holding the same signal history, so it acts as the
answer key.
Any decoder claim about the `.wdb` can be checked against the `.vcd`
written from the same run.

The design is deliberately small and its behaviour is fully known:
one 8-bit counter, one clock, one reset, one enable, one wrap flag, and a
run that covers exactly one full wrap of the counting range.
The value at every simulation time is therefore predictable without
reading either file.


## First pass: measure, do not guess

```sh
bazel run //cmd/wdbcvt -- -in "$PWD/bazel-bin/hdl/counter/sim.wdb"
```

`wdbcvt` reports four things, and no interpretation of them:

* the file size,
* the first 64 bytes as a hex dump,
* Shannon entropy per 4096-byte block, and over the whole file,
* every run of printable ASCII, longest first.

Each of those answers one question that has to be settled before any
parsing is attempted.

| Measurement | The question it settles |
| :--- | :--- |
| Header hex dump | Is there a magic number, and a version field next to it? |
| Whole-file entropy | Is the payload compressed? Near 8 bits per byte means yes, and means no field-level parsing will work until it is inflated. |
| Per-block entropy | Is the file part structured and part compressed? A low-entropy head followed by a high-entropy tail is the usual shape of a directory plus a compressed payload. |
| Printable runs | Do signal and scope names appear in the clear? If they do, the name table is the way into the structure, because the names are known from the VHDL source. |


## Findings

Every row is a measurement that reproduces.
All of them are from Vivado 2025.2, and are scoped to it.
See [provenance.md](provenance.md) for what guards these claims.

| Offset | Length | Meaning | How it was confirmed |
| :--- | :--- | :--- | :--- |
| `0x00` | 24 | ASCII `Xilinx WAVE DATABASE 01`, NUL terminated | `head -c 24 sim.wdb`. Present in all ten databases. |
| `0x18` | 17 | ASCII `Xilinx Simulator`, NUL terminated | Same hex dump. |
| `0x30` | 8 | `uint64` little endian, value `0x40` | Constant across all ten databases. |
| `0x38` | 4 | `uint32` little endian, Unix epoch seconds, the time the database was written | Decoded `1788417066` as `2026-09-03 08:31:06 CEST`, equal to the file's own mtime to the second. Differs between two runs of one design, which is how it was found. |
| `0x158` | 26 | ASCII `Xilinx ISim TYPE FILE 001`, NUL terminated | A nested section with its own magic. |

Whole file properties, also measured:

* **The payload is not compressed.** Mean entropy is 3.508 bits per byte
  over the whole file. A compressed payload sits near 8.
* **Signal names are stored in the clear, as plain ASCII.** The 40
  character name in `t1_bit_long_name` appears verbatim at `0x3db`.
* **Absolute source paths are stored in the clear**, including the path
  of the Vivado installation that produced the file and the build
  machine paths AMD compiled the standard libraries on.
* **Adding one transition grows the file by 15 bytes and perturbs 83
  places.** Many of the perturbed bytes change by exactly `+8`
  (`0x2b` to `0x33`, `0xeb` to `0xf3`, `0xc0` to `0xc8`), which is the
  signature of internal offset fields moving because a record grew.
  That, with the low entropy, says the format is a structured file full
  of internal offsets rather than an opaque blob.


## The noise mask

Two runs of `t1_bit_one_edge` differ in about 11 bytes across five
regions, and the regions are stable across pairs:

| Offset | Bytes | What it holds |
| :--- | :--- | :--- |
| `0x38` | 4 | Unix timestamp, confirmed above |
| `0xc4` | 4 | a per-run duration; `32194` and `28919` in two runs |
| `0x172` | 6 | two varying bytes followed by the same timestamp as `0x217` |
| `0x217` | 4 | a second Unix timestamp, a few seconds before `0x38` |
| `0x3c4` | 4 | a second per-run duration |

Every one of them is a clock or a duration.
Nothing else in the file varies between runs, which is a strong
property: **outside these five regions the format is deterministic.**

One caveat that matters. The two pairs measured so far reported 11 and
10 differing bytes at the same five regions. A byte of a timestamp only
shows up as noise when its two values happen to differ, so a mask built
from a single pair understates the true noise. Build the mask from
several pairs and take the union before trusting it.


## Open questions

1. Is `sim.wdb` a single file or a directory-shaped container?
   `xsim` also writes an `xsim.dir` tree, and the `vivado_view` rule
   opens the `.wdb` alongside it, which hints that the `.wdb` may not be
   self-contained.
2. Is the payload compressed, and with what?
   Check the entropy profile first, then look for `zlib` and `zstd`
   frame headers at the block boundaries the entropy profile suggests.
3. Do the signal names survive in the clear?
   The names to look for are known exactly: `tb`, `dut`, `ctl`, `stat`,
   `clk`, `reset`, `enable`, `value`, `wrapped`, `counter`, and
   `counter_types`.
4. How is time encoded?
   The testbench uses a 10 ns period and stops at a known time, so the
   largest timestamp in the file is known before it is found.
5. Does the format change between Vivado versions?
   Only 2025.2 is in use here. Any claim is version-scoped until a
   second version has been measured.


## What the conversion writes out

VCD first, through `github.com/filmil/go-vcd-parser`.
It already parses VCD, including the pragmatic extensions real simulators
emit, and it is a Bazel module, so it serves both jobs at once: it reads
the `sim.vcd` answer key, and it supplies the signal model that a `.wdb`
decoder fills in.
Add it as a `bazel_dep` at the point where the first decoder claim has to
be checked against the VCD, not before.
Until then the VCD can be read by eye, because the design is small.

FST is the better output format and comes later.
It is compressed, it holds large traces without the size blowup VCD
suffers, and the waveform viewers read it.
The reason to start with VCD anyway is that it is the format the answer
key is already in, so the first end to end check compares like with like.
Keep the decoder's output model separate from the VCD writer, so that
adding an FST writer later is a new writer and not a rewrite.


## Method

Work one open question at a time, and in this order: container shape,
compression, name table, time encoding, sample encoding.
Each answer becomes a row in the findings table plus a test in
`//pkg/wdb` built on a fixture, never on a file that has to be
regenerated by Vivado.
Where a claim can be cross-checked against `sim.vcd`, the test does that
check rather than asserting bytes.

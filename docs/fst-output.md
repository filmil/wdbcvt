<!-- SPDX-License-Identifier: Apache-2.0 -->

# Writing FST


## Why this document exists

VCD is the first output format, because the answer key is already a VCD.
FST is the one worth arriving at.
This records what was found about FST, and which route to take, so the
decision does not have to be made again under time pressure.


## What FST is

FST, for Fast Signal Trace, was written by Tony Bybell in 2014 as a
replacement for VCD in GTKWave.
It is a block format: a few metadata blocks, then value change blocks.
Every value change block holds the starting value of every variable, so
a block decodes on its own without reading the ones before it.
That is what makes seeking cheap, and it is the property VCD lacks.

GTKWave reads it, and so do Verilator, CVC, GHDL, nvc and Surfer.

There is no standard.
The nearest thing to a specification is `doc/block_format.txt` in the
libfst sources, plus an unofficial write-up by Tim Hutt.


## The licence question is settled

An open GTKWave issue asks whether the FST library is GPLv2, like
GTKWave, or MIT, as its file headers say.
The issue never got an answer, so it is worth stating the current
position plainly.

The library now lives in its own repository, `gtkwave/libfst`, and that
repository carries an **MIT** licence file, copyright Tony Bybell,
2009 to 2025.
GTKWave consumes it as a meson subproject.
The bundled compressors keep their own permissive licences: LZ4 is BSD
2-clause, FastLZ is MIT.

So libfst can be vendored into an Apache 2.0 project.
The stale GitHub issue is not evidence to the contrary.

The whole library is five source files: `fstapi.c`, `fstapi.h`,
`lz4.c`, `lz4.h`, `fastlz.c`, `fastlz.h`.


## Measured, not assumed

GHDL writes FST through libfst.
Running the `//hdl/counter` sources under GHDL produces both formats
from one simulation, which gives a real FST file for this exact design:

```sh
ghdl -a --std=08 counter.pkg.vhdl counter.ent.vhdl tb.ent.vhdl
ghdl -r --std=08 tb --fst=out.fst --vcd=out.vcd
```

| File | Bytes |
| :--- | ---: |
| `out.vcd` | 13050 |
| `out.fst` | 884 |

Roughly fifteen times smaller, on a trace this small.
The gap widens with trace length, which is the entire point of the
format.

The first bytes of that file:

```
00000000  fe 00 00 00 00 00 00 03  73 00 00 00 00 00 00 07
00000016  09 1f 8b 08 00 00 00 00  00 00 03 ed 95 79 74 13
```

Reading it against the block format: `fe` is block type 254, the
whole-file gzip wrapper. The two eight-byte big-endian fields are the
compressed length, 883, and the uncompressed length, 1801. Then `1f 8b`
begins the gzip stream. So a writer that repacks on close wraps the
entire file, and a reader has to unwrap before it sees any block.


## The four routes

| Route | What it costs | What it gives |
| :--- | :--- | :--- |
| Vendor libfst and bind it with cgo | cgo in the shipped binary, so cross compilation needs a C toolchain per target | The reference writer. Output is correct by construction. |
| Write the FST writer in Go | The block format has to be implemented | Stays pure Go. Cross compiles for free. No C toolchain. |
| Shell out to GTKWave's `vcd2fst` | GTKWave becomes a runtime dependency, and the data is walked twice | No new code at all |
| Use the Rust `fstapi` crate | Wrong language for this repository | Nothing this project needs |

## The route to take

**Write the writer in Go, and use libfst only to check it.**

Three things make the Go writer smaller than it sounds:

*   Only a writer is needed. Reading FST is somebody else's problem;
    this project reads `.wdb`.
*   A writer picks its own compression. Block payloads may be zlib, and
    `compress/zlib` is in the Go standard library. LZ4 and FastLZ are
    reader-side requirements, not writer-side ones, so neither has to be
    ported.
*   The whole-file gzip wrapper seen above is `compress/gzip`.

Keeping it pure Go matters because the release workflow cross compiles,
and because `rules_go` with cgo drags a C toolchain into every target
platform.

Checking it is the part that must not be skipped, and it needs no C in
the shipped binary:

*   GHDL and nvc both write FST through libfst, and both are already
    Bazel modules in the registry, as `rules_ghdl` and `rules_nvc`.
    Simulating the same `//hdl/counter` sources under one of them
    produces a golden FST for a design whose behaviour is known exactly.
*   Vendor libfst as a **test-only** dependency, build a small reader
    with `cc_binary`, and read back what the Go writer produced. The
    shipped `wdbcvt` stays pure Go; only the test needs a C toolchain.

Do not compare our output byte for byte against GHDL's.
Two correct writers make different legal choices about block boundaries
and compression. Compare decoded content.


## Order of work

1. Decode enough of `.wdb` to produce VCD. Check against `sim.vcd`.
2. Split the decoder's output model away from the VCD writer.
3. Add the FST writer against that model.
4. Add the libfst read-back test.

Step 2 is the one that is cheap now and expensive later.


## Sources

* FST file format, GTKWave documentation:
  https://gtkwave.github.io/gtkwave/internals/fst-file-format.html
* Unofficial FST specification, Tim Hutt:
  https://blog.timhutt.co.uk/fst_spec/
* libfst, MIT:
  https://github.com/gtkwave/libfst
* libfstwriter, a separate MIT C++14 writer:
  https://github.com/gtkwave/libfstwriter
* The unanswered licence question:
  https://github.com/gtkwave/gtkwave/issues/309

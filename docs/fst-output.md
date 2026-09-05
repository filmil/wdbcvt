<!-- SPDX-License-Identifier: Apache-2.0 -->

# Writing FST


## Why this document exists

FST is not a nicer output format to arrive at eventually.
**It is the only one of the two that can hold what the database holds**,
so it is required for the conversion to be correct at all.
This records what was found about FST, and which route to take.


## What VCD loses, measured

Vivado writes `sim.vcd` from the same run that writes `sim.wdb`.
For eight of the fifteen types in the corpus, that VCD holds no `$var`
and no value changes: `boolean`, `integer`, `real`, `time`,
`character`, user enumerations, records and arrays are simply absent.
Only `std_ulogic`, `bit` and the vector types survive.
The measurement is in [format.md](format.md).

A `.wdb` to VCD converter would therefore drop most of a real design's
signals, and drop them **silently**, because VCD has nowhere to put
them and no way to say so.

FST has somewhere to put them. From `fstapi.h`:

| Need | FST provides |
| :--- | :--- |
| integers | `FST_VT_VCD_INTEGER`, `FST_VT_SV_INT` and the sized integer types |
| reals | `FST_VT_VCD_REAL`, `FST_VT_SV_SHORTREAL` |
| time | `FST_VT_VCD_TIME`, `FST_VT_VCD_REALTIME` |
| strings and characters | `FST_VT_GEN_STRING`, variable length |
| enumerations with their literal names | `FST_VT_SV_ENUM`, plus `fstWriterCreateEnumTable` and `fstWriterEmitEnumTableRef` |

That last row matters more than it looks. The database stores
enumeration literals as text, `'U' 'X' '0' '1' 'Z' 'W' 'L' 'H' '-'` for
`std_ulogic` and `crimson viridian cobalt` for a user type. FST can
carry those names through to the viewer. VCD cannot.


## The order of work, revised

VCD still comes first, but for a smaller reason than before: it is the
format the answer key is already in, so the first end to end check of
the decoder compares like with like on bit and vector signals.

It is a checking step, not the deliverable.
The deliverable is FST, because a converter that silently drops every
integer, real, enumeration and record in a design is not a converter
anyone should use.


## What FST is

FST, for Fast Signal Trace, was written by Tony Bybell in 2014 as a
replacement for VCD in GTKWave.
It is a block format: a few metadata blocks, then value change blocks.
Every value change block holds the starting value of every variable, so
a block decodes on its own without reading the ones before it.
That is what makes seeking cheap, and it is the property VCD lacks.

GTKWave reads it, and so do Verilator, CVC, GHDL, nvc and Surfer.

There is no standard.
The nearest thing to a specification is `docs/block_format.txt` in the
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


## Measured, not assumed, before the route was chosen

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
| Vendor libfst and bind it with cgo | cgo in the shipped binary, so the build needs a C toolchain | The reference writer. Output is correct by construction. |
| Write the FST writer in Go | The block format has to be implemented, and kept working | Stays pure Go. Cross compiles for free. No C toolchain. |
| Shell out to GTKWave's `vcd2fst` | GTKWave becomes a runtime dependency, and the data is walked twice | No new code at all |
| Use the Rust `fstapi` crate | Wrong language for this repository | Nothing this project needs |


## The route taken: libfst through cgo

The first two rows were weighed again, and the first won.
The argument for the Go writer was that cgo drags a C toolchain into
every target platform and that the release cross compiles.
The release does not cross compile: `.forgejo/workflows/release.yml`
builds one artifact, `wdbcvt-linux-amd64`.
And the C toolchain is a `bazel_dep` away, so it is not the host's.

The argument against the Go writer is stronger, and it is the reason
this document exists at all: FST has no specification, so a second
writer is a second thing to maintain against a moving definition, and
nothing here would ever tell us that it had drifted.
libfst is the definition.

**Measured, by building it.** The library is three source files,
`fstapi.c` at 7051 lines, `lz4.c` at 2829 and `fastlz.c` at 549, plus
zlib. The upstream build generates a `config.h` from two function
probes and a thread switch; leaving `FST_INCLUDE_CONFIG` undefined
skips the include, and `-D_GNU_SOURCE -DHAVE_FSEEKO -DHAVE_REALPATH`
is the whole configuration. `third_party/libfst/libfst.BUILD` is a
`cc_library` of those files, and the archive is pinned by hash in
`MODULE.bazel`.

Two things had to be settled to get there, and both are recorded
because they cost time:

*   Bazel 9 removed the built in `cc_library`, so the build file loads
    it from `rules_cc`.
*   The host has `gcc` but no libstdc++ development files, and the cgo
    link step asks for `-lstdc++`:
    `/usr/bin/ld: cannot find -lstdc++`. A `cc_library` also builds a
    shared object by default, whose link fails the same way.
    `hermetic_cc_toolchain`, the zig toolchain, answers both, and it is
    pinned to a glibc old enough for the release. `linkstatic` keeps
    the shared object from being built at all.

`pkg/fst` is the binding: ten writer calls behind a Go API, and a
reader used only by tests. `pkg/fst:fst_test` writes a small file,
reads it back through libfst and checks the times, the variables,
their scope paths and their widths.


## What the converter writes

`bazel run //cmd/wdbcvt -- -in <file>.wdb -fst <file>.fst` converts a
database. What follows is the mapping it applies; `pkg/fstout` is the
code, and `pkg/fstout:fstout_test` checks it by reading the output back
through libfst and comparing against every case's `truth.json`.

**Scopes.** The instance tree of the database becomes the scope tree of
the file, one `fstWriterSetScope` per scope in preorder.

**Objects that share a handle.** Two objects on one handle are one
signal seen from two scopes, which FST calls an alias: the second
`fstWriterCreateVar` passes the first one's handle and no value change
is written twice. That is the same relationship the VCD expresses by
giving two variables one identifier code, and the reader already knows
which objects share a handle.

**Records and arrays are flattened.** FST has no record type and no
array type. A record becomes one variable per leaf field, in a scope
named after the record, recursively, which is what `File.Leaves`
already produces for the truth files: `s.delta_f.bravo` is
`s` a scope, `delta_f` a scope and `bravo` a variable. An array
becomes one variable per element, in a scope named after the array,
with the element's index as the variable name.
The alternative was one `FST_VT_GEN_STRING` per record holding the
whole aggregate as text. Flattening was chosen because a viewer can
then show, search and compare a field on its own, which is the reason
a waveform viewer is opened at all, and because the aggregate can
always be read back from the leaves while the reverse is not true.

**Scalars.** `std_ulogic`, `bit`, Verilog `logic` and their vectors
are wires, told apart by the class word of the enumeration entry, and
their values are the nine state characters FST already carries, in
lower case. An integer is `FST_VT_VCD_INTEGER` of 32 bits holding
two's complement, a real is `FST_VT_VCD_REAL`, which libfst stores as
the eight bytes of the double, and a `character` or a `string` is
`FST_VT_GEN_STRING`.

**Enumerations.** A user enumeration is written as
`FST_VT_GEN_STRING` holding the literal name, which every reader
displays. `fstWriterCreateEnumTable` and `FST_VT_SV_ENUM` are the
format's own way to say the same thing, and are the better answer once
a reader that uses the table matters; the model keeps the literal
either way, so the choice can change without touching the decoder.

**Physical types.** A physical value is a count and the unit it
counts, `3000 um`, and FST has no type for that pair, so it goes as a
string. `FST_VT_VCD_TIME` would hold the count and lose the unit.

**Names.** An extended identifier keeps its bars in the database,
`\g(0)\` for a VHDL generate iteration and `\g[0].r ` for a Verilog
variable declared in a generate block. The converter strips the
decoration, so a viewer shows `g(0)` and `g[0].r`.

**Times.** The file's time unit is a power of ten of a second, which
is what FST calls a timescale, so the times go over unchanged. The
last time written is the simulation end time, so a viewer shows the
run's full length and not the last change.


## What it costs

Measured on the machine that built it, `bazel build -c fastbuild`:

| Design | Objects | Changes | `.wdb` | `.vcd` | `.fst` | Time | Peak RSS |
| :--- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `//hdl/counter:sim` | 14 | 2388 | 14423 | 5560 | 1566 | 0.02 s | 7 MB |
| `//hdl/uart:sim` | 59 | 16718 | 44569 | 3259 | 2311 | 0.03 s | 10 MB |
| `//hdl/picorv32:sim` | 280 | 40589 | 268032 | 286721 | 13737 | 0.25 s | 18 MB |
| `//hdl/potato:sim` | 557 | 7441 | 197470 | 28448 | 38074 | 0.06 s | 18 MB |
| `//hdl/serv:sim` | 968 | 9233899 | 15.9 MB | 29.6 MB | 653905 | 8.8 s | 591 MB |
| `//hdl/neorv32:sim` | 5696 | 18875466 | 24.1 MB | 17.2 MB | 690347 | 71 s | 1.08 GB |

On the two long traces the FST is 45 times smaller than the VCD of the
same run, and 25 times, while holding signals the VCD leaves out
entirely. On `//hdl/potato:sim` it is larger than the VCD, and that is
the same fact from the other side: the VCD holds none of that design's
integers, enumerations, records or its two 32768 byte memories, and
the FST holds all of them.

The memory is the whole database: `ReadFile` inflates every page and
keeps it. The conversion itself streams, and opening a file now
decodes every object once through Stream rather than once per object,
which took the peak of `//hdl/neorv32:sim` from 1.5 GB to 1.08 GB.
Reading the pages lazily is the next thing to do if a database is ever
too large for memory.

The time is dominated by decoding a value per change and spelling its
leaves, about 3.8 us per change.

## Every database has a conversion test

`bazel test //...` converts every waveform database in the repository
and fails if any conversion does. Each corpus case gets a `:convert`
target from `wdb_case`, and each design outside the corpus has one of
its own, whose only question is whether that database converts at all,
to FST and to the SQLite output of
[sqlite-output.md](sqlite-output.md).

They say less than `//pkg/fstout:fstout_test`, which reads the output
back through libfst and compares it against the case's truth file, and
they say it per database: when a change breaks conversion for one
design, the name of the failing test is the name of that design.

## What is left to build

*   Spelling straight from the record bytes for a wire or a vector,
    which is where the 3.8 us per change goes; the value tree is built
    for every change and thrown away.
*   `fstWriterCreateEnumTable` for user enumerations, once a reader
    that uses the table matters.
*   An unlogged object, a VHDL generic or constant, is left out. The
    database holds no records for one, so there is nothing to write
    over time; a viewer would still show it as a constant, and FST has
    no place for that but a one change variable.

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

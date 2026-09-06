<!-- SPDX-License-Identifier: Apache-2.0 -->

[![Build and test](https://git.hdlfactory.com/HDL/wdbcvt/actions/workflows/test.yml/badge.svg)](https://git.hdlfactory.com/HDL/wdbcvt/actions?workflow=test.yml)
[![Release](https://git.hdlfactory.com/HDL/wdbcvt/actions/workflows/release.yml/badge.svg)](https://git.hdlfactory.com/HDL/wdbcvt/actions?workflow=release.yml)

# wdbcvt

Tools for reading the Vivado `xsim` waveform database (`.wdb`).

Based on the fraction of understood data sematics, **it should
support 95%+ of WDB features** as of this writing.

The WDB format is not publicly documented by AMD.
This repository is an independent study of the WDB format, resulting in
a program which is able to read a wide range of WDB files, and write
FST files.

See [docs/format.md](docs/format.md) for what has been measured so
far and how the work proceeds.

That 95% counts one kind of file: behavioural VHDL, Verilog and
SystemVerilog, simulated by Vivado 2025.2.
Not supported:

* gate level netlists with SDF back annotation, untried
* encrypted IP models, untried
* UPF power intent, untried
* more than one top in an elaboration, untried
* any other Vivado version, untried
* values of SystemVerilog strings, queues, dynamic arrays,
  associative arrays and class handles: the file holds none
* `trireg` and `let`: `xsim` rejects both
* writing a database: this tool only reads


## Where this lives

Development is on [git.hdlfactory.com/HDL/wdbcvt][forge], and so is the
CI: every corpus case is a Vivado simulation, and the runner that can
do that answers to that forge.

`main` is mirrored, read only, to
[github.com/filmil/wdbcvt][mirror], so the work can be read and linked
from a place people already look.
Issues and pull requests belong on the forge; the mirror has its issue
tracker turned off so that a report there cannot go unread.
The push is made by a deploy key that can write to the mirror and
nothing else.

[forge]: https://git.hdlfactory.com/HDL/wdbcvt
[mirror]: https://github.com/filmil/wdbcvt


## How this format knowledge was obtained

**This is an AI-first exploration.**
The format was worked out by an AI agent running experiments against
`.wdb` files and reading the bytes that came out.
It is not a port of AMD code, not a decompilation, and not a
specification.

An agent that infers a format from examples produces measurements that
reproduce and plausible stories about bytes, and the two look identical
on the page.
So nothing here rests on the derivation being trustworthy.
Every claim is guarded by software this project did not write, and by
references whose content is known before a `.wdb` is opened:

* the `sim.vcd` that Vivado writes from the same simulation run, in the
  IEEE 1800 text format,
* `github.com/filmil/go-vcd-parser`, an existing parser with its own
  tests, which reads that answer key,
* a `truth.json` per corpus case, derived from the design rather than
  from the database,
* GHDL and nvc, simulators that share no code with Vivado,
* `libfst`, the reference implementation, which writes the FST output
  and reads it back for the check.

Read [docs/provenance.md](docs/provenance.md) before relying on
`dewdb` for anything. It states what the guards do and do not cover, and
where the tool should not be used.


## Layout

* `hdl/counter/` holds a small VHDL design and its testbench.
  Simulating it produces the reference `sim.wdb`, and a `sim.vcd` from the
  same run that acts as the answer key.
* `hdl/uart/` holds a larger VHDL design, a UART looped back into a
  FIFO, that confirms the reader on a hierarchy not written to ask one
  question.
* `hdl/serv/` holds a Verilog bench around SERV, the bit serial RISC-V
  core, running its `hello_uart` program: a design nobody wrote for
  this repository.
  `third_party/serv/` holds the build file and the patch for the
  pinned SERV archive fetched in `MODULE.bazel`.
* `hdl/potato/` holds the VHDL counterpart: Potato, a RV32I processor
  in VHDL, under its own bench and a hand assembled program.
  `third_party/potato/` holds the build file and the patch for the
  pinned Potato archive.
* `hdl/picorv32/` holds PicoRV32 under the project's own
  `testbench_ez.v`, which holds the program it runs, so nothing but
  the build file in `third_party/picorv32/` is written here.
* `hdl/ibex/` holds Ibex under the `simple_system` example lowRISC
  ships with it, a SystemVerilog core with a bus, a memory and a timer.
  Only the bench and a hand assembled program are written here;
  `third_party/ibex/` holds the build file and the patch for the
  pinned archive.
* `hdl/neorv32/` holds NEORV32, a dual core RISC-V processor in VHDL,
  under the project's own testbench, booting the instruction memory
  image of the release.
  Only the wrapper that stops the run is written here;
  `third_party/neorv32/` holds the build file for the pinned archive.
* `cmd/wdbcvt/` is the command line tool.
  Today it probes a `.wdb` and reports measurements; it grows into a
  converter as the format becomes known.
  VCD comes first as a checking step, through
  `github.com/filmil/go-vcd-parser`, but the deliverable is FST.
  VCD cannot represent integers, reals, enumerations, records or arrays,
  so a WDB to VCD converter drops most of a real design silently.
  [docs/fst-output.md](docs/fst-output.md) records the measurement and
  the plan.
* `pkg/wdb/` holds the library the tool is built on.
* `pkg/fst/` writes FST through libfst, the reader and writer GTKWave
  uses, over cgo.
  FST has no specification, so the library is the definition of the
  format and this project does not keep a second writer.
  `third_party/libfst/` holds the build file for the pinned archive.
* `pkg/fstout/` maps a decoded database onto FST variables: what a
  record or an array flattens into, and how each value is spelled.
  `wdbcvt -in <file>.wdb -fst <file>.fst` writes one.
* `pkg/sqlout/` writes the same decoded database as an SQLite file, in
  the schema `go-vcd-parser` writes from a VCD, so a query written
  against one reads the other.
  `wdbcvt -in <file>.wdb -sqlite <file>.db` writes one; see
  [docs/sqlite-output.md](docs/sqlite-output.md).
  Every database in the repository has a `:convert` test that writes
  both outputs, so a conversion that breaks names the design it broke
  on.
* `docs/latex/` holds the report on the whole effort, in the IEEEtran
  class, built by `bazel build //docs/latex:report` and attached to
  every release as `wdbcvt-report.pdf`.
* `docs/` holds everything known about the format, written down as it is
  discovered. [docs/README.md](docs/README.md) is the index;
  [docs/format.md](docs/format.md) is the findings table.


## Building

Everything is built with Bazel.
The pinned version is in `.bazelversion`, so `bazelisk` picks it up on its
own.

```sh
bazel build //...
bazel test //...
```

Go is the hermetic toolchain: the SDK is fetched by `rules_go` from the
version in `go.mod`, and nothing needs Go installed on the machine.
Do not run `go` directly; run it through Bazel instead.

```sh
bazel run @rules_go//go -- mod tidy
bazel run //:gazelle      # regenerate the Go BUILD files
bazel run //:buildifier   # format the Bazel files
```


## Vivado

The Vivado rules come from
[`rules_vivado`](https://github.com/filmil/bazel_rules_vivado), and this
repository runs them in **hermetic** mode:

```
build --@rules_vivado//:vivado_mode=hermetic
```

Bazel installs Vivado itself, from the AMD unified installer archive named
in `MODULE.bazel`.
No Docker image and no host Vivado installation are involved.

Two things make that practical rather than a multi-hour tax on every
build:

* The installation lands in the **shared install cache**
  `/data/cache/vivado-install`, set through the `install_cache` attribute
  of the `vivado.install` tag.
  Every workspace and every user on the machine reuses the one
  installation, and `bazel clean --expunge` does not remove it.
  The install happens once per host, not once per checkout.
* `/data/cache/ci.bazelrc`, pulled in by `try-import`, adds the shared
  Bazel disk cache and repository cache, so build outputs are shared with
  the CI runner as well.
  On a machine without that file the `try-import` does nothing.

Bazel 9.2.0 or later is required, and `.bazelversion` pins it.
Earlier versions crash on the `file://` URL that names the installer
archive; see `AGENTS.md` for the detail.

Budget for a cold install. Measured once, on this machine, with the
archive on the same disk as the output base:

| Phase | Reached at |
| :--- | ---: |
| Copy the archive, write it to the shared repository cache, unpack it | 112 min |
| Batch install begins | 121 min |

Peak disk during the fetch was about 290 GB, because the archive is
moved three times: copied in, written to the repository cache, and
unpacked. The archive is deleted once unpacking finishes, returning
roughly 96 GB. This happens once per host, not once per workspace.

A host that has never built this repository needs:

* the installer archive at
  `/data/tools/archives/FPGAs_AdaptiveSoCs_Unified_SDI_2025.2_1114_2157_1.tar`
  (adjust `urls` in `MODULE.bazel` if it lives elsewhere), and
* enough transient disk space for the extraction, roughly 200 GB.

Only the `Artix-7` device family is installed.
Simulation needs no device family at all, and one small family keeps the
installation from ballooning.


## Simulating

```sh
bazel build //hdl/counter:sim
ls -l bazel-bin/hdl/counter/sim.wdb bazel-bin/hdl/counter/sim.vcd
bazel run //cmd/wdbcvt -- -in "$PWD/bazel-bin/hdl/counter/sim.wdb"
```


## CI

Three workflows live in `.forgejo/workflows/`, and all of them use
`runs-on: vivado`, because building anything here means running Vivado:

| Workflow | Trigger | What it does |
| :--- | :--- | :--- |
| `test.yml` | pull request, push to `main`, weekly | `bazel build //...`, then `bazel test //...`, then checks that the simulation wrote a non-empty `.wdb` |
| `release.yml` | manual dispatch, monthly | publishes the `wdbcvt` binaries for Linux amd64 and arm64 and for macOS, the report, and a reference `.wdb` and `.vcd` to the rolling `nightly` release, here and on the GitHub mirror; a run stops early when the tag already names the commit |
| `mirror.yml` | push to `main`, daily, on request | pushes `main` to the read-only GitHub mirror |

The `vivado` runner must have `bazelisk` on its `PATH`, the installer
archive at the path above, and write access to `/data/cache`.

Two repository secrets, and each job runs without its own:

| Secret | Used by | What it is |
| :--- | :--- | :--- |
| `GH_MIRROR_KEY` | `mirror.yml` | the private half of a deploy key the mirror accepts for writing |
| `GH_RELEASE_TOKEN` | `release.yml` | a fine grained GitHub token with `Contents: read and write` on the mirror, for the release it copies there |

A job whose secret is missing writes a warning and stops, rather than
failing.

That label runs jobs directly on the host, and the host has no `node`.
No JavaScript action can run there, `actions/checkout` included.
Every workflow here checks out with plain `git` for that reason.
An action is safe only when every step of it, and of anything it
`uses`, is bash.
Upstream `forgejo-release` is not: it embeds a node cache action behind
a guard that never skips, so the release workflow uses the vendored,
bash only copy in `.forgejo/actions/forgejo-release`.

The non amd64 release binaries are cross compiled through the zig based
hermetic C toolchain.
Vivado exists only for Linux amd64, so those platforms build the Go
binary alone, and no tests run for them.


## License

Apache 2.0. See [LICENSE](LICENSE).


## Prior Art

* https://github.com/Ross0907/Vivado-WDB-Waveform-Converter: direct inspiration to revive some prior work.
* https://adaptivesupport.amd.com/s/question/0D52E00006hpSPUSA2/export-xsim-waveform-wdb-to-other-format: one of
  many similar threads where people lament the lack of a converter tool

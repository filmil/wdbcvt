<!-- SPDX-License-Identifier: Apache-2.0 -->

[![Build and test](https://git.hdlfactory.com/HDL/wdbcvt/actions/workflows/test.yml/badge.svg)](https://git.hdlfactory.com/HDL/wdbcvt/actions?workflow=test.yml)
[![Release](https://git.hdlfactory.com/HDL/wdbcvt/actions/workflows/release.yml/badge.svg)](https://git.hdlfactory.com/HDL/wdbcvt/actions?workflow=release.yml)

# wdbcvt

Tools for reading the Vivado `xsim` waveform database (`.wdb`).

By measured data semantics, **wdbcvt supports over 95% of WDB features**
as of this writing.

AMD publishes no documentation for the WDB format.
This repository studies the format independently.
It provides tooling that parses `.wdb` files and converts them into FST and
SQLite formats.

See [docs/format.md](docs/format.md) for what has been measured so far and how
the work proceeds.

That 95% metric applies to one class of files: behavioural VHDL, Verilog, and
SystemVerilog, simulated by Vivado 2025.2.
What is not supported:

* gate level netlists with SDF back annotation: untried
* encrypted IP models: untried
* UPF power intent: untried
* more than one top in an elaboration: untried
* any other Vivado version: untried
* values of SystemVerilog strings, queues, dynamic arrays,
  associative arrays, and class handles: the file holds none
* `trireg` and `let`: `xsim` rejects both
* writing a database: this tool only reads


## Where this lives

Development happens on [git.hdlfactory.com/HDL/wdbcvt][forge].
Continuous integration runs on this forge as well: every corpus case requires a
Vivado simulation, and the runner capable of executing Vivado answers to this
forge.

The `main` branch is mirrored read-only to
[github.com/filmil/wdbcvt][mirror], so the repository can be read and linked
from GitHub.
Issues and pull requests belong on the forge.
The GitHub mirror disables its issue tracker so that bug reports do not go
unread.
A dedicated deploy key with scoped write access pushes commits to the mirror.

[forge]: https://git.hdlfactory.com/HDL/wdbcvt
[mirror]: https://github.com/filmil/wdbcvt


## How this format knowledge was obtained

**This is an AI-first exploration.**
An AI agent derived the format layout by running experiments against `.wdb`
files and reading the emitted bytes.
The result is not a port of AMD code, not a decompilation, and not a
specification.

An agent that infers a format from examples produces two kinds of output:
measurements that reproduce, and plausible stories about bytes.
The two look identical on the page.
Nothing here rests on assuming the derivation is trustworthy.
Every claim is guarded by software this project did not write, and by
references whose content is known before a `.wdb` file is opened:

* the `sim.vcd` file that Vivado writes from the same simulation run, in IEEE
  1800 text format,
* `github.com/filmil/go-vcd-parser`, an existing parser with independent
  tests that reads the answer key,
* a `truth.json` file per corpus case, derived directly from the design rather
  than the database,
* GHDL and nvc, open source simulators that share no code with Vivado,
* `libfst`, the reference implementation that writes FST output and reads it
  back for verification.

Read [docs/provenance.md](docs/provenance.md) before relying on `dewdb` or
`wdbcvt` for anything.
It states what the verification guards cover and what they do not, and explains
where the tool should not be used.


## Layout

* `hdl/counter/` holds a small VHDL design and its testbench.
  Simulating it produces the reference `sim.wdb`, and a `sim.vcd` from the
  same run that acts as the answer key.
* `hdl/uart/` holds a larger VHDL design, a UART looped back into a
  FIFO, that validates the reader on a hierarchy not tailored to a single test.
* `hdl/serv/` holds a Verilog testbench around SERV, the bit-serial RISC-V
  core, executing its `hello_uart` program: an external design not written for
  this repository.
  `third_party/serv/` holds the build file and patch for the pinned SERV
  archive fetched in `MODULE.bazel`.
* `hdl/potato/` holds the VHDL counterpart: Potato, an RV32I processor in
  VHDL, under its own testbench and a hand-assembled program.
  `third_party/potato/` holds the build file and patch for the pinned Potato
  archive.
* `hdl/picorv32/` holds PicoRV32 under the project's own `testbench_ez.v`,
  which provides the execution program, so only the build file in
  `third_party/picorv32/` is written here.
* `hdl/ibex/` holds Ibex under the `simple_system` example lowRISC ships with
  it: a SystemVerilog core with a bus, memory, and timer.
  Only the testbench and a hand-assembled test program are written here;
  `third_party/ibex/` holds the build file and patch for the pinned archive.
* `hdl/neorv32/` holds NEORV32, a dual-core RISC-V processor in VHDL, under
  the project's own testbench, booting the upstream release instruction image.
  Only the wrapper that terminates the simulation is written here;
  `third_party/neorv32/` holds the build file for the pinned archive.
* `cmd/wdbcvt/` is the command line converter.
  It reads a `.wdb` file, extracts signals and transitions, and writes FST or
  SQLite output files.
  VCD verification runs through `github.com/filmil/go-vcd-parser` as a check,
  while FST is the primary output format.
  Because VCD cannot represent integers, reals, enumerations, records, or
  arrays, converting WDB to VCD drops most types in real designs silently.
  [docs/fst-output.md](docs/fst-output.md) records the measurement and the
  conversion plan.
* `pkg/wdb/` holds the library the tool is built on.
* `pkg/fst/` writes FST through `libfst`, the reader and writer GTKWave uses,
  over cgo.
  FST has no written specification, so the reference library defines the
  format and this project does not maintain a separate writer.
  `third_party/libfst/` holds the build file for the pinned archive.
* `pkg/fstout/` maps a decoded database onto FST variables: what a record or
  an array flattens into, and how each value is formatted.
  Running `wdbcvt -in <file>.wdb -fst <file>.fst` writes one.
* `pkg/sqlout/` writes the decoded database as an SQLite file using the schema
  `go-vcd-parser` generates from VCD, so queries written against one format
  read the other.
  Running `wdbcvt -in <file>.wdb -sqlite <file>.db` writes one; see
  [docs/sqlite-output.md](docs/sqlite-output.md).
  Every database in the repository has a `:convert` test that writes both
  outputs, so any conversion regression names the exact design that failed.
* `docs/latex/` holds the report on the exploration in IEEEtran format,
  built with `bazel build //docs/latex:report` and attached to every release as
  `wdbcvt-report.pdf`.
* `docs/` holds findings about the format, recorded as they are discovered.
  [docs/README.md](docs/README.md) is the index;
  [docs/format.md](docs/format.md) is the findings table.


## Building

Bazel builds everything in this repository.
The pinned version is recorded in `.bazelversion`, so `bazelisk` selects it
automatically.

```sh
bazel build //...
bazel test //...
```

Go is managed through a hermetic toolchain.
The Go SDK is fetched by `rules_go` matching the version pinned in `go.mod`.
Nothing requires Go installed on the host machine.
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

Bazel installs Vivado directly from the AMD unified installer archive named
in `MODULE.bazel`.
No Docker image and no host Vivado installation are involved.

Two mechanisms make this setup practical rather than an expensive tax on every
build:

* The installation populates the **shared install cache** at
  `/data/cache/vivado-install`, configured through the `install_cache`
  attribute of the `vivado.install` tag.
  Every workspace and every user on the host reuses the single installation,
  and `bazel clean --expunge` does not remove it.
  The installation occurs once per host rather than once per checkout.
* `/data/cache/ci.bazelrc`, imported by `try-import`, attaches the shared
  Bazel disk cache and repository cache, sharing build outputs with the CI
  runner.
  On a host without that configuration file, the `try-import` does nothing.

Bazel 9.2.0 or later is required, and `.bazelversion` pins it.
Earlier versions fail on the `file://` URL that references the installer
archive; see `AGENTS.md` for details.

Budget for a cold installation.
Measured once on this machine, with the archive located on the same disk as
the output base:

| Phase | Reached at |
| :--- | ---: |
| Copy archive, write to shared repository cache, unpack | 112 min |
| Batch install begins | 121 min |

Peak disk usage during the fetch reached roughly 290 GB because the archive is
handled three times: copied into the workspace, written to the repository cache,
and unpacked.
The archive is deleted once unpacking completes, returning roughly 96 GB.
This occurs once per host rather than once per workspace.

A host that has never built this repository requires:

* the installer archive at
  `/data/tools/archives/FPGAs_AdaptiveSoCs_Unified_SDI_2025.2_1114_2157_1.tar`
  (adjust `urls` in `MODULE.bazel` if located elsewhere), and
* sufficient transient disk space for extraction, roughly 200 GB.

Only the `Artix-7` device family is installed.
Simulation requires no device family at all, and one small family prevents the
installation footprint from expanding needlessly.


## Simulating

```sh
bazel build //hdl/counter:sim
ls -l bazel-bin/hdl/counter/sim.wdb bazel-bin/hdl/counter/sim.vcd
bazel run //cmd/wdbcvt -- -in "$PWD/bazel-bin/hdl/counter/sim.wdb"
```


## CI

Three workflows live in `.forgejo/workflows/`.
All workflows specify `runs-on: vivado`, because building and testing in this
repository requires Vivado:

| Workflow | Trigger | What it does |
| :--- | :--- | :--- |
| `test.yml` | pull request, push to `main`, weekly | `bazel build //...`, then `bazel test //...`, then checks that the simulation wrote a non-empty `.wdb` |
| `release.yml` | manual dispatch, daily | publishes the `wdbcvt` binaries for Linux amd64 and arm64 and for macOS, the report, the waveform and documentation archives, and a reference `.wdb` and `.vcd` to the rolling `nightly` release, here and on the GitHub mirror; a run stops early when the tag already names the commit |
| `mirror.yml` | push to `main`, daily, on request | pushes `main` to the read-only GitHub mirror |

The `vivado` runner host requires `bazelisk` on its `PATH`, the installer
archive at the path documented above, and write permissions for `/data/cache`.

Two repository secrets configure external integrations:

| Secret | Used by | What it is |
| :--- | :--- | :--- |
| `GH_MIRROR_KEY` | `mirror.yml` | the private half of a deploy key the mirror accepts for writing |
| `GH_RELEASE_TOKEN` | `release.yml` | a fine grained GitHub token with `Contents: read and write` on the mirror, for the release it copies there |

A job whose secret is missing writes a warning and stops, rather than failing.

The `runs-on: vivado` runner executes jobs directly on the host, which has no
`node` installed.
No JavaScript action can run there, including `actions/checkout`.
Every workflow checks out code with plain `git` for that reason.
An action is safe only when every step of it, and of anything it `uses`,
is bash.
Upstream `forgejo-release` is not: it embeds a node cache action behind a guard
that never skips, so the release workflow uses the vendored, bash-only copy in
`.forgejo/actions/forgejo-release`.

The non-amd64 release binaries are cross-compiled using the zig-based hermetic
C toolchain.
Vivado exists only for Linux amd64, so non-amd64 targets compile the Go binary
alone, and run no simulation tests.


## License

Apache 2.0. See [LICENSE](LICENSE).


## Prior Art

* [Vivado-WDB-Waveform-Converter][ross0907]:
  direct inspiration to revive earlier reverse-engineering efforts.
* [AMD Support discussion][amd-thread]:
  one of many community threads where users lament the lack of a converter tool.

[ross0907]: https://github.com/Ross0907/Vivado-WDB-Waveform-Converter
[amd-thread]: https://adaptivesupport.amd.com/s/question/0D52E00006hpSPUSA2/export-xsim-waveform-wdb-to-other-format

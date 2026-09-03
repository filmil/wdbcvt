<!-- SPDX-License-Identifier: Apache-2.0 -->

[![Build](https://git.hdlfactory.com/HDL/wdbcvt/actions/workflows/build.yml/badge.svg)](https://git.hdlfactory.com/HDL/wdbcvt/actions?workflow=build.yml)
[![Test](https://git.hdlfactory.com/HDL/wdbcvt/actions/workflows/test.yml/badge.svg)](https://git.hdlfactory.com/HDL/wdbcvt/actions?workflow=test.yml)
[![Release](https://git.hdlfactory.com/HDL/wdbcvt/actions/workflows/release.yml/badge.svg)](https://git.hdlfactory.com/HDL/wdbcvt/actions?workflow=release.yml)

# wdbcvt

Tools for reading the Vivado `xsim` waveform database (`.wdb`).

AMD does not document that format.
This repository builds a known-good sample of one, and the tooling used to
take it apart.
See [docs/format.md](docs/format.md) for what has been measured so
far and how the work proceeds.


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
* `libfst`, the reference implementation, for checking FST output.

Read [docs/provenance.md](docs/provenance.md) before relying on
`dewdb` for anything. It states what the guards do and do not cover, and
where the tool should not be used.


## Layout

* `hdl/counter/` holds a small VHDL design and its testbench.
  Simulating it produces the reference `sim.wdb`, and a `sim.vcd` from the
  same run that acts as the answer key.
* `cmd/wdbcvt/` is the command line tool.
  Today it probes a `.wdb` and reports measurements; it grows into a
  converter as the format becomes known.
  The first output format is VCD, written through
  `github.com/filmil/go-vcd-parser`.
  FST comes later; [docs/fst-output.md](docs/fst-output.md) records how.
* `pkg/wdb/` holds the library the tool is built on.
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
| `build.yml` | pull request, push to `main`, weekly | `bazel build //...` |
| `test.yml` | pull request, push to `main`, weekly | `bazel test //...`, then checks that the simulation wrote a non-empty `.wdb` |
| `release.yml` | manual dispatch, monthly | publishes the `wdbcvt` binary and a reference `.wdb` and `.vcd` to the rolling `nightly` release |

The `vivado` runner must have `bazelisk` on its `PATH`, the installer
archive at the path above, and write access to `/data/cache`.

That label runs jobs directly on the host, and the host has no `node`.
No JavaScript action can run there, `actions/checkout` included.
Every workflow here checks out with plain `git` for that reason.
`forgejo-release` is fine: it is a composite action built from bash steps.


## License

Apache 2.0. See [LICENSE](LICENSE).

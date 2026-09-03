<!-- SPDX-License-Identifier: Apache-2.0 -->

[![Build](https://git.hdlfactory.com/HDL/wdbcvt/actions/workflows/build.yml/badge.svg)](https://git.hdlfactory.com/HDL/wdbcvt/actions?workflow=build.yml)
[![Test](https://git.hdlfactory.com/HDL/wdbcvt/actions/workflows/test.yml/badge.svg)](https://git.hdlfactory.com/HDL/wdbcvt/actions?workflow=test.yml)
[![Release](https://git.hdlfactory.com/HDL/wdbcvt/actions/workflows/release.yml/badge.svg)](https://git.hdlfactory.com/HDL/wdbcvt/actions?workflow=release.yml)

# wdbcvt

Tools for reading the Vivado `xsim` waveform database (`.wdb`).

AMD does not document that format.
This repository builds a known-good sample of one, and the tooling used to
take it apart.
See [docs/wdb-format.md](docs/wdb-format.md) for what has been measured so
far and how the work proceeds.


## Layout

* `hdl/counter/` holds a small VHDL design and its testbench.
  Simulating it produces the reference `sim.wdb`, and a `sim.vcd` from the
  same run that acts as the answer key.
* `cmd/wdbcvt/` is the command line tool.
  Today it probes a `.wdb` and reports measurements; it grows into a
  converter as the format becomes known.
  The first output format is VCD, written through
  `github.com/filmil/go-vcd-parser`.
  FST comes later, once there is something to write.
* `pkg/wdb/` holds the library the tool is built on.
* `docs/wdb-format.md` records the findings.


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

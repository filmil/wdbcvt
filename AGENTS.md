<!-- SPDX-License-Identifier: Apache-2.0 -->

# Instructions

Read `README.md` first.
It states what this repository is for, how it is built, and what the
`vivado` runner needs.

This repository has one purpose: work out the layout of the Vivado `xsim`
waveform database format, and build a converter for it.
`docs/wdb-format.md` is the working record of that effort.
Add a row to its findings table only for something reproduced, and keep
guesses in the open questions section where they are labelled as guesses.


# Build rules

* Bazel builds everything.
  The version is pinned in `.bazelversion`.
* Build with `bazel build //...` and test with `bazel test //...`.
* Never run `go` directly.
  Run `bazel run @rules_go//go -- <args>` instead.
* Run `bazel run //:gazelle` after adding or moving Go files.
* Run `bazel run //:buildifier` before committing Bazel files.
  Do not run it over the VHDL filegroups: it reorders `srcs`, and the
  VHDL compilation order matters.
  The `# do not sort` comment in `hdl/counter/BUILD.bazel` protects that
  list; keep it.
* Prefer the latest version of every `bazel_dep`.
  `bazel mod tidy` updates them.


# Vivado

Vivado runs in hermetic mode, set in `.bazelrc`:

```
build --@rules_vivado//:vivado_mode=hermetic
```

Bazel performs the installation itself, into the shared install cache
`/data/cache/vivado-install`.
Do not switch this to `docker` or `host` mode, and do not remove the
`install_cache` attribute from the `vivado.install` tag in
`MODULE.bazel`.
Without it, every workspace redoes an installation that takes hours.

`rules_vivado` must stay at 3.10.0 or later.
Hermetic mode does not exist in earlier releases.


# VHDL rules

* Entity ports do not take loose `std_ulogic` or `std_ulogic_vector`
  signals.
  Related signals are bundled into records declared in a package.
  `hdl/counter/counter.pkg.vhdl` is the pattern.
* Document the public API with Doxygen `--!` comments.
  Document entities and packages.
  Do not document architectures or package bodies.
* Entity documentation includes an ASCII timing diagram for a typical
  transaction.
* Do not put the project name into entity names.
* `use lib.pkg.all` is fine for the standard and IEEE libraries.
  For everything else, refer to elements through their library qualified
  name.
* Every testbench ends with `std.env.stop`, so that `run -all` terminates
  and xsim writes the waveform database out.


# CI rules

All three workflows use `runs-on: vivado`.
That label runs the job directly on the runner host, which is the only way
to reach the Vivado installation and the shared caches.

The host has no `node`, so a JavaScript action cannot run in these
workflows.
Do not add `actions/checkout` or any other JavaScript action.
Check out with plain `git`, the way the existing workflows do.
A composite action made of bash steps is fine; `forgejo-release` is one.


# Engineering standards

* Every new piece of functionality documents its public interface.
* Every new piece of functionality has unit tests, and passes them.
* Never ignore an error.
  Propagate it with context, or log it.


# Git rules

* Conventional Commits 1.0.0 for the commit title.
* Append this note as the last line of the commit message, below any
  summary:

  ```
  This commit has been created by an automated coding assistant,
  with human supervision.
  ```

  Then append the exact prompt that produced the commit, in full.
* Rebase, do not merge.
  Always pass `--no-edit` to `git merge`, `git cherry-pick` and
  `git rebase`.
* `git.hdlfactory.com` is a Forgejo instance.
  Drive it with `fj`, never with `gh`, and pass
  `-H git.hdlfactory.com` on every invocation.
  The remote is named `hd`.
* Every completed unit of work lands as a pull request.
  Do not push to `main`.


# Writing style

Short declarative sentences.
No em-dashes or en-dashes anywhere.
In markdown, start every sentence on its own line and wrap at 80 columns.

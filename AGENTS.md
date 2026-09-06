<!-- SPDX-License-Identifier: Apache-2.0 -->

# Instructions

Read `README.md` first.
It states what this repository is for, how it is built, and what the
`vivado` runner needs.

This repository has one purpose: work out the layout of the Vivado `xsim`
waveform database format, and build a converter for it.
`docs/format.md` is the working record of that effort.
Add a row to its findings table only for something reproduced, and keep
guesses in the open questions section where they are labelled as guesses.


# Build rules

* Bazel builds everything.
  The version is pinned in `.bazelversion`.
* Build with `bazel build //...` and test with `bazel test //...`.
* Never run `go` directly.
  Run `bazel run @rules_go//go -- <args>` instead.
* `bazel test //docs:docs_test` holds every document to the rules of
  the writing style section below.
  It fails on data given a verb, on a verb chosen for texture where a
  plainer one fits, on a term this repository coined, on a word that
  adds nothing, and on an em dash or an en dash.
  A passage that quotes a violation so as to correct it marks itself
  with `<!-- prose-lint: allow -->`, which holds to the next blank
  line.
  `bazel run //tools/prose/lint -- <file>` checks one file, and the
  rules live in `//tools/prose`.
* Never run a script directly either.
  Every tool in `//tools` is a target: `bazel run //tools/corpus/gen_t67`
  writes a tier, `bazel run //tools/corpus/typetab -- <file>` splits a
  type table, and `bazel run //tools:noise_mask -- <target>` measures
  the noise mask.
  A new tool arrives with the target that runs it.
* Do not add Python.
  The tools are Go programs, and the repository has no Python
  toolchain to run a script with.
  A one off calculation belongs in a Go program under `//tools`, or in
  a test.
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

Bazel must stay at **9.2.0 or later**, and `.bazelversion` pins it.
Bazel 9.1.0 crashes while fetching the hermetic installation, because
the installer archive is named by a `file://` URL and
`ProgressInputStream.reportProgress` calls `String.equals` on
`URI.getHost()`, which is `null` for a `file://` URL. The crash arrives
on the first progress report, tens of megabytes into a hundred gigabyte
download, as
`java.lang.NullPointerException: Cannot invoke "String.equals(Object)"
because the return value of "java.net.URI.getHost()" is null`.
Do not lower `.bazelversion` below 9.2.0 while the archive is served
from a `file://` URL.


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

Every workflow uses `runs-on: vivado`.
That label runs the job directly on the runner host, which is the only way
to reach the Vivado installation and the shared caches.

The host has no `node`, so a JavaScript action cannot run in these
workflows.
Do not add `actions/checkout` or any other JavaScript action.
Check out with plain `git`, the way the existing workflows do.
A composite action is fine only if every one of its steps is bash, and
that includes the steps of any action it `uses` itself.
A skipped step does not make it safe: an `if` guard of the form
`if: ${{ inputs.x }}` is true even when the input is `false`, because
the string `false` is truthy in the expression language.
Upstream `forgejo-release` fails exactly this way: its v2.13.4 embeds an
`actions/cache` step, a node action, behind such a guard.
The release workflow therefore uses the vendored, bash only copy in
`//.forgejo/actions/forgejo-release`.
Vendor and trim any other action before using it here.

A credential reaches a job through a repository secret and no other
way.
Do not generate, store, copy or name key material, and do not put the
location of a key in a commit message, a pull request or a comment.
When a job needs one, write the step to read
`${{ secrets.<NAME> }}`, make the job exit zero with a warning when
that is empty, and ask the repository owner to create the secret.
`//.forgejo/workflows/mirror.yml` is the pattern: it writes the secret
into a `mktemp -d` under `umask 077`, removes it on exit, and never
puts it on a command line or in a URL.

Host mode also gives the job a **pseudo-terminal**: a step's stdin is a
`/dev/pts` device, not `/dev/null`.
So any tool that decides to be interactive because it sees a terminal
will block forever, and the job sits in `running` until the six hour
timeout while holding the single serial runner.
`git log` is the one that has already done this: it started
`/usr/bin/pager`, which waited on input that never came.

Every command in a workflow step must therefore be non-interactive by
construction, not by luck.
Pass `--no-pager` to `git`, keep `GIT_PAGER=cat` in the step
environment, and give the same treatment to anything else that pages or
prompts.
Redirecting a step's stdin from `/dev/null` is the blunt version of the
same rule.


# The report

`//docs/latex` holds the report on the whole effort, in the IEEEtran
class, built by `rules_latex_host` from the host `pdflatex`.
Run `bazel build //docs/latex:report`; the PDF lands in
`bazel-bin/docs/latex/report.pdf`, and the release workflow attaches it
as `wdbcvt-report.pdf`.
The runner therefore needs a TeX Live with `IEEEtran.cls`.
Keep the report in step with `docs/format.md`: it states the numbers,
so a tier that changes the case count or a measurement changes the
report too.


# Provenance rules

This is an AI-first exploration of an undocumented format.
`docs/provenance.md` states that plainly, names the guards that stand in
for a specification, and says where `dewdb` should not be used.
Keep it accurate as the guards change.

Rules that follow from it, and that apply to every change here:

* Do not soften the framing.
  The README and `dewdb --help` both say that the format knowledge came
  from an agent running experiments, not from documentation.
  A reader who finds the tool without finding the docs still learns how
  the knowledge was obtained.
* All findings about the format live in `//docs`. Write them down as
  they are discovered, not at the end. `docs/README.md` is the index and
  says how a finding is recorded.
* A decoded field enters the findings table in `docs/format.md` only
  with the command that reproduces it.
  Everything else stays in the open questions section, labelled a guess.
* A test asserts against a `truth.json`, or against the VCD as read by
  `go-vcd-parser`.
  A test that asserts against bytes the decoder itself produced proves
  nothing, and must not be written.
* Every claim is scoped to the Vivado version that produced the file.
  Only 2025.2 has been observed.
* Never describe this work as a specification, a port, or a
  decompilation. It is none of those.


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

* Short declarative sentences.
* No em-dashes or en-dashes anywhere.
* In markdown, start every sentence on its own line and wrap at 80 columns.
* No mannered phrases.
* Name the thing that acts.
  The file, the writer, the reader or the simulator does something; a
  value, an object or a type does not.
  <!-- prose-lint: allow -->
  Write "the handle space grows by 8 bytes for one integer constant",
  not "each object takes its value size rounded up to 8".
* State the measurement before the shorthand, or instead of it.
  A sentence that only a reader who already knows the numbers can
  follow is not a finding, it is a note to yourself.
  <!-- prose-lint: allow -->
  "A package costs its objects and nothing else" says nothing; the
  numbers it stood for, 8 bytes for one integer, 16 for four, 64 for
  sixteen, say it.
* Define a term the first time this repository uses it, in the page
  that uses it.
  Handle space, arena, chunk and value class are this project's own
  words, not the reader's.
* Prefer the concrete number to the category.
  <!-- prose-lint: allow -->
  "0x50 for a sixteen element array, its 64 bytes and 16 more" beats
  "an array costs extra".
* A word that could be cut is cut.
  <!-- prose-lint: allow -->
  "in order to", "it should be noted that", "essentially", "simply",
  "of course" add nothing.

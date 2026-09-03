<!-- SPDX-License-Identifier: Apache-2.0 -->

# Where the WDB knowledge comes from, and what guards it


## Read this before trusting `dewdb`

The Vivado `xsim` waveform database format is undocumented.
AMD publishes no specification for it.
Nothing in this repository comes from AMD documentation, AMD source
code, or a disassembly of an AMD binary.

Everything here was worked out by an **AI agent**, running experiments
against `.wdb` files produced by simulating designs written for that
purpose, and reading the bytes that came out.
That is the honest description of the method, and it is stated first
because it is the fact a reader most needs when deciding how much weight
to put on a decoded field.

An agent that infers a format from examples produces two kinds of
output.
Some of it is a measurement that reproduces.
The rest is a plausible story about bytes.
The two look identical on the page.
So this project does not ask anyone to trust the derivation.
It guards the result with software written by other people, and with
references whose content is known before the `.wdb` is ever opened.


## What the derivation is not

* It is not a port of any AMD code.
* It is not a decompilation. No AMD binary was disassembled.
* It is not a specification. Where this repository states a fact about
  the format, that fact is a description of files observed on one
  machine, with one Vivado version.

The inputs are files that Vivado wrote, on a machine licensed to run
it, from source this repository owns.
The method is black box observation of those output files.


## The guards

No claim about the format is load bearing until at least one guard
below agrees with it.

**1. The VCD from the same simulation run.**
Every corpus case writes `sim.vcd` beside `sim.wdb`, from one run of one
design. VCD is specified in IEEE 1800 and is text. It is the answer key:
whatever the `.wdb` says about a signal, the `.vcd` from the same run
says it too, in a form that has been readable for thirty years.

**2. `go-vcd-parser`, which this project did not write.**
The answer key is read by `github.com/filmil/go-vcd-parser`, an existing
implementation with its own test suite. A decoder claim is checked
against what that parser reports, not against this project's own reading
of the same file. A bug shared between a decoder and its checker is the
failure this avoids.

**3. Ground truth declared before the file is opened.**
Each corpus case ships a `truth.json` stating the signals, widths,
times and values that the simulation was written to produce. It is
derived from the design, not from the database. The decoder has to
reproduce it. See [corpus.md](corpus.md).

**4. Independent simulators.**
GHDL and nvc simulate the same VHDL and write their own waveforms. They
share no code with Vivado. Where they agree with the decoded `.wdb`, the
agreement is evidence about the design rather than about one tool's
quirks. Every corpus case was already checked this way before any `.wdb`
existed: all nine analyzed, elaborated and ran under `ghdl --std=08`,
and every `truth.json` matched the VCD from that run.

**5. `libfst` for the output side.**
The FST writer is checked by reading its output back with `libfst`, the
reference implementation, vendored test only. See
[fst-output.md](fst-output.md).


## The rules that follow

* A decoded field is recorded in the findings table of
  [format.md](format.md) only with the command that reproduces
  it. Anything else stays in the open questions section, labelled as a
  guess.
* A test asserts against `truth.json` or against the VCD read by
  `go-vcd-parser`. A test that asserts against bytes the decoder itself
  produced proves nothing.
* Every claim is scoped to the Vivado version that produced the file.
  Only 2025.2 has been observed. A second version is a new experiment,
  not a confirmation.
* `dewdb` states this provenance in its own `--help` and package
  documentation. A user who finds the tool without finding this file
  still learns how the knowledge was obtained.


## What this means in practice

Use `dewdb` where a wrong answer is visible and cheap: looking at
waveforms, converting a trace for a viewer, diffing two runs.

Do not use it where a wrong answer is silent and expensive: sign off,
compliance evidence, or anything where a missing transition would not be
noticed. For those, open the database in Vivado, which is the only
implementation that is definitionally correct about its own format.

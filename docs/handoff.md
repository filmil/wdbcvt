<!-- SPDX-License-Identifier: Apache-2.0 -->

# Hand-off: where the exploration stands

This page is for whoever continues the work, agent or person.
It says what the reader decodes, what it does not, why, and what to do
next.
The findings themselves live in [format.md](format.md) and under
`format/`; nothing here replaces a row there.
Everything is scoped to Vivado 2025.2.


## What we have

The reader, `wdbcvt`, opens every one of the 1172 corpus cases of tiers
1 to 83 and reproduces their `truth.json` and Vivado's own VCD.
Run `bazelisk test //pkg/... --test_output=errors` to check.

Decoded and confirmed, with the case that found each in `format.md`:

* The container: fixed header, arena table, directory, page directory,
  trailer and the noise mask that separates content from timestamps.
* The type table: enumerations, integers, reals, physical types,
  arrays, records, aliases, access and file types, for VHDL, Verilog
  and SystemVerilog; and the string, queue, dynamic array, associative
  array and class entries that `xelab -debug all` adds.
* The debug section: scopes, units, declarations, files, instance
  records, handles, generics, generates, packages, ports bound to
  slices, and the declarations `-debug all` brings back.
* Values over time: arenas, pages, chunked writes, the encoding of
  every type above, forced and deposited values, 64 bit times, the
  placeholder record of a string or class handle, and what the VCD
  keeps of the same run.

Tiers 60 and 61 established what `-debug all` adds to a SystemVerilog
file and how its numbering runs; see `format/types.md` "Under -debug
all" and "The numbering", `format/hierarchy.md` "What -debug all
brings back" and `format/values.md` "Placeholder records under -debug
all".
Tier 62 established that drive strengths are resolved and not
recorded, that gates and pull sources are `Forked` process scopes,
and that a net with two or more drivers records bit by bit; see
`format/values.md` "Drive strength, pull sources and gates".
Tier 63 established that a driver of part of a net writes the pairs
its bits fall in, whole, that the first record holds `X` on the driven
bits and `Z` on the rest, and that an output port bound to part of a
net shares the net's handle with a bit offset; see
`format/values.md` "Partial drivers on a net".
Tier 64 established that several partial drivers write one record
each in the scheduler's order, that the port position word is written
on the first instance of a unit only, and that an `assign` in a
generate block gets no block scope; see `format/values.md` "Several
partial drivers on one net" and `format/hierarchy.md` on the word at
`40` and on generates.
Tier 65 put the tier 44 reading of the 8 byte times across a page
boundary, into another unit and up to 1 s.
Tier 66 established what the SystemVerilog constructs the corpus had
not seen leave: a `final` block and an `always_latch` are `Always`
scopes, assertions and `specify` blocks leave nothing, a `covergroup`
leaves nine generated scopes, a `program` ends the run, a `bind`
places the instance at the `bind` line, and a `specify` path delay
delays the records; see `format/hierarchy.md` on process scopes.


Two designs from outside the repository were added after tier 66:
`//hdl/picorv32:sim`, PicoRV32 under its own `testbench_ez.v`, which
read and matched its VCD unchanged, and `//hdl/neorv32:sim`, NEORV32
1.11.7 under its own testbench, 5696 objects, which found that the
chunk map of an object's records belongs to the signal at its handle
and not to the object; see `docs/corpus.md`.

FST output works: `wdbcvt -in <file>.wdb -fst <file>.fst` converts a
database, `pkg/fst` writes through `gtkwave/libfst` over cgo, and
`pkg/fstout` holds the mapping, with records and arrays flattened into
one variable per leaf. `pkg/fstout:fstout_test` converts every corpus
case, reads it back through libfst and compares against its
`truth.json`, which is the second guard for the types Vivado's VCD
leaves out. `docs/fst-output.md` states the mapping, what it costs and
what is left.

SQLite output works too: `wdbcvt -in <file>.wdb -sqlite <file>.db`
writes the same signals and changes as rows, in the schema
`go-vcd-parser` writes from a VCD, so one query reads either database.
`pkg/sqlout` holds the mapping and `pkg/sqlout:sqlout_test` compares
its rows against Vivado's own VCD, through that project's own writer.
`docs/sqlite-output.md` states the schema, the two places it departs
from the upstream DDL and why.

`//hdl/ibex:sim` is the SystemVerilog design, added after tier 67:
lowRISC's Ibex under its own `simple_system` example, 3287 objects,
running a hand assembled program that ends the run. It found that a
SystemVerilog enumeration takes its width from the range on its own
entry, which the reader had right and the VCD cross check had wrong,
and tier 67 pins that.

## What we do not have, and why

**The content of dynamic objects.**
`-debug all` declares a string, a queue, a dynamic array, an
associative array and a class handle, but records none of their
values: a string or class handle holds one 32 bit record of zeros at
time 0, and a container is never logged.
Nothing tried, a write from the source, `new` at time 0, `log_wave` by
name, changes that, and `set_value` on a string ends the batch script.
So the values of these objects are not in the database as far as any
case shows, and the reader cannot recover what the file does not hold.
Open question 24 of `format.md` keeps the two words it leaves
unexplained: the numbers that replace `-99` after an array's triples,
and a class's id word.

**Older open leads**, all in the open questions of `format.md`: the
`0x738` before the first signal, the per-package tail, the 8 of an
object-less scope, the generate versus instance difference of `0x18` to
`0x20`, word 10 of the header, interface port storage, the 12 then 16
of protected locals and the record `+4`.
Each is a single word with no visible effect on decoding, which is why
whole classes of objects came first.


## Plan for the next agent

Work on branch `ai-dev-20260904-rtd-designs`, PR #13, until it merges;
then branch from `hd/main`.
The generators of tiers 57 to 83 are in `tools/corpus/`, see its
README; a new tier starts by copying `gen_t83/main.go`.
The registration anchor for tier 84 in `hdl/corpus/BUILD.bazel` is the
last tier 83 case in sorted order, `t83_rec_nested__`.
`//tools/pagegrep` answers "is this value anywhere in the file": it
searches the bytes as they lie and every record of every inflated
page, and tier 68 is what it was written for.

1. Pick the next lead from the open questions of `format.md` and
   design minimal pairs for it, one variable per case, the case names
   exactly 16 characters.
   Tier 61 settled the order of the numbering of question 24; what
   reads it, and the hidden numbers of an associative array, are the
   open part.
2. Generate, register, build, dump, and write the truths from the
   dump only after reading the raw records; then run
   `bazelisk test //pkg/wdb:wdb_test --test_filter='TestCorpus/t68|TestVCD/t68'`.
3. Document in the `format/` page for the area, then `format.md`:
   findings rows before "Whole file properties, also measured:",
   comparison rows after the tier 83 rows, guesses in the open
   questions; a tier section in `corpus.md` before "Record which
   comparison produced which finding"; and the count 1172 upward
   everywhere (`docs/format/*.md`, `docs/corpus.md`, `README.md`,
   `docs/format.md`, `docs/corpus.md` "through tier NN").
   Keep lines at 80 columns and check with the awk loop in
   `docs/corpus.md`.
4. Commit in topical order: `feat(wdb): ...` for the reader,
   `test(corpus): tier NN, ...` for the cases, `docs: tier NN, ...`
   for the pages, each with the assistant note and the prompt.
   Run `bazelisk run //:buildifier` and the full test suite first,
   check the PR is open, push, and comment on it.

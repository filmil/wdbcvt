<!-- SPDX-License-Identifier: Apache-2.0 -->

# Hand-off: where the exploration stands

This page is for whoever continues the work, agent or person.
It says what the reader decodes, what it does not, why, and what to do
next.
The findings themselves live in [format.md](format.md) and under
`format/`; nothing here replaces a row there.
Everything is scoped to Vivado 2025.2.


## What we have

The reader, `wdbcvt`, opens every one of the 1014 corpus cases of tiers
1 to 63 and reproduces their `truth.json` and Vivado's own VCD.
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

Work on branch `ai-dev-20260904-mzq-tier61`, PR #11, until it merges;
then branch from `hd/main`.
The generators of tiers 57 to 63 are in `tools/corpus/`, see its
README; a new tier starts by copying `gen_t63.py`.
The registration anchor for tier 64 in `hdl/corpus/BUILD.bazel` is the
last tier 63 case in sorted order, `t63_pdr_w64_hi__`.

1. Pick the next lead from the open questions of `format.md` and
   design minimal pairs for it, one variable per case, the case names
   exactly 16 characters.
   Tier 61 settled the order of the numbering of question 24; what
   reads it, and the hidden numbers of an associative array, are the
   open part.
2. Generate, register, build, dump, and write the truths from the
   dump only after reading the raw records; then run
   `bazelisk test //pkg/wdb:wdb_test --test_filter='TestCorpus/t64|TestVCD/t64'`.
3. Document in the `format/` page for the area, then `format.md`:
   findings rows before "Whole file properties, also measured:",
   comparison rows after the tier 63 rows, guesses in the open
   questions; a tier section in `corpus.md` before "Record which
   comparison produced which finding"; and the count 1014 upward
   everywhere (`docs/format/*.md`, `docs/corpus.md`, `README.md`,
   `docs/format.md`, `docs/corpus.md` "through tier NN").
   Keep lines at 80 columns and check with the awk loop in
   `docs/corpus.md`.
4. Commit in topical order: `feat(wdb): ...` for the reader,
   `test(corpus): tier NN, ...` for the cases, `docs: tier NN, ...`
   for the pages, each with the assistant note and the prompt.
   Run `bazelisk run //:buildifier` and the full test suite first,
   check the PR is open, push, and comment on it.

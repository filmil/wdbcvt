<!-- SPDX-License-Identifier: Apache-2.0 -->

# Hand-off: where the exploration stands

This page is for whoever continues the work, agent or person.
It says what the reader decodes, what it does not, why, and what to do
next.
The findings themselves live in [format.md](format.md) and under
`format/`; nothing here replaces a row there.
Everything is scoped to Vivado 2025.2.


## What we have

The reader, `wdbcvt`, opens every one of the 938 corpus cases of tiers
1 to 59 and reproduces their `truth.json` and Vivado's own VCD.
Run `bazelisk test //pkg/... --test_output=errors` to check.

Decoded and confirmed, with the case that found each in `format.md`:

* The container: fixed header, arena table, directory, page directory,
  trailer and the noise mask that separates content from timestamps.
* The type table: enumerations, integers, reals, physical types,
  arrays, records, aliases, access and file types, for VHDL, Verilog
  and SystemVerilog.
* The debug section: scopes, units, declarations, files, instance
  records, handles, generics, generates, packages, ports bound to
  slices.
* Values over time: arenas, pages, chunked writes, the encoding of
  every type above, forced and deposited values, 64 bit times, and what
  the VCD keeps of the same run.

Tier 59, the last complete tier, established how Tcl forces, `set_value`
deposits and SystemVerilog `force` statements appear in the records; see
`format/values.md`, "Values the script or the source imposes".


## What we do not have, and why

**SystemVerilog under `-debug all` (tier 60, in progress).**
Every SystemVerilog case before tier 60 ran under xsim's typical
debugging level.
`xelab -debug all` makes xsim keep objects that typical drops: strings,
queues, dynamic arrays, associative arrays and class handles.
The 14 cases of tier 60 are in the corpus and build, and the reader
parses their type tables, but the tests of nine of them fail because
their truths are the guesses written before the files were inspected,
and because the value decoder rejects the new kinds.
This is the state of the branch, on purpose: the observations below are
reproducible today, the decoding of the new objects is not written yet.

Observed so far, each reproducible with
`wdbcvt -dump -in bazel-bin/hdl/corpus/<case>/sim.wdb`:

* Five new type kinds appear: `0x13` dynamic array, `0x14` queue,
  `0x15` associative array, `0x17` class and `0x18` string.
  The reader names and parses them (`pkg/wdb/types.go`).
  A string entry holds only the origin word.
  A queue or dynamic array holds the element type and the word 1.
  An associative array holds the element type, a word, and the key
  type: the word is 2 for a string key, `t60_dbg_assoc___`, and 3 for
  an int key, `t60_dbg_assoc_i_`.
  A class holds the parent class as a type index, -1 for none, then a
  word, then its fields as a record does, each followed by a word 0.
* The word after an array entry's last range triple is -99 in every
  file before tier 60 and stays -99 for an ordinary variable under
  `-debug all`, `t60_dbg_int_____`.
  It becomes a small non-negative number for a type that a class or
  dynamic object refers to: `int` holds 0 under a queue, dynamic array
  or associative array, and 1 or 2 under a class.
  The class's own word is 0 for the variable's class and 1 for its
  parent, `t60_dbg_class_d_`; in `t60_dbg_class_2_` the `logic [3:0]`
  field type holds 1 and `int` 2.
  The reader stores the word as `Type.Tail`.
  Guess: one numbering over the types the class machinery registers,
  starting at the variable's class.
  The order that assigns it is not established.
* A string or class handle variable gets a declaration of 32 bits and
  value class 0, an object at the usual second handle `0x828`, and it
  is logged: one 8 byte record of zeros at time 0 and nothing at the
  write at 50 ns, `t60_dbg_str_____`, `t60_dbg_class___`.
* A queue, dynamic array or associative array gets a declaration of 32
  bits and value class 3 with a range (0 to 0), an object at `0x828`,
  and is not logged; its arena is never written, `t60_dbg_queue___`.
* None of the new objects appears in the VCD of the default script.
* The handle space cost of a string is `0x98` with or without the flag,
  `t60_dbg_str_____` against `t60_dbg_vec_____`, `t11_sv_str______`
  against `t11_sv_logic____`.
* The file size jump of about 1500 bytes for a second variable is the
  second arena, not the flag; `t25_sv_two_same_` shows the same.

Why it is unfinished: the session was interrupted at the point of
teaching the value decoder the new kinds.
The decoder (`pkg/wdb/value.go`, `pkg/wdb/verilog.go`) sizes a Verilog
object from its type, and a string or class type has no bits of its own.

**Older open leads**, all in the open questions of `format.md`: the
`0x738` before the first signal, the per-package tail, the 8 of an
object-less scope, the generate versus instance difference of `0x18` to
`0x20`, word 10 of the header, interface port storage, the 12 then 16
of protected locals and the record `+4`.


## Plan for the next agent

Work on branch `ai-dev-20260904-kpr-tier51`, PR #10, until it merges;
then branch from `hd/main`.
The generators of tiers 57 to 60 are in `tools/corpus/`, see its
README.

1. Teach the decoder the new kinds.
   In `bitsOf` (`verilog.go`) return 32 for `KindString` and
   `KindClass`, and 32 for `KindQueue`, `KindDynArray` and `KindAssoc`,
   which matches the declared size of every tier 60 case.
   In `decodeBits` decode those 32 bits as an unnamed vector, so a
   string or class handle reads as 32 zero bits at time 0.
   `Open` then passes, because the logged flag and the records agree.
2. Fix the truths.
   `t60_dbg_int_____` lists `i` as a 32 bit vector but the reader
   prints an int as decimal: write `7` and `9`.
   `t60_dbg_str_____`, `t60_dbg_class___`, `t60_dbg_class_2_` and
   `t60_dbg_class_d_` need a second signal, width 32, with one
   transition at 0 to 32 zero bits.
   `t60_dbg_queue___`, `t60_dbg_dynarr__`, `t60_dbg_assoc___` and
   `t60_dbg_assoc_i_` need the second signal with `"logged": false`.
   Test with
   `bazelisk test //pkg/wdb:wdb_test --test_filter='TestCorpus/t60|TestVCD/t60'`.
3. Perturb once more before documenting: a class with two handles, a
   string written from Tcl with `set_value` under `-debug all`, and a
   queue logged with `log_wave` under `-debug all`, to learn whether
   any of them ever records a value.
   Also a VHDL case under `-debug all` beside `t22_dbg_all` to confirm
   VHDL is unchanged.
4. Document.
   `format/types.md`: rows for the five kinds in the kinds table, the
   entry layouts, and the trailer word with its observed values.
   `format/hierarchy.md`: the declarations and objects of the new
   kinds.
   `format/values.md`: the placeholder record.
   `format.md`: findings rows before "Whole file properties, also
   measured:", comparison rows after the tier 59 rows, and the tail
   numbering as a guess in the open questions.
   `corpus.md`: a "Tier 60" section before "Record which comparison
   produced which finding", and the count 938 to 952 everywhere
   (`docs/format/*.md`, `docs/corpus.md`, `README.md`,
   `docs/format.md`).
   Keep lines at 80 columns and check with the awk loop in
   `docs/corpus.md`.
5. Commit in topical order: `feat(wdb): ...` for the decoder,
   `test(corpus): tier 60, ...` for the truths, `docs: tier 60, ...`
   for the pages, each with the assistant note and the prompt.
   Run `bazelisk run //:buildifier` and the full test suite first.
6. Then continue the exploration with the older open leads above.

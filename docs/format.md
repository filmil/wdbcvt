<!-- SPDX-License-Identifier: Apache-2.0 -->

# Decoding the xsim waveform database (`.wdb`)


## What this document is

AMD does not document the `.wdb` container that `xsim` writes.
This file is the index of what has been measured about it, the findings
table, and the record of which comparison led to which discovery.
Everything here is either a measurement or a statement marked as a
guess.
Nothing gets promoted from guess to fact without a reproduction.

The layout itself is described in four documents, one per area:

| Document | Area |
| :--- | :--- |
| [format/container.md](format/container.md) | the fixed header, the arena table, the trailer, the directory, the page directory and the marker |
| [format/types.md](format/types.md) | the type table: enumerations, integers, reals, physical types, arrays and records |
| [format/hierarchy.md](format/hierarchy.md) | the debug section: scopes, units, declarations, source files, instance records and handles |
| [format/values.md](format/values.md) | arenas, pages, value records, encodings and alignment |

All of it is from Vivado 2025.2 and is scoped to that version.
See [provenance.md](provenance.md) for what guards the claims and
[corpus.md](corpus.md) for the cases named below.


## Reading a database, in one paragraph

A database is one flat file.
The header at `0x48` points at three directory entries, and each entry
names a section that sits directly before it.
The `Xilinx RTTI` entry covers the type table.
The `Xilinx DBG` entry covers the design hierarchy, which ends with one
instance record per logged object, and each record carries the object's
handle and the index of its declaration and scope.
The page directory follows the DBG entry: one record per arena, listing
the zlib pages of that arena.
A page inflates to 10240 bytes of `[time][key][length][value]` records.
An object's values are the records whose key is `handle & 0x7ff`, in
the pages of arena `handle >> 11`.
The trailer after the arena table holds the end time.

```sh
bazel build //hdl/corpus:all_wdb
bazel run //cmd/wdbcvt -- -dump -in "$PWD/bazel-bin/hdl/corpus/t2_flat3________/sim.wdb"
```

The dump prints every structure in the order above, with offsets, and
every row of the findings table can be checked against it.


## Findings

Every row is a measurement that reproduces with

```sh
bazel run //cmd/wdbcvt -- -dump -in "$PWD/bazel-bin/hdl/corpus/<case>/sim.wdb"
```

for the case the row names, and `//pkg/wdb:wdb_test` asserts every row
against the `truth.json` of all 63 cases.
The offsets are in the documents linked in the last column.

| Finding | Found by | Confirmed by | Where |
| :--- | :--- | :--- | :--- |
| Magic `Xilinx WAVE DATABASE 01`, producer `Xilinx Simulator` | hex dump of `t1_bit_one_edge` | all 63 cases | container |
| `0x38` is a Unix timestamp | noise mask, two runs of `t3_tr1` | equals the file mtime | container |
| `0x48` holds three pointers to 48 byte directory entries | `strings -t d` on `t2_flat3`, then reading the values | all 63 cases, each pointer lands on a name | container |
| The arena table at `0xc8` grows with the object count | `t5_sig10` shifted every trailer field by 8 | 3, 4, 6 slots in `t6_sig05`, `t6_sig12`, `t6_sig20` | container |
| The arena table has `ceil(handle space / 0x800)` slots | `t7_sig07` broke the `ceil(objects / 4) + 1` guess | the reader checks it in 63 of 63 | container |
| Trailer `+0x0c` is the arena table slot count | sweep of every fixed word over 63 cases | 63 of 63, checked by the reader | container |
| Trailer `+0x18` is the handle space size | `t7_sig07` to `t7_sig24` against the slot count | the slot rule, 63 of 63 | container |
| Arena records sit in first write order, not arena order | `t7_gen_for`, arena 2 first | 63 of 63 with the reader accepting any order | container |
| The trailer is the 0x48 bytes before the first directory pointer | `t5_sig10` against `t6_sig05` | end time correct in all 63 | container |
| The end time is a uint64 in ps at trailer `+0` | correlation sweep over 33 cases | 63 of 63, 20 ns to 1310 ns | container |
| The marker offset is at trailer `+0x38` | `t5_tr1000`, where the marker moved | 63 of 63 | container |
| The marker is `[0][logged objects minus 1]` | `t6_var_int` broke the earlier `objects minus 1` reading | 63 of 63 | container |
| Each directory entry follows the section it describes | `t2_flat3`: `WDB.Event` at `0xe0+0x48`, RTTI and DBG the same | 63 of 63 | container |
| The page directory starts 48 bytes after the DBG entry | `t2_flat3`, reading the offsets | 63 of 63 | container |
| An arena record is `0x4c0` bytes: 100 page offsets, 100 lengths, a count | `t5_tr1000`, two pages in one arena | `t6_tr1300`, three pages | container |
| A page is a zlib stream that inflates to 10240 bytes | entropy profile, then `zlib` on `t1_bit_one_edge` | 63 of 63 | values |
| Page header `[t0][last minus t0][n]` | `t5_tr1000` page 1 | all pages of all cases | values |
| A record is `[uint64 time][uint32 key][uint32 length][value]` | `t1_bit_two_edges` against `t1_bit_one_edge` | every record of every case matches `truth.json` | values |
| `handle >> 11` is the arena, `handle & 0x7ff` the key | `t5_sig10` | `t6_sig20`, four arenas | values |
| A page holds 10240 bytes of records and overflows into a new page | `t5_tr1000`, 600 one byte records | `t7_int700` 510 and `t7_wide700` 425 | values |
| Records at one time are in simulation order, not key order | `t7_gen_for` | `t2_flat3` | values |
| A delta cycle leaves two records at one time | `t7_delta` | | values |
| An overflowed page precedes the marker | `t5_tr1000` | `t6_tr1300` | values |
| Enumeration values are one byte, the literal's index | `t3_valz` against `t3_tr1`, same size | `t1_nine_state` walks all nine | values |
| Integers are int32, reals float64, time int64 ps | `t2_integer`, `t2_real`, `t2_time` | `truth.json` | values |
| Arrays are elements back to back, left index first | `t1_vec8` | `t2_array2d`, `t5_int_arr` | values |
| Record fields are aligned to their size, records to 8 | `t5_rec_real` against `t2_record` | `t5_arr_rec`, `t5_rec_sub5` | values |
| A signal has one record at time 0 and one per change | `t0_bit_const` | `t3_late`, 63 of 63 | values |
| The type table starts with `Xilinx ISim TYPE FILE 001` | `strings` on `t1_bit_one_edge` | 63 of 63 | types |
| `+32` of the type table is the number of types | correlation sweep | 63 of 63 | types |
| Type entries are `[len][tag]` name body | `t2_enum` against `t1_bit_one_edge` | 63 of 63 | types |
| Enumerations list their literals; `character` has 256 | `t2_enum`, `t2_character` | `truth.json` names | types |
| Integer entries carry the bounds, reals the bounds as float64 | `t2_integer`, `t2_real` | 63 of 63 | types |
| Physical entries list units with scales | `t2_time` | | types |
| Arrays carry element, index type and constraint triples | `t1_vec8` against `t2_array2d` | `t5_int_arr` | types |
| Records list fields with types and ranges | `t2_record` | `t2_record_nested`, `t5_rec_sub5` | types |
| A record field of record type lists one range per inner field, the scalar's own range, only when the inner record has an array field | `t7_rec_vfirst`, `t7_rec_bitv`, `t7_rec_intv`, `t7_rec_in2` | `t7_rec_in2v` | types |
| Types are shared between signals of the same type | `t2_record_two` | `t6_sig20`, one `STD_ULOGIC` | types |
| The DBG section starts with `Xilinx ISim DBG 006` and 18 region offsets | `t1_hier1` against `t2_hier3` | 63 of 63 | hierarchy |
| Scope records: name, parent, children, first object, unit, file, line | `t2_hier3` | 63 of 63 | hierarchy |
| Unit records: entity, architecture, kind, declaration count, file, line | `t2_hier3` | `t4_gen_diff_two` | hierarchy |
| Declaration records: name, file, line, size, type, ranges, kind | `t2_flat3` | 63 of 63 | hierarchy |
| Declaration kinds `0x0e` signal, `0x0f` variable, `0x12` generic, `0x13` loop index | `t4_gen_default`, `t5_tr1000`, `t6_var_int` | `t6_proc2` | hierarchy |
| Unit kind `0x0c` is a generate; iteration scopes are named `\g(0)\` | `t7_gen_for` | | hierarchy |
| The file table holds compile and local paths | `t2_slv8` against `t1_vec8` | 63 of 63 | hierarchy |
| Regions 14 and 15 are executable statement lines per file | `t6_proc2` | `t2_hier3` | hierarchy |
| Instance records: handle, second handle, scope, kind, declaration | `t2_flat3` | 63 of 63 | hierarchy |
| The second handle is the handle plus the value size rounded to 8 | `t2_record_two` against `t1_two_bits` | `t2_array2d`, `t2_record_nested` | hierarchy |
| Equal generics share a unit; different generics duplicate it | `t4_gen_same_two` against `t4_gen_diff_two` | | hierarchy |
| A generic is an object with one record at time 0 | `t4_gen_default` | `t4_gen_explicit` | hierarchy |
| A process variable is an object with no records | `t6_var_int` | `t6_proc2` | hierarchy |
| A loop index is an object with one record at time 0 holding 0 | `t5_tr1000` | `t6_tr1300` | hierarchy |
| A generate index is an object whose record holds the iteration value | `t7_gen_for` | | hierarchy |

Whole file properties, also measured:

* Signal names, scope names, type names and absolute source paths are
  stored in the clear.
  The paths include the Vivado installation and the machine paths AMD
  compiled the standard libraries on.
* Outside the noise mask, the file is deterministic.
  Two runs of the same design differ only at timestamps and durations.
  The pages are byte identical.
  See the noise mask section of
  [format/container.md](format/container.md).
* The `xsim.dir` tree beside the database is not needed to read it.


## Which comparison led to which discovery

The corpus is built of minimal pairs, and most of the findings above
came from one pair.
This table is the record of that, so that a reader can see what each
claim rests on and rerun the comparison.

| Comparison | Differs in | Discovery |
| :--- | :--- | :--- |
| two runs of `t3_tr1` | nothing | the noise mask: timestamps and durations only |
| `t1_bit_two_edges` against `t1_bit_one_edge` | one transition | a transition costs 15 bytes; the record layout |
| `t3_late` against `t3_tr1` | transition at 1000 ns not 10 ns | time is fixed width |
| `t3_valz` against `t3_tr1` | value `Z` not `1` | a value is an index, one byte |
| `t2_enum` against `t1_bit_one_edge` | a three literal type | enumerations are ordinary types; the entry framing |
| `t2_unsigned8` against `t2_signed8` | the type name | names cost one byte per character; the corpus needs padded directory names |
| `t2_slv8` against `t1_vec8` | resolved type, `use ieee.numeric_std` | the 447 bytes are two file table entries, not the resolved type |
| `t1_hier1` against `t2_hier3` | two more levels | scope records and the parent links |
| `t2_record_two` against `t2_record` | a second signal of the same type | types are shared; a second object gets a new handle |
| `t2_record_two` against `t1_two_bits` | value size 16 not 1 | the handle stride grows with the value size |
| `t2_flat3` against `t1_two_bits` | a third signal, three types | the directory entries follow their sections |
| `t4_gen_same_two` against `t4_gen_diff_two` | the generic values | units are shared only for equal generics; names never change |
| `t4_gen_default` against `t4_gen_explicit` | how the generic is set | no difference in the file |
| `t5_rec_real` against `t2_record` | a real field | fields align to their size |
| `t5_rec_sub5` against `t2_record_nested` | a 5 byte inner record | record fields align to 8 |
| `t5_sig10` against `t2_flat3` | ten signals | the arena table grows; the trailer moves; the handle split |
| `t5_tr1000` against `t3_tr16` | 1000 transitions | pages overflow at 600 records; the marker moves; `t1` |
| `t6_sig05`, `t6_sig12`, `t6_sig20` | the object count | arena table slot counts 3, 4, 6 |
| `t6_tr1300` against `t5_tr1000` | 1300 transitions | three pages, marker after the second |
| `t6_var_int` against `t3_tr1` | a process variable | variables are objects with no records; the marker counts logged objects |
| `t6_proc2` against `t6_var_int` | a second process | statement lines per file; the object order per scope |
| `t7_sig07` against `t6_sig05` | seven signals | four slots for seven objects; the slot guess was wrong |
| `t7_sig14`, `t7_sig16`, `t7_sig24` | the object count | slot boundaries at `0x1800`, `0x2000`, `0x2800` of trailer `+0x18` |
| `t7_int700`, `t7_wide700` against `t5_tr1000` | value size 4 and 8 | pages hold 510 and 425 records; the limit is bytes |
| `t7_delta` against `t3_tr1` | a `wait for 0 ns` between two assignments | two records at one time |
| `t7_rec_in2`, `t7_rec_in16` against `t5_rec_sub5` | inner record of scalars only | no extra triple at all |
| `t7_rec_vfirst` against `t5_rec_sub5` | inner field order | the `(0, 8, 1)` follows its field |
| `t7_rec_bitv`, `t7_rec_intv` against `t5_rec_sub5` | the inner scalar's type | the triple is the scalar's range; `8` is `std_ulogic`'s last literal |
| `t7_gen_for` against `t4_gen_diff_two` | a for generate | generate scopes and units; arena records in write order; records at one time unsorted |

Three findings were not found by a pair.
The end time, the type count and the has-objects flag came from the
correlation sweep: read the same offset in every case and keep the
offsets whose value equals a property known from `truth.json` in every
case.
A field that is correct in every case never differs between two cases,
so the sweep finds what a diff cannot.
It has one failure mode: a property that is nearly constant across the
corpus matches almost anything.
Design a case that moves a property before trusting a sweep hit on it.


## VCD cannot hold what the database holds

This is a property of VCD, not of any one writer, and it decides what a
converter can honestly produce.

Vivado writes `sim.vcd` from the same simulation run that writes
`sim.wdb`.
For eight of the fifteen types measured, that VCD contains no `$var`
declaration and no value changes at all.
The signal is absent, not degraded.

| Type | In Vivado's own VCD |
| :--- | :--- |
| `std_ulogic`, `bit` | present, as `wire 1` |
| `std_ulogic_vector`, `std_logic_vector`, `unsigned`, `signed` | present, as `wire N` |
| `boolean` | absent |
| `integer` | absent |
| `real` | absent |
| `time` | absent |
| `character` | absent |
| user enumeration | absent |
| record | absent |
| array | absent |

The whole VCD for the `integer` case is 123 bytes of header:

```
$date
   Thu Sep  3 08:52:42 2026
$end
$version
  2025.2
$end
$timescale
  1ps
$end
$enddefinitions $end
$dumpvars
$end
```

Two consequences follow.

The VCD answer key only covers bit and vector signals.
For the other eight types there is no independent reading of the same
run to check a decoder against.
[provenance.md](provenance.md) says which guard applies where.

A `.wdb` to VCD converter is lossy, and silently so.
It would drop every integer, real, enumeration, record and array in a
design without reporting anything, because VCD has nowhere to put them.
FST does: it has `FST_VT_VCD_INTEGER`, `FST_VT_VCD_REAL`,
`FST_VT_VCD_TIME`, `FST_VT_GEN_STRING` and `FST_VT_SV_ENUM`, and
`fstWriterCreateEnumTable` for enumerations with their literal names.
See [fst-output.md](fst-output.md).


## Open questions

Everything here is a guess or a gap, and stays here until a case
separates the readings.

1. Trailer `+0x18` is the handle space size, which fixes the slot count.
   Where its first `0x1088` bytes and the `0x58` per signal beyond the
   `0xf0` handle stride go is open.
2. Trailer `+0x10`, `0x800`, and `+0x20`, `0xc8`, read as the arena
   span and the arena table offset by their values.
   Both are constant, so that is a reading, not a finding.
3. An inner record with no array field carries no constraint at all.
   Whether a `real` field would contribute a range when a vector is
   present has not been asked.
4. DBG header words 14 to 16 are `0x101`, `0x101`, `0x10000` in every
   case, and the three `0x30` words at `0x98` and the `3` at `0xc0` in
   the fixed header are constant too.
   No case has moved them.
5. Word 10 of a declaration record varies between runs for a signal and
   is 0 for a variable.
   It is masked as noise and not read.
6. Handles of generics, variables and loop indexes follow no pattern
   seen yet.
   They are read from the instance record, so nothing depends on it.
7. `t7_delta` shows two records at one time for one delta.
   Whether a chain of deltas writes one record per delta, or only the
   ones that change the value, has not been asked.
8. Whether `0xc4` and the other per-run durations mean anything is
   open.
   They are masked.
9. Does the format change between Vivado versions?
   Only 2025.2 is in use here.
   Any claim is version scoped until a second version has been
   measured.
10. Verilog designs have not been simulated.
    Every case is VHDL, and the unit kinds and type classes may have
    Verilog values the corpus has never produced.


## What the conversion writes out

VCD first, through `github.com/filmil/go-vcd-parser`, as the checking
step: it reads the `sim.vcd` answer key, and a decoded database can be
compared against it for the bit and vector signals.

FST is the deliverable, because it holds every type in the table above.
Keep the decoder's output model separate from the VCD writer, so that
adding an FST writer is a new writer and not a rewrite.
See [fst-output.md](fst-output.md).


## Method

Build a minimal pair, mask the noise, diff, and write the answer into
the document for that area of the file before moving on.
Each answer becomes a row in the findings table plus a check in
`//pkg/wdb:wdb_test`, which asserts against `truth.json` and never
against bytes the decoder itself produced.
Where a claim can be cross-checked against `sim.vcd`, the test does that
check rather than asserting bytes.
The corpus and the pairing rules are in [corpus.md](corpus.md).

<!-- SPDX-License-Identifier: Apache-2.0 -->

# Values over time: arenas, pages and records

The values of every object live in zlib compressed pages at the end of
the file.
Each page holds a run of records, and each record is a time, an object
key and the object's whole new value.
The pages are grouped into arenas, and an arena is found through the
page directory described in [container.md](container.md).

Everything below is a measurement on files written by Vivado 2025.2.
Reproduce any of it with:

```sh
bazel build //hdl/corpus:all_wdb
bazel run //cmd/wdbcvt -- -dump -in "$PWD/bazel-bin/hdl/corpus/t5_tr1000_______/sim.wdb"
```


## Handles, arenas and keys

Every object carries a 64 bit handle in its instance record, see
[hierarchy.md](hierarchy.md).
The handle splits in two:

| Bits | Meaning |
| :--- | :--- |
| `handle >> 11` | arena index, into the arena table and the page directory |
| `handle & 0x7ff` | key, the number a value record carries |

Found by `t5_sig10`, the first case with more than one arena for
signals.
Its ten handles run from `0x768` in steps of `0xf0`, so `s0` is
`0x768` in arena 0 with key `0x768`, and `s1` is `0x858` in arena 1
with key `0x58`.
The dump of `t2_flat3` shows the same split at three objects:
`tb.alpha` in arena 0 with key `0x768`, `tb.bravo` and `tb.charlie` in
arena 1 with keys `0x58` and `0x148`.

An arena is a window of `0x800` handle values.
Which objects share an arena follows from their handles alone.
With one byte signals, arena 0 holds `s0`, arena 1 holds `s1` to `s9`,
arena 2 holds `s10` to `s17`, and arena 3 starts at `s18`.
Found by `t6_sig20`.

A generic or a loop index has a handle outside the signal sequence,
and lands in whatever arena that handle names.
`t5_tr1000` has the loop index at `0xdf8`, arena 1, key `0x5f8`, and
arena 1 holds nothing else.
`t7_gen_for` puts its three generate indexes and three generics in
arena 2 at `0x1040`, `0x1130`, `0x1070`, `0x1248`, `0x10a0` and
`0x1360`, in elaboration order, while its three signals take `0x768`,
`0x858` and `0x948` in arenas 0 and 1.
The elaboration time objects were written first, so arena 2's record is
the first in the page directory; see [container.md](container.md).


## Arena records and pages

The page directory has one `0x4c0` byte record per arena, described in
[container.md](container.md).
The record lists the file offset and the compressed length of each page
of that arena, 100 pages per record, and names a continuation record
when the arena has more.
`t9_tr70000` has 117 pages in two records.

A page inflates to exactly 10240 bytes, the page size named in the
trailer:

| Offset | Size | Meaning |
| ---: | ---: | :--- |
| `0` | 8 | `t0`, the time of the first record, in the file's time unit |
| `8` | 8 | `t1`, the time of the last record minus `t0` |
| `16` | 4 | `n`, the number of records |
| `20` | | `n` records, then zero padding to 10240 |

A record is:

| Offset | Size | Meaning |
| ---: | ---: | :--- |
| `0` | 8 | time, in the file's time unit |
| `8` | 4 | key |
| `12` | 4 | length of the value in bytes |
| `16` | | the value |

There is no alignment between records.
A one byte value makes a 17 byte record.

`t1` was found by `t5_tr1000`.
Its page 0 has `t0` 0 and `t1` 599000, with records at every 1000 ps
from 0 to 599000.
Its page 1 has `t0` 600000 and `t1` 400000, with the last record at
1000000.
A page holding one record at time 0 has both words 0.

Records are written in simulation order, and nothing sorts them
afterwards.
Two objects that change at the same time appear in the order the
simulator processed them.
`t2_flat3`, where `bravo` and `charlie` change together at 50 ns, has
them in declaration order, which is also handle order, and that was the
first reading.
`t7_gen_for` shows that it is not a rule: at 10 ns arena 1 holds key
`0x148`, the third instance's signal, before key `0x58`, the second
instance's, while at time 0 the same two keys come the other way round.
A reader must not assume that records at one time are sorted by key.

A delta cycle leaves two records at the same time.
`t7_delta` assigns `'1'`, waits `0 ns`, and assigns `'0'`, and its page
holds `t=10000 03` followed by `t=10000 02`.
The database keeps every delta, in order, and a VCD conversion that
keeps only the last value at a time loses one of them.

*Found by* `t7_gen_for` for the order and `t7_delta` for the delta,
each written to ask that one question.

Only a change gets a record.
`t8_delta3` assigns `'1'`, `'0'`, `'1'` across three deltas at 10 ns
and holds three records there, `03 02 03`.
`t8_delta_same` assigns `'1'` twice across two deltas and holds one
record at 10 ns.
`t8_same` assigns the initial value `'0'` at 10 ns and holds only the
time 0 record.
So the rule is one record per delta in which the value changes, and no
record for an assignment of the value already held.

*Found by* `t8_delta3`, `t8_delta_same` and `t8_same` against
`t7_delta`.

A page written out before the end of the run keeps one record per key
and time, the last one.
`t14_v_spill_dd` toggles a `reg clk = 1'b0` every nanosecond for 430 ns
beside a `reg d` without an initialiser, written `1` and then `0`
across `#0` at 5 ns, and the two share an arena that spills into a
second page.
Page 0 holds `d` as `X` at time 0 and `0` at 5000, and nothing of the
`1`.
`t14_v_page_dd`, the same two writes at 190 ns in a run that stays in
one page, holds both records at 190000, and `t14_v_spill_dd2`, the same
two writes at 428 ns in the second page of the spilling arena, holds
both as well.
So the loss belongs to a page the simulator wrote out while it was
still running, not to the arena or to the run length, and the last page
of an arena, written at the close, keeps every delta.
Page 0 still holds 425 records, the most that fit, so the lost record
left no gap.
The missing `X` record of tier 13 is the same loss: the `X` and the
initial value of `clk` share time 0, and `t14_v_spill_d`, which adds
`d` to that arena, keeps the `X` of `d` and loses the one of `clk`,
whichever of the two is declared first.

*Found by* `t14_v_spill_dd` against `t14_v_page_dd`, one record at
5000 against two at 190000.
*Confirmed by* `t14_v_spill_dd2` against `t14_v_spill_dd`, two records
in the last page; `t14_v_spill_d` and `t14_v_spill_dfst` against
`t13_v_tr430_2`, the `X` of `d` kept in the spilling arena.

The time unit is the simulation precision, and nothing finer is kept.
`t8_ps` waits `1 ps`, `998 ps` and `1500 fs`, and its records sit at
1, 999 and 1000, so the third wait was cut to 1 ps, and the end time is
11000.
That is the simulator's resolution as much as the file's: xsim runs the
VHDL corpus at its default 1 ps precision, and the VCD it writes agrees.

*Found by* `t8_ps` against `t1_bit_two_edges`.

The unit is written in the DBG section, as the power of ten of a
second after the timestamp, and it follows the Verilog precision.
`t21_v_ts_1ns_1ns` holds its change at 50 and its end time at 100
under `-9`, `t21_v_ts_1ns_100` holds 506 and 1001 under `-10`, for a
delay of `50.55` rounded to `50.6 ns` and an end of `100.1 ns`, and
`t21_v_ts_1ps_1fs` holds 50500 and 100000 under `-15`, for a delay of
`50.5 ps`.
The page bounds `t0` and `t1` are in the same unit.
The VCD is written in the same precision, `$timescale 1ns`, `100ps`
and `1fs`, so the two agree without conversion.
A source without a `timescale` directive, `t21_v_ts_none`, is at
picoseconds, and a VHDL child under a `1ns / 1ns` testbench,
`t21_mix_ts_1ns`, is at picoseconds as well: the finest precision in
the design wins.
The time scale unit, `10ns` in `t21_v_ts_10ns`, does not reach the
file; `#5` there is 50 under `-9`.
So the picosecond of the earlier tiers was the precision of every case,
not a property of the format, and the reader now takes the unit from
the section.

*Found by* `t21_v_ts_1ns_1ns` against `t11_v_bit_edge`, where the
reader refused the file over the word it had read as a constant `-12`.
*Confirmed by* `t21_v_ts_1ps_1ps`, `t21_v_ts_10ns`, `t21_v_ts_1ns_100`,
`t21_v_ts_1ps_1fs`, `t21_v_ts_none` and `t21_mix_ts_1ns`.

A VHDL design follows `xelab --timeprecision_vhdl` the same way.
`t22_vh_fs` at `1fs` holds `-15`, a change at 10000000 and an end time
of 20001500, and its `wait for 1500 fs` is kept whole.
`t22_vh_ns` at `1ns` holds `-9`, and the same wait rounds to nothing:
`s` records `1` and then `0` at time 10, two records at one time in
the order they were written, and the end time is 20.
A `TIME` signal counts the same unit: `t <= 1500 fs` records 1500 in
`t22_vh_fs`, 1 in `t22_base` at picoseconds, and nothing in
`t22_vh_ns`, where 0 ns is the initial value.
*Found by* `t22_vh_fs` against `t22_base`.
*Confirmed by* `t22_vh_ns`.

A net shared by a signal and the ports connected to it is one handle,
see [hierarchy.md](hierarchy.md), and its records show the connection
twice over.
`t8_port_chain` has a signal, a port and a port below that, all on
handle `0x768`, and the page holds three records at time 0, one per
object on the net, then one record for the change at 10 ns.
`t8_port_in` holds two at time 0 for its two objects.
So the count of time 0 records on a handle is the number of objects
that share it, and a reader that keys the initial values by handle
alone sees duplicates.
After time 0 a change is one record, whichever object drove it.

*Found by* `t8_port_chain` and `t8_port_in` against `t1_hier1`.

A procedure with a `signal` parameter writes the change twice.
`t9_proc_sig` declares `procedure set(signal s : out std_ulogic)` in
the process and calls it once at 10 ns, and the page holds `t=0 02`,
`t=10000 03`, `t=10000 03`.
The value is the same in both, so it is not a delta with a change, see
`t8_delta_same`, and the VCD holds one edge.
It costs `0x48` of handle space, where a procedure with a variable
costs `0x8`, so the parameter has some object of its own that is not
listed and that writes on the signal's handle.

*Found by* `t9_proc_sig` against `t9_proc_local`.


## Wide values: chunks

A value of 275 bytes or more is not one record.
It is written as a run of records with consecutive keys, each carrying
a chunk of the value, and a chunk is addressed by its handle: the
object's handle plus the chunk's byte offset in the value.
So a wide object occupies a range of handle space as long as its value,
and its chunks land in whichever arenas that range crosses.

`t9_vec292`, one `std_ulogic_vector(291 downto 0)` on handle `0x768`,
holds at time 0:

| Arena | Key | Bytes | Value bytes |
| ---: | ---: | ---: | ---: |
| 0 | `0x768` | 73 | 0 to 72 |
| 0 | `0x7b1` | 73 | 73 to 145 |
| 0 | `0x7fa` | 6 | 146 to 151 |
| 1 | `0x000` | 67 | 152 to 218 |
| 1 | `0x043` | 73 | 219 to 291 |

The logical chunks are four of 73 bytes.
The third is split at the arena boundary `0x800`, 6 bytes in arena 0
and 67 in arena 1, which is the only reason arena 1 exists in that
file.
A reader joins the pieces by address: it collects every record whose
key range overlaps the object's `[handle, handle + size)` from every
arena the range crosses, sorts them by address, and concatenates the
i-th record at each address into the i-th value.
The reader checks that the pieces are contiguous, that every address
holds the same number of records, and that the records joined into one
value carry the same time.

The split is in bytes and does not respect elements.
`t9_int73`, an array of 73 `integer`, 292 bytes, is chunked exactly as
`t9_vec292`: four chunks of 73 bytes, so a chunk boundary falls inside
an integer.

The chunk sizes depend on the value size alone, and the same chunking
is used at every time, including the duplicate time 0 records of a net
with several objects.
The rule, fitted to the 55 wide values of the `t9_vec*`, `t10_vec*`,
`t9_int73` and `t10_real40` cases and holding for every one:

| Value bytes | Chunks | Chunk bytes |
| :--- | :--- | :--- |
| `size < 275` | 1 | `size` |
| `size >= 275` | `n = 2 * ceil((size + 24) / 299)` | `floor(size / n)`, the last takes the rest |

So the chunks of one value are equal, and the last is up to `n - 1`
bytes longer.
The reader computes the run from the size and refuses a file whose
records sit at other addresses, which is how the rule is checked on
every case.

| Case | Bytes | Chunks | Sizes |
| :--- | ---: | ---: | :--- |
| `t10_vec274` | 274 | 1 | 274 |
| `t10_vec275` | 275 | 2 | 137, 138 |
| `t10_vec276` | 276 | 4 | 69 four times |
| `t9_vec292`, `t9_int73` | 292 | 4 | 73 four times |
| `t10_vec574` | 574 | 4 | 143 three times, then 145 |
| `t10_vec575` | 575 | 6 | 95 five times, then 100 |
| `t10_vec872` | 872 | 6 | 145 five times, then 147 |
| `t10_vec874` | 874 | 8 | 109 seven times, then 111 |
| `t9_vec1168` | 1168 | 8 | 146 eight times |
| `t10_vec1200` | 1200 | 10 | 120 ten times |
| `t9_vec3000` | 3000 | 22 | 136 twenty one times, then 144 |
| `t9_vec4096` | 4096 | 28 | 146 twenty seven times, then 154 |
| `t9_vec12000` | 12000 | 82 | 146 eighty one times, then 174 |
| `t10_vec20000` | 20000 | 134 | 149, then 183 |
| `t10_vec30000` | 30000 | 202 | 148, then 252 |

The count steps up by two every 299 bytes from 276: at 276, 575, 874,
1173 and so on, and `t10_vec574` against `t10_vec575` and
`t10_vec872` against `t10_vec874` are the pairs that pinned the step.
`t10_vec274` against `t10_vec275` pinned the first split, and
`t10_vec275` against `t10_vec276` showed the one size that gives two
chunks.
The first reading, chunks of at most 146 bytes with an even count, came
from the tier 9 sizes alone and was wrong twice over: `t10_vec20000`
has chunks of 149, and the count is even because the rule doubles it,
not because 146 divides anything.
What 24 and 299 are is open.
`t10_real40`, 40 reals of 8 bytes, is chunked exactly as
`t10_vec320`, 80 bytes four times, so the element type does not enter.

An arena is `0x800` handles and a value can be longer than that.
`t9_vec12000` spans arenas 0 to 6, and the page directory lists seven
arena records for one object.

*Found by* `t9_vec292` against `t9_vec256` and `t9_vec257`: the
widest value before tier 9 was the 32 bytes of `t2_array2d`, and the
first records longer than 257 bytes came back in pieces.
*Confirmed by* the reader reading all 55 wide values back against the
truth with the rule's addresses enforced.
`t9_int73` separates a byte split from an element split.

A page holds 10240 bytes: 20 bytes of header and then whole records.
The limit is the byte count, not the record count.

| Case | Value size | Record size | Records per full page | Pages |
| :--- | ---: | ---: | ---: | :--- |
| `t5_tr1000` | 1 | 17 | 600 | 600, 401 |
| `t6_tr1300` | 1 | 17 | 600 | 600, 600, 101 |
| `t7_int700` | 4 | 20 | 510 | 510, 191 |
| `t7_wide700` | 8 | 24 | 425 | 425, 276 |
| `t9_tr70000` | 1 | 17 | 600 | 600 times 116, then 401 |

`20 + 600 * 17` is 10220, `20 + 510 * 20` is 10220, and `20 + 425 * 24`
is 10220: in each case the next record would not fit in 10240.

*Found by* `t5_tr1000`, whose 1001 records split 600 and 401.
That fitted a byte limit and a record limit equally well.
*Confirmed by* `t7_int700` and `t7_wide700`, written to separate the
two, which split at 510 and 425.

The page that overflowed was written before the end of the simulation,
so it sits before the marker in the file, and the pages written at the
end follow the marker.
`t5_tr1000` has page 0 at `0x133b`, the marker at `0x1bac`, and page 1
at `0x1bbc`.
`t6_tr1300` has pages 0 and 1 first, the marker at `0x23f8`, and page 2
at `0x2408`.
The page directory is written last and points at all of them, so a
reader that follows the directory does not care about the order.


## Value encodings

The value bytes of a record are the object's whole value after the
change.
There is no delta encoding.
The length is the declaration's value size, word 4 of the declaration
record, in every VHDL case.
Verilog values are stored differently, in word pairs and in partial
records; see the Verilog section below.

| Type | Encoding |
| :--- | :--- |
| enumeration | the literal's index in the type's list, 1 byte up to 256 literals and 4 bytes from 257 on |
| `integer` and its subtypes | int32 |
| `real` | float64 |
| `time` | int64, in the unit whose scale is 1: `ps` for `TIME`, `um` for `dist_t` of `t21_phys_user` |
| array | elements back to back, left index first |
| record | fields in declaration order, aligned, total rounded up to 8 |

Enumeration covers `bit`, `std_ulogic`, `boolean`, `character` and user
types, found by `t1_nine_state`, `t2_boolean`, `t2_character` and
`t2_enum`.
The integer encoding was found by `t2_integer`, with 165 stored as
`a5 00 00 00`.
`t2_real` stores 1.5 as `0x3ff8000000000000`.
`t2_time` stores `7 ns` as 7000.
`t21_phys_user` declares `units um; mm = 1000 um; m = 1000 mm` and
stores `3 mm` as 3000, so a physical value counts the base unit and the
type entry gives every unit's scale in it.
`t21_int_neg` stores -165 as `5b ff ff ff`, two's complement, and
`t21_real_neg` stores -1.5 as `0xbff8000000000000`.
An integer subtype `range 0 to 7`, `t21_int_sub`, and a new integer
type of the same range, `t21_int_newtype`, are 4 bytes like `INTEGER`,
and their entries carry the narrow bounds under the subtype's name;
the two cases are byte identical outside the noise mask.
`bit_vector(7 downto 0)`, `t21_bitvec8`, is one byte per `BIT` like a
`std_ulogic_vector`.
Arrays were found by `t1_vec8`, `t2_array2d` and `t5_int_arr`.
Records were found by `t2_record`, `t5_rec_real` and `t5_arr_rec`.

`std_ulogic` indexes are `U X 0 1 Z W L H -`, so `0` is 2 and `1` is
3.
That is the order in the type table and the order in the IEEE source.

A type of 300 literals stores its index as a little endian `uint32`:
`t20_enum_300` holds `e299` as `2b 01 00 00`, and `t20_enum_256` and
`t20_enum_257` put the boundary between 256 and 257 literals.
The type entry's last word carries the size, so the reader does not
count literals.
Inside a record such a value is aligned to 4: the `w` field of
`t20_enum_300_rec` sits at offset 4 after a `std_ulogic`, and the
record is 8 bytes.
An array of the wide type is 4 bytes per element, `t20_enum_300_arr`.

A 2D array is stored with the first index outermost.
`t2_array2d` is `array (0 to 3) of std_ulogic_vector(7 downto 0)`, 32
bytes, and row 0 comes first, with its bit 7 first.

Alignment inside a record was separated by the tier 5 cases:

| Case | Record | Layout | Size |
| :--- | :--- | :--- | ---: |
| `t5_arr_rec` | `std_ulogic`, `integer` | `0`, `4` | 8 |
| `t5_rec_real` | `std_ulogic`, `real`, `integer` | `0`, `8`, `16` | 24 |
| `t5_rec_sub5` | record of 5 bytes, `std_ulogic` | `0`, `8` | 16 |
| `t2_record_nested` | record of 9 bytes, `std_ulogic` | `0`, `16` | 24 |

So an integer is aligned to 4, a real to 8, and a record field is
aligned to 8 whatever it holds.
The total is rounded up to 8.
An array of records uses that rounded size as its stride: the three
`pair_t` of `t5_arr_rec` are at 0, 8 and 16.

A record field of record type carries an extra range triple `(0, 8, 1)`
in the type table, and the 8 matches the alignment.
Whether it is the alignment or something else is open, because both a
5 byte and a 9 byte inner record produce 8.
See [types.md](types.md).


## Which objects get records

| Object | Records |
| :--- | :--- |
| signal | one at time 0 with the initial value, then one per change |
| port connected to a signal | one at time 0 on the signal's handle, then nothing of its own |
| port left `open` | as a signal, on its own handle |
| generic | one at time 0 with the value |
| constant of an architecture | one at time 0 with the value |
| loop index of a process `for` loop | one at time 0 holding 0 |
| loop index of a `for generate` | one at time 0 holding the iteration's value |
| port bound to a slice of a signal | one at time 0 on the signal's handle plus the slice's byte offset, then nothing of its own |
| port bound to a literal | as a signal, on its own handle |
| signal of a block | as a signal |
| variable declared in a process | none |
| constant of a package | none |
| signal of a package | none under `log_wave -recursive *`; as a signal under `log_wave -recursive /sig_pkg` |
| parameter of a SystemVerilog package | none |
| a `signal` parameter of a procedure | the change twice, on the signal's handle |
| signal of a null range | none, marked not logged |
| `std_logic` signal with two drivers | the initial value once per driver at time 0, then one per change |
| signal read through an external name | its change twice at the same time |
| the implicit signal of `'delayed`, `'stable`, `'quiet` or `'transaction` | none: there is no object |

A signal that never changes has exactly one record, found by
`t0_bit_const`.
The package signal row depends on the xsim script.
`t9_pkg_sig` runs the default script, `log_wave -recursive *`, and
its `sig_pkg.g` is an object marked not logged, with no records, and
arena 0 is unused.
`t13_pkg_log_all` runs the same design with
`log_wave -recursive /sig_pkg` added, and `g` is marked logged, holds
`0` at time 0 and `1` at 10 ns in arena 0 at its handle `0x768`, and
the logged range table holds `[0 1]` where `t9_pkg_sig` holds `[0 0]`.
The handle space is `0x1430` in both, so logging changes the records
and the marks, not the layout.
`get_objects -r /*` in that script lists `/tb/x` only, and
`get_objects /sig_pkg/*` lists `/sig_pkg/g`, so the default script
never sees the package.
The package parameter `p.W` of `t13_sv_pkg` is an object marked not
logged with no record under the default script.
`t15_sv_pkg_log` runs the same design with `log_wave -recursive /p`
added, and `W` is marked logged and holds one record at time 0,
`08 00 00 00 00 00 00 00`, the value 8 in the pair encoding of an
`int`, in arena 1 at its handle `0x890`; the logged range table holds
`[0 1]` where `t13_sv_pkg` holds `[0 0]`.
So a package parameter records like a module parameter once the
package is logged.

*Found by* `t13_pkg_log_all` against `t9_pkg_sig`, the same design
under two scripts.
*Confirmed by* `t15_sv_pkg_log` against `t13_sv_pkg`, the same for a
SystemVerilog package parameter.
`t3_late` puts its only change at 1000 ns, and the page holds two
records, at 0 and 1000000.
`t7_gen_for` has one object `i` per generate iteration, and their
records hold 0, 1 and 2, where the process loop index of `t5_tr1000`
holds 0 whatever the loop later does.
The generate index is a constant of the elaborated design and is
written like a generic.
`t8_gen_if` shows an architecture constant `with_dut : boolean := true`
written the same way, one record at time 0 holding `01`, under the
same declaration kind `0x13`.
`t8_port_open` shows an open `in` port with a default `'1'` recording
`03` at time 0 and the open `out` port it drives recording `00`, the
`'U'` it starts from, and then `03` in the first delta, each on its own
handle.

`t9_gen_types` has generics of `boolean`, `string(1 to 3)`,
`std_ulogic_vector(3 downto 0)` and `real`, and each is an object with
one record at time 0 in the encoding of its type: `01`, `61 62 63`,
`03 02 03 02` and `0x3ff8000000000000`.
A generic is not an integer only thing.
`t9_port_expr` binds an `in` port to `'1'` and the port takes a handle
of its own with a record `03` at time 0, as an open port with a default
does.
`t9_block` puts a signal inside a `block` and it records like any
signal of an architecture.

The tier 9 slice cases bind a port to part of a signal:

| Case | Signal | Binding | Port | Offset |
| :--- | :--- | :--- | :--- | ---: |
| `t9_port_slice` | `std_ulogic_vector(1 downto 0)` | `a => x(0)` | `std_ulogic` | 1 |
| `t9_port_slice2` | `std_ulogic_vector(3 downto 0)` | `a => x(2 downto 1)` | `std_ulogic_vector(1 downto 0)` | 1 |
| `t9_port_sliceto` | `std_ulogic_vector(0 to 1)` | `a => x(0)` | `std_ulogic` | 0 |

The port carries the signal's handle and, in the instance record word
at `+20`, the byte offset of its first element from the signal's left
element, see [hierarchy.md](hierarchy.md).
The offset counts bytes from the left, not index values: `x(0)` of
`1 downto 0` is byte 1, and `x(0)` of `0 to 1` is byte 0.
The port's value is the signal's record bytes `[offset, offset +
size)`, and the reader reads it out of the signal's records.

*Found by* `t9_port_slice` against `t8_port_in`: the port's value came
back as the whole vector until the word at `+20` was read.
*Confirmed by* `t9_port_slice2` and `t9_port_sliceto`.

A package constant and a package signal are objects without records.
`t9_port_rec` declares `constant zero : pair_t` in a package, and the
object is listed with a handle and no record, and the logged ranges of
[container.md](container.md) skip it.
`t9_pkg_sig` declares `signal g : std_ulogic` in a package, and it
takes the first handle `0x768` and no record, because
`log_wave -r /tb` covers the design under `/tb` and a package is not
under it.
Whether a wider `log_wave` records it is not tested.

*Found by* `t9_port_rec` and `t9_pkg_sig` against `t2_record`, which
uses a package type but declares no package object.

A resolved signal writes its initial value once per driver.
`t24_two_drivers` has `signal r : std_logic := 'Z'` driven by two
processes, `p` writing `'1'` at 50 ns and `q` writing `'Z'` at time 0
and `'0'` at 70 ns, and `r` holds four records: `Z` twice at time 0,
`1` at 50 ns and `X` at 70 ns, the resolved value, where `s` with
one driver in the same file holds its initial value once.
So the count at time 0 is the driver count, as the Verilog net rule
below counts objects, and the value is the resolved one, not each
driver's.
`t24_ext_name` reads `tb.dut.s` from `tb` through an external name,
`a <= << signal .tb.dut.s : std_ulogic >>;`, and `tb.dut.s` holds
its change at 10 ns twice, `03` and `03`, where `t24_config_spec`
with the same child and no external name holds it once; `a` holds
its own change at 10 ns once.
A null range signal, `z : std_ulogic_vector(0 downto 1)` of
`t24_null_range`, is an object marked not logged with no record.
The `records` field of `truth.json` pins these counts.

*Found by* `t24_two_drivers`, `r` against `s` in one file, and
`t24_ext_name` against `t24_config_spec`.

The initial value at time 0 is the declared default or the type's
leftmost literal.
`t1_nine_state` starts at `U`, index 0.

The trailer's end time is the time of `std.env.stop`, not the last
change.
`t3_late` ends at 1010000 with its last record at 1000000, and
`t6_tr1300` ends at 1310000 with its last record at 1300000.


## Verilog values

Tier 11 repeats the value ladder in Verilog and SystemVerilog, and the
pages of a Verilog design differ from the VHDL pages in three ways: the
declared size is in bits, a record holds word pairs rather than bytes,
and a record may cover part of the value.
Every claim reproduces with `wdbcvt -dump` on the case named.

**Word pairs.**
A Verilog value of `n` bits takes `8 * ceil(n / 32)` bytes of record:
one pair `[u32 a][u32 b]` per 32 bits, pair 0 holding bits 31 to 0.
Bit `i` of the value is `a[i] + 2 * b[i]`, which indexes the four
literals of `logic`: `0`, `1`, `Z`, `X`.
Bits above the width are 0 in both words.

| Case | Value | Record |
| :--- | :--- | :--- |
| `t11_v_bit_edge` | `X` | `01 00 00 00 01 00 00 00` |
| `t11_v_bit_edge` | `1` | `01 00 00 00 00 00 00 00` |
| `t11_v_vec8` | `8'ha5` | `a5 00 00 00 00 00 00 00` |
| `t11_v_vec100` | 100 bits of `X` | 24 bytes of `ff`, then `0f 00 00 00 0f 00 00 00` |
| `t11_v_vec64x` | bit 40 set to `x` | `ef bf ad de 00 01 00 00` at handle plus 8 |

`t11_v_vec64x` is the case that separated the two words: the `a` word
kept bit 8 of the high half at 1 and the `b` word gained it, so `X` is
`a` and `b` both set, and `Z` is `b` alone, the order of the literal
list.
The declaration's size, word 4, is the width in bits, and the second
handle is the handle plus the record size in bytes; see
[hierarchy.md](hierarchy.md).

*Found by* `t11_v_bit_edge` against `t1_bit_one_edge`, whose one bit
takes 8 bytes of record where the VHDL bit takes 1.
*Confirmed by* `t11_v_vec33` and `t11_v_vec100`, 16 and 32 bytes, and
by every tier 11 record dividing by 8.

**Partial records.**
A record's key is the handle plus 8 times the index of the first pair
it holds, and its length is 8 times the number of pairs it holds.
A record that covers the whole value starts at the handle.
An assignment to part of a variable writes only the pairs it touched:
`s[40] = 1'bx` in `t11_v_vec64x` writes one pair at handle plus 8,
and each `m[i] = 8'h00` in `t11_v_mem8` writes the one pair its
element lives in.
So a Verilog record is one change, of the pairs it holds, and the
reader overlays each record on the value the earlier records built.
A value wider than 275 bytes of record is chunked; see below.

*Found by* `t11_v_mem8`, whose `initial` block writes eight elements
and produces nine records at time 0, eight of them 8 bytes long, four
at the handle and four at the handle plus 8.
*Confirmed by* `t11_v_vec64x`, and by `t11_v_mem2w40`, whose write of
`m[0]` lands on pairs 1 and 2 and whose write of `m[1]` on pairs 0
and 1.

**Chunks.**
The chunk rule above applies to a Verilog value's record bytes,
`8 * ceil(bits / 32)`, with the same constants as for VHDL: a value
over 275 bytes is split into `2 * ceil((size + 24) / 299)` chunks, and
a chunk that crosses an arena boundary is split there.

| Case | Bits | Record bytes | Chunks | Found |
| :--- | ---: | ---: | :--- | :--- |
| `t12_v_vec1088` | 1088 | 272 | one, split at `0x800` into 152 and 120 | the last whole record |
| `t12_v_vec1089` | 1089 | 280 | 4 of 70 | the threshold is the same 275 |
| `t12_v_vec2272` | 2272 | 568 | 4 of 142 | one byte under the next step |
| `t12_v_vec2304` | 2304 | 576 | 6 of 96 | the step |
| `t12_v_vec4800` | 4800 | 1200 | 10 of 120 | ten chunks |
| `t12_v_vec12000` | 12000 | 3000 | 22 of 136, the last 144 | the same chunks as `t9_vec3000` |

A chunk boundary falls inside a word pair when the chunk size is not a
multiple of 8: 70 and 142 both split a pair between two chunks, and
the reader joins the chunks of a time before it reads the pairs.
A chunk holds the value's bytes in order, pair 0 first, so a whole
write of a chunked value is its chunks at the chunk addresses of the
VHDL rule, `handle + chunk * chunk_size`, with the arena splits on
top.
`t12_v_vec12000` spans three arenas: arena 0 holds a chunk of 136 and
16 bytes of the next, arena 1 the remaining 120 then fourteen of 136
and 24 bytes of the next, and arena 2 the remaining 112, four of 136,
and the last of 144.

A partial write into a chunked value is not chunked.
It is the 8 byte pair record of the partial records rule, at the
handle plus 8 times the pair index, wherever that address falls:
`t12_v_vec4800x` sets bit 2400 at 75 ns and writes one 8 byte record
at the handle plus 600, in arena 1.
`t12_v_mem40w32`, `reg [31:0] m [0:39]`, 320 bytes in four chunks of
80, writes `m[i]` one element per ns after its whole `X` write, and
each is an 8 byte record.
So a record on a chunked value is either a chunk piece, at a piece
address with a piece length, or some whole pairs, and the reader
classifies each record by address and length.
The two shapes meet at an arena boundary: the 8 byte rest of the
chunk split at `0x800` in `t12_v_mem40w32` has the address and length
of a write of `m[21]`.
The reader counts the records at each piece address in a time and
takes the smallest count as the number of whole writes, so a surplus
record at the split address is read as a pair write.

*Found by* `t12_v_vec1089` against `t12_v_vec1088`, 280 bytes of
record in four records of 70 where 272 bytes were two records of 152
and 120.
*Confirmed by* the other four wide cases, and by `t12_v_vec12000`
chunking as `t9_vec3000`, the same 3000 bytes in VHDL.
The pair write inside a chunked value was found by `t12_v_vec4800x`
against `t12_v_vec4800`, and the split rest ambiguity by
`t12_v_mem40w32`, which the first reader refused with two records at
`0x800` against one at the other chunk addresses.

**Order within a time.**
Three writes to one variable in one time step are three records in
write order: `t13_v_same_t` writes `1`, `2` and `3` to an 8 bit `reg`
at time 0 and the arena holds `X`, `01`, `02`, `03`.
An arena's records are in write order, and a time step's writes that
land in two arenas come back in arena order, not write order.
`t12_v_mem40_t0` writes `m[0]` to `m[39]` at time 0 after the whole
`X` write; `m[0]` is the top pair, at the handle plus 312, in arena 1,
and `m[21]` to `m[39]` are in arena 0.
The file holds the arena 0 records first, so a reader that replays
the arenas in file order sees `m[21]` to `m[39]` written before `m[0]`
to `m[20]`.
The values are unaffected, because no pair is written twice, but a
design that writes one pair twice in one time step across an arena
boundary would be read wrongly, and nothing in the file says which
write came first.
The test for the case compares the final value of each time rather
than the sequence.

*Found by* `t12_v_mem40_t0` against `t12_v_mem40w32`, the same forty
writes at one time and at forty times.

**Z and X.**
`t12_v_vec8_z` writes `8'bz0z1xx01` and records `1d 00 00 00 ac 00 00
00`: `a` is `0001 1101` and `b` is `1010 1100`, so a `Z` bit is `b`
alone, an `X` bit is both, and each bit is independent of its
neighbours.

*Found by* `t12_v_vec8_z` against `t11_v_vec8`.
*Confirmed by* `t11_v_vec64x`, one `X` bit.

**Where the elements go.**
A vector's leftmost bit is its most significant, whatever the declared
direction: `reg [0:7]` in `t11_v_vec8_asc` records `8'ha5` exactly as
`reg [7:0]` in `t11_v_vec8` does.
A memory is contiguous bits with its leftmost element at the top, so
`m[0]` of `reg [7:0] m [0:3]` is bits 31 to 24 and `m[3]` is bits 7 to
0, and `m[0]` of `reg [7:0] m [3:0]` is bits 7 to 0.
Elements are not padded: `reg [4:0] m [0:3]` is 20 bits in one pair,
and `reg [39:0] m [0:1]` is 80 bits in three pairs with `m[1]`
straddling pairs 0 and 1.
A two dimensional packed array, `logic [1:0][3:0]`, is 8 bits with
element `[1]` at the top.
An unpacked array of packed structs is contiguous the same way:
`s_t m [0:1]` over `struct packed { logic a; logic [3:0] b; }` in
`t13_sv_struct_ar` is 10 bits, `m[0]` in bits 9 to 5, and records
`ff 03` for all `X`, then `1a` for `m[1] = '{1, 4'b1010}`.
A typedef of an unpacked array, `t13_sv_tdef_ua`, changes nothing in
the records: `arr_t m` over `typedef logic [3:0] arr_t [0:1]` is the
8 bits of `logic [3:0] m [0:1]`.
An unpacked array of `real` is the exception: `real r [0:1]` in
`t13_sv_real_arr` declares 64 bits and takes one pair per element with
the last element lowest, as an unpacked struct does with its fields,
so `r[1] = 1.5` at 50 ns is an 8 byte record of pair 0 holding the
`float64`.
Its record at time 0 is 16 zero bytes, one for the whole array, and
the implicit process that sets `r[0] = 0.0` writes nothing, because
the value is already held.

*Found by* `t11_v_mem4` against `t11_v_mem4_desc`, the same four
writes under `[0:3]` and `[3:0]`, which moved the written byte from
the top to the bottom.
*Confirmed by* `t11_v_mem4w5`, `t11_v_mem3w5` and the `t11_v_mem2w*`
sweep from 9 to 64 bits per element, and by `t13_sv_struct_ar`,
`t13_sv_tdef_ua` and `t13_sv_real_arr` for the array forms above.

**Integral types and real.**
`integer`, `int`, `byte`, `longint` and `time` are vectors of their
width and record as such; `t11_v_integer` stores 165 as
`a5 00 00 00 00 00 00 00`, and the reader prints them in decimal,
signed except for `time`.
A `real` takes one pair holding the `float64`: `t11_v_real` stores 1.5
as `0x3ff8000000000000` and its declared size is 32.
A `shortreal` is the same: `t12_sv_shortreal` declares 32 bits with
the predefined `real` entry and stores 1.5 as the same `float64`, so
the 32 bit width of the source does not reach the database.
A `bit` type is two state and its `b` words are 0 throughout.

**Structs.**
A packed struct is a vector with its first field at the top:
`struct packed { logic a; logic [3:0] b; }` in `t11_sv_struct` records
`'{a: 1, b: 4'b1010}` as `1a`, `a` in bit 4.
`t11_sv_pstruct40` is 41 bits in two pairs, `a[39:0]` in bits 40 to 1
and `b` in bit 0.
An unpacked struct gives each field its own pairs, as many as the
field's width needs, and puts the last field lowest: in
`t11_sv_struct3`, `{a; b[3:0]; c[7:0]}` is 96 bits, with `c` in pair
0, `b` in pair 1 and `a` in pair 2.
A field wider than 32 bits takes the pairs a standalone value would:
`t11_sv_struct40` stores `a[39:0]` in pairs 1 and 2, low word first,
and `b` in pair 0.
A `real` field takes one pair: `t11_sv_struct_r` has `a` in pair 0
and the `float64` in pair 1.
The declared size follows: 96 for `t11_sv_struct3` and
`t11_sv_struct40`, 64 for `t11_sv_struct_r`, 41 for the packed
`t11_sv_pstruct40`.

*Found by* `t11_sv_ustruct` against `t11_sv_struct`, the same two
fields unpacked, which grew from one pair to two and put `b` first.
*Confirmed by* `t11_sv_struct3`, `t11_sv_struct40` and
`t11_sv_struct_r`.
A packed union records as one vector of its width: `t24_sv_union`
stores `u.b = 8'ha5` as `a5` in one pair, 8 bits declared, and both
fields read the same bits, `b` and `c` both `10100101`.
The VCD writes it as one `reg 8`.
*Found by* `t24_sv_union` against `t11_sv_struct`.

**Enums.**
A SystemVerilog enum records as its base type: `t11_sv_enum` stores
`DONE` as `02 00 00 00` in an `int`, and `t11_sv_enum4` stores `C`,
declared `4'd9`, as `09` in a 4 bit vector.
The reader looks the value up in the kind `0x04` entry to print the
name; see [types.md](types.md).

**Records at time 0.**
The first record of a Verilog variable holds its value before any
process runs: all `X` for a four state type, 0 for `bit`, `int`,
`byte`, `longint` and `real`.
An initializer in a `.v` file runs as an implicit `initial` process,
so a `reg s = 1'b0` records `X` and then `0` at time 0, and so does
every `.v` case with an initializer.
An `initial` block that writes at time 0 does the same:
`t11_v_mem8` records all `X`, then one element at a time.
A SystemVerilog `logic s = 1'b0` takes its initializer at declaration
and records `0` once, and a `bit`, `int` or struct with an initializer
records once too.
`t11_sv_enum4`, whose initializer gets an implicit process, records
`XXXX` and then `A`, while `t11_sv_enum` over `int` records `IDLE`
once, because the implicit process writes the value it already holds.
A `real` records `0` once for the same reason.
Without an initializer there is no implicit process: `t12_v_noinit`,
a `reg` first written at 50 ns, and `t12_sv_noinit`, the same with
`logic`, each hold one `X` record at time 0 and the write.
`t12_sv_enum_noin`, an enum over `int` without an initializer, holds
`IDLE` at time 0, the first literal, and no `X`.
A parameter holds one record at time 0 with its value:
`t12_v_params` records `K = 7`, `P = 8'h5a`, `Q = 9`, `R = 1.5` and
`L = 8`, and `t12_v_param64` a 64 bit value in two pairs, low word
first.
`t13_v_str_param` records `parameter P = "hello"` as 40 bits in two
pairs, `6f 6c 6c 65` low and `68` high, so the first character is at
the top and the last at bit 0.

The `X` record is absent when the variable's arena spills into a
second page, because the page holding time 0 was then written out
during the run, and such a page keeps one record per key and time; see
the page section above.
`t13_v_tr420`, a `reg clk = 1'b0` toggled every nanosecond for 420 ns,
holds 421 records in one page, `X`, `0` and 419 toggles.
`t13_v_tr430` holds 430 records over two pages, `0` and 429 toggles,
with no `X`, and `t13_v_tr2000` holds 2000 over five pages the same
way.
`t13_v_tr430_2` adds a `reg d` written once at 5 ns beside that clock;
`d` sits in an arena of its own, holds one page, and keeps its `X`,
`0`, `1`.
`t14_v_spill_d` puts a `reg d` without an initialiser into the clock's
arena, and `d` keeps its `X`, the only record of its key at time 0,
whether it is declared after the clock or before it,
`t14_v_spill_dfst`.
So the loss goes with a page written during the run, and the reader
takes the records as they are.
The same clock in VHDL, `t13_tr430`, holds 431 one byte records in one
page, because a VHDL record is 13 bytes and a Verilog one 24; see the
page table above.


*Found by* `t11_v_bit_edge` against `t1_bit_one_edge`, three records
where two were expected.
*Confirmed by* `t11_sv_logic` against `t11_v_bit_edge`, two records
and no implicit process; see [hierarchy.md](hierarchy.md).

**Nets.**
A `wire` records like a variable on its own handle: `t11_v_wire` has
`X`, `0`, `1` for both the `reg` and the wire that follows it.
A net shared by a wire and an output port holds one `X` record at
time 0 per object on it, then the value, as a VHDL net does: `y` and
`tb.dut.b` in `t11_v_port` share `0x768` and hold `X`, `X`, `1`, `0`.
The input port `a`, on its own handle, also holds two `X` records
before its `0`, though one object is on it.
When the input port is driven by a `wire`, `t12_v_port_wire`, the
port shares the wire's handle and the net holds three `X` records
before its `0`; the output side, `y` and `b`, holds two.
`t12_v_port_vec8` does the same with 8 bit ports.
An `output reg` port, `t12_v_port_reg`, keeps its own handle and holds
`X` then its value, while the wire it drives in the parent holds `X`
then the value, and the input port holds `X`, `X`, `0`.
So the count is one `X` per object on the handle plus one for an
input port, and the extra one was open until tier 16.
`t13_v_hier3_net` runs a net through three levels and holds three `X`
records on `w0`, a wire and the input port of `mid` at `0x768`, three
on `w1`, a wire of `mid` and the input port of `leaf` at `0x8e8`, two
on `w2`, a wire of `leaf` alone at `0x9a8`, and three on `y`, a wire
of `tb` and the output ports of `mid` and `leaf` at `0x828`.
Each of the four nets is driven by one `assign`.
An `inout` port, `t13_v_inout`, shares the wire's handle as an input
port does, with mode `0` in its declaration, and the net holds `X`,
`X`, then `Z` when the driver is `1'bz`, then `1`; the `reg` driver
holds `X`, `Z`, `1`.
An interface signal, `t13_sv_iface`, is on one handle under the
instance, `tb.b.d`, and under the child's interface port,
`tb.dut.p.d`, and holds `X`, `X`, `0`, `1`: one `X` per object and
nothing more, as a variable rather than a net.

The extra `X` goes with a reader of the net.
`t16_v_wire_rd1` adds `wire w2; assign w2 = w;` to `t11_v_wire`, and
`w` holds `X`, `X`, `0`, `1` where it held `X`, `0`, `1`, while `w2`,
which nothing reads, holds `X`, `0`, `1`.
`t16_v_wire_rd2` reads `w` from two assignments and `w` still holds
two `X` records, so the extra is one, not one per reader.
`t16_v_wire_rdp` reads `w` from `always @(w) q = w;` and `w` holds
two `X` records, so a process reader counts as an assignment does.
`t16_v_wire_rdi` connects `w` to the input port of a child that reads
nothing, and the net holds two `X` records for its two objects and
no extra, so a port connection alone is not a reader.
That is the input port of `t11_v_port`, read inside the child by
`assign b = ~a`, and every net of `t13_v_hier3_net`: `w0`, `w1` and
`w2` are each read by one `assign` and hold one `X` beyond their
objects, and `y`, which nothing reads, holds exactly one per object.
A `reg` gets no extra: `s` is read by `assign w = s` in every one of
these cases and holds one `X`.
So a net holds one `X` record at time 0 per object on its handle, and
one more when anything reads it, was the tier 16 reading.
The `truth.json` of these cases carries the count as `records` on
the signal, and `TestCorpus` holds the object to it.

*Found by* `t16_v_wire_rd1` against `t11_v_wire`, one record more.
*Confirmed by* `t16_v_wire_rd2`, `t16_v_wire_rdp` and
`t16_v_wire_rdi` against `t16_v_wire_rd1`, and the counts of
`t11_v_wire`, `t12_v_port_wire` and `t13_v_hier3_net` pinned in
their truths.

Tier 19 counted drivers and corrected the reading.
`t19_v_wand` drives a `wand` from two `assign` statements and nothing
reads it, and the net holds two `X` records, where the tier 16 rule
gave one.
`t19_v_wire_3drv` drives a `wire` from three assignments of one
value and holds two, not three, and `t19_v_wand_rd` adds a reader to
the two drivers and still holds two.
`t19_v_2drv_port` connects a wire with two drivers to an input port
that nothing reads and the net holds three: two objects and one more.
`t19_v_wire_nodrv` declares a wire with no driver and the net holds
one record, `Z`, and no `X`; `t19_v_nodrv_rd` reads that wire from an
assignment and the wire still holds the one `Z`, while the reading
wire holds `X` then `Z`.
So the records of a net before its first value are one per object on
its handle, holding the net's initial value, `X` when a driver exists
and `Z` for a `wire` with none, plus one more when the drivers and
readers of the net together number two or more.
A port connection from a `reg` counts as a driver, which is the
second `X` of the input port `a` in `t11_v_port`, and a connection to
a net joins the handles and counts as nothing.
The pulled and supply nets hold `X` then their value at time 0 with no
driver: `t19_v_tri0` and `t19_v_supply0` hold `X`, `0`, and
`t19_v_tri1` and `t19_v_supply1` hold `X`, `1`.
Why one extra record and not one per driver is open; the count is
what the corpus pins.

*Found by* `t19_v_wand` against `t11_v_wire`, two `X` records against
one with nothing reading either net.
*Confirmed by* `t19_v_wor`, `t19_v_triand` and `t19_v_trior` with the
same two drivers, `t19_v_wire_3drv`, `t19_v_wand_rd`,
`t19_v_2drv_port`, `t19_v_wire_nodrv` and `t19_v_nodrv_rd`, each
with its count pinned in `truth.json`, and by every earlier count
under the new rule.

**Writes of the value held.**
A Verilog write of the value already held produces no record, as a
VHDL one does not, see `t8_same` above.
`t17_v_reg_same` writes `s = 1'b0` at 50 ns to a `reg s = 1'b0` and
holds `X`, `0` and nothing at 50 ns.
`t17_v_net_same` drives `wire w` from `assign w = s & 1'b0` and
toggles `s` at 50 ns; `w` holds `X`, `0` and nothing at 50 ns, and
nothing more at time 0 either, though the assignment is evaluated
again when `s` takes its initial value.
`t17_v_mem_same` writes `m[2] = 8'h00` at 50 ns to an element that
holds `8'h00`, and the memory holds its five time 0 records and
nothing at 50 ns.
A nonblocking write records as a blocking one: `t17_v_reg_nb`, with
`s <= 1'b1`, holds the three records of `t11_v_bit_edge`, and
`t17_v_nb_swap`, with `a <= b; b <= a;` at 50 ns, holds one record
per `reg` at 50 ns with the swapped values.

*Found by* `t17_v_reg_same` against `t11_v_bit_edge`, two records
where the edge case holds three.
*Confirmed by* `t17_v_net_same` against `t11_v_wire`, and
`t17_v_mem_same` against `t11_v_mem4`; `t17_v_reg_nb` and
`t17_v_nb_swap` for the nonblocking form.

**What is not recorded.**
The `always #25 clk = ~clk` of `t11_v_always` is due at 100 ns, the
`$finish` time, and that toggle is not in the pages; the last `clk`
record is at 75 ns.
`t13_v_tr430` ends at 430 ns and its last record is at 429 ns, where
`t13_tr430`, the same clock in VHDL ended by `std.env.stop` at 430 ns,
records the toggle at 430 ns.
A `string` variable has no records, because it has no object; see
[hierarchy.md](hierarchy.md).
Vivado's VCD for a Verilog design is fuller than for VHDL: `real`,
`integer`, `time`, a struct and an enum all get a `$var`.
A memory and a `string` do not, so the `t11_v_mem*` cases have no VCD
guard and their `truth.json` is the only check.
A task's and a function's arguments and locals are recorded:
`t12_v_task` holds `v` and `tmp` of `task inc` as objects under
`tb.inc`, all `X` at time 0 and their values at the call, and
`t12_v_func` holds the return variable `inc` as well, written after
the locals.
See [hierarchy.md](hierarchy.md).


## What VCD cannot hold

Vivado's own VCD writer emits nothing for `boolean`, `integer`, `real`,
`time`, `character`, user enumerations, records or arrays other than
bit vectors.
Every one of those types is present in the pages with the encodings
above.
So the pages are the only record of those values, and the VCD guard in
[../provenance.md](../provenance.md) covers only the bit and vector
rows of the table.
For a Verilog design the VCD holds more, and `TestVCD` compares every
value it holds with the pages; see [vcd.md](vcd.md).

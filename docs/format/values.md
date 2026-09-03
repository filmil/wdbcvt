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
| `0` | 8 | `t0`, the time of the first record, in ps |
| `8` | 8 | `t1`, the time of the last record minus `t0` |
| `16` | 4 | `n`, the number of records |
| `20` | | `n` records, then zero padding to 10240 |

A record is:

| Offset | Size | Meaning |
| ---: | ---: | :--- |
| `0` | 8 | time in ps |
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

The time unit is the picosecond, and nothing finer is kept.
`t8_ps` waits `1 ps`, `998 ps` and `1500 fs`, and its records sit at
1, 999 and 1000, so the third wait was cut to 1 ps, and the end time is
11000.
That is the simulator's resolution as much as the file's: xsim runs the
corpus at its default 1 ps precision, and the VCD it writes agrees.

*Found by* `t8_ps` against `t1_bit_two_edges`.

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
| enumeration | 1 byte, the literal's index in the type's list |
| `integer` and its subtypes | int32 |
| `real` | float64 |
| `time` | int64 in ps |
| array | elements back to back, left index first |
| record | fields in declaration order, aligned, total rounded up to 8 |

Enumeration covers `bit`, `std_ulogic`, `boolean`, `character` and user
types, found by `t1_nine_state`, `t2_boolean`, `t2_character` and
`t2_enum`.
The integer encoding was found by `t2_integer`, with 165 stored as
`a5 00 00 00`.
`t2_real` stores 1.5 as `0x3ff8000000000000`.
`t2_time` stores `7 ns` as 7000.
Arrays were found by `t1_vec8`, `t2_array2d` and `t5_int_arr`.
Records were found by `t2_record`, `t5_rec_real` and `t5_arr_rec`.

`std_ulogic` indexes are `U X 0 1 Z W L H -`, so `0` is 2 and `1` is
3.
That is the order in the type table and the order in the IEEE source.

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
| signal of a package | none, under `log_wave -r /tb` |
| a `signal` parameter of a procedure | the change twice, on the signal's handle |

A signal that never changes has exactly one record, found by
`t0_bit_const`.
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
No Verilog value in the corpus is wide enough to test the chunk rule
above; the widest is `t11_v_vec100`, 32 bytes.

*Found by* `t11_v_mem8`, whose `initial` block writes eight elements
and produces nine records at time 0, eight of them 8 bytes long, four
at the handle and four at the handle plus 8.
*Confirmed by* `t11_v_vec64x`, and by `t11_v_mem2w40`, whose write of
`m[0]` lands on pairs 1 and 2 and whose write of `m[1]` on pairs 0
and 1.

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

*Found by* `t11_v_mem4` against `t11_v_mem4_desc`, the same four
writes under `[0:3]` and `[3:0]`, which moved the written byte from
the top to the bottom.
*Confirmed by* `t11_v_mem4w5`, `t11_v_mem3w5` and the `t11_v_mem2w*`
sweep from 9 to 64 bits per element.

**Integral types and real.**
`integer`, `int`, `byte`, `longint` and `time` are vectors of their
width and record as such; `t11_v_integer` stores 165 as
`a5 00 00 00 00 00 00 00`, and the reader prints them in decimal,
signed except for `time`.
A `real` takes one pair holding the `float64`: `t11_v_real` stores 1.5
as `0x3ff8000000000000` and its declared size is 32.
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
Why is open.

**What is not recorded.**
The `always #25 clk = ~clk` of `t11_v_always` is due at 100 ns, the
`$finish` time, and that toggle is not in the pages; the last `clk`
record is at 75 ns.
A `string` variable has no records, because it has no object; see
[hierarchy.md](hierarchy.md).
Vivado's VCD for a Verilog design is fuller than for VHDL: `real`,
`integer`, `time`, a struct and an enum all get a `$var`.
A memory and a `string` do not, so the `t11_v_mem*` cases have no VCD
guard and their `truth.json` is the only check.


## What VCD cannot hold

Vivado's own VCD writer emits nothing for `boolean`, `integer`, `real`,
`time`, `character`, user enumerations, records or arrays other than
bit vectors.
Every one of those types is present in the pages with the encodings
above.
So the pages are the only record of those values, and the VCD guard in
[../provenance.md](../provenance.md) covers only the bit and vector
rows of the table.

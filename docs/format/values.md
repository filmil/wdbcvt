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
of that arena, up to 100 pages.

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


## Overflow

A page holds 10240 bytes: 20 bytes of header and then whole records.
The limit is the byte count, not the record count.

| Case | Value size | Record size | Records per full page | Pages |
| :--- | ---: | ---: | ---: | :--- |
| `t5_tr1000` | 1 | 17 | 600 | 600, 401 |
| `t6_tr1300` | 1 | 17 | 600 | 600, 600, 101 |
| `t7_int700` | 4 | 20 | 510 | 510, 191 |
| `t7_wide700` | 8 | 24 | 425 | 425, 276 |

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
record, in every corpus case.

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
| variable declared in a process | none |

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

The initial value at time 0 is the declared default or the type's
leftmost literal.
`t1_nine_state` starts at `U`, index 0.

The trailer's end time is the time of `std.env.stop`, not the last
change.
`t3_late` ends at 1010000 with its last record at 1000000, and
`t6_tr1300` ends at 1310000 with its last record at 1300000.


## What VCD cannot hold

Vivado's own VCD writer emits nothing for `boolean`, `integer`, `real`,
`time`, `character`, user enumerations, records or arrays other than
bit vectors.
Every one of those types is present in the pages with the encodings
above.
So the pages are the only record of those values, and the VCD guard in
[../provenance.md](../provenance.md) covers only the bit and vector
rows of the table.

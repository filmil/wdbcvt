<!-- SPDX-License-Identifier: Apache-2.0 -->

# The type table

The section named by the `Xilinx RTTI` directory entry.
It starts with the magic `Xilinx ISim TYPE FILE 001`.
Reproduce every table below with `wdbcvt -dump`, which prints the
decoded entries under `types`.


## Framing

| Offset | Len | Content |
| :--- | ---: | :--- |
| `0` | 26 | `Xilinx ISim TYPE FILE 001`, NUL terminated |
| `26` | 2 | `uint16`, noise |
| `28` | 4 | `uint32` Unix time, noise |
| `32` | 4 | `uint32` number of type entries |
| `36` | 4 | `uint32` offset, from the magic, where the entries end |
| `40` | | the entries, back to back |
| end | 8 per type | `uint64` offset of each entry, from the magic |

Every entry is `[uint32 length][uint32 tag] name NUL body`.
The length covers the whole entry.
The low byte of the tag is the kind; the high bytes have been `0xa0`
in every entry.
The offset list at the end is the one cross reference the section
offers, and the decoder checks that it names every entry.

*Found by* the correlation sweep, which matched the word at `32` to the
number of type names in every case, once `TRUE` and `FALSE` were
classified as `BOOLEAN`'s literals rather than as types.
*Confirmed by* 49 of 49 cases decoding with the entry lengths chaining
exactly to the word at `36`.


## Kinds

| Tag | Kind | Body |
| :--- | :--- | :--- |
| `0x03` | enumeration | `[u32 2][u32 2][u32 class][u32 n]` then `n` NUL terminated literals, then `[u32 1]` |
| `0x05` | integer | `[u32 2][i32 low][i32 high][u32 1]` |
| `0x06` | real | `[u32 2][u32 0][f64 low][f64 high][u32 1]` |
| `0x0d` | physical | `[u32 0xa][u32 n]` then `n` times `name NUL [u64 scale]` |
| `0x10` | array | `[u32 2][u16 1][u16 0xa0][u32 element][u32 dims][u32 index][u32 1 or 2]` then range triples, then `-99` |
| `0x11` | record | `[u32 2][u16 1][u16 0xb][u32 n]` then `n` fields, then `-99` |

A range triple is `[i32 left][i32 right][i32 dir]`, with `dir` `1` for
`to` and `-1` for `downto`.
The trailing `-99` is `0xffffff9d` and closes the triple list.

**Enumeration.**
The class word is `2` for `BIT`, `3` for `STD_ULOGIC`, and `5` for
`BOOLEAN`, `CHARACTER` and a user enumeration.
`std_ulogic` is not a builtin to this format.
It is an ordinary enumeration whose nine literals are written out as
`'U' 'X' '0' '1' 'Z' 'W' 'L' 'H' '-'`, and a user type
`(crimson, viridian, cobalt)` is written the same way.
`CHARACTER` is an enumeration of 256 literals, which is why `t2_character`
is 1461 bytes larger than the one bit baseline where the other scalar
types cost about 400.

*Found by* `strings -a -t d` on `t1_bit_one_edge`, which shows
`STD_ULOGIC` followed by the nine literals 4 bytes apart.
*Confirmed by* `t2_enum`, whose literal names appear verbatim, and by
the class word being the only difference between `BIT` and a user
enumeration's entry shape.

**Integer.**
`INTEGER` is `-2147483648 to 2147483647` and `NATURAL` is
`0 to 2147483647`.
A subtype with its own bounds is a separate entry.

**Real.**
`REAL` is `-1e308 to 1e308`.

**Physical.**
`TIME` lists eight units with their size in picoseconds:
`fs=0 ps=1 ns=1000 us=1000000 ms=1000000000 sec=1000000000000
min=60000000000000 hr=3600000000000000`.
So the database's time unit is the picosecond, and `fs` rounds to zero.
There is no trailer word.

*Found by* `t2_time`, the only case with a physical type.

**Array.**
`element` and `index` are indexes into this table.
The constraint word is `1` for an unconstrained type such as
`STD_ULOGIC_VECTOR`, whose one triple is `(0, 0, -2)`, and `2` for a
constrained type, whose triples are the bounds.
A vector signal's own bounds are not here; they are in the declaration
record in the debug section.
`t2_array2d` declares `array (0 to 3) of std_ulogic_vector(7 downto 0)`
and its entry holds `(0 to 3) (7 downto 0)`: two triples, one per
dimension of the flattened shape.

**Record.**
Each field is `name NUL [u32 type][u32 nranges]` then that many
triples.
A field of a scalar type has no triples.
A field of a vector type carries its bounds, `(7 downto 0)`.
A field whose type is itself a record carries one extra triple first,
`(0, 8, 1)`, before the inner record's own triples.
It is `8` for an inner record of one scalar and one 8 bit vector, and
still `8` for one scalar and one 4 bit vector, so it does not count
scalars or bits.
Its meaning is open.

*Found by* `t2_record_nested` against `t2_record2`.
*Confirmed by* `t5_rec_sub5`, written to move the inner width, which
left the `8` unchanged.


## What the earlier size measurements meant

Before the entries were decoded, the table was studied through file
sizes, and those readings still hold:

* `t2_unsigned8` and `t2_signed8` differ by exactly 2 bytes, the
  difference between the names `unsigned` and `signed`.
  Names are stored one byte per character, NUL terminated.
  This was measured twice, because the first pair also differed by two
  characters of directory name; see [../corpus.md](../corpus.md).
* `t2_slv8` costs 447 bytes more than `t1_vec8`, and that was first
  read as the cost of a resolved type.
  It is not.
  The two type tables have the same three entries, and the difference
  is one byte of name.
  The 447 bytes are two source files in the debug section's file table:
  `t2_slv8` also uses `ieee.numeric_std`, and each file costs its
  compile path and its local path.
  A `use` clause is a confound to hold fixed; see
  [../corpus.md](../corpus.md).
* `t2_record_two` against `t2_record2`: two signals of one record type
  hold the type once.
  `grep -aoc alpha sim.wdb` prints `1`.
  Types are shared between signals, and the declaration record points
  at the type by index.
* `t2_record` is 1447 bytes smaller than `t2_flat3`, three signals with
  the same fields.
  A record is one object, and its field names live here, not beside the
  signal.

The table is a flat list in dependency order.
An outer record comes first, then the inner record, then the leaf
types, and a field or an element refers to a later entry by index.

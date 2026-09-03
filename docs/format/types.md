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
The first word of the body, called the origin below, says which
language the type came from.
It is `2` for a VHDL type and `0xa` for VHDL `TIME`, and was read as a
constant `2` until tier 11 simulated Verilog; see the Verilog section
below for the other values.
The offset list at the end is the one cross reference the section
offers, and the decoder checks that it names every entry.

*Found by* the correlation sweep, which matched the word at `32` to the
number of type names in every case, once `TRUE` and `FALSE` were
classified as `BOOLEAN`'s literals rather than as types.
*Confirmed by* 281 of 281 cases decoding with the entry lengths chaining
exactly to the word at `36`.


## Kinds

| Tag | Kind | Body |
| :--- | :--- | :--- |
| `0x03` | enumeration | `[u32 origin][u32 variant][u32 class][u32 n]` then `n` NUL terminated literals, then `[u32 1]` for VHDL or `[u32 0]` for Verilog |
| `0x04` | named values | `[u32 origin][u32 base][u32 n][u32 8]` then `n` times `name NUL [u64 value]`, then `[u32 nranges]` and that many triples, no `-99` |
| `0x05` | integer | `[u32 origin][i32 low][i32 high][u32 1]` |
| `0x06` | real | `[u32 origin][u32 variant][f64 low][f64 high][u32 1]` for VHDL, `[u32 0]` for Verilog |
| `0x07` | alias | `[u32 origin][u32 target][u32 nranges]` then that many range triples |
| `0x0d` | physical | `[u32 origin][u32 n]` then `n` times `name NUL [u64 scale]` |
| `0x10` | array | `[u32 origin][u16 layout][u16 0xa0][u32 element][u32 dims]` then `dims` index type words, then `[u32 nranges]` and that many range triples, then `-99` |
| `0x11` | record | `[u32 origin][u16 layout][u16 0xb][u32 n]` then `n` fields, then `-99` |

Kinds `0x04` and `0x07` come from SystemVerilog only, and are described
with the other Verilog entries below.

A range triple is `[i32 left][i32 right][i32 dir]`, with `dir` `1` for
`to` and `-1` for `downto`.
The trailing `-99` is `0xffffff9d` and closes the triple list.

**Enumeration.**
The variant word is `2` for every VHDL enumeration.
The class word is `2` for `BIT`, `3` for `STD_ULOGIC`, and `5` for
`BOOLEAN`, `CHARACTER` and a user enumeration.
`std_ulogic` is not a builtin to this format.
It is an ordinary enumeration whose nine literals are written out as
`'U' 'X' '0' '1' 'Z' 'W' 'L' 'H' '-'`, and a user type
`(crimson, viridian, cobalt)` is written the same way.
`CHARACTER` is an enumeration of 261 literals, which is why `t2_character`
is 1461 bytes larger than the one bit baseline where the other scalar
types cost about 400.

*Found by* `strings -a -t d` on `t1_bit_one_edge`, which shows
`STD_ULOGIC` followed by the nine literals 4 bytes apart.
*Confirmed by* `t2_enum`, whose literal names appear verbatim, and by
the class word being the only difference between `BIT` and a user
enumeration's entry shape.

An enumeration entry is named after the subtype the signal is declared
with: a `std_logic` signal gets an entry named `STD_LOGIC` with the nine
`STD_ULOGIC` literals, `t8_port_inout` against `t8_port_in`, and
nothing else in the file names `STD_ULOGIC`.

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

The origin word of `TIME` is `0xa`, which no other VHDL type has.

*Found by* `t2_time`, the only VHDL case with a physical type.

**Array.**
`element` and `index` are indexes into this table.
The word before the triples counts them.
An unconstrained type such as `STD_ULOGIC_VECTOR` has one triple,
`(0, 0, -2)`, and a constrained type has one triple of bounds per
dimension.
Through tier 10 the word read as `1` for unconstrained and `2` for
constrained, because every constrained VHDL array in the corpus had two
dimensions, and the tier 11 `time` entry with one constrained triple and
a `1` in the word corrected that.
A vector signal's own bounds are not here; they are in the declaration
record in the debug section.
`t2_array2d` declares `array (0 to 3) of std_ulogic_vector(7 downto 0)`
and its entry holds `(0 to 3) (7 downto 0)`: two triples, one per
dimension of the flattened shape.

**Index dimensions.**
The `dims` word counts the index dimensions of the type itself, and as
many index type words follow it, one per dimension.
`t2_array2d` has `dims` `1` and one index word, `3` for `natural`,
under two triples, because its second triple belongs to the element.
`t18_arr_2dim` declares `array (0 to 1, 0 to 2) of std_ulogic`, and
its entry has `dims` `2`, two index words `1 1`, then `2` triples
`(0 to 1) (0 to 2)`.
`t18_arr_3dim` has `dims` `3`, three index words and three triples.
Every entry through tier 17 had one index dimension, which is why the
word read as a constant `1` and the index word as a fixed field.
The value of a multidimensional array is its elements in row major
order, the last index fastest, one byte per `std_ulogic`: the
`t18_arr_2dim` record `03 02 03 02 03 03` is `(('1','0','1'),
('0','1','1'))`.

The index word names the integer entry of the index subtype:
`INTEGER` in `t18_arr_2dim`, `NATURAL` for `STD_ULOGIC_VECTOR`.
The triples are the type's own index ranges followed by the
constraints of an unconstrained element: `t19_arr_2d_vec`, an
`array (0 to 1, 0 to 2) of std_ulogic_vector(3 downto 0)`, has `dims`
`2` and three triples, `(0 to 1) (0 to 2) (3 downto 0)`, and its
value is 24 bytes in row major order with each vector's leftmost
element first.
An element type that is constrained adds nothing: `stack_t`, an
`array (0 to 1) of mat_t` in `t19_arr_of_2dim`, has `dims` `1` and
the one triple `(0 to 1)`, while its declaration record carries all
three ranges `(0 to 1) (0 to 1) (0 to 2)`.

*Found by* `t18_arr_2dim` against `t2_array2d`, where the reader ran
out of bytes 4 words early on the entry.

*Confirmed by* `t18_arr_3dim`, `t19_arr_2d_vec` and `t19_arr_of_2dim`,
and by `truth.json` of each against the decoded value.

**Record.**
Each field is `name NUL [u32 type][u32 nranges]` then that many
triples.
A field of a scalar type has no triples.
A field of a vector type carries its bounds, `(7 downto 0)`.
A field whose type is itself a record carries the inner record's
constraint, flattened: one triple per inner field, in the inner
record's field order.
An inner vector field contributes its bounds.
An inner scalar field contributes its type's range: `(0, 8, 1)` for
`std_ulogic`, whose nine literals are numbered 0 to 8, `(0, 1, 1)` for
`bit`, and `(-2147483648, 2147483647, 1)` for `integer`.
An inner `real` field contributes nothing, so the list can be shorter
than the field count.
The list is written only when the inner record has at least one array
field.
An inner record of scalars alone, whatever its size, gives the outer
field `nranges` 0.

| Case | Inner record | Triples on the outer field |
| :--- | :--- | :--- |
| `t5_rec_sub5` | `std_ulogic`, `std_ulogic_vector(3 downto 0)` | `(0 to 8) (3 downto 0)` |
| `t7_rec_vfirst` | `std_ulogic_vector(3 downto 0)`, `std_ulogic` | `(3 downto 0) (0 to 8)` |
| `t7_rec_in2v` | `std_ulogic`, two vectors | `(0 to 8) (3 downto 0) (1 downto 0)` |
| `t7_rec_bitv` | `bit`, a vector | `(0 to 1) (3 downto 0)` |
| `t7_rec_intv` | `integer`, a vector | `(-2147483648 to 2147483647) (3 downto 0)` |
| `t7_rec_in2` | two `std_ulogic` | none |
| `t7_rec_in16` | `std_ulogic`, `integer`, `real` | none |
| `t8_rec_realv` | `real`, `std_ulogic_vector(3 downto 0)` | `(3 downto 0)` |

*Found by* `t2_record_nested` against `t2_record2`, which showed the
extra `(0, 8, 1)`.
`t5_rec_sub5` moved the inner width and left the `8` unchanged, which
ruled out a size.
`t7_rec_in2` and `t7_rec_in16` removed the vector and the triple went
with it.
`t7_rec_vfirst` moved the vector before the scalar and the `(0, 8, 1)`
moved after the bounds, which made it a per field entry.
`t7_rec_bitv` and `t7_rec_intv` replaced the scalar and the triple
became that scalar's range.
*Confirmed by* `t7_rec_in2v`, whose two vectors and one scalar give
exactly three triples in field order.
`t8_rec_realv` put a `real` beside the vector and got the vector's
bounds alone, so a `real` has no range to contribute and is skipped.

## Verilog and SystemVerilog types

Tier 11 of the corpus repeats the type ladder in Verilog and
SystemVerilog.
The source language reaches the table in three places: the origin
word, the layout half word of an array or record, and two kinds, `0x04`
and `0x07`, that no VHDL case produces.
Every case below reproduces with `wdbcvt -dump`.

**Origin.**
The first word of an entry body:

| Origin | Types | Found by |
| ---: | :--- | :--- |
| `0x2` | every VHDL type but `TIME` | tiers 1 to 10 |
| `0xa` | VHDL `TIME` | `t2_time` |
| `0x1` | a Verilog type without a name of its own: a vector, a memory, a struct, an enum, a typedef | `t11_v_vec8`, `t11_sv_struct` |
| `0x5` | a Verilog predefined type: `logic`, `bit`, `real`, `scalar_int`, `integer`, `int`, `byte`, `longint` | `t11_v_bit_edge` |
| `0xd` | Verilog `time` | `t11_v_time` |

Read as bits: bit 1 is VHDL, bit 0 is Verilog, bit 2 is a predefined
type and bit 3 is a time type.
That is a reading, and only these five values have been seen.
The reader keeps the word as `Type.Origin` and refuses any other value.

*Found by* `t11_v_bit_edge` against `t1_bit_one_edge`, the same one
transition design in the two languages, whose one type entry begins
with `5` where the VHDL one begins with `2`.
*Confirmed by* every tier 11 case: the word is `1`, `5` or `0xd` in all
44 and never `2`.

**Layout.**
The `u16` after the origin of an array or record entry is `1` for a
VHDL type, `3` for a packed Verilog type and `2` for an unpacked one.
A vector `reg [7:0]` is packed.
A memory `reg [7:0] m [0:3]` is an unpacked array of a packed vector,
and its entry says `2` while the vector's says `3`.
A `struct packed` says `3` and a plain `struct` says `2`.

*Found by* `t11_sv_struct` against `t11_sv_ustruct`, the same two
field struct packed and unpacked, which differ in this half word and
in the declaration size.
*Confirmed by* `t11_v_mem4`, whose two array entries carry both values.

**Scalars.**
`logic` and `bit` are enumerations of four literals, with a different
variant word:

| Type | Entry |
| :--- | :--- |
| `logic` | `[5][0][1][4]` `0 1 Z X` `[0]` |
| `bit` | `[5][1][1][4]` `0 1 0 0` `[0]` |
| `real` | `[5][1][f64 0][f64 0][0]` |
| `scalar_int` | `[5][-2147483647][2147483647][1]` |

`logic` is the type of a `reg`, a `wire` and a SystemVerilog `logic`
alike.
`bit` lists its two literals and then two more `0`, so the literal list
is four long for both and the variant tells them apart.
What the variant and class words mean beyond that is open.
`real` has no bounds: both are `0`.
`scalar_int` is the index type of every Verilog array, as `NATURAL` is
the index type of `STD_ULOGIC_VECTOR`, and its low bound is
`-2147483647`, one above `INTEGER`'s.

**Vectors and integral types.**
A vector such as `reg [7:0]` is an array entry with an empty name,
origin `1`, layout `3`, element `logic`, index `scalar_int`, and one
triple `(0, 0, -2)`.
The bounds are in the declaration record, as for a VHDL vector.
So every unnamed vector in a design shares one entry, whatever its
width or direction: `t11_v_vec8`, `t11_v_vec8_asc`, `t11_v_vec33`,
`t11_v_vec100` and `t11_v_two_w64` all hold the same three entries.
The named integral types are the same array entry with a name and
their own bounds:

| Type | Origin | Element | Range | Found by |
| :--- | ---: | :--- | :--- | :--- |
| `integer` | `0x5` | `logic` | `(31, 0, -1)` | `t11_v_integer` |
| `time` | `0xd` | `logic` | `(63, 0, -1)` | `t11_v_time` |
| `int` | `0x5` | `bit` | `(31, 0, -1)` | `t11_sv_int` |
| `byte` | `0x5` | `bit` | `(7, 0, -1)` | `t11_sv_byte` |
| `longint` | `0x5` | `bit` | `(63, 0, -1)` | `t11_sv_longint` |

A declaration of a named type has no range records of its own.
Signedness is not recorded anywhere: the dump of `t11_v_signed8`,
`reg signed [7:0]`, is the dump of `t11_v_vec8` outside the
timestamps, and the two files differ in 17 bytes.

*Found by* `t11_v_integer` against `t11_v_vec8`, where the vector's
bounds moved from the declaration into the type entry and took a name.

**Two dimensional packed arrays.**
`logic [1:0][3:0]` in `t11_sv_arr2d` is an array of an array: the
outer entry has `dims` `1`, element the inner vector entry, and two
`(0, 0, -2)` triples, and the declaration carries `(1 downto 0)
(3 downto 0)` and size 8.
A memory `reg [7:0] m [0:3]` is the same shape with layout `2` on the
outer entry.

**Unpacked dimensions.**
Each unpacked dimension is an array entry of its own, layout `2`,
whose element is the entry for the dimensions inside it.
`logic [3:0] m [0:1][0:2]` in `t12_sv_unp2d` has three entries: the
shared vector entry, an entry of layout `2` over it with two
`(0, 0, -2)` triples, and an entry of layout `2` over that with three.
The declaration points at the outermost and carries every bound,
`(0 to 1) (0 to 2) (3 downto 0)`, with size 24.
An entry's triple count is one more than the entry it wraps, so the
count says how deep the entry is, and the triples themselves hold no
bounds.
The value is contiguous bits with `m[0][0]` at the top; see
[values.md](values.md).

A typedef of an unpacked array, `typedef logic [3:0] arr_t [0:1]` in
`t13_sv_tdef_ua`, is the alias of the outer entry, and the alias holds
every bound, the unpacked one first: `(0 to 1) (3 downto 0)`.
The declaration of an `arr_t` variable has size 8 and no ranges, as
the declaration of a `byte_t` variable has none below.
So the rule of the next section holds for any typedef: the bounds go
with the name.

*Found by* `t12_sv_unp2d` against `t11_v_mem4`, which has one unpacked
dimension and two entries.
*Confirmed by* `t13_sv_tdef_ua` against `t12_sv_unp2d`, which moved
the bounds from the declaration into the alias.

**A typedef of a vector.**
`typedef logic [7:0] byte_t` in `t12_sv_typedef` is an alias entry
whose third word is `1` and which is followed by one triple,
`(7, 0, -1)`.
The vector entry it names is the shared unnamed one, and the
declaration of a `byte_t` variable has no ranges of its own and size
8.
So the bounds of a named vector type live in the alias, where the
bounds of an anonymous vector live in the declaration, and the reader
takes the ranges from the innermost alias that has any when the
declaration has none.
The aliases of `t11_sv_struct` and `t11_sv_enum` count 0 ranges,
which is what the earlier reading, `[1][target][0]`, saw.

A typedef declared in a package is the same entry: `byte_t` of the
package `p` in `t13_sv_pkg` is an alias with `(7 downto 0)`, and the
variable of `tb` that uses it declares 8 bits and no ranges.
A typedef of a struct is an alias with no ranges wrapping the record
entry, and an unpacked array of that typedef is an array entry of
layout `2` over the alias: `s_t m [0:1]` in `t13_sv_struct_ar` has
the record, the alias `s_t` and the array in that order, and the
declaration carries `(0 to 1)` with size 10.

*Found by* `t12_sv_typedef` against `t11_v_vec8`, whose declaration
carries `(7 downto 0)` and whose type has no alias.
*Confirmed by* `t13_sv_pkg` and `t13_sv_struct_ar`.

**Parameters.**
A parameter's type follows its declaration: `parameter K = 5` and
`parameter [7:0] P` use the unnamed vector entry, with 32 and 8 bits
and a declaration range, `parameter integer Q` uses the `integer`
entry with 32 bits and no range, and `parameter real R` uses the
`real` entry.
The `real` parameter declares 16 bits, where a `real` variable
declares 32; both record one `float64` pair, and the 16 is open.
A `localparam` is a parameter with no difference in the table.
`reg [-4:3]` in `t12_v_neg_range` declares `(-4 to 3)`, so a bound is
a signed word, and the record is that of `reg [0:7]`.
A `shortreal` has no entry of its own; `t12_sv_shortreal` uses the
`real` entry.
A string parameter, `parameter P = "hello"` in `t13_v_str_param`,
uses the unnamed vector entry with 8 bits per character, 40 bits and
`(39 downto 0)`, so the string is a vector to the database.
A package parameter, `parameter int W = 8` in `t13_sv_pkg`, uses the
`int` entry with 32 bits and no range, and is a declaration of the
package unit.

*Found by* `t12_v_params` against `t11_v_param`, and `t12_v_neg_range`
against `t11_v_vec8_asc`.

**Structs.**
A `struct` is a record entry with an empty name, origin `1`, layout
`2` or `3`, and fields written as in VHDL: `name NUL [u32 type]
[u32 nranges]` then triples.
A vector field carries its bounds, `(39 downto 0)` in
`t11_sv_struct40`, and a scalar field none.
The `typedef` that names it is a separate entry of kind `0x07`:
`[1][target][0]`, where `target` is the index of the struct entry,
and the declaration points at the alias, not the struct.
The reader follows aliases before it reads a type.

*Found by* `t11_sv_struct` against `t2_record`, which showed the
record entry unnamed and a new entry after it holding the name.

**Enums.**
A SystemVerilog `enum` is an entry of kind `0x04`:
`[1][base][n][8]` then `n` times `name NUL [u64 value]`, then
`[u32 nranges]` and that many triples, with no `-99`.
`base` is the index of the entry the values are stored in.
An `enum {IDLE, RUN, DONE}` has base `int`, values 0, 1, 2 and no
triples.
An `enum logic [3:0] {A=1, B=5, C=9}` has base the unnamed vector
entry, the three values as given, and one triple `(3, 0, -1)`, the
bounds the vector entry does not hold.
The `8` has not moved; it may be the byte size of a value.
The typedef is a kind `0x07` alias as for a struct.

*Found by* `t11_sv_enum` against `t2_enum`.
*Confirmed by* `t11_sv_enum4`, which moved the base and added the
triple.

**What does not reach the table.**
A `string` variable, `t11_sv_str`, produces no type entry, no
declaration and no object, though its `initial` scope is there.
See [hierarchy.md](hierarchy.md).

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

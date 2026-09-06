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
*Confirmed by* 1172 of 1172 cases decoding with the entry lengths chaining
exactly to the word at `36`.


## Kinds

| Tag | Kind | Body |
| :--- | :--- | :--- |
| `0x03` | enumeration | `[u32 origin][u32 variant][u32 class][u32 n]` then `n` NUL terminated literals, then `[u32 size]`: the value's byte size for VHDL, `1` or `4`, and `0` for Verilog |
| `0x04` | named values | `[u32 origin][u32 base][u32 n][u32 8]` then `n` times `name NUL [u64 value]`, then `[u32 nranges]` and that many triples, no `-99` |
| `0x05` | integer | `[u32 origin][i32 low][i32 high][u32 1]` |
| `0x06` | real | `[u32 origin][u32 variant][f64 low][f64 high][u32 1]` for VHDL, `[u32 0]` for Verilog |
| `0x07` | alias | `[u32 origin][u32 target][u32 nranges]` then that many range triples |
| `0x08` | access | `[u32 origin][u32 designated][u32 8][u32 48]` |
| `0x0c` | file | `[u32 origin][u32 element][u32 8][u32 40]` |
| `0x0d` | physical | `[u32 origin][u32 n]` then `n` times `name NUL [u64 scale]` |
| `0x10` | array | `[u32 origin][u16 layout][u16 0xa0][u32 element][u32 dims]` then `dims` index type words, then `[u32 nranges]` and that many range triples, then `-99` |
| `0x11` | record | `[u32 origin][u16 layout][u16 0xb][u32 n]` then `n` fields, then `-99` |
| `0x13` | dynamic array | `[u32 origin][u32 element][u32 number]`, see "The numbering" |
| `0x14` | queue | `[u32 origin][u32 element][u32 number]` |
| `0x15` | associative array | `[u32 origin][u32 element][u32 number][u32 key]` |
| `0x17` | class | `[u32 origin][i32 parent][u32 number][u32 n]` then `n` fields, each `name NUL [u32 type][u32 nranges]` triples `[u32 0]` |
| `0x18` | string | `[u32 origin]` |

Kinds `0x04` and `0x07` come from SystemVerilog only, and are described
with the other Verilog entries below.
Kinds `0x13` to `0x18` come from SystemVerilog elaborated with
`xelab -debug all` only, and are described under "Under -debug all".

A range triple is `[i32 left][i32 right][i32 dir]`, with `dir` `1` for
`to` and `-1` for `downto`.
The trailing `-99` is `0xffffff9d` and closes the triple list.
An array entry of a SystemVerilog file elaborated with `-debug all`
holds a small non-negative number there instead when a class or a
dynamic container refers to the type; see "The numbering".
The reader keeps the word as `Type.Tail` and does not require `-99`.

**Enumeration.**
The variant word is `2` for every VHDL enumeration.
The class word goes with the shape of the literals, not with the type's
name:

| Class | Literals | Seen on |
| :--- | :--- | :--- |
| `2` | `'0'` and `'1'` | `BIT`, `mybit_t` of `t20_enum_bitlike` |
| `3` | the nine `STD_ULOGIC` literals | `STD_ULOGIC`, `STD_LOGIC`, `my9_t` of `t20_enum_ul_like` |
| `4` | any other set with a character literal in it | `CHARACTER`, `sym_t` `('a', 'b', 'c')`, `mix_t` `(alpha, 'b', gamma)` |
| `5` | identifiers only | `BOOLEAN`, `colour_t`, `one_t` `(only)`, `flag_t` `(no, yes)`, `wide_t` |

*Found by* `t20_enum_bitlike` against `t2_bit` and `t20_enum_ul_like`
against `t1_bit_one_edge`: a user type with `BIT`'s literals gets class
`2` and one with `STD_ULOGIC`'s literals gets class `3`, so the class
does not come from the name.
*Confirmed by* `t20_enum_chars`, `t20_enum_mixed`, `t20_enum_one` and
`t20_enum_two_id` against `t2_enum`, which sort the remaining shapes
into `4` and `5`, and by the tabulation of every enumeration entry in
the corpus, which shows no other class.
An earlier version of this page put `CHARACTER` in class `5`; the dump
shows `4`.

The last word is the byte size of a value: `1` for up to 256 literals
and `4` from 257 on.
*Found by* `t20_enum_300` against `t2_character`, where a 300 literal
type declares 4 bytes and stores `e299` as `2b 01 00 00`.
*Confirmed by* `t20_enum_256` and `t20_enum_257`, one byte and four
bytes either side of the boundary, by `t20_enum_300_arr`, whose pair of
the wide type is 8 bytes, and by `t20_enum_300_rec`, where the wide
field sits at offset 4 after a `std_ulogic`.
The size is a property of the type, so the reader takes it from this
word rather than from the literal count.
See [values.md](values.md).

`std_ulogic` is not a builtin to this format.
It is an ordinary enumeration whose nine literals are written out as
`'U' 'X' '0' '1' 'Z' 'W' 'L' 'H' '-'`, and a user type
`(crimson, viridian, cobalt)` is written the same way.
`CHARACTER` is an enumeration of 261 literals, which is why `t2_character`
is 1461 bytes larger than the one bit baseline where the other scalar
types add about 400.

*Found by* `strings -a -t d` on `t1_bit_one_edge`, which shows
`STD_ULOGIC` followed by the nine literals 4 bytes apart.
*Confirmed by* `t2_enum`, whose literal names appear verbatim, and by
the class word being the only difference between `BIT` and a user
enumeration's entry shape.

An enumeration entry is named after the subtype the signal is declared
with: a `std_logic` signal gets an entry named `STD_LOGIC` with the nine
`STD_ULOGIC` literals, `t8_port_inout` against `t8_port_in`, and
nothing else in the file names `STD_ULOGIC`.

**File.**
A file type is kind `0xc`: origin `2`, the index of the element type,
then the two words `8` and `40`.
It was first seen when `-debug all` brought the `textio` package in,
as `TEXT` of `t22_dbg_all` over `STRING`, and a file declared in the
design puts it in the table under the default debug level as well:
`text` in `t23_file_text`, which also brings `character`, `POSITIVE`
and `STRING` in, and the user types `int_file` over `INTEGER` and
`sul_file` over `STD_ULOGIC`.
The two words are `8` and `40` for all four, so they do not depend on
the element type; what they mean is open.
A file variable declares 0 bytes and has no record.
*Found by* `t22_dbg_all` against `t22_base`.
*Confirmed by* `t23_file_text`, `t23_file_int` and `t23_file_sul`.

**Access.**
An access type is kind `0x8`: origin `2`, the index of the designated
type, then the two words `8` and `48`.
The words are the same for `int_ptr` over `INTEGER` in `t23_access`
and `vec_ptr` over the unconstrained `STD_ULOGIC_VECTOR` in
`t23_access_vec`.
A variable of an access type declares 48 bytes, the second word, and
has no record, while a file variable declares 0 with its second word
at 40, so the second word is not simply the declared size.
The reader crashed on the kind before this case, and reads it with
the same shape as a file type.

Tier 75 puts both kinds over every shape of type the language allows,
and neither pair of words moves:

| Type | Words | Case |
| :--- | :--- | :--- |
| `access integer` | `8 48` | `t23_access` |
| `access std_ulogic_vector`, unconstrained | `8 48` | `t23_access_vec` |
| `access rec_t`, a record | `8 48` | `t75_acc_rec_____` |
| `access arr_t`, forty integers | `8 48` | `t75_acc_arr40___` |
| `access int_ptr`, another access | `8 48` | `t75_acc_acc_____` |
| `file of integer` | `8 40` | `t23_file_int` |
| `file of std_ulogic` | `8 40` | `t23_file_sul` |
| `TEXT` | `8 40` | `t23_file_text` |
| `file of rec_t` | `8 40` | `t75_fil_rec_____` |
| `file of arr_t` | `8 40` | `t75_fil_arr_____` |

The variable of each declares what its kind declares, whatever it
points at or holds: 48 bytes for every access variable, including the
one over a record and the one over another access, and 0 for every
file variable.
So the two words are constants of the kind and not a measurement of
the designated or element type, and the reading that the second is the
runtime size of the object survives only as a coincidence of the
access side: a file object would then be 40 bytes and declare 0.

*Found by* `t23_access` against `t6_var_int`, whose process variable is
an integer.
*Confirmed by* `t23_access_vec`, and by the five tier 75 cases above.

**Integer.**
`INTEGER` is `-2147483648 to 2147483647` and `NATURAL` is
`0 to 2147483647`.
A subtype with its own bounds is a separate entry under the subtype's
name, `small_t` `0 to 7` in `t21_int_sub`, and a new integer type of
the same range, `t21_int_newtype`, produces the same entry.
The value stays 4 bytes.

**Real.**
`REAL` is `-1e308 to 1e308`.

**Physical.**
`TIME` lists eight units with their size in picoseconds:
`fs=0 ps=1 ns=1000 us=1000000 ms=1000000000 sec=1000000000000
min=60000000000000 hr=3600000000000000`.
So `TIME` counts picoseconds, and `fs` rounds to zero.
A user physical type lists its units in its own base unit:
`t21_phys_user` declares `um`, `mm = 1000 um` and `m = 1000 mm`, and
the entry holds `um=1 mm=1000 m=1000000`.
The scales follow the simulation precision.
`t22_vh_fs`, elaborated with `--timeprecision_vhdl 1fs`, lists
`fs=1 ps=1000 ns=1000000` and so on up to `hr=3600000000000000000`,
and `t22_vh_ns`, at `1ns`, lists `fs=0 ps=0 ns=1 us=1000`.
So a `TIME` value counts the unit the DBG section names, and the entry
gives every unit's size in it.
There is no trailer word.
*Found by* `t22_vh_fs` against `t22_base`.
*Confirmed by* `t22_vh_ns`.

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

**An array of an unconstrained element.**
An array type whose element is unconstrained writes every triple as
`(0, 0, -2)`, its own index included, even when the index is written
with bounds.
`arr_t`, an `array (0 to 1) of std_ulogic_vector` in
`t42_arr_unc_elem`, is an entry with `dims` `1` and two triples
`(0, 0, -2) (0, 0, -2)`, and `array (natural range <>) of
std_ulogic_vector` in `t42_arr_unc_both` is the same entry; the two
files differ in the index word, an `INTEGER` entry for the `(0 to 1)`
of the first and `NATURAL` for the second, and in nothing else.
The declaration carries `(0 to 1) (3 downto 0)` in both, from
`arr_t(open)(3 downto 0)` and `arr_t(0 to 1)(3 downto 0)`.
An `array (0 to 1) of bundle_t` over a record with an unconstrained
field, `t42_rec_unc_arr`, is the same shape, `(0, 0, -2) (0, 0, -2)`
with the record as the element, and declares `(0 to 1) (3 downto 0)`.
*Found by* `t42_arr_unc_elem` against `t2_array2d`, `array (0 to 3)
of std_ulogic_vector(7 downto 0)`, whose entry holds `(0 to 3)
(7 downto 0)`.
*Confirmed by* `t42_arr_unc_both` and `t42_rec_unc_arr`.

**Bounds below zero.**
A bound is a signed 32 bit word in a triple and in a declaration
range alike.
`vec_t(3 downto -4)` in `t41_neg_vec`, an `array (integer range <>)
of std_ulogic`, puts `(3 downto -4)` in the declaration record where
`vec_t(7 downto 0)` of `t41_uvec` puts `(7 downto 0)`; the type entry
is the unconstrained triple in both, and the index word names
`INTEGER` for the one and `NATURAL` for the other.
`t41_neg_asc` has `(-4 to 3)`, and `int_array_t`, an
`array (-2 to 1) of integer` in `t41_neg_arr_type`, holds
`(-2 to 1)` in its entry.
The value is the elements in index order, the leftmost first,
whatever the bounds.
The predefined `std_ulogic_vector` cannot take such a bound, since its
index subtype is `NATURAL`, which is why the tier declares its own
array type.
*Found by* `t41_neg_vec` against `t41_uvec`.
*Confirmed by* `t41_neg_asc`, `t41_neg_arr_type`, `small_t` of
`t41_neg_int_sub`, an integer subtype `-8 to 7`, and the `sfixed`,
`ufixed` and `float32` entries below.

**Constrained subtype of an array.**
A signal declared with a constrained subtype of an unconstrained array
type, `subtype byte_t is vec_t(3 downto -4)` in `t41_arr_subtype`,
gets an entry named `byte_t` that carries the subtype's range
`(3 downto -4)` as its one triple, and the base type `vec_t` is not in
the table.
The declaration record carries the same range.
So the entry is named after the subtype the signal is declared with,
as the enumeration entries are, and holds the subtype's constraint
where the base type held none.
*Found by* `t41_arr_subtype` against `t41_neg_vec`, the same signal
declared through a subtype.
*Confirmed by* `float32` of `t41_float32`, which `ieee.float_pkg`
declares as a constrained subtype of `float` and whose entry is
`float32` with `(8 downto -23)`.

**The IEEE fixed and float packages.**
`sfixed(3 downto -4)` of `t41_sfixed` and `ufixed(3 downto -4)` of
`t41_ufixed` are unconstrained array entries named `sfixed` and
`ufixed` over `STD_ULOGIC`, indexed by `INTEGER`, with the bounds in
the declaration record, and their values are one byte per element
with the leftmost first: `to_sfixed(1.5, 3, -4)` is `00011000` and
`to_sfixed(-2.25, 3, -4)` is `11721100`.
`float32` of `t41_float32` is the constrained entry above, 32 bytes,
`to_float(1.5)` recorded as the IEEE 754 bits `0x3fc00000` one byte
per bit.
Nothing marks a fixed or floating point vector as one: the entries
have the shape of any array of `STD_ULOGIC`, and a reader that wants
the number has to know the package.
The names are spelled as the package source spells them: `sfixed`,
`ufixed` and `float32` in lower case, where `STD_ULOGIC_VECTOR` is in
upper case, and Vivado's `data/vhdl/src/ieee_2008` sources declare
`subtype sfixed is (resolved) UNRESOLVED_sfixed` and
`type STD_ULOGIC_VECTOR`.
A resolved subtype of an unconstrained type is named for the subtype,
as `STD_LOGIC_VECTOR` is, and the unresolved base does not appear.
*Found by* `t41_sfixed` against `t41_neg_vec`, the same bounds over a
user type.
*Confirmed by* `t41_ufixed` and `t41_float32`.

*Confirmed by* `t18_arr_3dim`, `t19_arr_2d_vec` and `t19_arr_of_2dim`,
and by `truth.json` of each against the decoded value.

**Predefined vectors.**
The VHDL 2008 `integer_vector`, `real_vector`, `time_vector` and
`boolean_vector` are unconstrained array entries named in upper case,
`INTEGER_VECTOR` and so on, over the scalar entry and indexed by
`NATURAL`, with the one triple `(0, 0, -2)`.
A signal `s : integer_vector(0 to 3)` carries `(0 to 3)` in its
declaration record and nothing in the type, where the user type
`int_array_t` of `t5_int_arr`, an `array (0 to 3) of integer`, holds
`(0 to 3)` in the type entry as well.
The values are 4, 8, 8 and 1 bytes per element as the scalar is, so
the four signals are 16, 32, 32 and 4 bytes, and `time_vector`
brings the `TIME` physical entry in under origin `0xa`.
*Found by* `t23_int_vector` against `t5_int_arr`.
*Confirmed by* `t23_real_vector`, `t23_time_vector` and
`t23_bool_vector`, each against `truth.json`.

**String.**
`STRING` is the same shape: an unconstrained array entry under origin
`2` over `character`, indexed by `POSITIVE`, with the one triple
`(0, 0, -2)`.
`t23_file_text` had brought the entry in behind `text`; a string
signal of the design brings it in on its own, and `character` and
`POSITIVE` with it, where `t2_character` had `character` alone.
The declaration of `s : string(1 to 5)` carries `(1 to 5)` and 5
bytes, and `(3 to 7)` under `string(3 to 7)`, so the bounds are the
ones written and not renumbered from 1.
The value is one byte per character, the `character` literal's index,
`68 65 6c 6c 6f` for `"hello"`, and the reader renders an array of a
character enumeration as one string.
A `string` process variable declares the same 5 bytes and `(1 to 5)`
and has no record, as an `integer` variable has none.
*Found by* `t44_str_sig` against `t2_character`.
*Confirmed by* `t44_str_sig_3to7` and `t44_str_var` against
`t6_var_int`.

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

**A field declared without bounds.**
VHDL 2008 lets a record field be an unconstrained array, with the
bounds given where a signal of the record is declared.
The field then carries the unconstrained triple `(0, 0, -2)`, and the
bounds are in the declaration record alone: `bravo : std_ulogic_vector`
constrained as `bundle_t(bravo(7 downto 0))` in `t42_rec_uncons` gives
`bravo:[3]((0, 0, dir -2))` in the entry and `(7 downto 0)` on the
declaration, where `t2_record2` with the bounds in the field has
`(7 downto 0)` in both places.
The declaration keeps its 16 bytes and the page records hold the same
bytes.
Two signals that constrain the type differently share the one entry
and differ only in their declarations: `s` with `(7 downto 0)` and 16
bytes beside `t` with `(3 downto 0)` and 8 bytes in
`t42_rec_two_cons`.
A constrained subtype of the record, `subtype b8_t is
bundle_t(bravo(7 downto 0))` in `t42_rec_subtype`, renames the entry
`b8_t` and leaves the field triple unconstrained, so the subtype's
bounds are still only in the declaration.
That is the opposite of an array subtype, whose entry holds the bounds
as its triple; see "Constrained subtype of an array" above.

The declaration record lists the bounds of every array dimension in
the record in field order, whether the field had them or not.
`t42_rec_two_unc`, two unconstrained fields constrained
`(alpha(3 downto 0), bravo(7 downto 0))`, and `t42_rec_mix_unc`, the
first field constrained in the record and the second at the signal,
both declare `(3 downto 0) (7 downto 0)`, and `t7_rec_in2v` with all
bounds in the fields declares `(3 downto 0) (1 downto 0)` the same way.
An unconstrained two dimensional field carries two `(0, 0, -2)`
triples and the declaration carries both ranges, `(0 to 1) (0 to 2)`
for `m : mat_t` constrained `rec_t(m(0 to 1, 0 to 2))` in
`t42_rec_unc_2dim`.
A field whose type is a record with an unconstrained field carries
the inner field's `(0, 0, -2)` in its flattened list, `i:[1]((0, 0,
dir -2))` in `t42_rec_unc_nest`, and the declaration carries the one
range `(3 downto 0)`.
The inner record is padded to 8 bytes on its own, so `outer_t` with a
four element vector inside `inner_t` and one `std_ulogic` after it is
16 bytes: `02 02 02 02 00 00 00 00 02 00 ...`.

So a reader decodes a VHDL record from its declaration's range list,
consuming one range per array dimension in field order, and uses the
field triples only when the declaration gives none.
`File.fieldConstraint` does that, and `Decode` reads `bravo` of
`t42_rec_uncons` as `10100101` where the field triple alone made it
one element.
*Found by* `t42_rec_uncons` against `t2_record2`, where the value
decoded as `bravo => 1`.
`t42_rec_two_unc` against `t42_rec_mix_unc` showed the declaration
list unchanged by where the bounds are written.
*Confirmed by* `t42_rec_subtype`, `t42_rec_two_cons`,
`t42_rec_unc_nest`, `t42_rec_unc_2dim` and `t42_rec_unc_arr`, each
against `truth.json`, and by the 700 case corpus after the change.
`t8_rec_realv` put a `real` beside the vector and got the vector's
bounds alone, so a `real` has no range to contribute and is skipped.

**Protected.**
A protected type is a record entry.
`counter` of `t23_protected`, a protected type whose body declares one
variable `cnt : integer`, is kind `0x11` named `counter` with the one
field `cnt` over `INTEGER`, and its methods leave nothing in the
table.
The shared variable of that type declares 8 bytes, the record's size
rounded to 8, and has no record.
*Found by* `t23_protected` against `t23_shared_int`.

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
A `union packed` says `6`.

*Found by* `t11_sv_struct` against `t11_sv_ustruct`, the same two
field struct packed and unpacked, which differ in this half word and
in the declaration size.
*Confirmed by* `t11_v_mem4`, whose two array entries carry both values.
The union is `t24_sv_union`, whose `union packed { logic [7:0] b;
logic [7:0] c; } u_t` is a record entry with an empty name, origin
`1`, layout `6` and the fields `b` and `c`, each `(7 downto 0)` over
the unnamed vector entry, under an alias `u_t`, exactly the shape of
the struct of `t11_sv_struct` but for the layout word.
The declaration says 8 bits, the width of one field, so the fields
share the bits and the entry is as wide as its widest field, not the
sum.
The reader accepts `6` as `LayoutUnion` and decodes every field over
the same bits.

*Found by* `t24_sv_union` against `t11_sv_struct`.

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
The class word is `1` for both, outside the VHDL classes above, and
what the variant words mean is open.
The last word is `0` where a VHDL enumeration carries its value size.
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
The `unsigned` qualifier of an integral type is not recorded either:
`int unsigned s` of `t27_sv_int_uns` is a declaration of the type
`int`, and the file is that of `t11_sv_int` outside the timestamps,
the line numbers and the scope name they change.
So a `byte unsigned` holding 165 reads back as `-91`, and the truth of
such a case says `unsigned` for the test to read the value back as an
unsigned number: `t27_sv_byte_uns`, `t27_sv_lng_uns`,
`t27_sv_intg_uns`, `t27_v_intg_uns`.
`shortint` is the same array entry over `bit` with `(15, 0, -1)`,
`t26_sv_shortint`.

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

A typedef of that array, `typedef rec_t arr_t [0:1]` in
`t35_sv_ust_tdef_` and `typedef s_t arr_t [0:1]` in
`t35_sv_pst_tdef_`, adds a second alias `arr_t` over the array entry,
carrying the `(0 to 1)` the declaration then drops, as
`t13_sv_tdef_ua` does for a vector element.
The unpacked struct itself is the same record entry of layout `2`
whether it stands alone, `t11_sv_ustruct`, sits in an array,
`t35_sv_ust_arr__`, or is a field of another unpacked struct,
`t35_sv_ust_nest_`, whose outer record lists `r:[4]` with the alias
`rec_t` as the field type.
An unpacked array field, `logic [3:0] v [0:1]` in `t35_sv_st_uarr__`,
is an unnamed array entry of layout `2` over the unnamed vector entry,
with two `(0, 0, -2)` placeholders, and the field carries
`(0 to 1)(3 downto 0)`.

*Found by* `t12_sv_typedef` against `t11_v_vec8`, whose declaration
carries `(7 downto 0)` and whose type has no alias.
*Confirmed by* `t13_sv_pkg`, `t13_sv_struct_ar`, `t35_sv_ust_tdef_`,
`t35_sv_pst_tdef_`, `t35_sv_ust_nest_` and `t35_sv_st_uarr__`.

**Parameters.**
A parameter's type follows its declaration: `parameter K = 5` and
`parameter [7:0] P` use the unnamed vector entry, with 32 and 8 bits
and a declaration range, `parameter integer Q` uses the `integer`
entry with 32 bits and no range, and `parameter real R` uses the
`real` entry.
The `real` parameter declares 16 bits, where a `real` variable
declares 32; both record one `float64` pair, and the 16 is open.
An untyped parameter with a real literal, `parameter K = 1.5` in
`t28_sv_prm_realu`, uses the `real` entry and declares 32 bits, so the
16 goes with the `real` keyword.
Tier 71 puts the same keyword in every other place it fits, and the
16 turns out to be the scalar parameter's alone:

| Declaration | Bits | Case |
| :--- | ---: | :--- |
| `parameter shortreal R = 1.5` | 16 | `t71_rlw_sreal_p_` |
| `parameter realtime R = 1.5` | 16 | `t71_rlw_rtime_p_` |
| `parameter real R` in a package | 16 | `t71_rlw_pkg_prm_` |
| `parameter real R` in a child module | 16 | `t71_rlw_kid_prm_` |
| `parameter R = 1.5`, untyped | 32 | `t71_rlw_untyped_` |
| `specparam d = 1.5` | 32 | `t71_rlw_specprm_` |
| `parameter real R [0:1]` | 64, so 32 an element | `t71_rlw_arr_prm_` |
| a VHDL `generic r : real` | 8 bytes, the float | `t71_rlw_vhdl_gen` |

So it is not the scope, not the keyword `parameter` against
`localparam`, not the value, and not the type entry, which the array
of two shares: a scalar parameter whose type names a real declares
half of what everything else holding the same `float64` declares.
A `realtime` is a real entry of its own, named `realtime`, with the
variant `1` of a Verilog `real`: a `realtime` variable declares 32
bits, `t28_sv_rtime_var`, and a `parameter realtime` 16,
`t28_sv_rtime_prm`.
`parameter time T = 10ns` uses the `time` entry, `t28_sv_prm_tmtyp`,
and `parameter T = 10ns` without a type uses the unnamed vector entry
with 64 bits, `t28_sv_prm_time`, whose record holds a `float64` and
not a vector; see [values.md](values.md).
A time expression, `parameter T = 10ns * 2` in `t30_sv_ptm_expr`,
uses the `real` entry with 32 bits instead.
A parameter of an enum type, `parameter state_t S = RUN` in
`t28_sv_prm_enum`, uses the alias entry of the typedef with 32 bits,
and `parameter int unsigned K` the `int` entry, `t28_sv_prm_int_u`,
without the qualifier.
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
A queue, a dynamic array, an associative array and a class,
`t24_sv_queue`, `t24_sv_dynarr`, `t24_sv_assoc` and `t24_sv_class`,
produce none either, and the `logic` entry is the only one in each.
See [hierarchy.md](hierarchy.md).

**Under -debug all.**
Every SystemVerilog case before tier 60 ran under xsim's typical
debugging level.
`xelab -debug all` keeps the objects that typical drops, and each
brings a kind of its own to the table.
Every case below is `t60_dbg_*` in the corpus, has
`xelab_args = ["-debug", "all"]` in its `BUILD.bazel`, and reproduces
with `wdbcvt -dump`.

| Source | Entries added | Case |
| :--- | :--- | :--- |
| `string str` | `0x18` `string`, origin `0x5`, no body | `t60_dbg_str_____` |
| `int q[$]` | `bit`, `scalar_int`, `int`, then `0x14` unnamed, origin `0x1`, element `int`, number `1` | `t60_dbg_queue___` |
| `int d[]` | the same three, then `0x13` unnamed, element `int`, number `1` | `t60_dbg_dynarr__` |
| `int a[string]` | the same three, `string`, then `0x15` unnamed, element `int`, number `2`, key `string` | `t60_dbg_assoc___` |
| `int a[int]` | the same three, then `0x15` unnamed, element `int`, number `3`, key `int` | `t60_dbg_assoc_i_` |
| `class c_t; int f = 1; endclass` | `0x17` `c_t`, origin `0x1`, parent `-1`, number `0`, one field `f` of `int` with no triple | `t60_dbg_class___` |
| `class c_t extends b_t` | `b_t` with parent `-1` and number `1`, then `c_t` with parent `b_t`'s index and number `0`, each with its own field only | `t60_dbg_class_d_` |

The class entry comes before the predefined entries of its field
types, the way an outer record comes before its fields; a derived
class's parent comes first.
A field of a class is written the way a record field is: name, type
index, range count and triples, `g` of `logic [3:0]` carrying
`(3, 0, -1)` in `t60_dbg_class_2_`; and one more word, `0` in every
field seen, follows the triples.
The inherited field `f` is listed under `b_t` only, not repeated under
`c_t`.

*Found by* `t60_dbg_str_____`, `t60_dbg_queue___`, `t60_dbg_dynarr__`,
`t60_dbg_assoc___` and `t60_dbg_class___` against `t60_dbg_none____`
and `t60_dbg_int_____`, which hold the ordinary entries only.
*Confirmed by* `t60_dbg_assoc_i_` against `t60_dbg_assoc___`, which
changes the key type and the word before it, by `t60_dbg_class_2_` and
`t60_dbg_class_d_` against `t60_dbg_class___`, and by
`t60_dbg_class_2h` and `t60_dbg_class_n_`, where a second handle and a
construction at time 0 leave the entries unchanged.

The ordinary entries are what they are under typical: `t60_dbg_vec_____`,
`t60_dbg_int_____`, `t60_dbg_real____`, `t60_dbg_struct__` and
`t60_dbg_mem_____` hold the same type tables as their tier 11
counterparts, `-99` included.

**The numbering.**
The word after an array entry's last triple, the id word of a class
and the word after a container's element type are one numbering,
tier 61.
It counts from `0` over the types the flag registers for the objects
it keeps, in this order:

* The variables of the module in declaration order:
  `int a[string]` then `int q[$]` gives the associative array `2` and
  the queue `3`, and the other way round the queue `1` and the
  associative array `3`, `t61_num_a_then_q` and `t61_num_q_then_a`.
* A container after its element: `int q[$]` gives `int` `0` and the
  queue `1`, `t60_dbg_queue___`; `int q[$][$]` gives `int` `0`, the
  inner queue `1` and the outer queue `2`, `t61_num_q_q_____`; a queue
  of a class gives the class `0`, its `int` field `1` and the queue
  `2`, `t61_num_q_cls___`.
* A class before its fields, and the fields last declared first:
  `int f; logic [3:0] g` gives the class `0`, the vector `1` and `int`
  `2`, `t60_dbg_class_2_`, and `real r; logic [3:0] g; int f` gives
  `int` `1` and the vector `2`, `t61_num_cls_rev_`.
  A parent class, or a class a field names, comes right after the
  class and before its own fields: `c_t extends b_t` gives `c_t` `0`,
  `b_t` `1` and `b_t`'s `int` `2`, `t60_dbg_class_d_`, and a field
  `b_t hb` the same, `t61_num_cls_cls_`.
  Two classes with a handle each give `a_t` `0`, `int` `1` and `b_t`
  `2`, `t61_num_two_cls_`.
* A type already numbered keeps its number: the `int` of `b_t` in
  `t61_num_two_cls_`, and two `int` fields, `t61_num_cls_2int`.
* An array is numbered by its element, not by itself: `int`, `byte`
  and `longint` in one class all hold `1`, `t61_num_cls_byte`,
  `t61_num_cls_byti` and `t61_num_cls_long`, and `int; byte;
  logic [3:0]` holds `2`, `2` and `1`, `t61_num_cls_ibv_`, where
  `logic [3:0]` and `logic [7:0]` share one entry and one number,
  `t61_num_cls_2vec`.
  The number is not written into the `bit` or `logic` entry, whose
  last word stays `0`.
* A `string` or a `real` takes no number and has no slot for one:
  `string q[$]` gives the queue `0`, `t61_num_q_str___`, and a
  `string` or `real` field leaves the others as they were,
  `t61_num_cls_str_`, `t61_num_cls_3f__`.
* An associative array is numbered twice with a `string` key and
  three times with an `int` key, and holds the last number:
  `int a[string]` gives `int` `0` and the array `2`, `int a[int]` gives `int` `0` and the
  array `3`, and a queue declared after either continues at `3` or
  `4`, `t61_num_a_then_q` and `t61_num_ai_thn_q`.
  Tier 70 separates the key from the element, which the tier 61 cases
  could not: their key was either a `string`, which takes no number,
  or the element's own type.

| Declaration | Element | Key | The array | Case |
| :--- | ---: | ---: | ---: | :--- |
| `int a[string]` | 0 | none | 2 | `t60_dbg_assoc___` |
| `logic [3:0] a[string]` | 0 | none | 2 | `t70_num_a_v_str_` |
| `byte a[string]` | 0 | none | 2 | `t70_num_a_b_str_` |
| `int a[byte]` | 0 | 1 | 3 | `t70_num_a_i_byte` |
| `byte a[int]` | 0 | 1 | 3 | `t70_num_a_b_int_` |
| `int a[int]` | 0 | unwritten | 3 | `t60_dbg_assoc_i_` |
| `int a[e_t]` | 2 | none | 4 | `t70_num_a_e_key_` |

  So the key is a numbered type in its own right: `byte` and `int`
  carry the number in their own entry when they are the key and not
  the element, and `int a[int]` still spends the number, on an entry
  that already holds the element's `0` and cannot hold a second.
  A `string` key spends nothing, as a `string` never does.
  An enumeration key spends nothing either, and its own declaration
  takes the two numbers before the element instead: the typedef of
  `t70_num_a_e_key_` sits before the variable and leaves `int` at
  `2`.
  One number is left over in every one of these, between the last
  numbered part and the array itself, and nothing in the file carries
  it; that it belongs to an iterator is the guess of question 24.
  The leftover is the associative array's alone: a dynamic array and a
  queue take the number after their element and no other, `int d[]`
  and `int q[$]` giving `int` `0`, the dynamic array `1` and the queue
  `2`, `t70_num_d_then_q`.
  Nesting repeats the whole rule: `int a[string][int]` numbers `int`
  `0`, the inner array `3` for an `int` key and the outer `5` for a
  `string` key, `t70_num_a_2dim__`, and a class with an associative
  field counts from the class, `c_t` `0`, `int` `1`, the array `3`,
  `t70_num_a_in_cls`.

The variable's own class or container therefore holds `0` when the
element takes no number, and the class of a handle variable always
does.
What the numbering is for is open; question 24 of
[../format.md](../format.md) keeps the guesses.
A container's declaration takes the value class of its element,
`3` for `int`, `0` for a class or a string, and one `(0 to 0)` range
per dimension, `t61_num_q_cls___`, `t61_num_q_str___` and
`t61_num_q_q_____`.

*Found by* `t61_num_a_then_q` against `t61_num_q_then_a`, which move
the same two containers past each other, and `t60_dbg_class_2_`
against `t61_num_cls_rev_`, which reverse the fields.
*Confirmed by* the other cases named above, 31 files with a number in
them and no exception to the rules.

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

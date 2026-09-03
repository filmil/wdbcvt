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
| [format/types.md](format/types.md) | the type table: enumerations, integers, reals, physical types, arrays and records, and the Verilog entries |
| [format/hierarchy.md](format/hierarchy.md) | the debug section: scopes, units, declarations, source files, instance records and handles, in both languages |
| [format/values.md](format/values.md) | arenas, pages, value records, encodings and alignment, and the Verilog word pairs |

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
against the `truth.json` of all 241 cases.
The offsets are in the documents linked in the last column.

| Finding | Found by | Confirmed by | Where |
| :--- | :--- | :--- | :--- |
| Magic `Xilinx WAVE DATABASE 01`, producer `Xilinx Simulator` | hex dump of `t1_bit_one_edge` | all 241 cases | container |
| `0x38` is a Unix timestamp | noise mask, two runs of `t3_tr1` | equals the file mtime | container |
| `0x48` holds three pointers to 48 byte directory entries | `strings -t d` on `t2_flat3`, then reading the values | all 241 cases, each pointer lands on a name | container |
| The arena table at `0xc8` grows with the object count | `t5_sig10` shifted every trailer field by 8 | 3, 4, 6 slots in `t6_sig05`, `t6_sig12`, `t6_sig20` | container |
| The arena table has `ceil(handle space / 0x800)` slots | `t7_sig07` broke the `ceil(objects / 4) + 1` guess | the reader checks it in 241 of 241 | container |
| Trailer `+0x0c` is the arena table slot count | sweep of every fixed word over the 63 cases of tier 7 | 241 of 241, checked by the reader | container |
| Trailer `+0x18` is the handle space size | `t7_sig07` to `t7_sig24` against the slot count | the slot rule, 241 of 241 | container |
| Arena records sit in first write order, not arena order | `t7_gen_for`, arena 2 first | 241 of 241 with the reader accepting any order | container |
| The trailer is the 0x48 bytes before the first directory pointer | `t5_sig10` against `t6_sig05` | end time correct in all 63 | container |
| The end time is a uint64 in ps at trailer `+0` | correlation sweep over 33 cases | 241 of 241, 20 ns to 1310 ns | container |
| The marker offset is at trailer `+0x38` | `t5_tr1000`, where the marker moved | 241 of 241 | container |
| The marker is a list of `[first][last]` object index ranges, as many as trailer `+0x30` counts, covering exactly the objects with records | `t9_port_rec` held two entries where `t6_var_int` had shown one; `t9_mark_gap` put an unlogged object first | 241 of 241, the reader checks every object against the ranges | container |
| An arena record's word 0 names a continuation record for pages past 100 | `t9_tr70000`, 117 pages | the reader reads the 70001 records back | container |
| Each directory entry follows the section it describes | `t2_flat3`: `WDB.Event` at `0xe0+0x48`, RTTI and DBG the same | 241 of 241 | container |
| The page directory starts 48 bytes after the DBG entry | `t2_flat3`, reading the offsets | 241 of 241 | container |
| An arena record is `0x4c0` bytes: 100 page offsets, 100 lengths, a count | `t5_tr1000`, two pages in one arena | `t6_tr1300`, three pages | container |
| A page is a zlib stream that inflates to 10240 bytes | entropy profile, then `zlib` on `t1_bit_one_edge` | 241 of 241 | values |
| Page header `[t0][last minus t0][n]` | `t5_tr1000` page 1 | all pages of all cases | values |
| A record is `[uint64 time][uint32 key][uint32 length][value]` | `t1_bit_two_edges` against `t1_bit_one_edge` | every record of every case matches `truth.json` | values |
| `handle >> 11` is the arena, `handle & 0x7ff` the key | `t5_sig10` | `t6_sig20`, four arenas | values |
| A page holds 10240 bytes of records and overflows into a new page | `t5_tr1000`, 600 one byte records | `t7_int700` 510 and `t7_wide700` 425 | values |
| Records at one time are in simulation order, not key order | `t7_gen_for` | `t2_flat3` | values |
| A delta cycle leaves two records at one time | `t7_delta` | | values |
| An overflowed page precedes the marker | `t5_tr1000` | `t6_tr1300` | values |
| A value of 275 bytes or more is written as chunks with consecutive keys, one record per chunk, addressed by handle plus byte offset | `t9_vec292` against `t9_vec256` and `t9_vec257` | the 18 `t9_vec*` sizes and `t9_int73`, read back against `truth.json` | values |
| A value of `size >= 275` bytes is `2 * ceil((size + 24) / 299)` chunks of `floor(size / n)` bytes, the last taking the rest | `t10_vec274` to `t10_vec276`, `t10_vec574` to `t10_vec575`, `t10_vec872` to `t10_vec874` | the reader enforces the addresses on all 55 wide values | values |
| Chunks split at an arena boundary and inside an element | `t9_vec292`, 6 plus 67 bytes across `0x800` | `t9_int73`, 73 byte chunks of 4 byte integers | values |
| A wide value spans as many arenas as its byte range crosses | `t9_vec12000`, arenas 0 to 6 | `t9_vec4096` | values |
| A procedure with a `signal` parameter writes the change twice | `t9_proc_sig` against `t9_proc_local` | | values |
| Enumeration values are one byte, the literal's index | `t3_valz` against `t3_tr1`, same size | `t1_nine_state` walks all nine | values |
| Integers are int32, reals float64, time int64 ps | `t2_integer`, `t2_real`, `t2_time` | `truth.json` | values |
| Arrays are elements back to back, left index first | `t1_vec8` | `t2_array2d`, `t5_int_arr` | values |
| Record fields are aligned to their size, records to 8 | `t5_rec_real` against `t2_record` | `t5_arr_rec`, `t5_rec_sub5` | values |
| A signal has one record at time 0 and one per change | `t0_bit_const` | `t3_late`, 241 of 241 | values |
| The type table starts with `Xilinx ISim TYPE FILE 001` | `strings` on `t1_bit_one_edge` | 241 of 241 | types |
| `+32` of the type table is the number of types | correlation sweep | 241 of 241 | types |
| Type entries are `[len][tag]` name body | `t2_enum` against `t1_bit_one_edge` | 241 of 241 | types |
| Enumerations list their literals; `character` has 256 | `t2_enum`, `t2_character` | `truth.json` names | types |
| Integer entries carry the bounds, reals the bounds as float64 | `t2_integer`, `t2_real` | 241 of 241 | types |
| Physical entries list units with scales | `t2_time` | | types |
| Arrays carry element, index type and constraint triples | `t1_vec8` against `t2_array2d` | `t5_int_arr` | types |
| Records list fields with types and ranges | `t2_record` | `t2_record_nested`, `t5_rec_sub5` | types |
| A record field of record type lists one range per inner field, the scalar's own range, only when the inner record has an array field | `t7_rec_vfirst`, `t7_rec_bitv`, `t7_rec_intv`, `t7_rec_in2` | `t7_rec_in2v` | types |
| Types are shared between signals of the same type | `t2_record_two` | `t6_sig20`, one `STD_ULOGIC` | types |
| The DBG section starts with `Xilinx ISim DBG 006` and 18 region offsets | `t1_hier1` against `t2_hier3` | 241 of 241 | hierarchy |
| Scope records: name, parent, children, first object, unit, file, line | `t2_hier3` | 241 of 241 | hierarchy |
| Unit records: entity, architecture, kind, declaration count, file, line | `t2_hier3` | `t4_gen_diff_two` | hierarchy |
| Declaration records: name, file, line, size, type, ranges, kind | `t2_flat3` | 241 of 241 | hierarchy |
| Declaration kinds `0x0e` signal, `0x0f` variable, `0x12` generic, `0x13` constant | `t4_gen_default`, `t5_tr1000`, `t6_var_int`, `t8_gen_if` | `t6_proc2`, `t7_gen_for` | hierarchy |
| Declaration word 9 is the port mode: 0 inout, 1 in, 2 out, 3 buffer, 4 linkage, 5 none | `t8_port_in`, `t8_port_out`, `t8_port_inout`, `t8_port_buffer`, `t9_port_lnk` | 241 of 241 against the `port` field in `truth.json` | hierarchy |
| Instance word `+16` is a `uint32` scope and `+20` a `uint32` byte offset into the value, for a port bound to a slice | `t9_port_slice`, offset 1 for `x(0)` of `1 downto 0` | `t9_port_slice2`, `t9_port_sliceto`, offset 0 for `x(0)` of `0 to 1` | hierarchy |
| A package with an object is a scope under the root with unit kind `0x0a` | `t9_port_rec` against `t2_record` | `t9_mark_two`, `t9_mark_gap`, `t9_pkg_sig` | hierarchy |
| A package constant or signal is an object with no records | `t9_port_rec`, `t9_pkg_sig` | `t9_mark_two`, `t9_mark_gap` | values |
| A block is a scope with unit kind `0x0c`, as a generate | `t9_block` | | hierarchy |
| A process variable of an entity instantiated `n` times over one unit is listed `n` times in each of the `n` process scopes | `t9_mark_two` against `t6_var_int` | `t9_var_inst3`, nine objects for three variables | hierarchy |
| Generics of `boolean`, `string`, vector and `real` type are objects with one record in the type's encoding | `t9_gen_types` | | hierarchy |
| A port bound to a literal owns a handle like an open port | `t9_port_expr` against `t8_port_in` | | values |
| A component, an alias, a function and a procedure add no scope, declaration or object | `t9_comp`, `t9_alias`, `t9_func`, `t9_proc_local` against `t8_port_in` and `t1_bit_one_edge` | | hierarchy |
| Unit kind `0x0c` is a generate; iteration scopes are named `\g(0)\` | `t7_gen_for` | `t8_gen_nest` | hierarchy |
| A nested generate repeats the shape: `\g(0)\.\h(0)\` plus an empty `h` per outer iteration | `t8_gen_nest` | | hierarchy |
| An if generate is one plainly named scope per branch label; the false branch is an empty scope | `t8_gen_if` | | hierarchy |
| A concurrent assignment is a process scope named `line__NN` | `t8_port_open` | `t8_port_vec8` | hierarchy |
| A connected port shares the handle of the signal on its net, down a chain | `t8_port_in` | `t8_port_chain`, `t8_port_out`, `t8_port_inout`, `t8_port_buffer` | hierarchy |
| An open port owns a handle and costs `0xb8` plus its rounded size | `t8_port_open3` | `t8_port_vec8`, `t8_port_vec16` | hierarchy |
| The file table holds compile and local paths | `t2_slv8` against `t1_vec8` | 241 of 241 | hierarchy |
| Regions 14 and 15 are executable statement lines per file | `t6_proc2` | `t2_hier3` | hierarchy |
| Instance records: handle, second handle, scope, kind, declaration | `t2_flat3` | 241 of 241 | hierarchy |
| The second handle is the handle plus the value size rounded to 8 | `t2_record_two` against `t1_two_bits` | `t2_array2d`, `t2_record_nested` | hierarchy |
| Equal generics share a unit; different generics duplicate it | `t4_gen_same_two` against `t4_gen_diff_two` | | hierarchy |
| A generic is an object with one record at time 0 | `t4_gen_default` | `t4_gen_explicit` | hierarchy |
| A process variable is an object with no records | `t6_var_int` | `t6_proc2` | hierarchy |
| A loop index is an object with one record at time 0 holding 0 | `t5_tr1000` | `t6_tr1300` | hierarchy |
| A generate index is an object whose record holds the iteration value | `t7_gen_for` | | hierarchy |
| An architecture constant is an object with one record at time 0 holding its value | `t8_gen_if` | | values |
| A net holds one time 0 record per object sharing its handle | `t8_port_chain` | `t8_port_in` | values |
| Only a value change gets a record; a same value assignment writes nothing | `t8_delta_same`, `t8_same` | `t8_delta3` | values |
| Times are in picoseconds and femtoseconds are truncated | `t8_ps` | 241 of 241 end times | values |
| A `real` field contributes no triple to an outer record field | `t8_rec_realv` against `t7_rec_intv` | `t7_rec_in16` | types |
| The first word of a type entry is the source language: `2` VHDL, `0xa` VHDL `TIME`, `1` Verilog, `5` Verilog predefined, `0xd` Verilog `time` | `t11_v_bit_edge` against `t1_bit_one_edge` | all 44 tier 11 cases, `t2_time` | types |
| The half word after it in an array or record is `1` VHDL, `2` unpacked, `3` packed | `t11_sv_struct` against `t11_sv_ustruct` | `t11_v_mem4`, both values in one file | types |
| The word before an array entry's triples counts them | `t11_v_time`, one constrained triple under a `1` | 241 of 241 | types |
| `logic` and `bit` are four literal enumerations `0 1 Z X` and `0 1 0 0`, told apart by the variant word | `t11_v_bit_edge`, `t11_sv_bit` | `t11_sv_int` | types |
| An unnamed Verilog vector is one shared array entry with `(0, 0, -2)`; `integer`, `time`, `int`, `byte`, `longint` are named entries with their own bounds | `t11_v_integer` against `t11_v_vec8` | `t11_v_time`, `t11_sv_byte`, `t11_sv_longint`, the five `t11_v_vec*` | types |
| Signedness is not recorded | `t11_v_signed8` against `t11_v_vec8`, identical dumps | | types |
| A `typedef` is an entry of kind `0x07` naming another entry | `t11_sv_struct` against `t2_record` | every `t11_sv_struct*`, `t11_sv_enum*` | types |
| An enum is an entry of kind `0x04`: base type, named values, then bounds with no `-99` | `t11_sv_enum` against `t2_enum` | `t11_sv_enum4` | types |
| Unit kinds `0x00` module, `0x05` named block, `0x07` process | `t11_v_bit_edge` against `t1_bit_one_edge` | `t11_v_always` | hierarchy |
| Process scopes are named `Initial<line>_<n>`, `Always<line>_<n>`, `NetRegassign<line>_<n>` | `t11_v_bit_edge`, `t11_v_always`, `t11_v_wire` | `t11_v_port` | hierarchy |
| A module with initializers gets one implicit `Initial` scope at the first initializer's line, costing `0x98` of handle space | `t11_v_bit_edge` against `t11_sv_logic` | `t11_v_two_w64`, `t11_sv_enum`, `t11_sv_str` | hierarchy |
| A generate loop folds into the instance name, `g[0].dut`, with no scope of its own | `t11_v_gen_for` against `t7_gen_for` | | hierarchy |
| Declaration kinds `0x00` variable, `0x01` parameter, `0x03` net and port; word 4 is bits | `t11_v_wire`, `t11_v_param` | `t11_v_port`, every tier 11 size | hierarchy |
| Nets get handles before variables; an output port shares the wire's handle; objects are `0xb8` plus the record size apart | `t11_v_wire`, `t11_v_port` | `t11_v_two_w64` | hierarchy |
| A `string` variable has no type, declaration or object, only its implicit scope | `t11_sv_str` | | hierarchy |
| A Verilog value is `8 * ceil(bits / 32)` bytes of `[u32 a][u32 b]` pairs; bit `i` is `a[i] + 2 b[i]` over `0 1 Z X` | `t11_v_bit_edge` against `t1_bit_one_edge` | `t11_v_vec64x`, `t11_v_vec100`, every tier 11 record | values |
| A record holds the pairs an assignment touched, keyed at the handle plus 8 per pair | `t11_v_mem8`, nine records at time 0 | `t11_v_vec64x`, `t11_v_mem2w40` | values |
| A memory is contiguous bits, leftmost element at the top | `t11_v_mem4` against `t11_v_mem4_desc` | `t11_v_mem4w5`, `t11_v_mem3w5`, the `t11_v_mem2w*` sweep | values |
| An unpacked struct gives each field its own pairs, last field lowest; a packed struct is a vector, first field at the top | `t11_sv_ustruct` against `t11_sv_struct` | `t11_sv_struct3`, `t11_sv_struct40`, `t11_sv_struct_r`, `t11_sv_pstruct40` | values |
| A `real` is one pair holding the `float64`, declared size 32 | `t11_v_real` | `t11_sv_struct_r` | values |
| The first record holds the value before any process runs, `X` for four state types, and a `.v` initializer records after it | `t11_v_bit_edge` against `t11_sv_logic` | `t11_v_mem8`, `t11_sv_enum4` against `t11_sv_enum` | values |
| A change due at the `$finish` time is not recorded | `t11_v_always` | | values |
| The chunk rule applies to a Verilog value's record bytes, `8 * ceil(bits / 32)`, with the VHDL constants, and a chunk boundary may fall inside a word pair | `t12_v_vec1089` against `t12_v_vec1088`, one chunk of 272 bytes then four of 70 | `t12_v_vec2272`, `t12_v_vec2304`, `t12_v_vec4800`, `t12_v_vec12000`, 22 chunks over three arenas | values |
| A pair write into a chunked value is one 8 byte record at the pair's own address, not a chunk | `t12_v_vec4800x`, bit 2400 at 75 ns, one record at the handle plus 600 | `t12_v_mem40w32`, forty element writes into a four chunk memory | values |
| Records of one time are kept in write order within an arena only; the order across arenas is lost | `t12_v_mem40_t0`, forty writes at time 0 split over two arenas | `t12_v_mem40w32`, the same writes at distinct times, in order | values |
| Without an initializer a `.v` `reg` and an `.sv` `logic` hold one `X` record at time 0 and no implicit scope; an `.sv` `enum` holds its first literal | `t12_v_noinit` against `t11_v_bit_edge` | `t12_sv_noinit`, `t12_sv_enum_noin` | values |
| `Z` is `b` alone and `X` both words, bit by bit | `t12_v_vec8_z`, `8'bz0z1xx01` as `a = 0x1d`, `b = 0xac` | `t11_v_vec64x` | values |
| A `shortreal` is the predefined `real` entry, declared 32 bits, one pair holding a `float64` | `t12_sv_shortreal` against `t11_v_real` | | values |
| An alias entry carries a range count and triples; a `typedef` of a vector holds the vector's bounds there and the declaration holds none | `t12_sv_typedef` against `t11_v_vec8` | `t11_sv_struct`, `t11_sv_enum`, whose aliases count 0 | types |
| An unpacked array has one array entry per unpacked dimension, each layout `2` of the entry inside it, with one `(0, 0, -2)` triple per dimension so far | `t12_sv_unp2d` against `t11_v_mem4` | `t11_v_mem4`, one entry for one dimension | types |
| A parameter's type is the vector entry for an untyped or ranged parameter, `integer` or `real` for a typed one; a `real` parameter declares 16 bits | `t12_v_params` against `t11_v_param` | `t12_v_param64` | types |
| A negative bound is stored as is, `(-4 to 3)`, and changes no record | `t12_v_neg_range` against `t11_v_vec8_asc` | | types |
| Unit kinds `0x03` task, `0x04` function; each is a scope holding its arguments and locals as objects, a function's return variable first | `t12_v_task` against `t11_v_bit_edge` | `t12_v_func` | hierarchy |
| A `reg` declared in a generate block is a declaration of the module under the escaped name `\g[0].r `, with no generate scope | `t12_v_gen_reg` against `t11_v_gen_for` | | hierarchy |
| Procedural processes are numbered module by module in post order, children before parents, source order then the implicit `Initial`; continuous assignments follow, parents first | `t12_v_proc_order` against `t11_v_port` and `t11_v_hier1` | `t11_v_gen_for`, `tb.Initial14_4` after two children | hierarchy |
| An input port driven by a `wire` shares the wire's handle; driven by a `reg` it does not; an `output reg` port does not share its wire | `t12_v_port_wire` against `t11_v_port` | `t12_v_port_vec8`, `t12_v_port_reg` | hierarchy |
| Parameters of one module sit at consecutive handles, 8 bytes per 32 bits of value, in a second arena | `t12_v_params`, five at `0x8c0` to `0x8e0` | `t12_v_param64`, 64 bits at `0x8c0` then `0x8d0` | hierarchy |
| Unit kind `0x01` is an interface; the instance is a scope of that unit, and an interface port of a child is a second scope of the same unit whose objects share the instance's handles | `t13_sv_iface` against `t11_v_port`, `tb.b.d` and `tb.dut.p.d` both at `0x768` | | hierarchy |
| Unit kind `0x08` is a SystemVerilog package, a scope beside `tb` under the root; its parameter is an object with no record and its typedef an alias entry | `t13_sv_pkg` against `t12_sv_typedef` | `t9_port_rec`, the VHDL package scope in the same place | hierarchy |
| `always_ff` and `always_comb` are `Always` scopes like `always` | `t13_sv_alwaysff` against `t11_v_always` | `t13_sv_iface`, `Always8_0` for the child's `always_comb` | hierarchy |
| An `inout` port has mode `0` and shares the net's handle; a `Z` driven onto the net is one record | `t13_v_inout` against `t12_v_port_wire` | | hierarchy |
| Over three levels the nets take handles in pre order and the variables follow them, and every port shares the handle of the net it is connected to | `t13_v_hier3_net` against `t12_v_port_wire` | `t12_v_proc_order`, two levels | hierarchy |
| A named block with a declaration is a unit of kind `0x05` holding the declaration, and its scope holds the object | `t13_v_blk_var` against `t11_v_always` | | hierarchy |
| A `reg` in an `if` generate is declared as `\g.r `, with one implicit `Initial` | `t13_v_gen_if_reg` against `t12_v_gen_reg` | | hierarchy |
| A `typedef` of an unpacked array carries every range in the alias, the unpacked one first, and the declaration none | `t13_sv_tdef_ua` against `t12_sv_unp2d` | `t12_sv_typedef` | types |
| A string parameter is an unnamed vector of 8 bits per character, the first character at the top | `t13_v_str_param` against `t12_v_params`, `"hello"` as 40 bits | | types |
| An unpacked array of `real` gives each element one pair, the last element lowest, and a write of the value already held produces no record | `t13_sv_real_arr` against `t11_v_real` | `t11_sv_ustruct`, the same order for fields | values |
| An unpacked array of packed structs is one contiguous value, element 0 at the top | `t13_sv_struct_ar` against `t11_sv_struct` | `t12_sv_unp2d` | values |
| Three writes to one variable at one time are three records, in write order | `t13_v_same_t` against `t11_v_bit_edge` | `t12_v_mem40_t0` | values |
| `log_wave -recursive /sig_pkg` logs a package signal; it records like a signal of `tb` in an arena of its own and the logged range count grows by one | `t13_pkg_log_all` against `t9_pkg_sig` | | values |
| A Verilog variable whose arena spills into a second page has no `X` record at time 0 | `t13_v_tr430` against `t13_v_tr420`, 430 records against 421 | `t13_v_tr2000`; `t13_v_tr430_2`, where `d` in its own arena keeps its `X` | values |
| A change due at the `std.env.stop` time is recorded, where one due at the `$finish` time is not | `t13_tr430` against `t13_v_tr430`, 431 records against 430 | `t11_v_always` | values |

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
| `t8_port_in` against `t1_hier1` | a port on the child | declaration word 9 is the port mode; a connected port shares the signal's handle |
| `t8_port_out`, `t8_port_inout`, `t8_port_buffer` against `t8_port_in` | the port mode | the mode values 2, 0, 3; the handle is shared whatever the mode |
| `t8_port_open` against `t8_port_in` | ports left open | an open port owns a handle; the `line__NN` process scope |
| `t8_port_open3` against `t8_port_open` | three open ports beside a signal | an open port's stride is `0xc0` where a signal's is `0xf0` |
| `t8_port_vec8`, `t8_port_vec16` against `t8_port_open` | the open port's width | the stride is `0xb8` plus the rounded size |
| `t8_port_chain` against `t8_port_in` | a port two levels down | every object on the net shares the handle and adds a time 0 record |
| `t8_delta3` against `t7_delta` | three deltas | one record per delta |
| `t8_delta_same`, `t8_same` against `t8_delta3` | assignments of the held value | no record without a change |
| `t8_ps` against `t1_bit_two_edges` | waits of 1 ps and 1500 fs | picosecond unit, femtoseconds truncated |
| `t8_rec_realv` against `t7_rec_intv` | a real beside the vector | a real contributes no triple |
| `t8_gen_if` against `t7_gen_for` | an if generate with a constant condition | plain branch scopes, an empty false branch, kind `0x13` is a constant |
| `t8_gen_nest` against `t7_gen_for` | a nested for generate | the iteration and empty label scopes repeat per level |
| `t9_vec292` against `t9_vec256`, `t9_vec257` | value size 292 not 256 or 257 | values over 257 bytes are chunked into records with consecutive keys |
| `t9_int73` against `t9_vec292` | 73 integers of the same 292 bytes | chunks split bytes, not elements |
| the 18 `t9_vec*` sizes | the value size | the chunk size table; a first reading of 146 byte chunks |
| `t10_vec274`, `t10_vec275`, `t10_vec276` | 274, 275, 276 bytes | one, two, four chunks: the split starts at 275 |
| `t10_vec574` against `t10_vec575`, `t10_vec872` against `t10_vec874` | one byte across a step | the count steps by two every 299 bytes from 276 |
| `t10_vec20000`, `t10_vec30000` against `t9_vec12000` | 20000 and 30000 bytes | chunks of 149 and 148, so 146 was not a limit; the rule holds to 202 chunks |
| `t10_real40` against `t10_vec320` | 8 byte elements | the element type does not enter the chunking |
| `t9_vec4096`, `t9_vec12000` against `t9_vec2048` | values wider than an arena | a value spans arenas; a slot in the middle of the table can be 0 |
| `t9_tr70000` against `t6_tr1300` | 70000 transitions | arena records continue through word 0 past 100 pages |
| `t9_port_slice` against `t8_port_in` | the port bound to `x(0)` of a 2 bit vector | instance word `+20` is a byte offset into the net's value |
| `t9_port_slice2`, `t9_port_sliceto` against `t9_port_slice` | a 2 bit slice, then a `to` range | the offset counts bytes from the left element |
| `t9_port_rec` against `t2_record` | a package constant beside the type | a package scope of kind `0x0a`; an unlogged object; the marker is a list of ranges |
| `t9_mark_gap` against `t9_port_rec` | the package constant first | the first marker word is an index, not 0 |
| `t9_mark_two` against `t6_var_int` | two instances with a process variable | variable objects multiply per instance; a second logged range |
| `t9_var_inst3` against `t9_mark_two` | three instances | nine objects for three variables, `0x118` apart |
| `t9_pkg_sig` against `t1_bit_one_edge` | a signal in a package | the package signal takes the first handle and is not logged; arena 0 unwritten |
| `t9_port_lnk` against `t8_port_in` | mode `linkage` | port mode 4 |
| `t9_port_expr` against `t8_port_in` | the port bound to `'1'` | a literal bound port owns a handle |
| `t9_gen_types` against `t4_gen_default` | four generics of other types | generics record in the type's encoding |
| `t9_block` against `t7_gen_for` | a block instead of a generate | a block is unit kind `0x0c` too |
| `t9_comp`, `t9_alias`, `t9_func`, `t9_proc_local` against `t8_port_in`, `t1_bit_one_edge` | a component, an alias, a function, a procedure | nothing in the file but 8 bytes of handle space per subprogram |
| `t9_proc_sig` against `t9_proc_local` | a `signal` parameter | the change is recorded twice; `0x48` of handle space |
| `t11_v_bit_edge` against `t1_bit_one_edge` | the source language | the origin word; unit kinds `0x00` and `0x07`; the implicit `Initial` scope; 8 byte word pair records; the `X` record at time 0 |
| `t11_sv_logic` against `t11_v_bit_edge` | `logic` in a `.sv` file for `reg` in a `.v` file | no implicit scope, `0x98` less handle space, no `X` record |
| `t11_v_vec8` against `t11_v_bit_edge` | an 8 bit vector | the unnamed vector entry; bounds in the declaration; size in bits |
| `t11_v_vec8_asc` against `t11_v_vec8` | `[0:7]` for `[7:0]` | the same record; the leftmost bit is the top |
| `t11_v_signed8` against `t11_v_vec8` | `reg signed` | nothing; signedness is not recorded |
| `t11_v_vec33`, `t11_v_vec100` against `t11_v_vec8` | 33 and 100 bits | one pair per 32 bits; bits above the width are 0 |
| `t11_v_vec64x` against `t11_v_vec8` | a bit set to `x` at 75 ns | `X` is both words; a partial record of one pair at handle plus 8 |
| `t11_v_integer` against `t11_v_vec8` | `integer` for `reg [7:0]` | a named array entry with its own `(31, 0, -1)`; no declaration ranges |
| `t11_v_time`, `t11_v_real` against `t11_v_integer` | `time`, `real` | origin `0xd`; the range count word is `1` under one constrained triple; a real is one pair of `float64` with size 32 |
| `t11_sv_int`, `t11_sv_byte`, `t11_sv_longint` against `t11_v_integer` | the SystemVerilog integral types | `bit` based named entries of 32, 8 and 64 bits; no `X` record |
| `t11_v_two_w64` against `t11_v_vec8` | a 64 bit and a 1 bit variable | the next handle is the second handle plus `0xb8` |
| `t11_v_wire` against `t11_v_bit_edge` | a `wire` driven by `assign` | declaration kind `0x03`; the `NetRegassign` scope; nets take handles first |
| `t11_v_hier1` against `t1_hier1` | a child module | an instance scope with the child's unit; both unit lines point at `module` |
| `t11_v_port` against `t8_port_in` | ports on the child | ports are nets with the mode in word 9; the output shares the wire's handle; the input does not share the `reg`'s |
| `t11_v_param` against `t11_v_hier1` | a `parameter` | declaration kind `0x01`; an object with no second handle, one record, the vector type |
| `t11_v_gen_for` against `t7_gen_for` | a Verilog generate loop | `g[0].dut` under `tb` with no generate scope; no `genvar` object |
| `t11_v_always` against `t11_v_bit_edge` | two `always` blocks and a named block | `Always` scopes; unit kind `0x05`; the toggle at `$finish` is unrecorded |
| `t11_v_mem4` against `t11_v_vec8` | a memory of four bytes | layout `2` on the outer entry; contiguous bits; one record per element write |
| `t11_v_mem4_desc` against `t11_v_mem4` | `[3:0]` for `[0:3]` | the leftmost element is the top, whatever its index |
| `t11_v_mem8` against `t11_v_mem4` | eight elements, two pairs | a record holds only the pair its element lives in |
| `t11_v_mem4w4`, `t11_v_mem4w5`, `t11_v_mem3w5` against `t11_v_mem4` | element widths that do not fill a byte | no padding between elements |
| `t11_v_mem2w9` to `t11_v_mem2w64` against `t11_v_mem4` | element widths across the 32 bit boundary | elements straddle pairs; records cover the pairs touched |
| `t11_sv_struct` against `t2_record` | a packed struct | the unnamed record entry, layout `3`; the `0x07` alias holding the name |
| `t11_sv_ustruct` against `t11_sv_struct` | the same struct unpacked | layout `2`; one pair per field, last field lowest; size rounds each field to 32 |
| `t11_sv_struct3`, `t11_sv_struct40`, `t11_sv_struct_r` against `t11_sv_ustruct` | three fields, a 40 bit field, a `real` field | fields take the pairs a standalone value would, low word first |
| `t11_sv_pstruct40` against `t11_sv_struct40` | the 41 bit struct packed | a 41 bit vector, first field at the top |
| `t11_sv_arr2d` against `t11_v_vec8` | `logic [1:0][3:0]` | an array of the vector entry, `dims` still `1`, two declaration ranges |
| `t11_sv_enum` against `t2_enum` | a SystemVerilog enum over `int` | kind `0x04` with named values and a base type; the implicit scope; one record |
| `t11_sv_enum4` against `t11_sv_enum` | the enum over `logic [3:0]` with values 1, 5, 9 | the base is the vector entry and the bounds follow the values; an `XXXX` record first |
| `t11_sv_str` against `t11_sv_logic` | a `string` | no type, declaration or object; the implicit scope remains; the handle space of a one pair variable |
| `t12_v_vec1089` against `t12_v_vec1088` | one more bit, 280 bytes of record against 272 | the chunk rule holds for Verilog with the VHDL constants, and splits inside a pair |
| `t12_v_vec4800x` against `t12_v_vec4800` | one bit of a ten chunk vector set at 75 ns | a pair write into a chunked value is an 8 byte record at the pair's address |
| `t12_v_mem40_t0` against `t12_v_mem40w32` | forty element writes at time 0, against one per ns | the order of one time's records across arenas is lost; an 8 byte split rest looks like a pair write |
| `t12_v_vec12000` against `t9_vec3000` | the same 3000 bytes in Verilog | the same 22 chunks over three arenas |
| `t12_v_noinit` against `t11_v_bit_edge` | no initializer | no implicit `Initial` scope, one `X` record |
| `t12_sv_enum_noin` against `t11_sv_enum` | no initializer on an `.sv` enum | no implicit scope; one record holding the first literal, no `X` |
| `t12_v_vec8_z` against `t11_v_vec8` | `8'bz0z1xx01` | `Z` is the `b` word alone, bit by bit |
| `t12_sv_shortreal` against `t11_v_real` | `shortreal` | the same predefined `real` entry and record |
| `t12_sv_typedef` against `t11_v_vec8` | `typedef logic [7:0] byte_t` | the alias entry carries the bounds; the declaration carries none |
| `t12_sv_unp2d` against `t11_v_mem4` | a second unpacked dimension | one array entry per unpacked dimension, nested |
| `t12_v_params` against `t11_v_param` | five parameters of five declarations | parameter types, the 16 bit `real` parameter, consecutive handles |
| `t12_v_param64` against `t12_v_params` | a 64 bit parameter | parameter handles advance by the record size |
| `t12_v_neg_range` against `t11_v_vec8_asc` | `reg [-4:3]` | the bounds are stored signed; the record is unchanged |
| `t12_v_task` against `t11_v_bit_edge` | a `task` | unit kind `0x03`; the arguments and locals are objects with records |
| `t12_v_func` against `t12_v_task` | a `function` | unit kind `0x04`; the return variable is declared first and written last |
| `t12_v_gen_reg` against `t11_v_gen_for` | a `reg` in the generate block, no instance | the escaped name `\g[0].r ` in the module; one implicit `Initial` per iteration |
| `t12_v_proc_order` against `t11_v_port` and `t11_v_hier1` | a child and a parent with an `initial`, an initializer and an `assign` each | the process counter order |
| `t12_v_port_wire` against `t11_v_port` | the input port driven by a `wire` | the input port shares the wire's handle; three `X` records on the shared net |
| `t12_v_port_reg` against `t11_v_port` | `output reg` | an output `reg` port has its own handle; the parent's wire keeps its own |
| `t13_sv_iface` against `t11_v_port` | an interface in place of ports | unit kind `0x01`; the interface port is a scope sharing the instance's objects |
| `t13_sv_pkg` against `t12_sv_typedef` | the typedef moved into a package with a parameter | unit kind `0x08` beside `tb`; the package parameter is an unlogged object |
| `t13_sv_alwaysff` against `t11_v_always` | `always_ff` and `always_comb` | `Always` scopes, nothing new |
| `t13_v_inout` against `t12_v_port_wire` | an `inout` port and a `Z` driver | port mode `0`; the same sharing; `Z` as one record |
| `t13_sv_tdef_ua` against `t12_sv_unp2d` | a typedef of an unpacked array | the alias carries both ranges; the declaration carries none |
| `t13_sv_real_arr` against `t11_v_real` | `real r [0:1]` | one pair per element, last lowest; no record for an unchanged value |
| `t13_sv_struct_ar` against `t11_sv_struct` | an unpacked array of a packed struct | one contiguous 10 bit value, element 0 at the top |
| `t13_v_str_param` against `t12_v_params` | `parameter P = "hello"` | a 40 bit unnamed vector, `h` at the top |
| `t13_v_same_t` against `t11_v_bit_edge` | three writes in one time step | three records in write order |
| `t13_v_hier3_net` against `t12_v_port_wire` | a third level of nets and ports | nets pre order over three levels, then variables; ports share upward |
| `t13_v_gen_if_reg` against `t12_v_gen_reg` | a `reg` in an `if` generate | `\g.r `, one implicit scope |
| `t13_v_blk_var` against `t11_v_always` | a `reg` declared in a named block | the block unit holds the declaration and the block scope the object |
| `t13_pkg_log_all` against `t9_pkg_sig` | `log_wave -recursive /sig_pkg` added to the script | the package signal records like a signal of `tb`; the logged range count grows |
| `t13_v_tr2000` against `t11_v_always` | two thousand toggles | five pages of 425 Verilog records; no `X` record at time 0 |
| `t13_v_tr430` against `t13_v_tr420` | 430 toggles against 420, one page against two | the `X` record goes with the spill into a second page |
| `t13_v_tr430_2` against `t13_v_tr430` | a second `reg` in its own arena | the `X` record of the arena that does not spill stays |
| `t13_tr430` against `t13_v_tr430` | the same clock in VHDL | one page of 431 one byte records; the toggle at `std.env.stop` is recorded |

Three findings were not found by a pair.
The end time, the type count and the trailer word at `+0x30`, then
read as a has-objects flag and now the logged range count, came from
the correlation sweep: read the same offset in every case and keep the
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

For a Verilog design the same VCD is fuller: `integer`, `real`,
`time`, a struct and an enum each get a `$var`, and only memories and
strings are absent, measured on tier 11.

Two consequences follow.

The VCD answer key only covers bit and vector signals of a VHDL design.
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
3. Word 28 of an instance record is 2 for every object that is not a
   signal, and word 44 is `-1`, 0, or a value that differs between
   runs of the same design.
   Both are read off and masked, and what they hold is open.
4. DBG header words 14 to 16 are `0x101`, `0x101`, `0x10000` in every
   case, and the three `0x30` words at `0x98` and the `3` at `0xc0` in
   the fixed header are constant too.
   No case has moved them.
5. Word 10 of a declaration record varies between runs for a signal and
   is 0 for a variable.
   It is masked as noise and not read.
6. Handles of generics, constants, variables and loop indexes follow
   no pattern seen yet.
   They are read from the instance record, so nothing depends on it.
7. The chunk rule of [format/values.md](format/values.md) holds for
   55 wide VHDL values and 8 wide Verilog values, and its constants,
   275, 24 and 299, have no explanation.
   The 275 and the 299 differ by the 24, so the rule may be one
   constant and one threshold in disguise.
8. Whether `0xc4` and the other per-run durations mean anything is
   open.
   They are masked.
9. Does the format change between Vivado versions?
   Only 2025.2 is in use here.
   Any claim is version scoped until a second version has been
   measured.
10. The variant and class words of an enumeration entry, `2` and a
    class per VHDL type, `0` or `1` and `1` for `logic` and `bit`, and
    the `1` after a Verilog `real`, separate the types they are seen
    on and nothing else is known about them.
    The `dims` word of an array entry is `1` in every entry, including
    the two dimensional ones.
11. A package signal is not logged under `log_wave -recursive *`
    and is under `log_wave -recursive /sig_pkg`, `t13_pkg_log_all`.
    A package parameter, `p.W` of `t13_sv_pkg`, is an object with no
    record under either script, so whether anything logs it is open.
12. Which of the `n` duplicated variable handles in an entity
    instantiated `n` times belongs to which instance is not readable
    from the file.
    Nothing depends on it while variables have no records.
13. The handle space costs of a subprogram, 8 bytes, and of a `signal`
    parameter, `0x48`, name objects the instance list does not
    contain.
    What they are is open.
14. A `real` parameter declares 16 bits, `t12_v_params`, where a
    `real` variable declares 32 and both hold one `float64` pair.
    Where the 16 comes from is open.
15. The records of one time keep their write order within an arena
    only.
    `t12_v_mem40_t0` writes forty elements at time 0 and the file holds
    the last nineteen in arena 0 before the first twenty one in arena
    1, so a same time write order across arenas cannot be read back.
    The reader replays the arenas in file order, and the test compares
    the final value of each time.
16. An 8 byte rest of a chunk split at an arena boundary has the shape
    of a pair write, `t12_v_mem40w32` at `0x800`.
    The reader tells them apart by count: a time with fewer records at
    one chunk address than at the others treats the surplus as pair
    writes.
    A design that writes the split pair and a whole value in the same
    time step would defeat that.
17. A `string` variable takes the handle space of a one pair variable
    and appears nowhere else; where its value goes, if anywhere, is
    open.
18. A net holds one `X` record at time 0 per object on its handle, plus
    one more when a port shares it: `t12_v_port_wire` holds three on
    the handle of `x` and `a`, and `t11_v_port` two on the input port
    alone.
    `t13_v_hier3_net` holds three on `w0`, a wire with an input port
    on it, three on `w1`, the same one level down, two on `w2`, a
    wire alone, and three on `y`, a wire with two output ports on it,
    and `t13_v_inout` holds two and then a `Z` on a wire with an
    `inout` port and a `Z` driver.
    What the extra record is written by is open.
19. The `X` record at time 0 of a Verilog variable is absent when its
    arena spills into a second page: `t13_v_tr430` holds 430 records
    starting at `0`, where `t13_v_tr420` holds 421 starting at `X`.
    The record may be dropped when the first page is written out, or
    written into a page still in memory at the close.
    The two readings differ for a variable written in the second page
    only, which no case has.

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

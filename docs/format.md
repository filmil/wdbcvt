<!-- SPDX-License-Identifier: Apache-2.0 -->

# Decoding the xsim waveform database (`.wdb`)


## What this document is

AMD does not document the `.wdb` container that `xsim` writes.
This file is the index of what has been measured about it, the findings
table, and the record of which comparison led to which discovery.
Everything here is either a measurement or a statement marked as a
guess.
Nothing gets promoted from guess to fact without a reproduction.

The layout itself is described in four documents, one per area, and a
fifth covers the VCD written beside every database:

| Document | Area |
| :--- | :--- |
| [format/container.md](format/container.md) | the fixed header, the arena table, the trailer, the directory, the page directory and the marker |
| [format/types.md](format/types.md) | the type table: enumerations, integers, reals, physical types, arrays and records, and the Verilog entries |
| [format/hierarchy.md](format/hierarchy.md) | the debug section: scopes, units, declarations, source files, instance records and handles, in both languages |
| [format/values.md](format/values.md) | arenas, pages, value records, encodings and alignment, and the Verilog word pairs |
| [format/vcd.md](format/vcd.md) | what Vivado's VCD holds of the same run, how it spells values, and the cross-check of every case against it |

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
against the `truth.json` of all 1119 cases, and against the `sim.vcd`
of every case for the objects a VCD can hold.
The offsets are in the documents linked in the last column.

| Finding | Found by | Confirmed by | Where |
| :--- | :--- | :--- | :--- |
| Magic `Xilinx WAVE DATABASE 01`, producer `Xilinx Simulator` | hex dump of `t1_bit_one_edge` | all 1119 cases | container |
| `0x38` is a Unix timestamp | noise mask, two runs of `t3_tr1` | equals the file mtime | container |
| `0x48` holds three pointers to 48 byte directory entries | `strings -t d` on `t2_flat3`, then reading the values | all 1119 cases, each pointer lands on a name | container |
| The arena table at `0xc8` grows with the object count | `t5_sig10` shifted every trailer field by 8 | 3, 4, 6 slots in `t6_sig05`, `t6_sig12`, `t6_sig20` | container |
| The arena table has `ceil(handle space / 0x800)` slots | `t7_sig07` broke the `ceil(objects / 4) + 1` guess | the reader checks it in 1119 of 1119 | container |
| Trailer `+0x0c` is the arena table slot count | sweep of every fixed word over the 63 cases of tier 7 | 1119 of 1119, checked by the reader | container |
| Trailer `+0x18` is the handle space size | `t7_sig07` to `t7_sig24` against the slot count | the slot rule, 1119 of 1119 | container |
| Arena records sit in first write order, not arena order | `t7_gen_for`, arena 2 first | 1119 of 1119 with the reader accepting any order | container |
| The trailer is the 0x48 bytes before the first directory pointer | `t5_sig10` against `t6_sig05` | end time correct in all 63 | container |
| The end time is a uint64 at trailer `+0`, in the file's time unit | correlation sweep over 33 cases | 1119 of 1119, 20 ns to 70010 ns; 100 for 100 ns in `t21_v_ts_1ns_1ns` | container |
| The marker offset is at trailer `+0x38` | `t5_tr1000`, where the marker moved | 1119 of 1119 | container |
| The marker is a list of `[first][last]` object index ranges, as many as trailer `+0x30` counts, covering exactly the objects with records | `t9_port_rec` held two entries where `t6_var_int` had shown one; `t9_mark_gap` put an unlogged object first | 1119 of 1119, the reader checks every object against the ranges | container |
| An arena record's word 0 names a continuation record for pages past 100 | `t9_tr70000`, 117 pages | the reader reads the 70001 records back | container |
| Each directory entry follows the section it describes | `t2_flat3`: `WDB.Event` at `0xe0+0x48`, RTTI and DBG the same | 1119 of 1119 | container |
| The page directory starts 48 bytes after the DBG entry | `t2_flat3`, reading the offsets | 1119 of 1119 | container |
| An arena record is `0x4c0` bytes: 100 page offsets, 100 lengths, a count | `t5_tr1000`, two pages in one arena | `t6_tr1300`, three pages | container |
| A page is a zlib stream that inflates to 10240 bytes | entropy profile, then `zlib` on `t1_bit_one_edge` | 1119 of 1119 | values |
| Page header `[t0][last minus t0][n]` | `t5_tr1000` page 1 | all pages of all cases | values |
| A record is `[uint64 time][uint32 key][uint32 length][value]` | `t1_bit_two_edges` against `t1_bit_one_edge` | every record of every case matches `truth.json` | values |
| `handle >> 11` is the arena, `handle & 0x7ff` the key | `t5_sig10` | `t6_sig20`, four arenas | values |
| A page holds 10240 bytes of records and overflows into a new page | `t5_tr1000`, 600 one byte records | `t7_int700` 510 and `t7_wide700` 425 | values |
| Records at one time are in simulation order, not key order | `t7_gen_for` | `t2_flat3` | values |
| A delta cycle leaves two records at one time | `t7_delta` | | values |
| An overflowed page precedes the marker | `t5_tr1000` | `t6_tr1300` | values |
| A value of 275 bytes or more is written as chunks with consecutive keys, one record per chunk, addressed by handle plus byte offset | `t9_vec292` against `t9_vec261` and `t9_vec257` | the 18 `t9_vec*` sizes and `t9_int73`, read back against `truth.json` | values |
| A value of `size >= 275` bytes is `2 * ceil((size + 24) / 299)` chunks of `floor(size / n)` bytes, the last taking the rest | `t10_vec274` to `t10_vec276`, `t10_vec574` to `t10_vec575`, `t10_vec872` to `t10_vec874` | the reader enforces the addresses on all 55 wide values | values |
| Chunks split at an arena boundary and inside an element | `t9_vec292`, 6 plus 67 bytes across `0x800` | `t9_int73`, 73 byte chunks of 4 byte integers | values |
| The chunk rule applies to a write, from the write's own address and for its own length | `t32_wide_slice__` against `t32_vec_slice___`, four records of 75 at the handle plus 300 for one of 4 | `t32_wide_top____`, `t32_wide_field__`, `t32_wide_tail___`; `t32_wide_small__`, 4 bytes in one record | values |
| A VHDL assignment to a field, slice or element writes a record of the part's bytes at the handle plus the part's byte offset | `//hdl/counter:sim` against the corpus, a record driven one field at a time | 28 t32 cases: `t32_rec_field___` at `+1`, `t32_vec_slice___` at `+4`, `t32_vec_to_slice` at `+0`, `t32_arr_row_bit_` at `+13`, `t32_arr2d_elem__` at `+3`, `t32_rec_intlast_` at `+4` | values |
| The parts one delta assigns are one record where they touch and two records where they do not; a whole assignment in the same delta folds them into one whole record | `t32_rec_two_adj_` against `t32_rec_two_gap_`, 2 bytes at `+1` for 1 byte at `+0` and 1 at `+2` | `t32_vec_adj_slc_`, `t32_vec_two_slc_`, `t32_rec_wthenf__`, `t32_rec_fthenw__`, `t32_vec_slc_over`; `t32_rec_delta___`, two records for two deltas | values |
| A wide value spans as many arenas as its byte range crosses | `t9_vec12000`, arenas 0 to 6 | `t9_vec4096` | values |
| A procedure with a `signal` parameter writes the change twice | `t9_proc_sig` against `t9_proc_local` | | values |
| Enumeration values are one byte, the literal's index | `t3_valz` against `t3_tr1`, same size | `t1_nine_state` walks all nine | values |
| Integers are int32, reals float64, physical values int64 in the base unit | `t2_integer`, `t2_real`, `t2_time` | `truth.json`; `t21_int_neg` two's complement, `t21_real_neg`, `t21_phys_user` in `um` | values |
| Arrays are elements back to back, left index first | `t1_vec8` | `t2_array2d`, `t5_int_arr` | values |
| Record fields are aligned to their size, records to 8 | `t5_rec_real` against `t2_record` | `t5_arr_rec`, `t5_rec_sub5` | values |
| A signal has one record at time 0 and one per change, when logged from the start | `t0_bit_const` | `t3_late`, 1119 of 1119; `t45_log_late____` for a log started at 10 ns | values |
| The type table starts with `Xilinx ISim TYPE FILE 001` | `strings` on `t1_bit_one_edge` | 1119 of 1119 | types |
| `+32` of the type table is the number of types | correlation sweep | 1119 of 1119 | types |
| Type entries are `[len][tag]` name body | `t2_enum` against `t1_bit_one_edge` | 1119 of 1119 | types |
| Enumerations list their literals; `character` has 261 | `t2_enum`, `t2_character` | `truth.json` names | types |
| Integer entries carry the bounds, reals the bounds as float64 | `t2_integer`, `t2_real` | 1119 of 1119 | types |
| Physical entries list units with scales | `t2_time` | | types |
| Arrays carry element, index type and constraint triples | `t1_vec8` against `t2_array2d` | `t5_int_arr` | types |
| Records list fields with types and ranges | `t2_record` | `t2_record_nested`, `t5_rec_sub5` | types |
| A record field of record type lists one range per inner field, the scalar's own range, only when the inner record has an array field | `t7_rec_vfirst`, `t7_rec_bitv`, `t7_rec_intv`, `t7_rec_in2` | `t7_rec_in2v` | types |
| Types are shared between signals of the same type | `t2_record_two` | `t6_sig20`, one `STD_ULOGIC` | types |
| The DBG section starts with `Xilinx ISim DBG 006` and 18 region offsets | `t1_hier1` against `t2_hier3` | 1119 of 1119 | hierarchy |
| Scope records: name, parent, children, first object, unit, file, line | `t2_hier3` | 1119 of 1119 | hierarchy |
| Unit records: entity, architecture, kind, declaration count, file, line | `t2_hier3` | `t4_gen_diff_two` | hierarchy |
| Declaration records: name, file, line, size, type, ranges, kind | `t2_flat3` | 1119 of 1119 | hierarchy |
| Declaration word 1 is the index of the region 17 entry holding the value class of the declaration's objects | `t31_sv_w1_swap` against `t31_sv_w1_i5`, where swapping `int i = 5` and `logic s` swapped the words `1`, `0` to `0`, `1` and the entries `[1 0 0] [3 0 0]` to `[3 0 0] [1 0 0]` | `t12_v_params`, six words `0 1 2 1 0 1` over three entries; the reader's range check in 1119 of 1119; `t31_sv_w1_i0` through `t31_sv_w1_own50`, where the value and the writes leave the word alone | hierarchy |
| Declaration kinds `0x0e` signal, `0x0f` variable, `0x12` generic, `0x13` constant | `t4_gen_default`, `t5_tr1000`, `t6_var_int`, `t8_gen_if` | `t6_proc2`, `t7_gen_for` | hierarchy |
| Declaration word 9 is the port mode: 0 inout, 1 in, 2 out, 3 buffer, 4 linkage, 5 none | `t8_port_in`, `t8_port_out`, `t8_port_inout`, `t8_port_buffer`, `t9_port_lnk` | 1119 of 1119 against the `port` field in `truth.json` | hierarchy |
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
| The file table holds compile and local paths | `t2_slv8` against `t1_vec8` | 1119 of 1119 | hierarchy |
| Regions 14 and 15 are executable statement lines per file | `t6_proc2` | `t2_hier3` | hierarchy |
| Instance records: handle, second handle, scope, kind, declaration | `t2_flat3` | 1119 of 1119 | hierarchy |
| The second handle is the handle plus the value size rounded to 8 | `t2_record_two` against `t1_two_bits` | `t2_array2d`, `t2_record_nested` | hierarchy |
| Equal generics share a unit; different generics duplicate it | `t4_gen_same_two` against `t4_gen_diff_two` | | hierarchy |
| A generic is an object with one record at time 0 | `t4_gen_default` | `t4_gen_explicit` | hierarchy |
| A process variable is an object with no records | `t6_var_int` | `t6_proc2` | hierarchy |
| A loop index is an object with one record at time 0 holding 0 | `t5_tr1000` | `t6_tr1300` | hierarchy |
| A generate index is an object whose record holds the iteration value | `t7_gen_for` | | hierarchy |
| An architecture constant is an object with one record at time 0 holding its value | `t8_gen_if` | | values |
| A net holds one time 0 record per object sharing its handle | `t8_port_chain` | `t8_port_in` | values |
| Only a value change gets a record; a same value assignment writes nothing | `t8_delta_same`, `t8_same` | `t8_delta3` | values |
| Times are in the simulation precision and nothing finer is kept | `t8_ps` | 597 end times at the default 1 ps precision | values |
| The DBG word after the timestamp is the power of ten of the time unit, and every time in the file counts that unit | `t21_v_ts_1ns_1ns` against `t11_v_bit_edge`, `-9` and a change at 50 | `t21_v_ts_1ps_1ps`, `t21_v_ts_10ns`, `t21_v_ts_1ns_100` at `-10`, `t21_v_ts_1ps_1fs` at `-15`; the VCD `$timescale` agrees in 1119 of 1119 | hierarchy, values |
| The finest precision in the design sets the unit; no `timescale` means picoseconds | `t21_mix_ts_1ns` against `t21_mix_vh_in_v` | `t21_v_ts_none` | values |
| Two Verilog instances with different parameter values share one unit record and one declaration set | `t21_v_param_diff` against `t21_v_param_same` | `t4_gen_diff_two`, where VHDL repeats the unit | hierarchy |
| A mixed language design keeps each unit's, declaration's and type's language markers, and the port at the boundary has a handle of its own | `t21_mix_vh_in_v` against `t9_comp` | `t21_mix_v_in_vh` | hierarchy |
| A VHDL port driven from Verilog holds `U` then the driven value at time 0; a Verilog port driven from VHDL holds `X` then the value | `t21_mix_vh_in_v` against `t9_comp` | `t21_mix_v_in_vh` against `t11_v_port` | values |
| An integer subtype or new integer type gets its own entry with the narrow bounds and keeps 4 bytes | `t21_int_sub` against `t2_integer` | `t21_int_newtype`, byte identical to `t21_int_sub` | types |
| A user physical type lists its units scaled in its base unit and the value counts that unit | `t21_phys_user` against `t2_time` | `truth.json` | types, values |
| `--timeprecision_vhdl` sets the DBG word and rescales the `TIME` entry, so a `TIME` value counts the precision unit | `t22_vh_fs` against `t22_base`, `fs=1 ps=1000`, `1500 fs` stored as 1500 | `t22_vh_ns`, `fs=0 ps=0 ns=1` | types, values |
| DBG regions 14 and 15 are the `line` debugging ability and empty under `-debug wave` | `t22_dbg_wave` against `t22_base` | `truth.json`, the values unchanged | hierarchy |
| A VHDL function is a unit of kind `0x11` and a procedure `0x12`, present only under `-debug subprogram`; their locals are declarations of kind `0x14` with the parameter mode, on frame offset handles, never logged | `t22_dbg_subprog` against `t22_base` | `t22_dbg_sub_proc`, `t22_dbg_all` | hierarchy |
| `-debug all` adds the library packages as root children of kind `0x0a` with their constants as unlogged objects and their types in the table | `t22_dbg_all` against `t22_base` | `truth.json`, 19 objects | hierarchy, types |
| A file type is kind `0xc`: origin, the element type, the words `8` and `40` whatever the element; a file object is a 0 byte variable with no record under the default debug level | `t22_dbg_all`, `TEXT` of `STRING` | `t23_file_text`, `t23_file_int`, `t23_file_sul` | types, hierarchy |
| `--O0` and `--mt off` change nothing; `--generic_top` changes the generic's record and what depends on it | `t22_o0`, `t22_mt_off`, `t22_gen_top`, each against `t22_base` | `truth.json` | hierarchy |
| An access type is kind `0x8`: origin, the designated type, the words `8` and `48`; its variable declares 48 bytes and has no record | `t23_access` against `t6_var_int` | `t23_access_vec` | types, hierarchy |
| A shared variable is kind `0x0f` in the entity's scope on a variable handle; one of a protected type has both handles in the signal handles, and the protected type is a record entry | `t23_shared_int` against `t6_var_int`, `t23_protected` against `t23_shared_int` | `truth.json` | hierarchy, types |
| Subprogram frame offsets follow each local's alignment, and a vector local takes 24 bytes whatever its length | `t23_sub_sizes` against `t22_dbg_subprog` | `t23_sub_vec16`, `t23_sub_vec32` | hierarchy |
| A signal parameter is declaration kind `0x15` with its mode, on a 64 byte frame slot | `t23_sub_sig_prm` against `t22_dbg_sub_proc` | `truth.json` | hierarchy |
| A subprogram declared in a process gets two scopes, under the entity and under the process, over one unit and one declaration | `t23_sub_in_proc` against `t22_dbg_sub_proc` | `truth.json`, three objects | hierarchy |
| Each instantiated architecture of an entity is its own unit, named `child(a)`, `child(b)`, with its own declarations and process units | `t23_arch_b` against `t4_gen_explicit` | `t23_arch_both` | hierarchy |
| The predefined `integer_vector`, `real_vector`, `time_vector` and `boolean_vector` are unconstrained array entries over the scalar, the bounds in the declaration | `t23_int_vector` against `t5_int_arr` | `t23_real_vector`, `t23_time_vector`, `t23_bool_vector` | types |
| A `real` field contributes no triple to an outer record field | `t8_rec_realv` against `t7_rec_intv` | `t7_rec_in16` | types |
| The first word of a type entry is the source language: `2` VHDL, `0xa` VHDL `TIME`, `1` Verilog, `5` Verilog predefined, `0xd` Verilog `time` | `t11_v_bit_edge` against `t1_bit_one_edge` | all 44 tier 11 cases, `t2_time` | types |
| The half word after it in an array or record is `1` VHDL, `2` unpacked, `3` packed, `6` packed union | `t11_sv_struct` against `t11_sv_ustruct` | `t11_v_mem4`, both values in one file | types |
| A packed union is a record entry of layout `6` as wide as its widest field, every field over the same bits | `t24_sv_union` against `t11_sv_struct` | `truth.json`, `TestVCD` | types, values, vcd |
| DBG header word 14 flags `-debug drivers` in byte 1 and `readers` in byte 2; word 15 flags `line` in byte 1 and `subprogram` in byte 2, byte 0 of each is `1` | `t24_dbg_readers` against `t24_two_drivers`, one byte at `0x303`; `t24_dbg_line` against `t24_dbg_drv_only` | `t24_dbg_sub_only`, `t24_dbg_xlibs`, `t24_dbg_drivers`, the `header words` line of every dump | hierarchy |
| Without `-debug line`, header word 11 is `0` and regions 15 and 16 are empty | `t24_dbg_drv_only` against `t24_dbg_line` | `t24_dbg_xlibs` | hierarchy |
| DBG header word `i`, for 0 to 13, counts region `i + 4`: records for a record region, bytes up to the last NUL for a name pool, words for region 15, and 0 for an empty region | a sweep of the 17 words against the region lengths over 762 databases, every one fitting | the reader's `checkCounts`, 1119 of 1119, `//hdl/serv:sim`, `//hdl/potato:sim` | hierarchy |
| `-debug xlibs` alone brings the four packages, without the `resolved` scope of `all` | `t24_dbg_xlibs` against `t24_dbg_drv_only` | `t22_dbg_all` | hierarchy |
| `-debug drivers` over `typical` and `readers` over `typical` change nothing but the flag byte | `t24_dbg_drivers`, `t24_dbg_readers`, each against `t24_two_drivers` | `cmp` outside the noise mask | hierarchy |
| A signal attribute in a concurrent assignment is an implicit process `line__N`, two of them for `'delayed`, and no object for the implicit signal | `t24_att_delayed` against `t24_att_stable` | `t24_att_quiet`, `t24_att_transact`, `t24_ext_name` | hierarchy |
| A signal of a null range declares 0 bytes and is an object marked not logged with no record | `t24_null_range` against `t1_bit_one_edge` | `truth.json` | hierarchy, values |
| A resolved signal with two or more drivers records every transaction, changed or not; a single driver's unchanged assignment is not recorded | `t24_two_drivers`, `r` against `s`, read as once per driver until `t34_res_two_drv0` against `t24_two_drivers` dropped the time 0 assignment and the second record | `t34_res_3drv____`, `X` again at 80 ns; `t34_res_txn_zero`; `t34_res_same_fld` against `t34_res_two_fld_`; `records` in `truth.json` | values |
| A signal read through an external name records its change twice | `t24_ext_name` against `t24_config_spec` | `records` in `truth.json` | values |
| A `case` generate is one scope with the taken alternative's declarations; the alternative label is not in the name and the other alternatives leave nothing | `t24_case_gen` against `t8_gen_if` | `truth.json` | hierarchy |
| A configuration specification chooses the architecture unit like a direct instantiation, and the component leaves no trace | `t24_config_spec` against `t23_arch_b` | `truth.json` | hierarchy |
| A queue, a dynamic array, an associative array and a class leave no type entry, declaration or object, and take `0xf8` of handle space each | `t24_sv_queue` against `t11_sv_logic` | `t24_sv_dynarr`, `t24_sv_assoc`, `t24_sv_class` | hierarchy, types |
| Each `fork` branch is a `vprocess` scope `ForkedN_i`; a clocking block leaves nothing | `t24_sv_fork` against `t11_v_always`; `t24_sv_clocking` against `t11_sv_logic` | `truth.json` | hierarchy |
| DBG region 17 holds one 3 word entry per distinct value class among the objects, header word 13 counts them, and the region is padded to 8 bytes | `t25_sv_two_class` against `t25_sv_two_same`, 24 bytes and word 13 `2` for 16 and `1` | the reader's length check, 1119 of 1119 | hierarchy |
| The first word of a value class entry is the class code and the other two are 0; every VHDL object is class 0, and a Verilog object is classed by type and initializer form: 1 sized literal, 3 integer types and parameters, 4 unsized literal into a vector and `time`, 6 string parameter | `t25_sv_vec8_sz` against `t25_sv_vec8_int`, `1` for `4` | the tier 25 sweep, `t12_v_params` `[0 0 0] [3 0 0] [1 0 0]` in object order | hierarchy |
| A SystemVerilog package enters the file only when a declaration uses one of its types; a package holding only a parameter leaves no unit, scope or object, used or not | `t25_sv_pkg_prm` against `t25_sv_pkg_tdef` | `t25_sv_pkg_unusd`, `t13_sv_pkg`; the `absent` list of `truth.json` | hierarchy |
| A time literal initializer in a `.sv` file runs as an implicit process, the default and then the value, where `time s = 0` and `time s = 64'h0` record once | `t25_sv_time_lit` against `t11_sv_int`; `t27_sv_time_uns` against `t25_sv_time_lit` | `t27_sv_int_time`, `t27_sv_time_szd`, `t25_sv_time_noin` | values |
| The integral types are class 3 and `time` class 4 whatever the initializer; a packed type takes the class of its initializer: 1 for a sized or fill literal or an expression of them, 3 or 4 for an unsized literal by the target's signedness, 6 for a string literal, 0 for none | `t27_sv_sgn8_pos` against `t27_sv_v8_neg`; `t27_sv_int_uns` against `t27_sv_byte_uns` | the tier 26 and 27 sweeps, 56 cases | hierarchy |
| The `unsigned` qualifier of an integral type is not recorded | `t27_sv_int_uns` against `t11_sv_int`, the same file outside timestamps and lines | `t27_sv_byte_uns` reads back `-91` for 165; the `unsigned` field of `truth.json` | types |
| A function called only from an initializer leaves no unit, scope or object, and the initializer records once | `t26_sv_logic_fn` against `t11_sv_logic` | `t12_v_func`, where a call from a process has the scope | hierarchy |
| A `parameter string` leaves no object and 8 bytes of handle space; an untyped string parameter is an object of class 6 in `.sv` as in `.v` | `t26_sv_str_prm` against `t27_sv_str_untyp` | `t13_v_str_param` | hierarchy |
| An `event` leaves no declaration or object and takes `0x2c0` of handle space before the next object | `t26_sv_event` against `t11_sv_logic`, `s` at `0xa28` for `0x768` | the `absent` list of `truth.json` | hierarchy |
| A cast in an initializer leaves a hidden logged variable `xilinx_isim_temp_0_ln<line>castingOp` of the cast's type, first in the module, with one record at time 0, and the initializer runs as an implicit process | `t28_sv_int_cast` against `t27_sv_int_real` | `t28_sv_enum_cast`, `t28_sv_v8_szcast`; the `0x190` of handle space | hierarchy, values |
| A time literal, `$time` or a cast into a vector or a real runs as an implicit process; a real literal, an assignment pattern or an enum literal into an `int` based enum does not | `t28_sv_v64_time` against `t25_sv_v64_unsz` | `t28_sv_v64_stime`, `t28_sv_real_time`, `t28_sv_rtime_var`, `t28_sv_v8_real`, `t28_sv_pstr_pat`, `t28_sv_uarr_pat` | values |
| An untyped parameter with a time literal declares an unnamed 64 bit vector of class 4 whose record holds the `float64` of the value in the time unit, in 8 bytes; the record is written 16 bytes long and the second half is the next parameter's value, or `a8 07 00 00 00 00 00 00` after the last | `t28_sv_prm_time` against `t28_sv_prm_tmtyp`, where `parameter time` is a `time` holding `10` in two pairs; `t30_sv_ptm_two` against `t28_sv_prm_time`, where `T1 = 10ns` records `float64(10)` then `float64(20)` of `T2` | the `dewdb` decode of `T` against `truth.json` in `t30_sv_ptm_20ns`, `t30_sv_ptm_10ps` `0.01`, `t30_sv_ptm_1us`, `t30_sv_ptm_1s`, `t30_sv_ptm_frac` `10.5`, `t30_sv_ptm_ps_ts` `10000`, `t30_sv_ptm_late` | values, types |
| An untyped parameter with a time expression, `10ns * 2`, is a `real` entry of 32 bits and class 0 with an 8 byte record | `t30_sv_ptm_expr` against `t30_sv_ptm_20ns` | `t28_sv_prm_realu` | types, values |
| An untyped parameter from a sized literal wider than 32 bits declares that width, records two pairs and takes 16 bytes of handle space where a parameter of 32 bits or fewer takes 8; an untyped expression is evaluated at 32 bits, `1 << 40` records `0` | `t30_sv_prm_wide` against `t30_sv_prm_ubase`, handle space `0x92c` for `0x924`; `t30_sv_prm_shft` against `t28_sv_prm_expr` | `t12_v_param64`; `t30_sv_ptm_two`, two 8 byte parameters at `0x92c` | values, container |
| The value class of a packed type follows the form of the initializer down to the operands: a based literal without a size, an expression with a sized operand, a comparison and `$unsigned(5)` are 1, `$signed(5)` is 4, a real expression is 0, a string concatenation is 6; classes 2 and 5 do not appear | `t30_sv_v8_sgnu` against `t30_sv_v8_uns`, `[4 0 0]` for `[1 0 0]` | `t30_sv_v8_ubase`, `t30_sv_v8_sbase`, `t30_sv_v8_negsz`, `t30_sv_v8_mixed`, `t30_sv_v8_szexp`, `t30_sv_v8_cnd`, `t30_sv_v8_cmp`, `t30_sv_v8_1fill`, `t30_sv_v8_realx`, `t30_sv_v8_str2`, `t30_sv_v16_strc`, and the parameters `t30_sv_prm_ubase`, `t30_sv_prm_szsgn`, `t30_sv_prm_neg8`, `t30_sv_prm_cmp`, `t30_sv_prm_cnd`, `t30_sv_prm_shft`, `t30_sv_prm_realx`, `t30_sv_prm_strc` | hierarchy |
| `realtime` is a real entry named `realtime` with the Verilog variant, 32 bits for a variable and 16 for a parameter, and an untyped real parameter declares 32 bits | `t28_sv_rtime_var` against `t11_v_real`; `t28_sv_prm_realu` against `t26_sv_real_prm` | `t28_sv_rtime_noi`, `t28_sv_rtime_prm` | types |
| The value class follows the target of an initializer, not its expression, and a string literal into a typed parameter is class 1 | `t28_sv_v8_prmneg` against `t28_sv_prm_neg`, `s` 4 beside `K` 3; `t28_sv_prm_lstr` against `t28_sv_v16_str` | `t28_sv_bit8_neg`, `t28_sv_v8_pow`, `t28_sv_pstr_int` | hierarchy |
| A cast leaves its hidden variable only in a declaration initializer, one per cast numbered `temp_0`, `temp_1` through the module and from 0 again in a child; a cast in a process, a function, a continuous assignment, `always_comb` or a parameter leaves none, though each costs handle space, and two casts in one initializer fold and leave nothing | `t29_sv_cast_proc` against `t29_sv_incr`, the same objects and `0x1f0` more handle space; `t29_sv_cast_same` against `t28_sv_int_cast` | `t29_sv_cast_two`, `t29_sv_cast_sub`, `t29_sv_cast_fn` against `t29_sv_fn_noc`, `t29_sv_cast_asgn` against `t29_sv_asgn_noc`, `t29_sv_cast_alwc` against `t29_sv_alwc_noc`, `t29_sv_cast_prm` against `t28_sv_prm_expr` | hierarchy, values |
| A continuous assignment with a cast has two `NetRegassign` process scopes at its line, where one without has one | `t29_sv_cast_asgn` against `t29_sv_asgn_noc` | the `0x198` of handle space | hierarchy |
| A `for` loop with an `int` index or a `foreach` in a process declares the index in a block unit and scope `Block<line>_<n>` as a logged object; the `foreach` index records every value, the `for` index its first and its last | `t29_sv_for_int` against `t29_sv_incr`; `t29_sv_foreach` against `t29_sv_for_int` | `t29_sv_for_modi`, where a module level `integer` index records every value | hierarchy, values |
| A SystemVerilog function called from a process has its return variable as a logged object in the function scope, recording the default at time 0 and the value at the call | `t29_sv_fn_noc` against `t29_sv_incr` | `t29_sv_cast_fn`; `t12_v_func` in Verilog | hierarchy, values |
| An untyped parameter with an integer expression, `$clog2` or a negative literal is a 32 bit vector of class 3, a sized expression gives its width and class 1, and an enum parameter is class 3 over the typedef's alias | `t28_sv_prm_expr` against `t26_sv_logic_prm` | `t28_sv_prm_clog`, `t28_sv_prm_neg`, `t28_sv_prm_szexp`, `t28_sv_prm_enum`, `t28_sv_prm_int_u` | types, hierarchy |
| The word before an array entry's triples counts them | `t11_v_time`, one constrained triple under a `1` | 1119 of 1119 | types |
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
| A write in a child through a port lands at the port's offset plus the part's offset on the actual's handle, and a port bound to a record field carries the field's offset as a slice binding does | `t34_pmap_slice__` against `t9_port_slice2__`, 2 bytes at `+6` for offset 4 and `v(1 downto 0)`; `t34_pmap_field__` against `t9_port_slice`, offset 1 for `a => r.b` | `t34_port_fld_out`, `truth.json` | values, hierarchy |
| A write is per driver: two processes assigning adjacent fields in one delta give two records, in the order the simulator ran the processes | `t34_two_prc_adj_` against `t32_rec_two_adj_`, two 1 byte records against one of 2 | `t34_two_prc_rev_`, `q` first whichever field it writes; `t34_gen_elems___`, `g(2)`, `g(1)`, `g(0)`; `//hdl/uart:sim`, `head`, `count`, `empty` at 1615 ns | values |
| A Verilog partial write covers whole word pairs, and is chunked from its own address by the rule when 275 bytes or more | `t33_v_wsl_hi____` against `t12_v_vec4800x`, six records of 100 at the handle plus 600 for one of 8 | `t33_v_wsl_272___` and `t33_v_wsl_280___` either side of 275; `t33_v_wsl_mid___`, bits 16 to 2415 as 608 bytes; `t33_v_wsl_lo____`, `t33_sv_wsl_hi___`, `t33_v_mem_row___`, `t33_sv_st_wide__` | values |
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
| Unit kind `0x02` is a modport, named `interface.modport`, declaring each modport signal with its port mode; the instance gets a scope of it and the child's port is a second scope of it, both sharing the interface signal's handle | `t15_sv_iface_mp` against `t13_sv_iface` | | hierarchy |
| A vector in an interface is one more declaration of the interface unit and one more shared object in the port scope | `t15_sv_iface_vec` against `t13_sv_iface`, `tb.b.v` and `tb.dut.p.v` both at `0x828` | | hierarchy |
| `always_ff` and `always_comb` are `Always` scopes like `always` | `t13_sv_alwaysff` against `t11_v_always` | `t13_sv_iface`, `Always8_0` for the child's `always_comb` | hierarchy |
| An `inout` port has mode `0` and shares the net's handle; a `Z` driven onto the net is one record | `t13_v_inout` against `t12_v_port_wire` | | hierarchy |
| Over three levels the nets take handles in pre order and the variables follow them, and every port shares the handle of the net it is connected to | `t13_v_hier3_net` against `t12_v_port_wire` | `t12_v_proc_order`, two levels | hierarchy |
| A named block with a declaration is a unit of kind `0x05` holding the declaration, and its scope holds the object | `t13_v_blk_var` against `t11_v_always` | | hierarchy |
| A `reg` in an `if` generate is declared as `\g.r `, with one implicit `Initial` | `t13_v_gen_if_reg` against `t12_v_gen_reg` | | hierarchy |
| A `typedef` of an unpacked array carries every range in the alias, the unpacked one first, and the declaration none | `t13_sv_tdef_ua` against `t12_sv_unp2d` | `t12_sv_typedef` | types |
| A string parameter is an unnamed vector of 8 bits per character, the first character at the top | `t13_v_str_param` against `t12_v_params`, `"hello"` as 40 bits | | types |
| An unpacked array of `real` gives each element one pair, the last element lowest, and a write of the value already held produces no record | `t13_sv_real_arr` against `t11_v_real` | `t11_sv_ustruct`, the same order for fields | values |
| An unpacked array of packed structs is one contiguous value, element 0 at the top | `t13_sv_struct_ar` against `t11_sv_struct` | `t12_sv_unp2d` | values |
| An unpacked array of unpacked structs gives each element a slot of the struct's pairs, last element lowest, and a field write is one pair at slot plus field | `t35_sv_ust_arr__` against `t13_sv_struct_ar` | `t35_sv_ust_fld__`, `t35_sv_ust_tdef_` | values |
| A nested unpacked struct flattens into the outer struct's slots; an unpacked array field packs into its own pairs and an element write rewrites the field's pair | `t35_sv_ust_nest_` and `t35_sv_st_uarr__` against `t35_sv_ust_fld__` | `t11_v_mem4` | values |
| A typedef of an unpacked struct array is a second alias over the array entry carrying the range the declaration drops | `t35_sv_ust_tdef_` against `t35_sv_ust_arr__` | `t35_sv_pst_tdef_`, `t13_sv_tdef_ua` | types |
| Three writes to one variable at one time are three records, in write order | `t13_v_same_t` against `t11_v_bit_edge` | `t12_v_mem40_t0` | values |
| `log_wave -recursive /sig_pkg` logs a package signal; it records like a signal of `tb` in an arena of its own and the logged range count grows by one | `t13_pkg_log_all` against `t9_pkg_sig` | `t15_sv_pkg_log` against `t13_sv_pkg`, a SystemVerilog package parameter with one record at time 0 | values |
| A net holds one `X` record at time 0 per object on its handle, and one more when anything reads it; a second reader adds nothing and a port connection alone is not a reader | `t16_v_wire_rd1` against `t11_v_wire`, four records against three | `t16_v_wire_rd2`, `t16_v_wire_rdp`, `t16_v_wire_rdi`; the counts of `t12_v_port_wire` and `t13_v_hier3_net`; narrowed by tier 19, below | values |
| A Verilog write of the value already held produces no record, for a `reg` written from an `initial` block, a net with no reader whose assignment re-evaluates to its value, and a memory element | `t17_v_reg_same` against `t11_v_bit_edge`, two records against three | `t17_v_net_same`, `t17_v_mem_same`; narrowed by tier 36, below | values |
| A nonblocking write records as a blocking one, and a nonblocking swap gives each `reg` one record with the swapped value | `t17_v_reg_nb` against `t11_v_bit_edge` | `t17_v_nb_swap` | values |
| The `dims` word of an array entry counts the type's index dimensions, and one index type word per dimension follows it before the range count | `t18_arr_2dim` against `t2_array2d`, `2` and two index words against `1` and one | `t18_arr_3dim`, `3` and three | types |
| A multidimensional VHDL array is one value in row major order, the last index fastest, and the VCD writes it as one `wire` over all elements | `t18_arr_2dim` against `t1_vec8`, `03 02 03 02 03 03` for `(('1','0','1'),('0','1','1'))` | `t18_arr_3dim`, `TestVCD` on both | values |
| A VHDL signal read by a concurrent assignment gets no extra record at time 0, where a Verilog net read by a driver gets one `X` | `t18_sig_read` against `t1_bit_one_edge`, two records against two | `t16_v_wire_rd1`, the Verilog contrast | values |
| The extra record of a net at time 0 comes with two or more drivers and readers together, not with a reader alone; a `wire` with no driver holds one `Z` and no `X` | `t19_v_wand` against `t11_v_wire`, two `X` against one with two drivers and no reader | `t19_v_wire_3drv`, `t19_v_wand_rd`, `t19_v_2drv_port`, `t19_v_wire_nodrv`, `t19_v_nodrv_rd`, and every tier 16 count | values |
| The declaration kind word names the Verilog net kind: `0x03` `wire` and `uwire`, `0x04` `wand`, `0x05` `wor`, `0x06` `tri`, `0x07` `triand`, `0x08` `trior`, `0x09` `tri0`, `0x0a` `tri1`, `0x0c` `supply0`, `0x0d` `supply1`; the VCD `$var` line carries the same keyword | `t19_v_wand` against `t11_v_wire`, one word | the other eight net kind cases, `TestCorpus` against `declared`, `TestVCD` against the `$var` kind | hierarchy |
| The triples of an array entry are its own index ranges then the constraints of an unconstrained element; a constrained element adds none | `t19_arr_2d_vec` against `t18_arr_2dim`, three triples under `dims` `2` | `t19_arr_of_2dim`, one triple over a two dimensional element | types |
| A Verilog variable whose arena spills into a second page has no `X` record at time 0 | `t13_v_tr430` against `t13_v_tr420`, 430 records against 421 | `t13_v_tr2000`; `t13_v_tr430_2`, where `d` in its own arena keeps its `X`; `t14_v_spill_d` and `t14_v_spill_dfst`, where `d` in the spilling arena keeps it | values |
| A change due at the `std.env.stop` time is recorded, where one due at the `$finish` time is not | `t13_tr430` against `t13_v_tr430`, 431 records against 430 | `t11_v_always` | values |
| A `std_logic` signal gets an enumeration entry named `STD_LOGIC` with the nine `STD_ULOGIC` literals | `t8_port_inout` against `t8_port_in` | `truth.json` names the subtype | types |
| The class word of an enumeration entry follows the literals: `2` for `'0'` `'1'`, `3` for the nine `STD_ULOGIC` literals, `4` for any other set with a character literal, `5` for identifiers only | `t20_enum_bitlike` against `t2_bit`, `t20_enum_ul_like` against `t1_bit_one_edge` | `t20_enum_chars`, `t20_enum_mixed`, `t20_enum_one`, `t20_enum_two_id`; every enumeration entry in the corpus | types |
| The last word of a VHDL enumeration entry is the value's byte size, `1` up to 256 literals and `4` from 257 on, and the value is the literal index at that size | `t20_enum_300` against `t2_character`, `e299` as `2b 01 00 00` | `t20_enum_256`, `t20_enum_257`, `t20_enum_300_arr`, `t20_enum_300_rec` with the wide field at offset 4 | types, values |
| An unconstrained two dimensional array type carries one `(0, 0, -2)` triple per dimension and the declaration carries the ranges | `t20_arr_2d_uncon` against `t18_arr_2dim` | `t20_rec_2dim`, two triples on a record field of a constrained two dimensional type | types |
| A `real` parameter declares 16 bits for a `localparam` and for a value that does not fit a `float32` alike | `t20_v_realp_big` and `t20_v_realp_lp` against `t12_v_params` | `t12_v_params` | types |
| The 16 bits of a `real` parameter belong to the scalar form alone: an array of two declares 64, so 32 an element, and an untyped parameter, a `specparam` and a variable holding the same `float64` declare 32 | `t71_rlw_arr_prm_` against `t71_rlw_sreal_p_`, 64 bits against 16 | `t71_rlw_untyped_`, `t71_rlw_specprm_`, `t71_rlw_pkg_prm_`, `t71_rlw_kid_prm_`, `t71_rlw_rtime_p_`, `t71_rlw_vhdl_gen`, whose VHDL generic declares the 8 bytes of the float | types |
| `-debug line` and `-debug subprogram` are one mode, which writes the same file to the byte; byte 1 of header word 15 is a subprogram's scope and byte 2 is its own declarations, and a narrow mode without `wave` writes no database at all | `t72_dbg_line____` against `t72_dbg_subprog_`, 4499 bytes each with the same counts, and `t72_dbg_typical_`, whose `tb.f` has no objects | `t72_dbg_all_____`, `t72_dbg_wave____`, `t72_dbg_drivers_`, `t72_dbg_readers_` | hierarchy |
| The second pair of a protected type's method scopes hangs from the last process of the architecture that declares the variable, and not from the last scope the writer visits: a generate, an instance or a block after that process does not take it | `t73_prt_gen_last` against `t55_prot_pkg_prc`, the pair still under `tb.p` with two generate iterations after it | `t73_prt_inst_lst`, `t73_prt_blk_last`, and `t73_prt_arch_gen`, where the type in the architecture keeps the pair under `tb` | hierarchy |
| The two words of an access or file type entry are constants of the kind, `8 48` and `8 40`, and do not move with the designated or element type; every access variable declares 48 bytes and every file variable 0 | `t75_acc_rec_____` and `t75_acc_acc_____` against `t23_access`, a record and another access behind the pointer | `t75_acc_arr40___`, `t75_fil_rec_____`, `t75_fil_arr_____` | types |
| A SystemVerilog subprogram local is storage class 3 whatever its type, where a VHDL local of an array, string or record type is 4; a file inside a subprogram, parameter or local, leaves no object at all | `t76_stc_sv_arr__` against `t50_sub_func_prm`, an `int a[2]` local at 3 where a VHDL vector local is 4 | `t76_stc_sv_ref__`, `t76_stc_sv_out__`, `t76_stc_sv_str__`, `t76_stc_file_prm`, `t76_stc_file_loc`, `t76_stc_acc_prm_` | hierarchy |
| A recursive wildcard matches a package scope and none of its objects, so no form of `log_wave` with a pattern logs a package signal, whatever the pattern is anchored to; naming the package, as a scope or through `get_objects /pkg/*`, is what logs it | `t74_lgw_root____` and `t74_lgw_cur_root` against `t74_lgw_pkg_name`, `/*` and the root made current logging nothing more than `*` | `t74_lgw_objects_`, `t74_lgw_pkg_obj_`, and the `OBJECTS:` and `SCOPES:` lines the six scripts print | values |
| A page written out before the end of the run keeps one record per key and time, the last; the last page of an arena keeps every delta | `t14_v_spill_dd` against `t14_v_page_dd`, one record at 5 ns against two at 190 ns | `t14_v_spill_dd2`, two records at 428 ns in the second page; the missing `X` of `t13_v_tr430` is the same loss | values |
| At every time the VCD lists a value, the last database value at that time spells it, for every VCD variable of every case; the changes of value in the VCD are the changes of value in the database, and the VCD keeps a few of the database's writes of the value held and drops the rest | `TestVCD`, `bazel test //pkg/wdb:wdb_test --test_filter=TestVCD` | 1119 of 1119, `//hdl/counter:sim`, `//hdl/uart:sim` and `//hdl/serv:sim`; one real field of `t11_sv_struct_r` and every untyped time parameter excepted, where the VCD writes the `float64` bytes as a vector; the times were equal in every case through tier 35 and differ from tier 36 on | vcd |
| Two VCD variables with one identifier code are two objects with one handle and one offset | `TestVCD` on `t12_v_port_wire` and `t13_sv_iface` | 1119 of 1119 | vcd |
| The VCD leaves out every VHDL generic, constant and non bit type, every signal outside `tb`, and every Verilog unpacked array not named by a typedef, and nothing else | `TestVCD`, the `vcdOmitted` rule | 1119 of 1119 | vcd |
| An unpacked struct is written to the VCD as 32 bit slots per field; the database holds the fields at their own widths | `TestVCD` on `t11_sv_ustruct` against the dump | `t11_sv_struct3`, `t11_sv_struct40` | vcd |
| A Verilog port bound to a slice of a net shares the net's handle, and the object offset word counts bits from bit 0 of the net; the port's value is the bits `[offset, offset + width)` of the net's word pairs, and a slice beyond pair 0 lies in the pair its bits fall in; a port bound to a slice of a `reg` is a net of its own with its own handle and records | `//hdl/serv:sim` against `t9_port_slice`, `i_wb_adr` bound to `wb_mem_adr[12:2]` at offset 2 and `i_shamt` bound to `o_dbus_dat[26:24]` at offset 24 | `t37_v_port_slc__`, `t37_v_port_bit__`, `t37_v_port_pair1`, `t37_v_port_span_`; `t37_v_port_reg__` for the `reg` case | hierarchy, values |
| A nonblocking assignment in `always @(posedge clk)` records every target of the block on its first run, and on every later run preceded by an event on an operand of the block, whether the value changes or not; a blocking clocked assignment, a combinational `always`, an `initial` block and a `#` delayed block record changes only | `//hdl/serv:sim` against `t17_v_reg_same`, 2965811 repeated records, then `t36_v_nb_clk_lit` against `t17_v_reg_same`, three records against two | `t36_v_nb_clk_tog`, `t36_v_nb_clk_150`, `t36_v_nb_clk_two`, `t36_v_nb_clk_x__`, `t36_v_nb_clk_net`, `t36_v_hier_p_nba`, and the control cases of tier 36 that record nothing: `t36_v_bl_clk_lit`, `t36_v_comb_and__`, `t36_v_nb_evt_lit`, `t36_v_nb_dly_lit`, `t36_v_nb_ini_evt`, `t36_v_nb_two_lit`, `t36_v_init_expr_` | values |
| A net with one driver and no reader records changes only; a net with two or more drivers and readers together records every evaluation of a driver, changed or not, and a child's net driving an output port through `assign` counts the port as a reader | `t36_v_net_copy__` against `t36_v_net_mux_w_`, seven records against three | `t36_v_net_rd_not`, `t36_v_net_rd_cat`, `t36_v_net_rd_alw`, `t36_v_net_2drv__`, `t36_v_hier_int__`, `t36_v_hier_i_sel`, `t36_v_hier_i_ab_`, `t36_v_hier_i_and`, `t36_v_hier_i_or_`, `t36_v_hier_i_noc`, `t36_v_hier_regs_`; `rreg` of `//hdl/serv:sim`, 54413 records for 40935 changes | values |
| A Verilog top elaborated with `-generic_top name=value` is named `tb(name="value")` in the scope, the unit and the VCD `$scope` line | `//hdl/serv:sim` against `t22_gen_top` | `TestVCD` on `//hdl/serv:sim` | hierarchy |
| `$readmemh` and `$readmemb` write one 8 byte element record per line of the file at time 0, in file order from the lowest address, after the whole `X` write; `@addr` and the range arguments move the start, and a second load of the same file records nothing | `//hdl/serv:sim` against `t12_v_mem40_t0`, 25 records for a 24 line file, then `t38_v_rmh_4w____` against `t38_v_mem4w32___`, the same six records | `t38_v_rmb_4w____`, `t38_v_rmh_2of4__`, `t38_v_rmh_at2___`, `t38_v_rmh_rng___`, `t38_v_rmh_desc__`, `t38_v_rmh_twice_`; the 24 SERV RAM words against `hello_uart.hex` | values |
| A memory over several arenas holds its whole write in the chunks of the chunk rule, split at every arena boundary, and each element in the pair its position gives, `m[0]` at the top | `t38_v_mem512____` against `t38_v_rmh_4w____`, 4096 bytes in 28 chunks over three arenas, `m[511]` at the handle | the SERV RAM, 16384 bytes in 110 chunks over nine arenas | values |
| The rest of a chunked value is chunked again by the same rule when it is 276 bytes or more; a rest of exactly 275 bytes stays one chunk, where a value of 275 bytes is two | `//hdl/potato:sim`, the 32768 byte instruction memory with its rest of 356 in four chunks of 89, then `t39_vec20120____` against `t39_vec20121____`, rests of 275 and 276 | `t39_vec30022____`, `t39_vec30023____`, `t39_vec20125____`, `t39_vec20561____`, `t39_vec22199____`, `t39_vec22347____`, `t39_vec22348____`, `t39_vec22349____`, `t39_vec22647____`, `t39_vec22791____`, `t39_vec32768____`; `t39_mem4096_____`, a 30720 byte partial write; the reader enforces the addresses in 1119 of 1119 | values |
| A string generic set by `--generic_top` is declared with the length and range `(1 to n)` of the value given, not of the default | `t40_gen_str_top_` against `t40_gen_uncons__`, `ks=hello` declared as 5 bytes `(1 to 5)` over 3 bytes `(1 to 3)` | `IMEM_FILENAME` and `DMEM_FILENAME` of `//hdl/potato:sim`, 19 and 44 bytes for defaults of 17 | hierarchy |
| An unconstrained array generic with a literal default is declared with the ascending range `(0 to n - 1)` of the literal; a constrained one keeps its declared range | `t40_gen_uncons__` against `t40_gen_cons____`, `x"A"` declared `(0 to 3)` against `(3 downto 0)`, the same records | `RESET_ADDRESS` of `//hdl/potato:sim`, `(0 to 31)` at the unconstrained top generic and `(31 downto 0)` at the constrained ones below | hierarchy |
| A VHDL index bound is a signed 32 bit word in a type entry's triple, and a 64 bit pair with the sign in the high word in a declaration range; the value is the elements in index order whatever the bounds | `t41_neg_vec_____` against `t41_uvec________`, `vec_t(3 downto -4)` against `vec_t(7 downto 0)` | `t41_neg_asc_____`, `t41_neg_arr_type`, `t41_neg_int_sub_`, `t41_sfixed______`, `t41_float32_____`; the `range` field of `truth.json`, now checked in every case that gives one | types |
| A signal declared with a constrained subtype of an unconstrained array type gets an entry named after the subtype, holding the subtype's range as its triple; the base type is not in the table | `t41_arr_subtype_` against `t41_neg_vec_____`, `byte_t` with `(3 downto -4)` against `vec_t` with the unconstrained triple | `float32` of `t41_float32_____`, `(8 downto -23)` | types |
| `sfixed`, `ufixed` and `float32` of the IEEE fixed and float packages are arrays of `STD_ULOGIC` indexed by `INTEGER` with nothing marking them as numbers, named in lower case as their sources spell them | `t41_sfixed______` against `t41_neg_vec_____`, the same bounds over a user type | `t41_ufixed______`, `t41_float32_____` | types |
| A record field declared without bounds carries the unconstrained triple `(0, 0, -2)`, and the bounds live in the declaration record alone; the declaration lists the bounds of every array dimension of a record in field order, so a reader decodes a record from its declaration and not from the field triples | `t42_rec_uncons__` against `t2_record2______`, `bravo : std_ulogic_vector` constrained at the signal against `(7 downto 0)` in the field; the value decoded as one element | `t42_rec_two_unc_` against `t42_rec_mix_unc_`, the same declaration list with the bounds in the field or at the signal; `t42_rec_two_cons`, `t42_rec_unc_nest`, `t42_rec_unc_2dim`, `t42_rec_unc_arr_`, each against `truth.json` | types |
| A constrained subtype of a record renames the entry and leaves the field triple unconstrained, unlike an array subtype whose entry holds the bounds | `t42_rec_subtype_` against `t42_rec_uncons__`, an entry `b8_t` with `(0, 0, -2)` on the field | `t41_arr_subtype_` for the array side | types |
| An array type whose element is unconstrained writes its own index as `(0, 0, -2)` even when the index is written with bounds; the declaration carries the index and the element bounds | `t42_arr_unc_elem` against `t2_array2d______`, `array (0 to 1) of std_ulogic_vector` against `array (0 to 3) of std_ulogic_vector(7 downto 0)` | `t42_arr_unc_both`, which differs only in the index word; `t42_rec_unc_arr_` | types |
| A type generic is neither a declaration nor an object; the actual type enters the table under the formal's name, `integer "data_t"`, and the actual's own name is absent | `t42_gen_type____` against `t4_gen_explicit_`, `data_t => integer` against a generic of `INTEGER` | `t42_gen_type_enu`, `enum "data_t"` with the nine `STD_ULOGIC` literals | hierarchy |
| A generic package instance is a package scope named after the uninstantiated package, at its file and line, with no trace of the instance name or the generic; two instances give two scopes of one name | `t42_gen_pkg_____` against `t42_pkg_subtype_`, the same subtype in a plain package; the files differ in the scope name and the paths only | `t42_gen_pkg_two_`, two scopes `gp` and two entries `word_t`; `t42_gen_pkg_cons`, the constant of `n` an unlogged object under `gp` | hierarchy |
| A package with only a type or subtype in it gets a scope with no object | `t42_pkg_subtype_` against `t42_gen_pkg_____` | `t34_pmap_field__`, whose `trio_pkg` holds one record type; corrects the "no scope" claim recorded against `t2_record`, which has no package | hierarchy |
| An unconstrained port is declared with the bounds and size of its actual, and two instances whose actuals differ in width repeat the unit and its declarations the way different generics do | `t43_port_unc_two` against `t43_port_unc_sam`, eight and four bit actuals against two eight bit ones: 7 units and 6 declarations against 5 and 4 | `t43_port_uncons_` against `t8_port_vec8____`; `t43_port_unc_asc`, `(0 to 7)`; `t43_port_unc_out`; `t43_port_unc_rec`, a record port with `(7 downto 0)` on the declaration and the field unconstrained in the type | hierarchy |
| Record times and page bounds are 64 bit: a change at `5 sec` records 5000000000000 | `t44_time_5ms____` against `t1_bit_one_edge_` | `t44_time_5s_____`, `t44_time_late___`, `t44_v_time_5ms__`, `t65_tim_1s______`, `t65_tim_cross___`, `t65_tim_ns_5s___` | values |
| A VHDL `string` signal is an array of `character` with the written bounds on the declaration, one byte per character | `t44_str_sig_____` against `t2_character____` | `t44_str_sig_3to7`, `t44_str_var_____` | types |
| A signal's records start at the time `log_wave` names it, with one record holding the value then held, and the VCD backdates that value to `#0` | `t45_log_late____` against `t45_log_base____` | `t45_log_dut_late`; `t45_log_twice___`, one record of the value held at the second `log_wave`; `t45_run_steps___`, no trace of a split run | values |
| The root scope holds one child per `--top`, in option order, and the default script logs the first top only | `t45_two_tops____` against `t45_log_base____` | `t45_two_tops_all`, both tops logged by name | hierarchy |
| Every count and index of the debug section and the arena table is a whole 32 bit word: 140004 scopes, 140000 objects and 18147 slots read by the same rules as 3 and 2 | `t46_gen_70000___` against `t46_sig_1000____` | `t46_v_gen_70000_`, 70000 objects in one scope; `t46_deep_100____`, a recursive entity 100 levels deep | hierarchy, container |
| A VHDL signal's handle stride is `0xb8` plus its storage rounded to 8 plus `0x30` per driver, and every generic, constant, variable and loop index of the design takes a handle after the last signal | `t46_sig_1000____` against `t45_log_base____`, 998 undriven signals `0xc0` apart after two driven ones `0xf0` on | `t46_gen_70000___` and `t46_deep_100____`, indexes and generics `0x640` past the last signal; `t46_drv_2_next__`, `0x140`, and `t46_drv_3_next__`, `0x178`, for two and three drivers; 109 dumped cases with both kinds of object | hierarchy |
| A signal nothing drives costs `0xf8` of handle space and a driven one `0x148`, exactly, over a thousand signals | `t46_sig_1000____` against `t0_bit_const____` and `t1_bit_one_edge_` | `t46_drv_2_next__` and `t46_drv_3_next__`, the handle space of `t24_two_drivers_` and `t34_res_3drv____` | container |
| A use clause costs handle space by the package, `0x604` for `std_logic_1164`, `0x1f8` for `numeric_std`, `0x400` for `math_real`, and adds the package and its body to the file table; the type of the signal costs nothing; a package of the design costs `0x80` plus the rounded storage of its constants, and its types and subprograms are free | `t47_use_numstd__` against `t1_bit_one_edge_`, the clause alone reproducing the `0x1f8` of the nine tier 2 cases that use it | `t47_use_1164_bit`, `t47_use_none____`, `t47_use_lib_only`, `t47_use_one_name`, `t47_use_textio__`, `t47_use_numbit__`, `t47_use_mathrl__`; `t47_use_pkg_emp_`, `t47_use_pkg_typ_`, `t47_use_pkg_4arr`, `t47_use_pkg_fn2_`, `t47_use_pkg_pr2_`, `t47_use_pkg_two_`, `t47_use_pkg_nul_` | container, hierarchy |
| A declaration range record is the two bounds as 64 bit pairs, the direction, `1` for `to` and `-1` for `downto`, and the distance between the bounds plus one, which is 2 for the null range `(0 downto 1)` | `t2_array2d`, `(0 to 3)` and `(7 downto 0)` in one declaration | `t41_neg_arr_type`, `(-2 to 1)` with `-1` high words; `t12_v_neg_range`, `[-4:3]`; `t24_null_range`, span 2 beside size 0 | hierarchy |
| A Verilog `reg` pays no stride for its writer, and a `wire` pays `0xc0`, `0xe8`, `0xf0` and `0xf0` for zero to one, two, three and four `assign` statements | `t46_v_gen_70000_`, 70000 registers `0xc0` apart, two of them written | `t46_v_wire_4asg_` against `t19_v_wire_3drv_`; `t19_v_wire_nodrv`, `t11_v_wire______`, `t19_v_2drv_port_` | hierarchy |
| The word at `40` of an instance record is the position of a Verilog port in its module's port list, from 0, by the port list and not by the connection or by the declaration order of a non ANSI header; every VHDL object holds 0 | `t48_v_port_nansi` against `t48_v_port_pos4_`, objects in declaration order `d`, `c`, `b`, `a` holding 3, 2, 1, 0 | `t48_v_port_rev__`, `t48_v_port_posit`, `t48_v_port_open_`, `t36_v_hier_and__`, `t13_v_hier3_net_`, `t11_v_port______`, checked against `position` in `truth.json`; `//hdl/serv:sim`, 0 to 4 on `servant_sim`; `//hdl/potato:sim`, 0 on 557 VHDL objects | hierarchy |
| The net of an input port connected to a `reg` takes its handle in the order of the connections written in the instantiation | `t48_v_port_rev__` against `t48_v_port_pos4_`, `0x8e8` and `0x9a8` swapped between `a` and `b` | `t48_v_port_posit` | hierarchy |
| The word at `28` of an instance record is a storage class: 0 for a signal, 1 for a port on a language boundary on either side, 2 for a generic, constant, variable or loop index, 3 for a scalar or access subprogram local, 4 for an array, string or record subprogram local, 6 for a signal parameter, whatever the class and mode | `t49_sub_rec_loc_` and `t49_sub_int_arr_` against `t23_sub_sizes___`, 4 on a record and an integer array local as on the vector local | `t49_sub_var_prm_`, `t49_sub_vec_prm_`, `t49_sub_sig_in__`, `t49_sub_sig_vec_`, `t49_mix_2port___`, `t49_mix_deep____`, the eight `t50_sub_` cases; `t21_mix_v_in_vh_`, `t21_mix_vh_in_v_`, `t22_dbg_sub_proc`, `t23_sub_sizes___`, `t23_sub_sig_prm_`, checked against `storage` in `truth.json` | hierarchy |
| A subprogram parameter of an unconstrained type has no declaration and no object, and still takes 24 bytes of the frame | `t49_sub_str_prm_` against `t23_sub_sig_prm_`, two declarations for three parameters and the same handle space; the signal parameter after it on `0xe8` for `0xd0` | `t50_sub_ivec_prm`, an `integer_vector` | hierarchy |
| A signal parameter's 64 byte frame slot starts on a multiple of 8, and a vector parameter takes the 24 byte descriptor a vector local does | `t50_sub_in_var__` against `t23_sub_sig_prm_`, `v` on `0xd0` and `q` on `0xd8`; `t50_sub_func_prm` against `t49_sub_vec_prm_`, `a` on `0x40` and `r` on `0x58` | `t50_sub_acc_loc_`, `t50_sub_str_loc_`, a local on `0x110` after the signal parameter | hierarchy |
| The declarations of a unit list the signals in source order, then the generics, constants and variables in source order; a subprogram lists its signal parameters first the same way | `t50_ord_const1st` against `t5_tr1000_______`, the signal before the constant declared above it; `t50_ord_proc_con` against `t6_var_int______`, the constant before the variable declared below it | `t50_ord_two_sig_`, `t4_gen_default__`, `t22_dbg_sub_proc`, `t49_sub_vec_prm_`, `t50_sub_in_var__` | hierarchy |
| An `automatic` SystemVerilog task or function keeps its unit and scope and lists no argument or local; a `static` local of one is listed; a static task's arguments hold 0 in the word at `40` whatever their place | `t51_sv_task_auto` against `t51_sv_task_stat`, the task unit with no declarations and the handle space `0xbb4` for `0xc14` | `t51_sv_task_ref_`, `t51_sv_func_auto`, `t51_sv_task_stvr`, `t51_sv_task_out_`, `t51_sv_task_inou` | hierarchy |
| A `for` loop index in a subprogram is not in the file | `t51_sub_loop_idx` against `t23_sub_sig_prm_`, the signal parameter and the variable alone | | hierarchy |
| A `file` parameter is absent and takes 8 bytes of the frame; a file object of an architecture is a variable of size 0 with the file type, one handle and no record | `t51_sub_file_prm` against `t23_sub_sig_prm_`, `q` on `0xd8` | | hierarchy |
| A procedure of a package is a scope under the package scope, with the frame offsets of an architecture procedure | `t51_sub_pkg_proc` against `t23_sub_sig_prm_` | | hierarchy |
| In the second handle region an object's stride is its value size, the next object on a multiple of its own size, and an array or string takes 16 bytes more than its elements; a constant and a generic stride as a variable | `t52_var_real____`, `t52_var_vec8____`, `t52_var_arr4____` against `t52_var_int_____`, 8, `0x18` and `0x20` for 4 | the other twelve `t52_var_`, `t52_con_` and `t52_gen_` cases; `t9_gen_types____` | hierarchy |
| An instance and a generate iteration cost the same in the second region: `0x30` bare, `0x90` more for a process, `0x38` more for a signal and `0x20` more for its driver, the signals before the data objects and the process after | `t52_inst2_proc__`, `t52_inst2_sig___`, `t52_inst2_sigprc` against `t52_inst2_empty_`, `k` strides `0xc0`, `0x68` and `0x118` for `0x30`; the three `t52_gi2_` cases against each other, the same on `i` | `t7_gen_for______`, `t46_gen_70000___`, `t46_deep_100____`, the `0xf8` and `0x50` of tier 46 | hierarchy |
| A for generate with an empty body elaborates to the plain label scope alone, with no iteration scope and no index | `t52_gi2_empty___` against `t52_gi2_proc____` | | hierarchy |
| A scope's block in the second region is `0x28`, plus its data objects rounded up to 8, `0x38` per signal or port, connected or open, `0x90` per process, a concurrent assignment being one, and `0x20` per driver, the signals first and the data objects next | `t53_inst2_2gen__`, `t53_inst2_const_`, `t53_inst2_2proc_`, `t53_inst2_var___`, `t53_inst2_2sig__`, `t53_inst2_conc__`, `t53_inst2_2drv__`, `t53_inst2_port__`, `t53_inst2_portop` against `t52_inst2_empty_` and the tier 52 bodies, `k` strides `0x30`, `0x30`, `0x150`, `0xc0`, `0xa0`, `0x118`, `0x1c8`, `0x68`, `0x68` | `t6_proc2________`, `t46_deep_100____` at `0x140` | hierarchy |
| The generate and block scopes of a unit are laid out depth first after the unit's own block, and the instances it holds follow, each with its own unit's scopes and instances the same way | `t53_ifgen_inst__` and `t53_blk_inst____` against `t52_inst2_empty_`, both wrappers before both children; `t53_inst2_nest__`, `d0`, `d0.e`, `d1`, `d1.e` | `t7_gen_for______`, `t8_gen_nest_____`, the first `dut` at `0x1300` | hierarchy |
| The signals' blocks run from `0x738`, each handle `0x30` into its block, the packages' blocks follow, and the root unit's block follows those | `t54_noenv_sig___` against `t54_none_noenv__`, the variable at `0x860` for `0x738`; `t53_inst1_empty_`, `t53_inst3_empty_` against `t52_inst2_empty_`, the first `k` at `0xeb8` with one, two or three children | `t54_nosig_var___`, `t46_deep_100____`, `t46_gen_70000___` | hierarchy |
| The packages' blocks cost `0xd8` for `textio` with `env`, `0x540` for `textio` with `std_logic_1164`, `0x40` for `env` on top, `0xf8` for `numeric_std`, `0x308` for `math_real` and `0x28` plus its constants for a package of the design, and the rest of each package's handle space lies past the second region | `t54_none_nosig__` against `t54_none_noenv__`; `t54_1164_noenv__`, `t54_lib_numstd_v`, `t54_lib_mathrl_v`, `t54_pkg_con_var_` against `t52_var_int_____` and each other | `t54_lib_none_var`, `t54_pkg_2con_var`, the tier 47 handle spaces | hierarchy |
| A package of the design is in the file when a use clause names it or a name refers into it, and absent otherwise | `t54_pkg_unused__` against `t54_pkg_con_var_` and `t54_pkg_use_var_` | | hierarchy |
| A bench without `std.env.stop` ends at its last event, and the file lists neither `env` nor `textio` | `t54_none_noenv__` against `t54_none_nosig__`, four files for six and `0x1c0` less handle space | `t54_1164_noenv__`, `t54_noenv_sig___` | hierarchy |
| A concurrent assignment costs `0x50` in the first region where a process driver costs `0x30` | `t53_inst2_conc__` against `t52_inst2_sigprc`, `c` `0x110` apart for `0xf0` | | hierarchy |
| A process constant has one record at time zero holding its value, where a process variable has none | `t50_ord_proc_con` against `t6_var_int______`, `c = 3` at 0 and nothing for `v` | checked against `value` in `truth.json` | hierarchy |
| A VHDL port under a Verilog net that carries a VHDL value from elaboration holds the value and no `U`; a VHDL signal driven by a Verilog `assign` holds `U`, the `X` of the assign, then the value | `t49_mix_deep____` against `t49_mix_2port___` and `t21_mix_vh_in_v_` | | values |
| A constant declared in a subprogram is a local of kind `0x14` and class 3 or 4 as a variable is, and takes its size plus 16 bytes of the frame where a scalar variable takes its size | `t55_sub_con_loc_` against `t55_sub_loop____`, `v` on `0x58` for `0x44` | `t55_sub_con_nori`, `t55_sub_2con____`, `t55_sub_con_real`, `t55_sub_var_init`, `t55_sub_con_arr_` | hierarchy |
| An alias, a file object and a protected type variable declared in a subprogram are absent from the file and take 0, 40, and 12 then 16 bytes of the frame | `t55_sub_alias___`, `t55_sub_file_loc`, `t55_sub_prot_loc` against `t51_sub_loop_idx`, the local on `0x110`, `0x138` and `0x11c` | `t55_sub_prot_2__`, `t55_sub_prot_3__` | hierarchy |
| A function inside a function gets two scopes on one unit, the frame numbered from `0x40` in each | `t55_sub_nested__` against `t50_sub_func_prm` | `t23_sub_in_proc` | hierarchy |
| The methods of a protected type with a variable of it are two pairs of subprogram scopes, both under the architecture when it declares the type, and under the package and the architecture's last process when a package declares it; the type without a variable, the body's variable and the local are absent | `t55_sub_prot_loc` against `t51_sub_loop_idx`; `t55_prot_pkg____` against `t55_prot_shared_`; `t55_prot_pkg_2p_` against `t55_prot_pkg_2pl` | `t55_sub_prot_typ`, `t55_prot_arch_pr`, `t55_prot_arch_2p`, `t55_prot_pkg_prc`, `t55_prot_pkg_sv_` | hierarchy |
| A subprogram's static composite values take handle space by their bytes: a composite local's literal or default initial value, and every aggregate or string literal in the body or the call; a type, a scalar local, and a value computed at the call take none | `t56_typ_arr_loc_` against `t56_typ_arr_unus`, `0x10` for four integers; `t56_typ_arr_lit_` against `t56_typ_arr_noin`, `0x10` more for an aggregate in the body; `t56_sub_arr_dyni`, `t56_typ_rec_prm_`, nothing for a computed initial value | `t56_typ_arr8_loc`, `t56_typ_vec4_loc`, `t56_typ_vec4_2lc`, `t56_typ_vec_noin`, `t56_typ_rec_loc_`, `t56_sub_rec_3int`, `t50_sub_func_prm`, `t50_sub_str_loc_` | hierarchy |
| A record's static value costs its declared size plus 4, and a record declares a multiple of 8 bytes, 8 for one integer | `t56_sub_rec_1int`, `t56_sub_rec_2int` against `t56_sub_rec_3int`, `t56_sub_rec_4int`, `0xc` then `0x14` | `t56_sub_rec_2rl_`, `t56_typ_rec_arr_` | hierarchy |
| `log_wave` logs a signal, a constant, a loop index, a generate iteration's signal and index, and the data objects directly in a named scope; it never logs a process or shared variable, and a bit, a slice or a field of a signal names the whole signal or nothing | `t57_log_var_____` against `t57_log_con_____`: the variable draws "No matching HDL object" and nothing is logged, the constant gets its record; `t57_log_slice___` against `t57_log_bit_____`: the slice logs the whole vector, the bit nothing | `t57_log_var_all_` under `-debug all`, `t57_log_shv_____`, `t57_log_loop____`, `t57_log_rec_fld_`, `t57_log_gen_sig_`, `t57_log_gen_idx_`, `t57_log_gen_it__`, `t57_log_gen_____`, `t57_log_proc____`, `t57_log_top_____` | values |
| A database whose script logs nothing keeps every scope, unit, declaration and object, marks each object not logged, and has four arena slots of `0`, a marker count and offset of `0`, no page, and the directory right after the debug section | `t57_log_none____` against `t57_log_all_____`: 4007 bytes against 8048, the same handle space `0x18dc` | `t57_log_var_____`, `t57_log_bit_____`, `t57_log_rec_fld_`, `t57_log_gen_____`, `t57_log_shv_____`, each 4007 bytes | container |
| In SystemVerilog `log_wave` logs every variable it can name, the module's `int` and `real`, a named block's variable, and a static task's argument and local, where a VHDL variable is refused; a bit, a memory element and a struct field log nothing and a slice the whole object; a generate block's wire is reached only through `get_objects -regexp` | `t58_sv_log_int__`, `t58_sv_log_blkv_` against `t57_log_var_____`; `t58_sv_log_gen_w` against `t58_sv_log_gen__` | `t58_sv_log_real_`, `t58_sv_log_tsk_a`, `t58_sv_log_tsk_l`, `t58_sv_log_tsk__`, `t58_sv_log_blk__`, `t58_sv_log_bit__`, `t58_sv_log_slc__`, `t58_sv_log_mem_e`, `t58_sv_log_st_fl`, `t58_sv_log_top__`, `t58_sv_log_none_` | values |
| A forced or deposited value is an ordinary record with no mark: a Tcl force records the value it imposes when applied, held or not, and then the value held at every driver transaction of another value; `remove_forces` and a second `add_force` record the new value twice; a deposit records once and does not hold | `t59_frc_s_const_` against `t59_frc_none____`, the dumps differ in the records of `s` and the noise words only; `t59_frc_release_`, `t59_frc_twice___`, `t59_frc_deposit_` | `t59_frc_s_cancel`, `t59_frc_s_pat___`, `t59_frc_v_const_`, `t59_frc_v_bit___`, `t59_frc_mid_____`, `t59_frc_mid_same`, `t59_frc_rel_same`, `t59_frc_dep_mid_`, `t59_frc_dep_same`, `t59_frc_sv_tcl__`, `records` in each `truth.json` | values |
| A SystemVerilog `force` statement records the value it imposes, held or not, a write to the forced variable and its `release` record nothing, and the statement costs `0x48` of handle space and no object | `t59_frc_sv_force`, `t59_frc_sv_frc_0` against `t59_frc_sv_none_`, `0x9fc` against `0x9b4` | `t59_frc_sv_long_`, `t59_frc_sv_norel`, `t59_frc_sv_relon` | values |
| `xelab -debug all` adds five type kinds to a SystemVerilog file: `0x18` string with no body, `0x14` queue and `0x13` dynamic array with an element type and the word `1`, `0x15` associative array with an element type, a word `2` for a `string` key or `3` for an `int` key, and the key type, and `0x17` class with a parent index or `-1`, an id, and record style fields each closed by a `0`; a class entry precedes the entries of its field types and its parent precedes it | `t60_dbg_str_____`, `t60_dbg_queue___`, `t60_dbg_dynarr__`, `t60_dbg_assoc___`, `t60_dbg_class___` against `t60_dbg_none____` and `t60_dbg_int_____` | `t60_dbg_assoc_i_`, `t60_dbg_class_2_`, `t60_dbg_class_d_`, `t60_dbg_class_2h`, `t60_dbg_class_n_` | types |
| The word after an array entry's last range triple, the id word of a class and the word after a container's element are one numbering under `-debug all`, from `0` over the registered types: variables in declaration order, a container after its element, a class before its fields with the fields last declared first, a parent or a field's class right after the class; an array numbered by its element, so `int`, `byte` and `longint` share one number; a type numbered once; `string` and `real` unnumbered; an associative array taking two numbers with a `string` key and three with an `int` key | `t61_num_a_then_q` against `t61_num_q_then_a`, `t60_dbg_class_2_` against `t61_num_cls_rev_` | the 31 files with a number: `t60_dbg_queue___`, `t60_dbg_dynarr__`, `t60_dbg_assoc___`, `t60_dbg_assoc_i_`, `t60_dbg_class___`, `t60_dbg_class_2_`, `t60_dbg_class_d_`, `t60_dbg_class_2h`, `t60_dbg_class_n_`, `t60_dbg_q_log___` and the 21 `t61_num_*` cases | types |
| Under `-debug all` a string or class handle is declared as 32 bits of value class 0, gets a handle and is logged with one 8 byte zero record per `log_wave` that names it at time 0 and none for any write or construction; a queue, dynamic array or associative array is declared as 32 bits of value class 3 with the range `(0 to 0)`, gets a handle, and is never logged, `log_wave` on it warning as under typical | `t60_dbg_str_____`, `t60_dbg_queue___` against `t60_dbg_none____` | `t60_dbg_dynarr__`, `t60_dbg_assoc___`, `t60_dbg_assoc_i_`, `t60_dbg_class___`, `t60_dbg_class_2_`, `t60_dbg_class_d_`, `t60_dbg_class_2h`, `t60_dbg_class_n_`, `t60_dbg_str_log_`, `t60_dbg_q_log___` | hierarchy, values |
| A gate primitive, a switch primitive or a pull source is a `Forked<line>_<n>` process scope, one per instance whatever its name or width, on the counter of the module's processes; a gate delay and a drive strength add no scope, unit or declaration, and `tran`, `tranif1` and `trireg` do not elaborate | `t62_str_and_____` against `t62_str_wire____`, `Forked11_1` against `NetRegassign11_1` | `t62_str_and_2___`, `t62_str_bufif___`, `t62_str_bufif_n_`, `t62_str_nmos____`, `t62_str_gate_dly`, `t62_str_pullup__`, `t62_str_pulldn__`, `t62_str_vec_pu__` | hierarchy |
| A drive strength is resolved by the writer and not recorded: a strength case holds the two bit values of every other case; a pull source and a gate write records as an `assign` does, a gate delay is applied to the record time; a net with two or more drivers writes one record per bit of the net at each evaluation, from bit 0 up, and a literal driver with a strength adds one record at time 0 where a plain literal driver adds none | `t62_str_strong__` against `t62_str_equal___`, `0` then `1` against `0` then `X` and one record more; `t62_str_vec_2drv` against `t62_str_vec_1drv`, four records per write against one | `t62_str_weak____`, `t62_str_mixed___`, `t62_str_supply__`, `t62_str_wand____`, `t62_str_pullup__`, `t62_str_pulldn__`, `t62_str_pu_drv__`, `t62_str_and_____`, `t62_str_gate_dly`, `t62_str_bufif___`, `t62_str_nmos____`, `t62_str_vec_pu__`, counts pinned, and their VCDs | values |
| A driver of part of a net writes the pairs its bits fall in, whole, with the resolved value of the bits it does not drive, at the pair's address through the chunk map; the net's first record holds `X` on the driven bits and `Z` on the rest; each partial driver writes its own record; an output port bound to part of a net shares the net's handle with the bit offset and holds no record | `t63_pdr_bit0____` against `t62_str_vec_1drv`, `ZZZX` then `ZZZ1` for `XXXX` then `1119`; `t63_pdr_port_bit` against `t37_v_port_bit__`, offset 1 on the output side | `t63_pdr_bit3____`, `t63_pdr_slice___`, `t63_pdr_w64_bit0`, `t63_pdr_w64_bit6`, `t63_pdr_w64_hi__`, `t63_pdr_2400_bit`, `t63_pdr_2400_hi_`, `t63_pdr_two_bits`, `t63_pdr_concat__`, `t63_pdr_port_slc`, `t63_pdr_port_hi_`, counts pinned, and their VCDs | values |
| The port position word is written on the first instance of a unit only; every port of every later instance of the same unit holds 0, and the first instance of another unit holds its own | `t64_ord_two_kids` against `t63_pdr_port_bit`, `tb.u1.o` at 0 for `tb.u.o` at 1 | `t64_ord_two_nets`, `t64_ord_two_same`, `t64_ord_three___`, `t64_ord_two_pos4`, `t64_ord_two_mods`, `t64_ord_gen_kids`, against `t64_ord_pos_expr` and `t64_ord_pos_bit3` | hierarchy |
| An `assign` in a generate block is a `NetRegassign` scope of the module on its process counter with nothing of the block in its name; an instance in the block keeps the block's name | `t64_ord_gen4____` against `t63_pdr_two_bits`, four `tb.NetRegassign11_n` and no `g[i]` | `t64_ord_gen_rev_`, `t64_ord_gen_kids` | hierarchy |
| Several partial drivers on one net write one record each at the pair's address, the source order at time 0, the scheduler's order later, a driver fed from the net repeating the value held; a driver of an element of an unpacked array of nets writes the element's pair; an `inout` port on a bit shares the handle with the offset | `t64_ord_src_rev_` against `t63_pdr_two_bits`, `1ZZX` then `1ZZ0` for `XZZ0` then `1ZZ0`; `t64_ord_two_same` against `t64_ord_two_kids`, `u1` first at 50 ns for `u0` first | `t64_ord_gen4____`, `t64_ord_gen_rev_`, `t64_ord_w64_two_`, `t64_ord_2400_two`, `t64_ord_self____`, `t64_ord_chain___`, `t64_ord_unp_elem`, `t64_ord_unp_whol`, `t64_ord_gen_kids`, `t64_ord_inout___`, counts pinned, and their VCDs | values |
| A `final` block and an `always_latch` are `Always` scopes; an immediate or concurrent assertion, a named `sequence` or `property` and a `specify` block leave no scope, unit or declaration; a `covergroup` leaves three generated `function` scopes named `xlnx_isim_covergroup_<name>::<part>`, two `Block` scopes and two `Forked` scopes on the module's process counter | `t66_prc_final___` against `t11_sv_logic____`, `tb.Always11_0` for nothing; `t66_prc_covgrp__` against `t11_sv_logic____`, nine scopes for one | `t66_prc_latch___`, `t66_prc_ass_imm_`, `t66_prc_ass_conc`, `t66_prc_prop____`, `t66_prc_specify_` | hierarchy |
| A `program` is a module unit whose instance is an ordinary scope, and the simulation ends when its `initial` block ends; a `bind` puts the bound instance under the target scope at the `bind` line; a `specify` path delay delays the records of the path's output | `t66_prc_program_` against `t11_sv_logic____`, `tb.p.Initial20_0` and an end time of 10 ns; `t66_prc_specify_` against `t66_prc_kid_____`, 1 ns and 51 ns for 0 and 50 ns | `t66_prc_bind____`, `t66_prc_spec_0__`, whose zero delay records at 0 and 50 ns | hierarchy, values |
| The records of an object belong to the signal at its handle, so the signal's size sets the chunk boundaries: a port bound to a slice of a chunked signal sees a chunk boundary inside its own bytes | `//hdl/neorv32:sim`, `bus_req_i` of 88 bytes at offset 1408 of a 1760 byte array whose initial write is chunked at 146 | the 39 objects of that shape in the same design, and every case and design of the corpus, unchanged | values |
| A SystemVerilog enumeration declares its width in the range its own entry carries after the literals; the base type of `enum logic [1:0]` is a plain `logic` and carries none | `t67_esz_pk_2bit_` against `t11_sv_struct___`, a 4 bit struct whose VCD reads `1011` | `t67_esz_pk_4bit_`, `t67_esz_pk_int__`, and `//hdl/ibex:sim`, whose `csr_op_e` is 2 bits and `x_debug_ver_e` 4 in one file | types |
| The characters of a SystemVerilog string are in no part of the file, at any length, under typical or `-debug all`, logged or not; the placeholder record under `-debug all` is the same eight zero bytes for four characters and for forty | a search of the bytes and of every record of every page, `bazel run //tools/pagegrep -- -pat ZQXJ`, over `t68_str_lit4____`, `t68_str_lit40___`, `t68_str_dbg_____`, `t68_str_dbg40___`, `t68_str_log_____` and `t68_str_arr_____`, against the control `t68_str_byte____`, where the same characters are found | the same search over `t11_sv_str______`, `t60_dbg_str_____` and `t60_dbg_str_log_` | values |
| An unpacked array of strings under `-debug all` is one object of one placeholder per element: `string a [0:1]` is a 64 bit declaration with one 16 byte record of zeros | `t68_str_dbg_arr_` against `t68_str_dbg_____` | `t68_str_arr_____`, absent under typical as the scalar is | values |
| A net's initializer is a continuous assignment and not an initial value, so it takes value class 0 where the same literal on a variable takes 1; a `const` variable takes the class of its initializer and nothing else, and a parameter overridden by `defparam` or by `-generic_top` looks like any other parameter | `t69_vcl_wire_ini` against `t11_sv_logic____`, class 0 for `wire w = 1'b1` | `t69_vcl_const_v_`, `t69_vcl_const_i_`, `t69_vcl_defparam`, `t69_vcl_gtop_prm` | hierarchy |
| A `specparam` is a parameter declaration with an object and a record; a variable of a `parameter type` carries the parameter's name as its type; a `chandle` leaves no declaration under typical | `t69_vcl_specprm_`, `t69_vcl_typeprm_` and `t69_vcl_chandle_` against `t11_sv_logic____` | `t69_vcl_bits_prm`, whose `$bits` parameter is an ordinary class 3 | hierarchy |
| The key of an associative array is a numbered type of the numbering under `-debug all`: it carries the number in its own entry when it is not the element's type, spends one on the shared entry when it is, and spends none when it is a `string` or an enumeration; one number per associative array is left over, and a dynamic array and a queue leave none | `t70_num_a_i_byte` against `t60_dbg_assoc_i_`, the `byte` key holding 1 | `t70_num_a_b_int_`, `t70_num_a_b_str_`, `t70_num_a_v_str_`, `t70_num_a_e_key_`, `t70_num_a_2dim__`, `t70_num_a_in_cls`, `t70_num_d_then_q` | types |

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
| `t8_ps` against `t1_bit_two_edges` | waits of 1 ps and 1500 fs | picosecond unit at the default precision, femtoseconds truncated |
| `t8_rec_realv` against `t7_rec_intv` | a real beside the vector | a real contributes no triple |
| `t8_gen_if` against `t7_gen_for` | an if generate with a constant condition | plain branch scopes, an empty false branch, kind `0x13` is a constant |
| `t8_gen_nest` against `t7_gen_for` | a nested for generate | the iteration and empty label scopes repeat per level |
| `t9_vec292` against `t9_vec261`, `t9_vec257` | value size 292 not 261 or 257 | values over 257 bytes are chunked into records with consecutive keys |
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
| `t8_port_inout` against `t8_port_in` | `std_logic` in place of `std_ulogic` | the enumeration entry is named after the subtype |
| `t14_v_spill_d` against `t13_v_tr430_2` | the second `reg` moved into the arena that spills | its `X` record stays; only the clock's goes |
| `t14_v_spill_dfst` against `t14_v_spill_d` | the second `reg` declared before the clock | the same; the loss does not go with the first record of the arena |
| `t14_v_spill_dd` against `t14_v_page_dd` | two writes across `#0` in a page that spills, against one that stays | one record where there were two: a page written during the run keeps the last record per key and time |
| `t14_v_spill_dd2` against `t14_v_spill_dd` | the two writes moved into the second page | both records stay; the last page of an arena keeps every delta |
| `t15_sv_pkg_log` against `t13_sv_pkg` | `log_wave -recursive /p` added to the script | the package parameter is marked logged and holds one record at time 0 |
| `t15_sv_iface_vec` against `t13_sv_iface` | a vector added to the interface | one more declaration and one more shared object per scope |
| `t15_sv_iface_mp` against `t13_sv_iface` | the child takes a modport | unit kind `0x02`; a modport scope under the instance and the port scope both of that unit; port mode `in` |
| `t16_v_wire_rd1` against `t11_v_wire` | a second wire assigned from the first | the read wire holds one `X` record more |
| `t16_v_wire_rd2` against `t16_v_wire_rd1` | a third wire assigned from the first | still one more, not two |
| `t16_v_wire_rdp` against `t16_v_wire_rd1` | an `always` process reading the wire in place of the assignment | the same one more |
| `t16_v_wire_rdi` against `t16_v_wire_rd1` | the wire connected to an input port that nothing reads | one `X` per object and no more |
| `t17_v_reg_same` against `t11_v_bit_edge` | the write made with the value held | no record at 50 ns |
| `t17_v_net_same` against `t11_v_wire` | the assignment masked to a constant 0 | no record when the driver re-evaluates to the value held |
| `t17_v_mem_same` against `t11_v_mem4` | the element written with its value | no record at 50 ns |
| `t17_v_reg_nb` against `t11_v_bit_edge` | `<=` in place of `=` | the same three records |
| `t17_v_nb_swap` against `t17_v_reg_nb` | two regs swapped by nonblocking writes | one record each at 50 ns |
| `t18_arr_2dim` against `t2_array2d` | two index dimensions on one array type | `dims` counts index dimensions and one index type word follows per dimension |
| `t18_arr_3dim` against `t18_arr_2dim` | a third index dimension | `dims` `3`, three index words, three triples |
| `t18_sig_read` against `t1_bit_one_edge` | a concurrent assignment reading the signal | no extra record; only Verilog nets count readers |
| `t19_arr_2d_vec` against `t18_arr_2dim` | vectors as the elements of a two dimensional array | three triples under `dims` `2`; row major value |
| `t19_arr_of_2dim` against `t18_arr_2dim` | an array of the two dimensional array | one triple; the constrained element adds none |
| `t19_v_tri` against `t11_v_wire` | `tri` for `wire` | the declaration kind word, `0x06` for `0x03`, and nothing else |
| `t19_v_wand` against `t11_v_wire` | two drivers on a `wand` | kind `0x04`; two `X` records with nothing reading |
| `t19_v_wire_3drv` against `t11_v_wire` | three drivers | two `X` records, not three |
| `t19_v_wand_rd` against `t19_v_wand` | a reader beside the two drivers | still two `X` records |
| `t19_v_2drv_port` against `t19_v_wire_3drv` | the net also on an input port | three `X` records: one per object plus one |
| `t19_v_wire_nodrv` against `t11_v_wire` | no driver | one `Z` record and no `X` |
| `t19_v_nodrv_rd` against `t19_v_wire_nodrv` | a reader on the undriven wire | still one `Z`; the reader holds `X` then `Z` |
| `t19_v_supply1` against `t11_v_wire` | a `supply1` | kind `0x0d`; `X` then `1` with no driver |
| `t19_sv_uwire` against `t11_v_wire` | `uwire` | kind `0x03`, the same as `wire` |
| `t20_enum_bitlike` against `t2_bit` | a user type with `BIT`'s literals | class `2`, the same as `BIT`; the class follows the literals |
| `t20_enum_ul_like` against `t1_bit_one_edge` | a user type with `STD_ULOGIC`'s literals | class `3`, the same as `STD_ULOGIC` |
| `t20_enum_chars` against `t2_enum` | character literals `'a'` `'b'` `'c'` | class `4`, the class of `CHARACTER` |
| `t20_enum_mixed` against `t2_enum` | identifiers with one character literal | class `4` |
| `t20_enum_one` against `t2_enum` | one literal | class `5`; assigning the only literal again records nothing |
| `t20_enum_two_id` against `t2_enum` | two identifiers | class `5`, the same as `BOOLEAN` |
| `t20_enum_300` against `t2_character` | 300 literals | the last word `4`; a 4 byte value |
| `t20_enum_256` against `t20_enum_300` | 256 literals | the last word `1`; a 1 byte value |
| `t20_enum_257` against `t20_enum_256` | one literal more | the last word `4`; the boundary |
| `t20_enum_300_arr` against `t20_enum_300` | an array of two of the wide type | 8 bytes; 4 per element |
| `t20_enum_300_rec` against `t20_enum_300` | a record of `std_ulogic` and the wide type | 8 bytes; the wide field at offset 4 |
| `t20_arr_2d_uncon` against `t18_arr_2dim` | the two dimensional type left unconstrained | two `(0, 0, -2)` triples; the ranges on the declaration |
| `t20_rec_2dim` against `t18_arr_2dim` | the two dimensional array as a record field | two triples on the field |
| `t20_v_realp_big` against `t12_v_params` | a `real` parameter of `123456.789` | still 16 bits |
| `t20_v_realp_lp` against `t12_v_params` | `localparam real` | still 16 bits |
| `t21_int_neg` against `t2_integer` | `-165` for `165` | two's complement int32 |
| `t21_int_sub` against `t2_integer` | `subtype small_t is integer range 0 to 7` | an entry `small_t` `0 to 7`; 4 bytes |
| `t21_int_newtype` against `t21_int_sub` | `type` for `subtype` | byte identical outside the noise mask |
| `t21_real_neg` against `t2_real` | `-1.5` for `1.5` | the float64 sign bit |
| `t21_phys_user` against `t2_time` | a user physical type | units scaled in `um`; `3 mm` stored as 3000 |
| `t21_bitvec8` against `t2_slv8` | `bit_vector` for `std_ulogic_vector` | `BIT_VECTOR` over `BIT`, one byte per element |
| `t21_v_param_same` against `t11_v_param` | a second instance with the same parameter | one unit, two `K` objects |
| `t21_v_param_diff` against `t21_v_param_same` | different parameter values | still one unit; the values differ in the `K` records |
| `t21_v_ts_1ns_1ns` against `t11_v_bit_edge` | `timescale 1ns / 1ns` | the DBG word `-9`; the change at 50, the end at 100 |
| `t21_v_ts_1ps_1ps` against `t21_v_ts_1ns_1ns` | `1ps / 1ps` | `-12`; 50 and 100 |
| `t21_v_ts_10ns` against `t21_v_ts_1ns_1ns` | `10ns / 1ns` with `#5` | `-9`; 50 and 100, the scale unit does not reach the file |
| `t21_v_ts_1ns_100` against `t21_v_ts_1ns_1ns` | `1ns / 100ps` with `#50.55` | `-10`; 506 and 1001 |
| `t21_v_ts_1ps_1fs` against `t21_v_ts_1ps_1ps` | `1ps / 1fs` with `#50.5` | `-15`; 50500 and 100000 |
| `t21_v_ts_none` against `t21_v_ts_1ns_1ns` | no `timescale` | `-12` |
| `t21_mix_ts_1ns` against `t21_mix_vh_in_v` | a `1ns / 1ns` testbench over a VHDL child | `-12`; the VHDL precision wins |
| `t21_mix_v_in_vh` against `t9_comp` | the child written in Verilog | a `module` unit and a `net` port beside the VHDL entries; the port on its own handle, `X` then `0` |
| `t21_mix_vh_in_v` against `t11_v_port` and `t9_comp` | the child written in VHDL | an `entity` unit and a `signal` port beside the Verilog entries; the port on its own handle, `U` then `0` |
| `t22_base` against `t8_ps` | a generic on `tb`, a `TIME` signal and a function | the generic as `tb.k`; the function absent |
| `t22_dbg_wave` against `t22_base` | `-debug wave` | regions 14 and 15 empty, 32 bytes less |
| `t22_dbg_subprog` against `t22_base` | `-debug subprogram` beside `typical` | a `0x11` unit and scope `tb.inc`; locals of kind `0x14` at `0x40` and `0x44` |
| `t22_dbg_sub_proc` against `t22_dbg_subprog` | a procedure beside the function | a `0x12` unit; `inout` and `in` modes; locals at `0xd0` to `0xd8` |
| `t22_dbg_all` against `t22_base` | `-debug all` | four library packages, 15 more objects, 16 more types, the `TEXT` kind |
| `t22_vh_fs` against `t22_base` | `--timeprecision_vhdl 1fs` | `-15`; `TIME` rescaled; `1500 fs` kept |
| `t22_vh_ns` against `t22_base` | `--timeprecision_vhdl 1ns` | `-9`; a zero length wait; two records at one time |
| `t22_o0` against `t22_base` | `--O0` | byte identical outside the noise mask |
| `t22_mt_off` against `t22_base` | `--mt off` | byte identical outside the noise mask |
| `t22_gen_top` against `t22_base` | `--generic_top k=9` | the records of `k` and `n` only |
| `t23_file_text` against `t22_dbg_all` | `file f : text` in the design | the `TEXT` kind under `-debug typical`; `f` a 0 byte variable |
| `t23_file_int` against `t23_file_text` | a file of integer | the same words `8` and `40` |
| `t23_file_sul` against `t23_file_int` | a file of std_ulogic | the same words again |
| `t23_shared_int` against `t6_var_int` | a shared variable | kind `0x0f` in scope `tb`, on `0xde0` |
| `t23_protected` against `t23_shared_int` | a protected type | a record entry `counter`; the object on `0x858` with two handles |
| `t23_access` against `t6_var_int` | an access type | kind `0x8`; the reader crashed; 48 bytes declared |
| `t23_access_vec` against `t23_access` | access to an unconstrained array | the same words `8` and `48` |
| `t23_sub_sizes` against `t22_dbg_subprog` | locals of 1, 4, 8, 8 and 4 bytes | offsets `0x40 0x44 0x48 0x50 0x68`; alignment |
| `t23_sub_vec16` against `t23_sub_sizes` | a 16 element vector local | `m` stays at `0x68` |
| `t23_sub_vec32` against `t23_sub_vec16` | a 32 element vector local | `m` stays at `0x68`; 24 bytes is a descriptor |
| `t23_sub_sig_prm` against `t22_dbg_sub_proc` | a signal parameter | kind `0x15` `port out`; the next local 64 bytes on |
| `t23_sub_in_proc` against `t22_dbg_sub_proc` | a procedure inside a process | scopes `tb.flip` and `tb.p.flip`; three objects |
| `t23_arch_b` against `t4_gen_explicit` | `entity work.child(b)` | the unit `child(b)`; no unit for `a` |
| `t23_arch_both` against `t23_arch_b` | both architectures instantiated | units `child(a)` and `child(b)`, nothing shared |
| `t23_int_vector` against `t5_int_arr` | `integer_vector(0 to 3)` | an unconstrained `INTEGER_VECTOR` entry; bounds in the declaration |
| `t23_real_vector` against `t23_int_vector` | `real_vector` | 8 bytes per element |
| `t23_time_vector` against `t23_int_vector` | `time_vector` | the `TIME` entry under origin `0xa` |
| `t23_bool_vector` against `t23_int_vector` | `boolean_vector` | 1 byte per element |
| `t24_att_delayed` against `t24_att_stable` | `'delayed` for `'stable` | two `tb.line__18` scopes for one |
| `t24_att_quiet` against `t24_att_stable` | `'quiet` | the same shape; `q` records like `b` |
| `t24_att_transact` against `t24_att_stable` | `'transaction` | one scope; `t` toggles at 50 and 70 ns |
| `t24_null_range` against `t1_bit_one_edge` | `std_ulogic_vector(0 downto 1)` | a 0 byte declaration, an unlogged object |
| `t24_two_drivers`, `r` against `s` | two drivers of a `std_logic` | `Z` twice at time 0, then `1`, then `X` |
| `t24_dbg_drivers` against `t24_two_drivers` | `-debug drivers` over `typical` | nothing |
| `t24_dbg_readers` against `t24_two_drivers` | `-debug readers` over `typical` | one byte at `0x303`: word 14 byte 2 |
| `t24_dbg_line` against `t24_dbg_drv_only` | `line` for `drivers` under `wave` | word 14 `0x101` to `0x1`, word 15 `0x1` to `0x10101`, region 15 filled |
| `t24_dbg_drv_only` against `t22_dbg_wave` | `drivers` over `wave` | word 14 byte 1 |
| `t24_dbg_sub_only` against `t24_dbg_line` | `subprogram` for `line` | the same words and regions |
| `t24_dbg_xlibs` against `t24_dbg_drv_only` | `xlibs` for `drivers` | the four packages, 15 objects, no `resolved` scope, word 14 byte 1 clear |
| `t24_case_gen` against `t8_gen_if` | `case k generate` | one scope `tb.g`, the taken alternative's signal, no scope for the other |
| `t24_ext_name` against `t24_config_spec` | an external name reading `tb.dut.s` | `tb.line__19`; the change of `tb.dut.s` twice |
| `t24_config_spec` against `t23_arch_b` | a component with `for dut : child use entity work.child(a)` | the unit `child(a)`, nothing else |
| `t24_sv_queue` against `t11_sv_logic` | `int q[$]` | nothing but `0xf8` of handle space |
| `t24_sv_dynarr` against `t24_sv_queue` | `int d[]` | the same |
| `t24_sv_assoc` against `t24_sv_queue` | `int a[string]` | the same |
| `t24_sv_class` against `t24_sv_queue` | a class handle | the same |
| `t24_sv_union` against `t11_sv_struct` | `union packed` for `struct packed` | layout `6`; 8 bits for two 8 bit fields |
| `t24_sv_fork` against `t11_v_always` | `fork ... join` | `tb.Forked13_1`, `tb.Forked14_2` |
| `t24_sv_clocking` against `t11_sv_logic` | a clocking block | nothing |
| `t25_sv_two_class` against `t25_sv_two_same` | `int i = 5` for `logic t = 1'b1` | word 13 `2` for `1`; region 17 `[1 0 0] [3 0 0]` for `[1 0 0]` |
| `t25_sv_vec8_sz` against `t25_sv_vec8_int` | `8'h00` for `0` | region 17 `[1 0 0]` for `[4 0 0]` |
| `t25_sv_logic_int` against `t11_sv_logic` | `0` for `1'b0` | `[4 0 0]` for `[1 0 0]` |
| `t25_sv_bit_unsz` against `t25_sv_logic_int` | `bit` for `logic` | the same `[4 0 0]` |
| `t25_sv_v64_unsz` against `t25_sv_vec8_int` | `[63:0]` for `[7:0]` | the same `[4 0 0]` |
| `t25_sv_logic_one`, `t25_sv_logic_x`, `t25_sv_logic_exp` against `t11_sv_logic` | `'1`, `1'bx`, `1'b0 \| 1'b0` for `1'b0` | the same `[1 0 0]` |
| `t25_sv_int_sized`, `t25_sv_int_noini` against `t11_sv_int` | `32'h0`, no initializer for `0` | the same `[3 0 0]` |
| `t25_sv_byte_szd` against `t11_sv_byte` | `8'h05` for `0` | `[1 0 0]` for `[3 0 0]` |
| `t25_sv_real_lit` against `t11_v_real` | `real s = 1.5` in `.sv` | `[0 0 0]`, one record |
| `t25_sv_time_lit` against `t11_v_time` | `time s = 10ns` in `.sv` | `[4 0 0]`; `X` then `10` at time 0 |
| `t25_sv_time_noin` against `t25_sv_time_lit` | no initializer | `[4 0 0]`; one `X` record |
| `t25_v_reg_int` against `t11_v_bit_edge` | `reg s = 0` for `1'b0` | nothing |
| `t25_v_vec8_sz` against `t25_sv_vec8_sz` | `.v` for `.sv` | `[0 0 0]` for `[1 0 0]`; the `X` record |
| `t25_v_int_sized`, `t25_v_int_noinit` against `t11_v_integer` | `32'h0`, no initializer for `0` | the same `[3 0 0]` |
| `t25_v_prm_real` against `t20_v_realp_big` | an `integer s` beside the `real` parameter | `[3 0 0] [0 0 0]`: the parameter is class 0 |
| `t25_v_prm_dflt`, `t25_v_prm_none`, `t25_v_defparam` against `t11_v_param` | the override form | nothing but the value |
| `t25_v_prm_two` against `t11_v_param` | a second parameter | a second object, still `[0 0 0] [3 0 0]` |
| `t25_v_prm_lp`, `t25_v_prm_lp_ind` against `t11_v_param` | a localparam | an object with the computed value, class 3 |
| `t25_v_prm_tb` against `t11_v_bit_edge` | a parameter of the root module | `tb.K` as an object of `tb` |
| `t25_sv_wire` against `t19_sv_uwire` | `wire` for `uwire` | nothing |
| `t25_sv_net_init` against `t25_sv_wire` | `wire w = s` for `assign` | nothing; the net is class 0 |
| `t25_sv_alw_ff`, `t25_sv_always` against `t13_sv_alwaysff` | one process kind alone | the `Always` scope, `[1 0 0]` |
| `t25_sv_alw_comb`, `t25_sv_alw_latch` against `t13_sv_alwaysff` | one process kind alone | the uninitialized output is class 0 |
| `t25_sv_pkg_tdef` against `t13_sv_pkg` | the typedef alone | the package unit and scope, no object |
| `t25_sv_pkg_prm` against `t25_sv_pkg_tdef` | the parameter alone, used in a cast | no package at all; `0xf8` of handle space for the cast |
| `t25_sv_pkg_unusd` against `t25_sv_pkg_prm` | the cast removed | `0xf8` less handle space |
| `t26_sv_logic_prm` against `t11_sv_logic` | `logic s = K`, `parameter K = 1'b0` | `[1 0 0]` for both objects |
| `t26_sv_int_prm` against `t11_sv_int` | `int s = K` | `[3 0 0]` |
| `t26_sv_logic_fn` against `t11_sv_logic` | `logic s = f()` | nothing: no unit, no scope, one record |
| `t26_sv_v8_unshex` against `t25_sv_vec8_int` | `'h00` for `0` | `[1 0 0]` for `[4 0 0]` |
| `t26_sv_logic_1` against `t25_sv_logic_int` | `1` for `0` | the same `[4 0 0]` |
| `t26_sv_int_neg`, `t26_sv_int_szd5`, `t26_sv_int_szd64`, `t26_sv_int_unhex` against `t11_sv_int` | `-1`, `5'd3`, `64'h0`, `'h0` for `0` | the same `[3 0 0]` |
| `t26_sv_shortint`, `t26_sv_lng_szd` against `t11_sv_int` | `shortint s = 0`, `longint s = 64'h0` | `[3 0 0]`; the `shortint` entry `(15, 0, -1)` |
| `t26_sv_bit8_szd`, `t26_sv_bit8_int` against `t25_sv_vec8_sz` | `bit [7:0]` from `8'h00`, from `0` | `[1 0 0]`, `[4 0 0]` |
| `t26_sv_v32_int`, `t26_sv_v32_szd` against `t25_sv_vec8_int` | `logic [31:0]` from `0`, from `32'h0` | `[4 0 0]`, `[1 0 0]`: the literal's width is not it |
| `t26_sv_sgn8_neg` against `t11_v_signed8` | `logic signed [7:0] s = -1` in `.sv` | `[3 0 0]`; the record `11111111` |
| `t26_sv_sgn8_szd` against `t26_sv_sgn8_neg` | `8'h00` for `-1` | `[1 0 0]` |
| `t26_sv_real_int` against `t25_sv_real_lit` | `real s = 1` | `[0 0 0]`; `1` once |
| `t26_sv_v8_str` against `t25_sv_vec8_sz` | `"a"` for `8'h00` | `[6 0 0]`; `01100001` |
| `t26_sv_v8_cat`, `t26_sv_v8_rep`, `t26_sv_logic_cnd` against `t25_sv_vec8_sz` | a concatenation, a replication, a conditional | the same `[1 0 0]` |
| `t26_sv_integer_x` against `t25_sv_int_sized` | `integer s = 'x` | `[3 0 0]`; all `X` once |
| `t26_sv_byte_neg` against `t11_sv_byte` | `-1` for `0` | the same `[3 0 0]` |
| `t26_sv_str_prm` against `t13_v_str_param` | `parameter string P` in `.sv` | no object; 8 bytes of handle space over `t11_sv_logic` |
| `t26_sv_bit_prm`, `t26_sv_v8_prm` against `t26_sv_logic_prm` | `parameter bit`, `parameter logic [7:0]` | `[1 0 0]` |
| `t26_sv_lp_int` against `t26_sv_logic_prm` | `localparam int L = 3` | `[1 0 0] [3 0 0]` |
| `t26_sv_real_prm` against `t25_v_prm_real` | `parameter real R` in `.sv` | `[1 0 0] [0 0 0]`; 16 bytes |
| `t26_sv_event` against `t11_sv_logic` | `event e` | no object; `s` at `0xa28`; `0x2f8` more handle space |
| `t27_sv_int_uns`, `t27_sv_int_unsni`, `t27_sv_lng_uns`, `t27_sv_intg_uns`, `t27_v_intg_uns` against their signed siblings | `unsigned` | nothing but line numbers; `[3 0 0]` |
| `t27_sv_byte_uns` against `t11_sv_byte` | `byte unsigned s = 0` | `[4 0 0]` for `[3 0 0]`; the same records |
| `t27_sv_v8_neg` against `t26_sv_sgn8_neg` | `logic [7:0] s = -1`, unsigned | `[4 0 0]` for `[3 0 0]` |
| `t27_sv_sgn8_pos` against `t26_sv_sgn8_neg` | `5` for `-1` | the same `[3 0 0]` |
| `t27_sv_v8_ssized`, `t27_sv_sgn8_szdn`, `t27_sv_v8_uns32` against `t25_sv_vec8_sz` | `8'sh05`, `-8'sd1`, `32'd5` | the same `[1 0 0]` |
| `t27_sv_time_szd`, `t27_sv_time_uns` against `t25_sv_time_lit` | `64'h0`, `0` for `10ns` | `[4 0 0]`; `0` once, no `X` |
| `t27_sv_int_str`, `t27_sv_int_real` against `t11_sv_int` | `"a"`, `1.5` for `0` | `[3 0 0]`; `97`, `2` once |
| `t27_sv_int_time` against `t25_sv_time_lit` | `int s = 10ns` | `[3 0 0]`; `0` then `10` at time 0 |
| `t27_sv_real_szd` against `t25_sv_real_lit` | `real s = 8'h05` | `[0 0 0]`; `5` once |
| `t27_sv_v8_xfill`, `t27_sv_v8_zfill`, `t27_sv_v8_0fill` against `t25_sv_vec8_sz` | `'x`, `'z`, `'0` | the same `[1 0 0]` |
| `t27_sv_bit_noini`, `t27_sv_byte_noin`, `t27_sv_bit8_noin`, `t27_sv_v8_noini` against `t12_sv_noinit` | `bit`, `byte`, `bit [7:0]`, `logic [7:0]` without an initializer | `[0 0 0]`; `0` or `X` once |
| `t27_v_sgn8_neg` against `t26_sv_sgn8_neg` | `.v` for `.sv` | `[0 0 0]`; the `X` record |
| `t27_sv_str_untyp` against `t26_sv_str_prm` | `parameter P` for `parameter string P` | the object `tb.P`, 40 bits, class 6 |
| `t28_sv_v8_real` against `t27_sv_int_real` | `logic [7:0] s = 1.5` | `[0 0 0]`; `2` once |
| `t28_sv_v64_time` against `t27_sv_int_time` | `logic [63:0] s = 10ns` | `[0 0 0]`; all `X` then `10` |
| `t28_sv_enum_pkd` against `t11_sv_enum` | an enum over `logic [1:0]` from `RUN` | `[0 0 0]`; `XX` then `RUN` |
| `t28_sv_enum_cast` against `t11_sv_enum` | `state_t'(1)` | the hidden variable, `[3 0 0] [0 0 0]`; `IDLE` then `RUN` |
| `t28_sv_pstr_pat`, `t28_sv_pstr_szd`, `t28_sv_pstr_int` against `t11_sv_struct` | a packed struct from `'{}`, `5'b00000`, `0` | `[0 0 0]`, `[1 0 0]`, `[4 0 0]` |
| `t28_sv_uarr_pat`, `t28_sv_uarr_dflt` against `t12_sv_unp2d` | an unpacked array from `'{2'b00, 2'b01}`, `'{default: 2'b01}` | `[0 0 0]`; one record |
| `t28_sv_str_pat` against `t11_sv_ustruct` | a positional pattern for a named one | nothing |
| `t28_sv_prm_expr`, `t28_sv_prm_clog`, `t28_sv_prm_neg` against `t26_sv_logic_prm` | `parameter K = 2 * 3`, `$clog2(8)`, `-1` | 32 bit vectors, `[1 0 0] [3 0 0]` |
| `t28_sv_prm_szexp` against `t28_sv_prm_expr` | `4'd5 + 4'd1` | 4 bits, `[1 0 0]` |
| `t28_sv_prm_time` against `t28_sv_prm_expr` | `parameter T = 10ns` | a 64 bit vector, `[1 0 0] [4 0 0]`, the `float64` in the record |
| `t28_sv_prm_tmtyp` against `t28_sv_prm_time` | `parameter time T = 10ns` | the `time` entry, `10`; 8 bytes more handle space |
| `t28_sv_prm_realu` against `t26_sv_real_prm` | `parameter K = 1.5` untyped | 32 bits for 16, `[1 0 0] [0 0 0]` |
| `t28_sv_prm_int_u` against `t26_sv_lp_int` | `parameter int unsigned K = 5` | the `int` entry, `[1 0 0] [3 0 0]` |
| `t28_sv_prm_enum` against `t28_sv_prm_expr` | `parameter state_t S = RUN` | the alias entry, 32 bits, `[1 0 0] [3 0 0]` |
| `t28_sv_v64_stime` against `t28_sv_v64_time` | `$time` for `10ns` | all `X` then `0` |
| `t28_sv_v8_signed` against `t27_sv_v8_ssized` | `$signed(8'h05)` | the same `[1 0 0]` |
| `t28_sv_int_cast` against `t27_sv_int_real` | `int'(1.5)` | the hidden variable; `0` then `2`; `0x190` more handle space |
| `t28_sv_v8_szcast` against `t25_sv_vec8_int` | `8'(0)` | the hidden variable, `[3 0 0] [0 0 0]` |
| `t28_sv_v16_str` against `t26_sv_v8_str` | `logic [15:0] s = "a"` | `[6 0 0]`; `0000000001100001` |
| `t28_sv_prm_lstr` against `t26_sv_v8_prm` | `parameter logic [7:0] P = "a"` | `[1 0 0]`, not 6 |
| `t28_sv_v8_prmneg` against `t27_sv_v8_neg` | `logic [7:0] s = K`, `parameter K = -1` | `[4 0 0] [3 0 0]` |
| `t28_sv_bit8_neg` against `t27_sv_v8_neg` | `bit [7:0] s = -1` | the same `[4 0 0]` |
| `t28_sv_v8_bitsel` against `t26_sv_logic_prm` | `logic [7:0] s = K[7:0]` | `[1 0 0]` |
| `t28_sv_v8_pow` against `t26_sv_logic_1` | `logic [7:0] s = 2 ** 3` | the same `[4 0 0]` |
| `t28_sv_real_time` against `t27_sv_int_time` | `real s = 10ns` | `[0 0 0]`; `0` then `10` |
| `t28_sv_rtime_var`, `t28_sv_rtime_noi` against `t11_v_real` | `realtime s = 10ns`, `realtime s` | the `realtime` entry, 32 bits; `0` then `10`; `0` |
| `t28_sv_rtime_prm` against `t28_sv_prm_time` | `parameter realtime T = 10ns` | the `realtime` entry, 16 bits, `[1 0 0] [0 0 0]` |
| `t29_sv_cast_proc` against `t29_sv_incr` | `s = int'(2.5)` in a process | no hidden variable; `0x1f0` more handle space |
| `t29_sv_cast_two` against `t28_sv_int_cast` | two initializers with casts | `temp_0_ln5` and `temp_1_ln6`, each before its variable; one implicit process |
| `t29_sv_cast_same` against `t28_sv_int_cast` | `int'(1.5) + int'(2.5)` in one initializer | nothing; `5` once |
| `t29_sv_cast_asgn` against `t29_sv_asgn_noc` | `assign w = 8'(s + 1)` | a second `NetRegassign` scope; `0x198` more handle space |
| `t29_sv_cast_fn` against `t29_sv_fn_noc` | `return int'(2.5)` in a function | no hidden variable; `0x1f0` more handle space |
| `t29_sv_fn_noc` against `t29_sv_incr` | `s = f()` with `function int f` | the scope `tb.f` and the object `tb.f.f`, `0` then `3` |
| `t29_sv_cast_alwc` against `t29_sv_alwc_noc` | `always_comb w = 8'(s + 1)` | no hidden variable; `0xf8` more handle space |
| `t29_sv_cast_prm` against `t28_sv_prm_expr` | `parameter K = int'(1.5)` | a 32 bit vector holding `2`; the same handle space |
| `t29_sv_cast_sub` against `t28_sv_int_cast` | the cast in a child module | `tb.u.xilinx_isim_temp_0_ln5castingOp` before `tb.u.s` |
| `t29_sv_cast_sgn` against `t28_sv_v8_signed` | `signed'(8'h05)` for `$signed(8'h05)` | the hidden variable, class 1; `s` class 0, `X` then the value |
| `t29_sv_cast_real` against `t28_sv_int_cast` | `real s = real'(3)` | a hidden `real`, `[0 0 0]`; `0` then `3` |
| `t29_sv_stream`, `t29_sv_bits` against `t28_sv_v8_signed`, `t28_sv_int_cast` | `{<<{8'h05}}`, `$bits(s)` | nothing; `10100000`, `32` once |
| `t29_sv_incr` against `t29_sv_cast_proc` | `s++` | nothing |
| `t29_sv_for_int` against `t29_sv_incr` | `for (int i = 0; i < 3; i++)` | the block scope `tb.Block7_1` and the object `i`, `0` then `3` |
| `t29_sv_for_modi` against `t29_sv_for_int` | `integer i` in the module | `i` in the module scope, `X`, `0`, `1`, `2`, `3` |
| `t29_sv_foreach` against `t29_sv_for_int` | `foreach (a[i])` | `tb.Block8_1.i`, `0`, `1`, `2`, `3`; `a` an unnamed array |
| `t30_sv_ptm_two` against `t28_sv_prm_time` | a second untyped time parameter | `T1` records `float64(10)` then `float64(20)`; the 16 byte record over-reads the next parameter |
| `t30_sv_ptm_20ns`, `t30_sv_ptm_10ps`, `t30_sv_ptm_1us`, `t30_sv_ptm_1s`, `t30_sv_ptm_frac` against `t28_sv_prm_time` | `20ns`, `10ps`, `1us`, `1s`, `10.5ns` | `20`, `0.01`, `1000`, `1e9`, `10.5`: the `float64` counts the time unit and keeps the fraction |
| `t30_sv_ptm_ps_ts` against `t28_sv_prm_time` | `timescale 1ps / 1ps` | `10000`; the value follows the unit |
| `t30_sv_ptm_late` against `t28_sv_prm_time` | the parameter after the variable | nothing |
| `t30_sv_ptm_expr` against `t30_sv_ptm_20ns` | `parameter T = 10ns * 2` | the `real` entry, 32 bits, class 0, one `float64` `20` |
| `t30_sv_v8_ubase`, `t30_sv_v8_sbase` against `t25_sv_logic_int` | `'d5`, `'sd5` into `logic [7:0]` | class 1 for 4; a based literal counts as sized |
| `t30_sv_v8_negsz`, `t30_sv_v8_mixed`, `t30_sv_v8_szexp`, `t30_sv_v8_cnd`, `t30_sv_v8_cmp`, `t30_sv_v8_1fill` against `t25_sv_logic_int` | `-8'd1`, `8'd5 + 1`, `4'd5 + 4'd1`, `1'b1 ? 8'd5 : 0`, `(1 < 2)`, `'1` | class 1 |
| `t30_sv_v8_sgnu` against `t30_sv_v8_uns` | `$signed(5)` for `$unsigned(5)` | class 4 for 1 |
| `t30_sv_v8_realx` against `t28_sv_v8_real` | `5 + 1.5` | class 0, like the real literal |
| `t30_sv_v8_str2`, `t30_sv_v16_strc` against `t26_sv_v8_str` | `"ab"` into 8 bits, `{"a", "b"}` into 16 | class 6; `01100010` and `0110000101100010` |
| `t30_sv_prm_ubase`, `t30_sv_prm_szsgn`, `t30_sv_prm_neg8`, `t30_sv_prm_cmp` against `t26_sv_logic_prm` | `'d5`, `8'sd5`, `-8'd1`, `(1 < 2)` untyped | class 1, 32, 8, 8 and 1 bits; the comparison is a `logic` |
| `t30_sv_prm_wide` against `t30_sv_prm_ubase` | `40'h1` | 40 bits, a 16 byte record, `0x92c` of handle space for `0x924` |
| `t30_sv_prm_shft`, `t30_sv_prm_cnd` against `t28_sv_prm_expr` | `1 << 40`, `1'b1 ? 3 : 4` | class 3, 32 bits; the shift records `0` |
| `t30_sv_prm_realx` against `t28_sv_prm_realu` | `5.0 * 2` | the `real` entry, class 0; nothing else |
| `t30_sv_prm_strc` against `t27_sv_str_untyp` | `{"a", "b"}` | class 6, 16 bits, `62 61` |
| `t31_sv_w1_swap` against `t31_sv_w1_i5` | `int i = 5` before `logic s` | declaration word 1 swapped with the entries: it indexes region 17 |
| `t31_sv_w1_i0`, `t31_sv_w1_i1`, `t31_sv_w1_i165`, `t31_sv_w1_nowrt`, `t31_sv_w1_own50` against `t31_sv_w1_i5` | the value and the writes of `i` | nothing in word 1 |
| `t31_sv_w1_s5` against `t31_sv_w1_i5` | the `int` alone | one entry `[3 0 0]`, word 1 `0` |
| `t31_sv_w1_v8_5` against `t31_sv_w1_i5` | `logic [7:0] i = 5` for `int i = 5` | `[1 0 0] [4 0 0]`, word 1 `1` either way |
| `t31_v_w1_int5` against `t31_sv_w1_i5` | the same pair in a `.v` file | `[0 0 0] [3 0 0]`, `s` class 0 as every `.v` variable |
| `//hdl/counter:sim` against the corpus | a record driven one field at a time | a VHDL record is one write, not the whole value: three record addresses for one object |
| `t32_rec_field___` against `t32_rec_whole___` | `r.b <= '1'` for the whole record | 1 byte at the handle plus 1 for 8 at the handle |
| `t32_rec_two_adj_` against `t32_rec_two_gap_` | fields `b`, `c` for `a`, `c` in one delta | adjacent parts merge into one record; a gap gives two records at one time |
| `t32_rec_delta___` against `t32_rec_two_adj_` | a `wait for 0 ns` between the two fields | two records for two deltas, even for adjacent bytes |
| `t32_rec_wthenf__`, `t32_rec_fthenw__` against `t32_rec_delta___` | a whole assignment beside a field in one delta | one whole record holding the result, in either order |
| `t32_rec_conc____`, `t32_vec_slc_conc` against `t32_rec_field___`, `t32_vec_slice___` | a concurrent assignment for a process | the same record |
| `t32_rec_vecfield`, `t32_rec_intfld__`, `t32_rec_intlast_` against `t32_rec_field___` | a vector field, an integer field, a field behind an integer | the offset is the field's aligned offset: `+5`, `+0`, `+4` |
| `t32_vec_slice___`, `t32_vec_elem____`, `t32_vec_to_slice` against `t32_rec_field___` | a slice and an element of an 8 bit vector, `downto` and `to` | the offset counts from the left bound: `+4`, `+5`, `+0` |
| `t32_vec_adj_slc_`, `t32_vec_two_slc_` against `t32_vec_slice___` | a second slice, touching or not | 6 bytes at `+2`; 4 bytes at `+4` then 2 at `+0`, in source order |
| `t32_vec_slc_over` against `t32_vec_two_slc_` | the whole vector then a slice | one 8 byte record, `03 03 03 03 02 02 02 02` |
| `t32_arr_elem____`, `t32_arr_row_____`, `t32_arr_row_bit_`, `t32_arr2d_elem__` against `t32_vec_elem____` | an integer element, a row, a bit of a row, a 2D element | `+4` of 4, `+8` of 8, `+13` of 1, `+3` of 1 |
| `t32_wide_slice__` against `t32_vec_slice___` | the slice on a 600 byte vector | a 300 byte write is four chunks of 75 from its own address |
| `t32_wide_top____`, `t32_wide_small__` against `t32_wide_slice__` | the high half; 4 bits | chunks at the handle; one 4 byte record at `+596` |
| `t32_wide_field__`, `t32_wide_tail___`, `t32_wide_tail_a_` against `t32_wide_slice__` | a 300 byte field before and behind a `std_ulogic` | chunks of 75 at `+0` and at `+1` where the whole 304 byte record was four of 76; 1 byte at `+0` |
| `t33_v_wsl_hi____` against `t12_v_vec4800x` | `s[4799:2400]` for `s[2400]` | a Verilog partial write of 600 bytes is six chunks of 100 from its own address |
| `t33_v_wsl_lo____`, `t33_v_wsl_4b____` against `t33_v_wsl_hi____` | the low half; four bits | six chunks at the handle, split at `0x800`; one pair |
| `t33_v_wsl_mid___` against `t33_v_wsl_lo____` | the slice moved up 16 bits | 608 bytes in six chunks: the write is whole pairs, 0 to 75 |
| `t33_v_wsl_272___` against `t33_v_wsl_280___` | 1088 bits against 1089 from pair 34 | one record of 272; four of 70: the threshold of a partial write is 275 |
| `t33_sv_wsl_hi___` against `t33_v_wsl_hi____` | a `.sv` file | the same six chunks |
| `t33_v_mem_row___` against `t33_v_wsl_hi____` | a row of `reg [2399:0] m [0:3]` | six chunks at the row's pairs, `+1200` for `m[1]`; arena 1 holds the time zero rows last slot first |
| `t33_sv_st_wide__` against `t33_sv_wsl_hi___` | a 2400 bit struct field with a `logic` after it | six chunks at `+8`, above the last field's pair |
| `t34_port_fld_out` against `t32_rec_field___` | the field written in a child through a record port | the same 1 byte record at `+1` on the shared handle |
| `t34_pmap_field__` against `t9_port_slice` | a port bound to a record field | offset 1 in the instance record; the child's write at `+1` |
| `t34_pmap_slice__` against `t9_port_slice2__` | a slice written in the child of a slice bound port | 2 bytes at `+6`: the port's offset 4 and the slice's 2 |
| `t34_two_prc_adj_` against `t32_rec_two_adj_` | the second field from a second process | two 1 byte records, `q`'s first, where one process gave one of 2 |
| `t34_two_prc_rev_` against `t34_two_prc_adj_` | the fields swapped between the processes | `q`'s record still first: source order of the processes, reversed, not address |
| `t34_gen_elems___` against `t32_vec_slc_conc` | three concurrent element assignments under a `for generate` | three 1 byte records at 50 ns, `g(2)` first |
| `t34_res_two_fld_` against `t34_two_prc_adj_` | `std_logic` fields | the same records: a field with one driver is not resolved |
| `t34_res_same_fld` against `t34_res_two_fld_` | both processes on `r.b`, one assigning `'Z'` at time 0 | the whole record, then `Z` at `+1` at time 0; `X` at 70 ns |
| `t34_res_two_drv0` against `t24_two_drivers` | no assignment at time 0 | `Z` once at time 0: the second record was the transaction, not the driver |
| `t34_res_txn_zero` against `t34_res_two_drv0` | one driver assigning `'Z'` at time 0 | `Z` once: a single driver's unchanged assignment is not recorded |
| `t34_res_3drv____` against `t24_two_drivers` | a third driver assigning `'Z'` at 80 ns | `Z` twice at time 0, then `1`, `X`, and `X` again at 80 ns |
| `//hdl/uart:sim` against the corpus | a transmitter, a receiver and a FIFO under one top, 59 objects in 8 arenas and 17 pages | nothing new: every value reads back, the VCD agrees, and the bench's `errors` holds 0 |
| `t35_sv_ust_arr__` against `t13_sv_struct_ar` | the struct of `m [0:1]` unpacked | 128 bits in four pairs, `m[1]` in pairs 0 and 1, `b` below `a` |
| `t35_sv_ust_tdef_` against `t35_sv_ust_arr__` | `typedef rec_t arr_t [0:1]` | a second alias with `(0 to 1)`; the same records |
| `t35_sv_pst_tdef_` against `t13_sv_struct_ar` | `typedef s_t arr_t [0:1]` over the packed struct | the same alias; one 10 bit word as before |
| `t35_sv_ust_fld__` against `t35_sv_ust_arr__` | `m[1].b` written alone | 8 bytes at pair 0 |
| `t35_sv_st_uarr__` against `t35_sv_ust_fld__` | an unpacked array field, `s.v[1]` written | 64 bits; `v` packed into pair 0; the pair rewritten whole |
| `t35_sv_ust_nest_` against `t35_sv_ust_fld__` | `rec_t r` as a field, `s.r.b` written | 96 bits, `r.b` in pair 0; 8 bytes at pair 0 |
| `//hdl/serv:sim` against the corpus | a RISC-V system of 23 module instances under one top, 968 objects in 281 scopes, `-generic_top memfile=...`, 3.3 ms of the `hello_uart` program | the bit offset of a port bound to a slice of a net, `t9_port_slice` holding VHDL bytes; the top named `tb(memfile="...")`; 2965811 records of the value held; a RAM loaded by `$readmemh` over nine arenas |
| `t36_v_nb_clk_lit` against `t17_v_reg_same` | `s <= 0` in `always @(posedge clk)` not in `initial` | three records against two: a record on the block's first run |
| `t36_v_nb_clk_tog` against `t36_v_nb_clk_lit` | `t <= ~t` in the same block | `s` records on every run: `t` is an operand of the block that had an event |
| `t36_v_nb_clk_150` against `t36_v_nb_clk_tog` | `t` toggled at 150 ns not every cycle | nothing at 125 ns: the event must precede the run |
| `t36_v_bl_clk_lit` against `t36_v_nb_clk_lit` | `s = 0`, blocking | two records: the rule is nonblocking only |
| `t36_v_nb_evt_lit` against `t36_v_nb_clk_lit` | `always @(a) s <= 1'b0`, `a` toggling at 30 and 60 ns | two records: the runs at 30 and 60 ns record nothing; the rule holds for the edge clocked block |
| `t36_v_nb_clk_rd_` against `t36_v_nb_clk_tog` | `t` read, not written, by the block | `t` unchanged: readers do not record |
| `t36_v_nb_clk_net` against `t36_v_nb_clk_tog` | `t` a net driven by `assign` | `s` two records: a same value record of a net is not an event |
| `t36_v_hier_p_nba` against `t36_v_nb_clk_lit` | the block in a child, its target an output port | `X`, `X`, `0` on the port: the same rule through a port |
| `t36_v_net_copy__` against `t36_v_net_mux_w_` | `wire c = w` added | seven records of `w` against three: every evaluation, once read |
| `t36_v_net_rd_not` against `t36_v_net_copy__` | `~w` | the same seven; the operator is not what matters |
| `t36_v_net_rd_alw` against `t36_v_net_copy__` | `always @(w)` reads it | the same seven; any reader counts |
| `t36_v_net_2drv__` against `t36_v_net_mux_w_` | a second identical `assign`, no reader | five records: two drivers count as a driver and a reader do |
| `t36_v_hier_int__` against `t36_v_hier_mux__` | `wire i` in the child, `assign w = i` to the port | seven records of `i` against three of the port: the port is the reader |
| `t36_v_hier_i_noc` against `t36_v_hier_int__` | `i` not driving the port | three: unread, changes only |
| `t37_v_port_slc__` against `t9_port_slice` | Verilog, `v[5:2]` bound to a port | offset `2` in bits, one handle; VHDL counts bytes |
| `t37_v_port_pair1` against `t37_v_port_slc__` | `v[39:34]` of 40 bits | offset `34`; the port's bits in pair 1 |
| `t37_v_port_span_` against `t37_v_port_pair1` | `v[35:28]` | offset `28`; the value spans pairs 0 and 1 |
| `t37_v_port_reg__` against `t37_v_port_slc__` | the actual a `reg`, not a `wire` | a handle of its own, offset 0, five records |
| `t38_v_rmh_4w____` against `t38_v_mem4w32___` | `$readmemh` in place of four element writes | the same six records: one 8 byte element write per line |
| `t38_v_rmh_desc__` against `t38_v_rmh_4w____` | `m [3:0]` | the first line in `m[0]`, the bottom pair: the lowest address first |
| `t38_v_rmh_twice_` against `t38_v_rmh_4w____` | the file loaded twice | the same six records: the second load writes the values held |
| `t38_v_mem512____` against `t38_v_rmh_4w____` | 512 elements, 4096 record bytes | 28 chunks over three arenas; `m[511]` written at the handle in arena 0 |
| `//hdl/potato:sim` against the corpus | a RISC-V processor in VHDL, 144 scopes, 557 objects, two 32768 byte memories, string generics from `-generic_top`, `textio` loading | the rest of a chunked value chunked again; the range of a string generic set on the command line; the range of an unconstrained generic |
| `t39_vec30023____` against `t39_vec30022____` | a rest of 275 bytes, not 274 | the same: one record for the rest either way |
| `t39_vec20121____` against `t39_vec20120____` | a rest of 276 bytes, not 275 | four records of 69 for the rest: the threshold for a rest is 276 |
| `t39_vec22347____` against `t39_vec22348____` | a rest of 295, not 296 | 73 three times then 76, against 74 four times: the rest's own rest stays whole below 276 |
| `t39_mem4096_____` against `t39_vec32768____` | a loop over elements 256 to 4095 of a 4096 byte array | one partial write of 30720 bytes at the handle plus 2048, in 205 chunks of 149 and a rest of 175 |
| `t40_gen_cons____` against `t40_gen_uncons__` | `std_ulogic_vector(3 downto 0)` over `std_ulogic_vector` for a generic with default `x"A"` | the declared range `(3 downto 0)` over `(0 to 3)`; nothing else |
| `t40_gen_str_top_` against `t40_gen_uncons__` | `--generic_top ks=hello` | `ks` declared as 5 bytes `(1 to 5)`, its record `hello`; the handle of `kv` moves from `0xdf4` to `0xdf8` |
| `t41_uvec________` against `t1_vec8` | `array (natural range <>) of std_ulogic` declared in the architecture | an entry `vec_t` of the `STD_ULOGIC_VECTOR` shape; the same records |
| `t41_neg_vec_____` against `t41_uvec________` | `(integer range <>)` constrained `(3 downto -4)` | `INTEGER` as the index word; `(3 downto -4)` in the declaration, the bound signed |
| `t41_arr_subtype_` against `t41_neg_vec_____` | the signal declared with `subtype byte_t is vec_t(3 downto -4)` | an entry `byte_t` with the range as its triple; no `vec_t` entry |
| `t41_sfixed______` against `t41_neg_vec_____` | `sfixed(3 downto -4)` from `ieee.fixed_pkg` | an entry `sfixed` in lower case of the same shape; `00011000` for 1.5 |
| `t41_float32_____` against `t41_sfixed______` | `float32` from `ieee.float_pkg` | a constrained entry `float32` `(8 downto -23)`; the IEEE 754 bits one per byte |
| `t42_rec_uncons__` against `t2_record2______` | `bravo : std_ulogic_vector` constrained at the signal | the field triple `(0, 0, -2)`; the bounds only in the declaration; the value misread as one element until the reader used the declaration |
| `t42_rec_subtype_` against `t42_rec_uncons__` | `subtype b8_t is bundle_t(bravo(7 downto 0))` | the entry renamed `b8_t`, its field still unconstrained |
| `t42_rec_two_cons` against `t42_rec_uncons__` | a second signal `bundle_t(bravo(3 downto 0))` | one entry for both; the declarations differ in range and size |
| `t42_rec_two_unc_` against `t42_rec_mix_unc_` | both fields unconstrained, or the first constrained in the record | the same declaration list `(3 downto 0) (7 downto 0)` |
| `t42_rec_unc_nest` against `t42_rec_uncons__` | the unconstrained field inside an inner record | `(0, 0, -2)` on the outer field's flattened list; the inner record padded to 8 bytes |
| `t42_rec_unc_2dim` against `t20_rec_2dim____` | an unconstrained two dimensional field | two `(0, 0, -2)` triples; `(0 to 1) (0 to 2)` in the declaration |
| `t42_rec_unc_arr_` against `t42_rec_uncons__` | `array (0 to 1) of bundle_t` constrained `arr_t(open)(bravo(3 downto 0))` | the array entry's own index written `(0, 0, -2)` |
| `t42_arr_unc_elem` against `t2_array2d______` | `array (0 to 1) of std_ulogic_vector` | both triples `(0, 0, -2)` where the constrained element gave `(0 to 3) (7 downto 0)` |
| `t42_arr_unc_both` against `t42_arr_unc_elem` | `array (natural range <>) of std_ulogic_vector` | the index word `NATURAL` for `INTEGER`; nothing else |
| `t42_gen_pkg_____` against `t1_vec8_________` | the vector's subtype from a generic package instance | a scope `gp` beside `tb`; a constrained `word_t` entry; no `n` |
| `t42_pkg_subtype_` against `t42_gen_pkg_____` | the same subtype in a plain package | the scope name and the paths; nothing else |
| `t42_gen_pkg_two_` against `t42_gen_pkg_____` | a second instance `gp4` | a second scope `gp`, a second unit, a second `word_t` |
| `t42_gen_pkg_cons` against `t42_gen_pkg_____` | `constant width : natural := n` in the package | an object `gp.width`, not logged |
| `t42_gen_type____` against `t4_gen_explicit_` | `generic (type data_t; ...)` mapped to `integer` | the entry named `data_t`; no `INTEGER`; the type generic has no declaration |
| `t42_gen_type_enu` against `t42_gen_type____` | `data_t => std_ulogic` | `enum "data_t"` with the nine literals; 1 byte |
| `t43_port_uncons_` against `t8_port_vec8____` | `a : in std_ulogic_vector` bound to an eight bit signal | the port declared `(7 downto 0)`, 8 bytes, the actual's handle |
| `t43_port_unc_two` against `t43_port_unc_sam` | the second instance bound to four bits instead of eight | a second `child(sim)` unit and a second pair of declarations |
| `t43_port_unc_asc` against `t43_port_uncons_` | the actual `(0 to 7)` | the port declared `(0 to 7)` |
| `t43_port_unc_out` against `t43_port_uncons_` | `a : out std_ulogic_vector` driven by the child | the same declaration with `port out`; the actual's handle |
| `t43_port_unc_rec` against `t43_port_uncons_` | a port of `bundle_t` with `bravo : std_ulogic_vector` | `(7 downto 0)` and 16 bytes on the port; the package scope `bundle_pkg` |
| `t44_time_5ms____` against `t1_bit_one_edge_` | the one edge at `5 ms`, 5000000000 ps | the record time, the page's `t1` and the end time above 2^32; nothing else moves |
| `t44_time_5s_____` against `t44_time_5ms____` | the edge at `5 sec` | 5000000000000 in the same three places |
| `t44_time_late___` against `t44_time_5ms____` | an edge at 1 ns before the one at 5 ms | `t0` 0 and `t1` 5000001000 |
| `t44_v_time_5ms__` against `t11_v_bit_edge__` | the Verilog `#5000000` under `1ns / 1ps` | 5000000000 as in VHDL |
| `t44_str_sig_____` against `t2_character____` | `s : string(1 to 5)` | the `STRING` entry over `character` by `POSITIVE`; `(1 to 5)` and 5 bytes on the declaration; one byte per character |
| `t44_str_sig_3to7` against `t44_str_sig_____` | `string(3 to 7)` | `(3 to 7)` on the declaration |
| `t44_str_var_____` against `t6_var_int______` | a `string(1 to 5)` process variable | a 5 byte declaration with `(1 to 5)` and no record |
| `t45_log_late____` against `t45_log_base____` | `log_wave -recursive *` after `run 10 ns` | the first record at 10000 holding the value held, `t0` 10000; the VCD backdates it to `#0` |
| `t45_run_steps___` against `t45_log_base____` | `run 10 ns`, `run 10 ns`, `run -all` | no difference |
| `t45_log_twice___` against `t45_log_base____` | a second `log_wave -recursive *` at 10 ns | one record at 10000 holding the value held |
| `t45_log_one_____` against `t45_log_base____` | `log_wave /tb/s` with a second signal `u` | `u` declared, listed, marked not logged; its arena unused |
| `t45_log_dut_____` against `t45_log_one_____` | `log_wave -recursive /tb/dut` with a signal in `tb` | `tb.s` not logged, `tb.dut.c` as usual, ranges `[1 1]` |
| `t45_log_dut_late` against `t45_log_dut_____` | the child log after `run 10 ns` | the child's first record at 10000 |
| `t45_two_tops____` against `t45_log_base____` | `--top corpus.tb2 --top corpus.tb` | two root children in option order; the default script logs the first top only |
| `t45_two_tops_all` against `t45_two_tops____` | `log_wave -recursive /tb2` and `/tb` | both tops record; `tb.s` in arena 1 at key `0x58`; ranges `[0 1]` |
| `t46_sig_1000____` against `t45_log_base____` | 1000 `std_ulogic` signals, two driven | undriven signals `0xc0` apart, driven `0xf0`; `0x3d9e8` of handle space is `0x1088 + 2 * 0x148 + 998 * 0xf8` |
| `t46_gen_70000___` against `t46_sig_1000____` | a for generate of 70000 iterations, a signal, an index and a process each | 140004 scopes and 140000 objects in whole 32 bit words; 49 MB; the 70000 indexes after the last signal, `0x118` apart |
| `t46_v_gen_70000_` against `t46_gen_70000___` | the same in Verilog | 70000 objects in one scope, `0xc0` apart, no stride for the `initial` writer |
| `t46_deep_100____` against `t8_gen_if_______` | an entity instantiating itself under an if generate, 100 levels | 306 scopes, paths of 101 names; 101 generics after the last signal, `0x140` apart |
| `t46_drv_2_next__` against `t24_two_drivers_` | a driven `std_ulogic` after the two driver `std_logic` | the next handle `0x140` on, `0xc0 + 0x80`; the same handle space |
| `t46_drv_3_next__` against `t46_drv_2_next__` | a third driver | `0x178`, `0xc0 + 0xb8`; the handle space of `t34_res_3drv____` |
| `t46_v_wire_4asg_` against `t19_v_wire_3drv_` | a fourth `assign` | the next handle `0xf0` on, as with three |
| `t47_use_numstd__` against `t1_bit_one_edge_` | `use ieee.numeric_std.all`, unused | `0x1f8` of handle space and two file table entries; the tier 2 "type cost" was the clause |
| `t47_use_1164_bit` against `t47_use_none____` | `library ieee` and `use ieee.std_logic_1164.all` over a `bit` signal | `0x604`; a `bit` and a `std_ulogic` cost the same |
| `t47_use_lib_only` and `t47_use_one_name` against `t47_use_none____` | the library clause alone; a clause naming one item | nothing; the price of `.all` |
| `t47_use_pkg_emp_`, `t47_use_pkg_typ_`, `t47_use_pkg_4arr`, `t47_use_pkg_fn2_`, `t47_use_pkg_pr2_` against `t1_bit_one_edge_` | a package of the design that is empty, or holds a subtype, four types, two functions, two procedures | `0x80` each: the scope and unit; types and subprograms are free |
| `t47_use_pkg_two_` and `t47_use_pkg_nul_` against `t47_use_pkg_emp_` | two constants; two null range constants | 8 and `0x20`: the rounded storage; the constants are unlogged objects of the package scope |
| `t41_neg_arr_type` and `t24_null_range__` against `t2_array2d______`, reread | a negative bound and a null range | the range record: 64 bit bound pairs, the direction word, and a span that is not the element count |
| a sweep of the words at `40` and `44` of the instance records over the corpus and `//hdl/serv:sim` | nothing; the same files | `40` is 0 to 29 on the ports of `serv` in port list order, and `44` holds stack and heap shaped values that differ between objects and runs |
| `t48_v_port_nansi` against `t48_v_port_pos4_` | a non ANSI header with the ports declared in reverse order | `40` counts the port list; the objects follow the declarations |
| `t48_v_port_rev__` and `t48_v_port_posit` against `t48_v_port_pos4_` | the connections by name in reverse order; by position | `40` unchanged; the input port handles follow the connection order |
| `t48_v_port_open_` against `t48_v_port_pos4_` | output `d` unconnected | `40` unchanged; `d` gets a handle of its own and the parent's wire one `Z` record |
| a sweep of the word at `28` of the instance records over the corpus | nothing; the same files | 1, 3, 4 and 6 beside the 0 and 2 the reader knew, on the mixed language and subprogram cases |
| `t49_sub_rec_loc_` and `t49_sub_int_arr_` against `t23_sub_sizes___` | a record local; an integer array local | 4, as the vector local: composite, not `std_ulogic_vector` |
| `t49_sub_var_prm_` and `t49_sub_vec_prm_` against `t23_sub_sizes___` | a `variable` class scalar parameter; a `constant` class vector parameter | 3 and 4: the type decides, not the class or mode |
| `t49_sub_sig_in__` and `t49_sub_sig_vec_` against `t23_sub_sig_prm_` | an `in` signal parameter; a vector signal parameter | 6 either way |
| `t49_sub_str_prm_` against `t23_sub_sig_prm_` | an unconstrained `string` parameter | no declaration, no object, no handle space |
| `t49_mix_2port___` against `t21_mix_v_in_vh_` | an output port at the boundary | 1 on both ports, positions 0 and 1; `U`, `X`, `0` on the VHDL signal the assign drives |
| `t49_mix_deep____` against `t49_mix_2port___` | a VHDL leaf under the Verilog child | 1 on the leaf's ports too; the leaf's input holds no `U` and its output no `X`; every boundary port has its own handle |
| `t50_sub_var_vec_` and `t50_sub_var_rec_` against `t49_sub_var_prm_` | an `inout` `variable` vector; record parameter | 4 on both, as on a `constant` vector |
| `t50_sub_in_var__` against `t49_sub_var_prm_` | an `in` `variable` scalar parameter beside a signal parameter | 3; the signal parameter on `0xd8` after the 4 byte scalar on `0xd0`; the signal parameter listed first |
| `t50_sub_acc_loc_` and `t50_sub_str_loc_` against `t23_sub_sig_prm_` | an access type local; a `string(1 to 4)` local | 3; 4; both on `0x110` after the signal parameter |
| `t50_sub_sig_rec_` against `t49_sub_sig_vec_` | a record signal parameter | 6 |
| `t50_sub_ivec_prm` against `t49_sub_str_prm_` | an unconstrained `integer_vector` parameter | absent as the `string` was; the signal parameter on `0xe8` in both |
| `t50_sub_func_prm` against `t49_sub_vec_prm_` | the vector parameter on a function | 4 on `0x40`, the scalar local on `0x58` |
| `t50_ord_const1st` against `t5_tr1000_______` | the constant declared above the signal | the signal still listed first |
| `t50_ord_proc_con` against `t6_var_int______` | a constant above the variable in a process | source order among the data objects; the constant has a record at 0 |
| `t50_ord_two_sig_` against `t5_tr1000_______` | signals `z`, `a`, `s` | source order, handles in that order |
| `t51_sv_task_auto` against `t51_sv_task_stat` | `task automatic` | no declarations; the handle space `0xbb4` for `0xc14` |
| `t51_sv_task_ref_`, `t51_sv_func_auto` against `t51_sv_task_auto` | a `ref` argument; an automatic function | nothing either |
| `t51_sv_task_stvr` against `t51_sv_task_auto` | a `static` local in the automatic task | the local listed, the argument not |
| `t51_sv_task_out_`, `t51_sv_task_inou` against `t51_sv_task_stat` | an `output`; an `inout` argument | the modes; 0 in the word at `40` on both arguments |
| `t51_sub_loop_idx` against `t23_sub_sig_prm_` | a `for` loop in the procedure | no index in the file |
| `t51_sub_file_prm` against `t23_sub_sig_prm_` | a `file` parameter and a file object | the parameter absent, `q` on `0xd8`; the file object a size 0 variable |
| `t51_sub_pkg_proc` against `t23_sub_sig_prm_` | the procedure in a package | the scope `pk.drive` under `pk`; the same frame offsets |
| `t52_var_real____`, `t52_var_time____`, `t52_var_rec_____`, `t52_var_vec4____`, `t52_var_str4____`, `t52_var_vec8____`, `t52_var_arr4____`, `t52_var_sul_____`, `t52_var_bool____` against `t52_var_int_____` | the type of the first of two process variables | strides 8, 8, 8, `0x14`, `0x14`, `0x18`, `0x20`, 4, 4: the size, 16 more for an array, the next on its alignment |
| `t52_con_int_____`, `t52_con_real____`, `t52_con_vec8____`, `t52_gen_int_____`, `t52_gen_vec8____` against the `t52_var_` case of the same type | constants; generics | the same strides and handle space |
| `t52_inst2_proc__`, `t52_inst2_sig___`, `t52_inst2_sigprc` against `t52_inst2_empty_` | a process; an undriven signal; a signal driven by a process in each of two children | `k` strides `0xc0`, `0x68`, `0x118` for `0x30`; the handle space up by twice `0x90`, twice `0xf8`, twice `0x1d8` |
| `t52_gi2_proc____`, `t52_gi2_sig_____`, `t52_gi2_sigprc__` against the `t52_inst2_` case of the same body | a generate iteration for an instance | the same `i` strides; `0x18` to `0x20` less handle space |
| `t52_gi2_empty___` against `t52_gi2_proc____` | the iteration empty | no iteration scope and no index, only `tb.g` |
| `t53_inst1_empty_`, `t53_inst3_empty_` against `t52_inst2_empty_` | one child; three children | the first `k` at `0xeb8` in all three; `0x30` per child |
| `t53_inst2_2gen__`, `t53_inst2_const_` against `t52_inst2_empty_` | a second generic; an architecture constant | 4 past `k`; the stride `0x30` still |
| `t53_inst2_2proc_`, `t53_inst2_var___` against `t52_inst2_proc__` | a second process; a variable in the process | `0x150`; `0xc0` with the variable 4 past `k` |
| `t53_inst2_2sig__` against `t52_inst2_sig___` | a second undriven signal | `0xa0` |
| `t53_inst2_conc__` against `t52_inst2_sigprc` | the driver a concurrent assignment | `0x118` still; `c` `0x110` apart in the first region for `0xf0` |
| `t53_inst2_2drv__` against `t52_inst2_sigprc` | a `std_logic` with two driving processes | `0x1c8`; `c` `0x140` apart; three records, one repeating the value |
| `t53_inst2_port__`, `t53_inst2_portop` against `t52_inst2_empty_` | an input port connected; open | `0x68` both; the connected port shares `0x768` |
| `t53_inst2_nest__` against `t52_inst2_empty_` | an empty grandchild with a generic | `0x60`; `d0`, `d0.e`, `d1`, `d1.e` at `0x30` each |
| `t53_ifgen_inst__`, `t53_blk_inst____` against `t52_inst2_empty_` | each child under an if generate; under a block | `k` `0x30` apart, `0x50` past `0xeb8`; the handle space `0x58` up |
| `t54_none_noenv__` against `t54_none_nosig__` | no `std.env.stop`, no library, no signal | the variable at `0x738` for `0x810`; `standard` alone in the file |
| `t54_noenv_sig___` against `t54_none_noenv__` | a `bit` signal | `s` at `0x768`, the variable at `0x860` |
| `t54_lib_none_var` against `t54_none_nosig__` | a `bit` signal, with `env` | the variable at `0x938`, `0x128` on as without `env` |
| `t54_1164_noenv__`, `t54_lib_1164_bit` against `t54_lib_none_var` | `std_logic_1164`; and `env` | the variable at `0xda0`, then `0xde0` |
| `t54_nosig_var___` against `t54_lib_1164_bit` | no signal | the variable at `0xcb8`, `0x738` plus `0x580` |
| `t54_lib_numstd_v`, `t54_lib_mathrl_v` against `t52_var_int_____` | `numeric_std`; `math_real` | `0xf8` and `0x308` on; `0x1f8` and `0x400` of handle space |
| `t54_pkg_con_var_`, `t54_pkg_2con_var` against `t52_var_int_____` | a package with one constant; two | `0x30` on both; the constants at `0xd40` and `0xd44` |
| `t54_pkg_use_var_`, `t54_pkg_unused__` against `t54_pkg_con_var_` | a use clause and no reference; neither | the same file; the package absent |
| `t55_sub_con_loc_`, `t55_sub_con_arr_` against `t55_sub_loop____` | a scalar constant local; an array one | kind `0x14`, class 3 and 4; 20 and 24 bytes of the frame |
| `t55_sub_con_nori`, `t55_sub_2con____`, `t55_sub_con_real` against `t55_sub_con_loc_` | the constant unread by initialisers; two constants; a real one | 20 bytes still; 20 each; 24 |
| `t55_sub_var_init` against `t55_sub_con_loc_` | a variable in the constant's place | 4 bytes |
| `t55_sub_loop____` against `t50_sub_func_prm` | a `for` loop in a function | no index, no frame room |
| `t55_sub_alias___`, `t55_sub_file_loc`, `t55_sub_prot_loc` against `t51_sub_loop_idx` | an alias; a file; a protected variable, each local to the procedure | absent; the next local on `0x110`, `0x138`, `0x11c` |
| `t55_sub_prot_2__`, `t55_sub_prot_3__` against `t55_sub_prot_loc` | two, three protected locals | `0x12c`, `0x13c` |
| `t55_sub_nested__` against `t50_sub_func_prm` | a function inside the function | `tb.g` and `tb.f.g` on one unit, both from `0x40` |
| `t55_sub_prot_typ`, `t55_prot_shared_` against `t55_sub_prot_loc` | the type without a variable; a shared variable of it | no method scopes; the same four, `0x100` more handle space |
| `t55_prot_arch_pr`, `t55_prot_arch_2p` against `t55_prot_shared_` | the process calling; two processes | the second pair under `tb` after the processes |
| `t55_prot_pkg____`, `t55_prot_pkg_prc` against `t55_prot_shared_` | the type in a package; the process calling | `pk.bump`, `pk.get`, and `tb.p.bump`, `tb.p.get` |
| `t55_prot_pkg_2p_`, `t55_prot_pkg_2pl` against `t55_prot_pkg____` | two processes, `p2` calling first; last | the second pair under `tb.p2` both times |
| `t55_prot_pkg_sv_` against `t55_prot_pkg_prc` | the shared variable in the package, behind package subprograms | `pk.bump` and `pk.get` twice under `pk`; the variable absent |
| `t56_typ_arr_unus`, `t56_typ_arr_loc_` against `t56_typ_none____` | an array type; a local of it | nothing; `0x10` |
| `t56_typ_arr_noin`, `t56_sub_arr_dyni` against `t56_typ_arr_loc_` | no initialiser; one from a parameter | `0x10` still; nothing |
| `t56_typ_arr_lit_`, `t56_typ_arr_dyn_` against `t56_typ_arr_noin` | an aggregate literal in the body; an aggregate of a variable | `0x10` more; nothing more |
| `t56_typ_arr_2loc`, `t56_typ_arr_2typ`, `t56_typ_arr8_loc` against `t56_typ_arr_loc_` | two locals; of two types; eight elements | `0x20` each |
| `t56_typ_vec4_loc`, `t56_typ_vec4_sub`, `t56_typ_vec4_2lc`, `t56_typ_vec_noin` against `t56_typ_none____` | a vector local; of a subtype; two; one with a body aggregate | `4`, `4`, `8`, `8` |
| `t56_typ_int_rng_`, `t56_typ_enum_loc` against `t56_typ_none____` | a ranged integer local; an enumeration local | nothing |
| `t56_typ_rec_loc_`, `t56_typ_rec_noin`, `t56_typ_rec_prm_`, `t56_typ_rec_arr_` against `t56_typ_none____` | a record local; uninitialised; from the parameter; with a vector field | `0xc`, `0xc`, nothing, `0xc` |
| `t56_sub_rec_1int`, `t56_sub_rec_2int`, `t56_sub_rec_3int`, `t56_sub_rec_4int`, `t56_sub_rec_2rl_` against each other | records of one to four integers, and of two reals | `0xc`, `0xc`, `0x14`, `0x14`, `0x14`; 8 or 16 bytes declared |
| `t56_typ_arr_prc_`, `t56_prc_arr_noin`, `t56_prc_vec_init`, `t56_prc_vec_noin` against `t52_var_arr4____` and `t52_var_vec4____` | process variables, initialised or not | the tier 52 strides and handle space |
| `t57_log_all_____` against `t7_gen_for______` | `log_wave -recursive *` over a design with every object kind | everything but the process and shared variables logged; ranges `[0 3] [5 8] [10 10]` |
| `t57_log_none____`, `t57_log_var_____`, `t57_log_var_all_`, `t57_log_shv_____`, `t57_log_bit_____`, `t57_log_rec_fld_`, `t57_log_gen_____` against `t57_log_all_____` | no `log_wave`; one naming a process variable, under `typical` and `all`; a shared variable; one bit; a record field; the generate statement | nothing logged; 4007 bytes; all slots `0` |
| `t57_log_con_____`, `t57_log_loop____`, `t57_log_gen_idx_` against `t57_log_all_____` | `log_wave` naming a constant, a loop index, a generate index | that object alone, one record at 0 |
| `t57_log_slice___`, `t57_log_rec_____`, `t57_log_gen_sig_` against `t57_log_all_____` | naming a slice of the vector, the record, one iteration's signal | the whole vector; the record; that signal |
| `t57_log_gen_it__`, `t57_log_proc____`, `t57_log_top_____` against `t57_log_all_____` | naming one iteration, the process, the top without `-recursive` | the scope's own data objects, variables excluded: `gs` and `i`; `k`; `s`, `v`, `r`, `c` |
| `t58_sv_log_all__` against `t57_log_all_____` | the tier 57 script over a SystemVerilog design with every object kind | all thirteen objects logged, `[0 12]` |
| `t58_sv_log_none_`, `t58_sv_log_bit__`, `t58_sv_log_mem_e`, `t58_sv_log_st_fl`, `t58_sv_log_gen__` against `t58_sv_log_all__` | no `log_wave`; one naming a bit, a memory element, a struct field, a generate block | nothing logged; 3402 bytes; all slots `0` |
| `t58_sv_log_int__`, `t58_sv_log_real_`, `t58_sv_log_blkv_`, `t58_sv_log_tsk_a`, `t58_sv_log_tsk_l` against `t57_log_var_____` | naming a Verilog variable: of the module, of a named block, of a static task | that variable logged, where the VHDL variable was refused |
| `t58_sv_log_slc__`, `t58_sv_log_mem__`, `t58_sv_log_st___`, `t58_sv_log_prm__`, `t58_sv_log_lprm_` against `t58_sv_log_all__` | naming a slice, the memory, the struct, the parameter, the localparam | the whole vector; that object alone |
| `t58_sv_log_blk__`, `t58_sv_log_tsk__`, `t58_sv_log_top__` against `t58_sv_log_all__` | naming the named block, the task, the module without `-recursive` | the scope's own objects: `bv`; `x` and `tmp`; the ten of `tb` |
| `t58_sv_log_gen_w` against `t58_sv_log_gen__` | `[get_objects -regexp ...]` for the generate wire, against its path | the wire logged; nothing |
| `t59_frc_s_const_` against `t59_frc_none____` | `add_force /tb/s 1` before the run, against no force | records `0`, `1` at time 0 and `1` at 20 ns, the value held at the driver's `'0'`; nothing else in the dump differs but the noise words |
| `t59_frc_v_const_`, `t59_frc_v_bit___` against `t59_frc_none____` | the vector forced whole, and one bit of it | `1111` at 0, 10 and 20 ns; `1000`, `1119`, `1010` once each |
| `t59_frc_s_cancel`, `t59_frc_s_pat___` against `t59_frc_s_const_` | `-cancel_after 5ns`; a `{0 0ns} {1 2ns} -repeat_every 4ns` pattern | the return recorded once at 5 ns; every step, and two records where the pattern and the driver write together |
| `t59_frc_mid_____`, `t59_frc_mid_same` against `t59_frc_none____` | a force after `run 15 ns` of `0`, and of the value held `1` | `0` at 15 ns and nothing at 20; `1` at 15 ns and the held `1` at 20 |
| `t59_frc_release_`, `t59_frc_rel_same`, `t59_frc_twice___` against `t59_frc_s_const_` | `remove_forces` at 15 ns on a force of `0`, on a force of `1` the driver agreed with, and a second `add_force` | the new value twice at 15 ns in each; `0` after the release of `1` |
| `t59_frc_deposit_`, `t59_frc_dep_mid_`, `t59_frc_dep_same` against `t59_frc_s_const_` | `set_value` before the run, after `run 15 ns`, and of the value held | one record at the deposit, the driver's next different value recorded: a deposit does not hold |
| `t59_frc_sv_force`, `t59_frc_sv_frc_0` against `t59_frc_sv_none_` | a `force s = 1'b1` and a `force s = 1'b0` at 5 ns in the source, released at 15 ns | the forced value at 5 ns, held or not; nothing at the driver's write while forced, nothing at the release; handle space `0x9fc` against `0x9b4` |
| `t59_frc_sv_long_`, `t59_frc_sv_norel`, `t59_frc_sv_relon` against `t59_frc_sv_force` | a release at 25 ns, no release, a release with no force | the driver's `0` at 20 ns unrecorded under a force; the `0x48` is the `force` statement's, the `release` costs nothing |
| `t59_frc_sv_tcl__` against `t59_frc_s_const_` | `add_force` on a SystemVerilog `logic` | time 0 once, the held value at 10 ns, nothing at 20; the VCD holds time 0 only |
| `t60_dbg_none____` against `t11_sv_bit______` | `xelab -debug all` | nothing in the type table, the declarations, the objects or the records of a `logic`; the file differs in noise words only |
| `t60_dbg_vec_____`, `t60_dbg_int_____`, `t60_dbg_real____`, `t60_dbg_struct__`, `t60_dbg_mem_____` against `t60_dbg_none____` | an ordinary second variable under `-debug all` | the same entries, declarations and records as the tier 11 counterparts, the array trailer `-99` |
| `t60_dbg_str_____` against `t60_dbg_none____` | a `string` under `-debug all` | type kind `0x18`, a 32 bit declaration of class 0, a logged handle at `0x828` with one zero record at time 0 and nothing for the write at 50 ns; handle space `0x98` as under typical |
| `t60_dbg_queue___`, `t60_dbg_dynarr__`, `t60_dbg_assoc___` against `t60_dbg_int_____` | a queue, a dynamic array and an associative array under `-debug all` | kinds `0x14`, `0x13` and `0x15`; a 32 bit declaration of class 3 with `(0 to 0)`, an unlogged handle, an arena never written; `int` closes with `0` instead of `-99` |
| `t60_dbg_assoc_i_` against `t60_dbg_assoc___` | an `int` key for a `string` key | the key type index changes and the word before it goes from `2` to `3`; no `string` entry |
| `t60_dbg_class___` against `t60_dbg_int_____` | a class with one `int` field and a handle of it | kind `0x17` before the field's predefined entries, parent `-1`, id `0`; the handle is declared as 32 bits of class 0 and logged with one zero record; `int` closes with `1` |
| `t60_dbg_class_2_`, `t60_dbg_class_d_` against `t60_dbg_class___` | a second `logic [3:0]` field; a parent class | the field carries its range triple and a `0`, the unnamed vector entry closes with `1` and `int` with `2`; the parent comes first with id `1`, the derived class names it by index and lists its own field only |
| `t60_dbg_class_2h`, `t60_dbg_class_n_` against `t60_dbg_class___` | two handles of one class; construction at time 0 | one class entry, the second handle at `0x8e8` with its own zero record; the zeros do not change with `new` |
| `t60_dbg_str_log_`, `t60_dbg_q_log___` against `t60_dbg_str_____`, `t60_dbg_queue___` | `log_wave` naming the string, the queue | the string's zero record is written once per `log_wave` that names it; the queue is refused with a warning and the file is unchanged |
| `t61_num_a_then_q` against `t61_num_q_then_a` | an associative array and a queue, declared in either order | the numbers follow the declaration order, `2` and `3` against `1` and `3`; the associative array takes two |
| `t61_num_ai_thn_q` against `t61_num_a_then_q` | an `int` key | the associative array holds `3` and the queue after it `4`; an `int` key takes one number more |
| `t61_num_q_q_____`, `t61_num_q_cls___`, `t61_num_q_str___`, `t61_num_q_vec___`, `t61_num_q_byte__` against `t60_dbg_queue___` | the element of the queue | the element is numbered before the queue, recursively; a string element takes no number and the queue holds `0`; the declaration takes the element's value class |
| `t61_num_cls_rev_` against `t60_dbg_class_2_` | the fields in reverse order | the numbers swap: the field declared last holds `1` |
| `t61_num_cls_byte`, `t61_num_cls_byti`, `t61_num_cls_long`, `t61_num_cls_2int`, `t61_num_cls_2vec`, `t61_num_cls_ibv_` against `t60_dbg_class___` | a second integral or vector field | `int`, `byte` and `longint` share one number and the unnamed vector has another; an array is numbered by its element |
| `t61_num_cls_3f__`, `t61_num_cls_str_` against `t60_dbg_class_2_`, `t60_dbg_class___` | a `real` field, a `string` field | neither takes a number nor moves the others |
| `t61_num_cls_q___`, `t61_num_cls_cls_`, `t61_num_two_cls_` against `t60_dbg_class___` | a queue field, a handle field, a second class | the queue follows its element; a field's class follows the class as a parent does, before its own fields; the second class continues the count after the first class's fields |
| `t61_num_cls_int_` against `t60_dbg_class___` | an `int` variable beside the handle | the ordinary object is declared, logged and recorded as under typical, at the next handle |
| `t62_str_and_____` against `t62_str_wire____` | an `and` gate for the `assign` | `Forked11_1` for `NetRegassign11_1`; the same records |
| `t62_str_and_2___`, `t62_str_bufif_n_`, `t62_str_vec_pu__` against `t62_str_and_____`, `t62_str_bufif___`, `t62_str_pullup__` | two gates in one statement; a named gate; a pull array | two `Forked` scopes; the name unused; one scope for four pulls |
| `t62_str_gate_dly` against `t62_str_and_____` | `and #3` | `0` at 3 ns and `1` at 53 ns; `0x50` of handle space |
| `t62_str_bufif___`, `t62_str_nmos____` against `t62_str_and_____` | a `bufif1`, an `nmos` | `X`, `Z`, `Z`, then `1`: two writes at time 0 |
| `t62_str_pullup__`, `t62_str_pulldn__` against `t62_str_wire____` | a pull source for the `assign` | `Forked11_1`; `X` then `1` or `0`; the same handle space |
| `t62_str_strong__` against `t62_str_equal___` | strengths on two literal and `s` drivers | `1` for `X` at 50 ns; `X`, `0`, `0` for `X`, `0`; 8 more of handle space |
| `t62_str_weak____`, `t62_str_mixed___`, `t62_str_supply__`, `t62_str_wand____`, `t62_str_pu_drv__` against `t62_str_strong__` | the strengths moved, a `wand`, a pullup for the weak literal | the resolved value; three records at time 0 in each |
| `t62_str_tri_____`, `t62_str_uwire___` against `t62_str_wire____` | `tri`, `uwire` | kind `0x06` and `0x03`; nothing else |
| `t62_str_vec_2drv` against `t62_str_vec_1drv` | a second literal driver on a 4 bit net | four records per write, one per bit, for one |
| `t63_pdr_bit0____` against `t62_str_vec_1drv` | `assign v[0] = s` for a whole driver | `ZZZX`, `ZZZ0`, `ZZZ1`: `Z` on the undriven bits from the first record |
| `t63_pdr_bit3____`, `t63_pdr_slice___` against `t63_pdr_bit0____` | bit 3, the low nibble of 8 | `XZZZ`; `ZZZZXXXX`; the same handle space |
| `t63_pdr_w64_bit0`, `t63_pdr_w64_bit6`, `t63_pdr_w64_hi__` against `t63_pdr_bit0____` | a 64 bit net | the first record 16 bytes; each write 8 bytes at the pair driven; handle space 8 more |
| `t63_pdr_w64_all_` against `t63_pdr_w64_hi__` | the whole 64 bits driven | three 16 byte records; 16 more of handle space |
| `t63_pdr_2400_bit`, `t63_pdr_2400_hi_` against `t63_pdr_w64_bit0` | a 2400 bit net | the first record in six chunks; a write of 8 bytes at the handle, of 104 at byte 96 of chunk 4 |
| `t63_pdr_2400_all` against `t63_pdr_2400_hi_` | the whole 2400 bits driven | three chunked records; `0x1f8` more of handle space |
| `t63_pdr_two_bits` against `t63_pdr_bit0____` | a second partial driver `v[3] = ~s` | five records for three, each a whole pair with the other bit as it stands |
| `t63_pdr_concat__` against `t62_str_wire____` | `assign {a, b} = {s, ~s}` | two nets on two handles, three records each |
| `t63_pdr_port_bit` against `t63_pdr_bit0____` | a child output on `v[1]` | `tb.u.o` on the net's handle with offset 1 and position 1; the first record twice |
| `t63_pdr_port_slc`, `t63_pdr_port_hi_` against `t63_pdr_port_bit` | the output on `v[7:4]`, `v[63:32]` | offset 4, 32; the write at the handle plus 8 for the word |
| `t64_ord_src_rev_` against `t63_pdr_two_bits` | the two assigns in the other source order | `1ZZX` then `1ZZ0` for `XZZ0` then `1ZZ0`: source order at time 0 |
| `t64_ord_gen4____`, `t64_ord_gen_rev_` against `t63_pdr_two_bits` | four `assign v[i] = s` in a generate loop, up and down | `tb.NetRegassign11_1` to `_4`, no `g[i]` scope; one record per driver in loop order |
| `t64_ord_w64_two_`, `t64_ord_2400_two` against `t63_pdr_two_bits` | the two drivers in two pairs, in two chunks | 8 bytes at the handle and at the handle plus 8; at the handle and at byte 92 of chunk 5 |
| `t64_ord_self____`, `t64_ord_chain___` against `t63_pdr_two_bits` | `v[1] = v[0]`; `w[1] = v[0]` | six records, `ZZX0` twice; three on each net |
| `t64_ord_unp_elem`, `t64_ord_unp_whol` against `t63_pdr_bit0____` | a bit, an element of `wire [3:0] v [0:1]` | `(ZZZZ, ZXZZ)`, `(ZZZZ, XXXX)`: the element's pair, the other element `Z` |
| `t64_ord_two_kids` against `t63_pdr_port_bit` | a second child on `v[3]` from `~s` | three first records; `u0` first at both times; `tb.u1.o` position 0 |
| `t64_ord_two_same`, `t64_ord_gen_kids` against `t64_ord_two_kids` | both inputs on `s`; the children from a generate loop | `u1` first at 50 ns; `tb.g[1].u.o` position 0 |
| `t64_ord_pos_expr`, `t64_ord_pos_bit3` against `t64_ord_two_kids` | one child, on a scalar net from `~s`, on `v[3]` from `s` | position 1 on `o` in both: neither the expression nor the bit moves it |
| `t64_ord_two_nets`, `t64_ord_three___`, `t64_ord_two_pos4` against `t64_ord_two_kids` | two children on two nets, three children, two children of four ports | position 0 on every port of every instance after the first |
| `t64_ord_two_mods` against `t64_ord_two_nets` | the second child of another module | position 1 on its `o`: the count is per unit |
| `t64_ord_inout___` against `t63_pdr_port_bit` | a child `inout` on `v[3]` driven inside | offset 3, mode 0; `XZXZ` twice, `XZ0Z`, `0Z0Z`, `0Z1Z` |
| `t65_tim_1s______`, `t65_tim_ns_5s___` against `t44_time_5s_____` and `t44_v_time_5ms__` | an end at 1 s in picoseconds, a write at 4.5 s in nanoseconds | the times read back whole, whatever the unit |
| `t65_tim_cross___` against `t44_v_time_5ms__` | 3000 writes every 1 ns across 2^32 ps | eight pages, the crossing inside page 3, every record read back |
| `t66_prc_final___`, `t66_prc_latch___` against `t11_sv_logic____` | a `final` block, an `always_latch` | an `Always` scope for each |
| `t66_prc_ass_imm_`, `t66_prc_ass_conc`, `t66_prc_prop____` against `t11_sv_logic____` | an immediate assertion, a concurrent one, a named property | no scope of their own; handle space `0x9bc`, `0xd04`, `0xf04` |
| `t66_prc_task____`, `t66_prc_func____` against `t11_sv_logic____` | a task called from an `initial`, a function called from an `assign` | the tier 12 `task` and `function` units, `tb.t` and `tb.f` |
| `t66_prc_program_` against `t11_sv_logic____` | a `program` instantiated in the module | `tb.p` of a module unit, `tb.p.Initial20_0`; the run ends with the program |
| `t66_prc_bind____` against `t66_prc_func____` | a child bound in with `bind` | `tb.b` at the `bind` line, its port on its own handle |
| `t66_prc_specify_`, `t66_prc_spec_0__` against `t66_prc_kid_____` | a `specify` path delay of 1 and of 0 | records at 1 and 51 ns, `0x120` of handle space; nothing for the zero delay |
| `t66_prc_covgrp__` against `t11_sv_logic____` | a `covergroup` with one coverpoint | nine scopes, three of them `xlnx_isim_covergroup_cg::` functions |
| `//hdl/picorv32:sim` against `//hdl/serv:sim` | a second Verilog core, 280 objects, from its own bench | nothing new: every value and the VCD on the first attempt |
| `//hdl/neorv32:sim` against `//hdl/potato:sim` | a VHDL processor of 5696 objects, 1395 arenas | the chunk map belongs to the signal at the handle, not to the object |
| `t67_esz_pk_2bit_`, `t67_esz_pk_4bit_` against `t11_sv_struct___` | an enumeration field between two logics | the field is as wide as the entry's range, not as wide as its base type |
| `t67_esz_pk_int__` against `t67_esz_pk_2bit_` | the enumeration over `int` | 32 bits from the base type, which carries the range in that case |
| `t68_str_lit40___` against `t68_str_lit4____` | forty characters instead of four | nothing moves: the same 2619 bytes and `0xaac` of handle space |
| `t68_str_noinit__` against `t68_str_lit4____` | the string without an initializer | `0xa14`, the `0x98` of the implicit initializer process |
| `t68_str_dbg40___` against `t68_str_dbg_____` | forty characters under `-debug all` | the same file and the same eight zero bytes |
| `t68_str_dbg_arr_` against `t68_str_dbg_____` | an array of two strings under `-debug all` | one 64 bit object with one 16 byte record of zeros |
| `t68_str_log_____` against `t68_str_lit4____` | `log_wave /tb/str` under typical | nothing: the warning `No matching HDL object or HDL scope found` |
| `t68_str_byte____` against `t68_str_lit4____` | four `byte` holding the characters of the string | the characters, in one record, in reverse element order |
| `t69_vcl_wire_ini` against `t11_sv_logic____` | a net with an initializer, `wire w = 1'b1` | value class 0, where the same literal on a variable is class 1 |
| `t69_vcl_const_v_`, `t69_vcl_const_i_` against `t11_sv_logic____` | `const` before the declaration | nothing: the class is the initializer's, 1 and 3 |
| `t69_vcl_defparam`, `t69_vcl_gtop_prm` against `t11_sv_logic____` | a parameter overridden from outside the module | an ordinary class 3 parameter; the command line override renames the top scope `tb(P=5)` |
| `t69_vcl_specprm_` against `t11_sv_logic____` | a `specparam` in a `specify` block | a parameter declaration, class 3, with an object and a record |
| `t69_vcl_typeprm_` against `t11_sv_logic____` | a variable of a `parameter type` | the parameter's name is the variable's type name |
| `t69_vcl_chandle_` against `t11_sv_logic____` | a `chandle` | nothing at all, as a `string` leaves nothing |
| `t70_num_a_i_byte` against `t60_dbg_assoc_i_` | a `byte` key in place of an `int` one | the key holds a number of its own, 1, where the `int` key had none to hold |
| `t70_num_a_v_str_`, `t70_num_a_b_str_` against `t60_dbg_assoc___` | another element under a `string` key | the numbers do not move: a `string` key spends none |
| `t70_num_a_e_key_` against `t60_dbg_assoc_i_` | an enumeration key | the enumeration's own declaration spends two before the element, and the key spends none |
| `t70_num_d_then_q` against `t61_num_a_then_q` | a dynamic array in place of the associative one | no number is left over: only an associative array leaves one |
| `t70_num_a_2dim__`, `t70_num_a_in_cls` against `t60_dbg_assoc___` | the array nested, and inside a class | the rule repeats per dimension, and counts from the class |
| `t71_rlw_arr_prm_` against `t71_rlw_sreal_p_` | two real parameters in an array instead of one scalar | 64 bits, 32 an element, where the scalar declares 16 |
| `t71_rlw_untyped_`, `t71_rlw_specprm_` against `t71_rlw_sreal_p_` | the same value without a type keyword | 32 bits, as a variable declares |
| `t71_rlw_pkg_prm_`, `t71_rlw_kid_prm_` against `t71_rlw_sreal_p_` | the parameter in a package and in a child | 16 either way: the scope does not move it |
| `t71_rlw_vhdl_gen` against `t71_rlw_sreal_p_` | a VHDL real generic | 8 bytes, the float itself, with no halving on that side |
| `t72_dbg_line____` against `t72_dbg_subprog_` | `line` in place of `subprogram` | the same file to the byte: one mode under two names |
| `t72_dbg_typical_` against `t72_dbg_line____` | the mode without byte 2 of word 15 | the function's scope without its objects |
| `t72_dbg_readers_` against `t72_dbg_drivers_` | `readers` in place of `drivers` | byte 2 of word 14 alone, so the two bytes are independent |
| `t73_prt_gen_last`, `t73_prt_inst_lst`, `t73_prt_blk_last` against `t55_prot_pkg_prc` | another scope after the last process | the method scopes stay under the process |
| `t73_prt_arch_gen` against `t73_prt_gen_last` | the protected type in the architecture | the pair moves back under `tb`, as tier 55 measured |
| `t75_acc_rec_____`, `t75_acc_acc_____`, `t75_acc_arr40___` against `t23_access______` | a record, an access and a forty element array behind the pointer | the same `8 48`, and the same 48 byte variable |
| `t75_fil_rec_____`, `t75_fil_arr_____` against `t23_file_int____` | a record and an array as the file's element | the same `8 40`, and the same 0 byte variable |
| `t76_stc_sv_arr__` against `t50_sub_func_prm` | an unpacked array local in SystemVerilog instead of VHDL | storage class 3, where the VHDL composite local is 4 |
| `t76_stc_file_prm`, `t76_stc_file_loc` against `t51_sub_file_prm` | a file inside the subprogram rather than the architecture | no object at all, where the architecture's file is class 2 |
| `t74_lgw_root____`, `t74_lgw_cur_root` against `t74_lgw_star____` | the wildcard rooted, and the root made current | nothing changes: the package stays unlogged |
| `t74_lgw_pkg_obj_`, `t74_lgw_pkg_name` against `t74_lgw_star____` | the package named, by its objects and as a scope | the package signal is logged, either way |
| `//hdl/ibex:sim` against `//hdl/neorv32:sim` | a SystemVerilog design of 3287 objects | the enumeration width above; nothing else the reader had wrong |
| the 17 header words against the region lengths, every database, reread | nothing; a sweep | word `i` counts region `i + 4` for 0 to 13; words 1 to 4, 8 and 12 are the counts of the empty regions |
| every `sim.vcd` against its `sim.wdb`, through `go-vcd-parser` | nothing; the same run | the VCD spelling rules, the omission rule, the code sharing rule, and one wrong VCD value |

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
`time`, a parameter, a struct and an enum each get a `$var`, and only
unpacked arrays declared without a typedef, and strings, are absent.
VHDL generics and constants are absent too, and so is any signal outside
`tb`.
The full rule, held by `TestVCD` over every case, and the way the VCD
spells each value, are in [format/vcd.md](format/vcd.md).

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
   Where its first `0x1088` bytes go is open.
   Tier 46 prices a signal at `0xf8` and each driver at `0x50`, and
   tier 52 finds the `0x38` and the `0x20` of those that the first
   region's strides lack in the second region; what remains open is
   what a library package costs by, `0x604` for `std_logic_1164`,
   `0x1f8` for `numeric_std` and `0x400` for `math_real`, when a
   package of the design costs
   `0x80` plus its constants; see
   [format/container.md](format/container.md).
2. Trailer `+0x10`, `0x800`, and `+0x20`, `0xc8`, read as the arena
   span and the arena table offset by their values.
   Both are constant, so that is a reading, not a finding.
3. Word 28 of an instance record is a storage class of six seen
   values, tier 49; 5 has not been seen in the 48 cases built with
   `-debug subprogram` through tier 55, whose constants, aliases,
   files and protected variables in a subprogram are class 3 or 4 or
   absent, nor in the seven forms of tier 76, where a file parameter
   and a file local leave no object, an access parameter and every
   SystemVerilog argument and local are 3, and a SystemVerilog
   `string` local is absent.
   What it would name is open.
   Word 44 is `-1`, 0, `0x7ffc` to `0x7fff` on the objects of every
   Verilog case, or a value that differs between objects and between
   runs of the same design, tier 48; that the writer copies it from
   memory it never set is a reading, and the reader masks it.
4. DBG header word 16 is `0x10000` in every case, and the three
   `0x30` words at `0x98` and the `3` at `0xc0` in the fixed header
   are constant too.
   No case has moved them.
   Words 14 and 15 are the `-debug` flags, which tiers 24 and 72
   move.
5. Word 10 of a declaration record varies between runs for a signal and
   is 0 for a variable.
   Within one run it is a multiple of 256 and grows from signal to
   signal by `0x300` to `0x500` in declaration order, 1000 distinct
   values in `t46_sig_1000____`, which is the shape of a heap address
   truncated to 32 bits; that reading is a guess.
   It is masked as noise and not read.
6. Handles of generics, constants, variables and loop indexes come
   after the last signal of the design, tier 46, and within that region
   the strides are the value sizes and the scope costs of tiers 52
   and 53, which also place the region's start, and tier 54 fills
   the `0x580` between the signals and it with the packages' blocks.
   What the `0x738` before the first signal holds, what each package
   leaves past the second region, the 8 of handle space a scope without
   data objects costs beyond its `0x28`, and the `0x18` to `0x20` of
   handle space a generate iteration saves against an instance of
   the same body, are open.
   A resolved signal's stride, `0x140` for two drivers and `0x178` for
   three, fits `0x30` per driver plus `0x10` plus 8 per driver from
   the second on, on two points, which is a guess.
   All of it is read from the instance record, so nothing depends on
   it.
7. The chunk rule of [format/values.md](format/values.md) holds for
   55 wide VHDL values, 8 wide Verilog values and the wide partial
   writes of tiers 32 and 33, and its constants, 275, 24 and 299,
   have no explanation.
   The 275 and the 299 differ by the 24, so the rule may be one
   constant and one threshold in disguise.
   Tier 39 adds a second threshold: the rest of a chunked value is
   chunked again from 276 bytes, one above the 275 of a value, and
   why the two differ is open.
   A reading that fits both is a writer that splits any piece over
   275 bytes, `size > 275`, and a value of exactly 275 that reaches
   it by another path; no case separates that from two thresholds.
8. Whether `0xc4` and the other per-run durations mean anything is
   open.
   They are masked.
9. Does the format change between Vivado versions?
   Only 2025.2 is in use here.
   Any claim is version scoped until a second version has been
   measured.
10. The variant word of an enumeration entry, `2` for VHDL, `0` for
    `logic` and `1` for `bit`, and the `1` after a Verilog `real`,
    separate the types they are seen on and nothing else is known
    about them.
    The class word follows the literals, tier 20, and its four VHDL
    values look like a display rule: binary, nine state, character and
    named, but that reading is a guess.
11. A package signal is not logged under `log_wave -recursive *`
    and is under `log_wave -recursive /sig_pkg`, `t13_pkg_log_all`.
    A package parameter, `p.W` of `t13_sv_pkg`, is logged the same
    way, `t15_sv_pkg_log`.
    Tier 74 puts it on the object query rather than on `log_wave`:
    `get_scopes -r /*` matches the package scope and
    `get_objects -r /*` returns nothing from it, so `/*`, the root
    made current first and `log_wave [get_objects -r /*]` all log as
    little as `*` does, and naming the package is the only way in.
    Why the object query skips a package is what is left.
12. Which of the `n` duplicated variable handles in an entity
    instantiated `n` times belongs to which instance is not readable
    from the file.
    Nothing depends on it while variables have no records.
13. The handle space costs of a subprogram, 8 bytes, and of a `signal`
    parameter, `0x48`, name objects the instance list does not
    contain.
    What they are is open.
    The frame of a subprogram has its own costs that name nothing in
    the file, tier 55: 16 bytes over the value for a constant, 40 for
    a file object, 12 for the first protected variable and 16 for each
    further one, where a file parameter takes 8 and an alias nothing.
    The handle space a subprogram's static composite values take,
    tier 56, is measured but not placed: where in the handle space the
    bytes of an aggregate literal lie is open, and so is the 4 a record
    costs over its size.
    Why the second pair of a protected type's method scopes hangs from
    the architecture when the architecture declares the type, and from
    the architecture's last process when a package does, is open too.
    It is the last process and not the last scope visited: tier 73
    follows that process with a generate, an entity instance and a
    block, and the pair stays under the process in all three, so the
    writer keeps the process rather than the scope it saw last.
14. A `real` parameter declares 16 bits, `t12_v_params`, where a
    `real` variable declares 32 and both hold one `float64` pair.
    Tier 71 bounds it: a `shortreal` and a `realtime` parameter, and
    one in a package or a child module, declare 16 as well, and an
    array of two declares 64, so 32 an element, on the same type
    entry.
    An untyped parameter, a `specparam` and a variable holding the
    same `float64` all declare 32, and a VHDL real generic declares
    the 8 bytes of the float.
    So the 16 is the scalar typed parameter's alone, and why that one
    form declares half is open.
15. The records of one time keep their write order within an arena
    only.
    `t12_v_mem40_t0` writes forty elements at time 0 and the file holds
    the last nineteen in arena 0 before the first twenty one in arena
    1, so a same time write order across arenas cannot be read back.
    The reader replays the arenas in file order, and the test compares
    the final value of each time.
16. An 8 byte rest of a chunk split at an arena boundary has the shape
    of a pair write, `t12_v_mem40w32` at `0x800`.
    The reader gives a whole write the first unused record at each of
    its chunk addresses, so a pair write at the split address before
    the whole write in that arena is read as the rest, and the rest as
    the pair write.
    The final value of the time is right either way, because the
    arena keeps write order; the values between are not.
17. Answered by tier 68: the value of a SystemVerilog `string` goes
    nowhere.
    A search of the whole file and of every record of every inflated
    page finds the characters in none of the cases, at four and at
    forty characters, under typical and under `-debug all`, logged
    and not, while the same search finds them in the control case
    that holds them in an array of `byte`; see
    [format/values.md](format/values.md).
    What the 32 zero bits of the `-debug all` placeholder stand for is
    still open, and is part of question 24.
    A VHDL `string` variable is an ordinary array variable with a
    declaration and a size, `t44_str_var`, so nothing was open on that
    side.
18. The two words after the element type of a file type entry are
    `8` and `40` for every file type, and `8` and `48` for every
    access type, whatever the element or designated type.
    Tier 75 holds them over a record, a forty element array and
    another access type, and over a file of a record and of an array,
    and neither pair moves; every access variable declares 48 bytes
    and every file variable 0, whatever it points at or holds.
    So they are constants of the kind.
    That the second is the runtime size of the object fits the access
    side and not the file side, where the object would be 40 bytes and
    declares 0, and that the first is the size of a pointer is a guess
    no case here can move.
19. The value class codes of region 17 are 0, 1, 3, 4 and 6 in the
    corpus, and 2 and 5 have not been seen after the tier 28 sweep
    over real and time literals, casts, patterns, enums, `$time`,
    `$signed`, `$clog2` and parameter expressions, nor after the tier
    69 sweep over a `specparam`, the supply nets, `const` variables,
    a `defparam` and a command line override, a type parameter,
    `$bits`, a two state bit from a real, an enumeration from a cast
    of `'x`, a net with an initializer and a `chandle`.
    The integral types are 3 and `time` 4 whatever the initializer,
    `real` and `realtime` 0, and a packed type takes 0, 1, 3, 4 or 6
    from its initializer as converted to the target: none, a real, a
    pattern or a process for 0, a sized or fill literal or an
    expression of them for 1, an unsized literal or expression into a
    signed target for 3, into an unsigned target for 4, a string
    literal into a variable for 6.
    That the code names the kind of constant the elaborator holds for
    the initial value, and that 2 and 5 are kinds no corpus
    declaration produces, are guesses.
20. The record of an untyped parameter with a time literal is the 16
    bytes a 64 bit vector takes, `t12_v_param64`, filled from the
    address of an 8 byte `float64`, so the second half is what follows
    the value in memory: the next parameter in `t30_sv_ptm_two`.
    What the `a8 07 00 00 00 00 00 00` after the last parameter is,
    is open.
    The value itself is resolved; see [format/values.md](format/values.md).
21. Answered by tier 72: `-debug line` and `-debug subprogram` are
    one mode.
    The two write the same file to the byte, and byte 2 of word 15
    marks that mode rather than being set by `line` in passing: the
    files that have it declare a subprogram's own objects, and
    `typical`, which has byte 1 alone, has the subprogram's scope and
    none of its objects.
    See [format/hierarchy.md](format/hierarchy.md).
22. Which writes of the value held reach the VCD is open.
    The database records them by the two tier 36 rules, and the VCD
    drops nearly all of them, but keeps `rf_wen` of `//hdl/serv:sim`
    at 93000 ps and `r` of `t34_res_3drv____` at 80 ns.
    Tier 59 adds that the VCD keeps every held value write of a Tcl
    force on a VHDL signal, `1` at 20 ns in `t59_frc_s_const_`, and
    none on a SystemVerilog `logic`, `t59_frc_sv_tcl__`.
    That the VCD writer reads a dirty flag the database writer does
    not is a guess.
23. Why a clocked nonblocking assignment records the value held after
    an event on an operand of its block, and why a shared net records
    every evaluation, is open.
    That both come from a write path that skips the compare when the
    value arrives through the scheduler's event queue is a guess.
24. What the numbering under `-debug all` is for is open.
    Tier 61 fixes the order it is assigned in, see
    [format/types.md](format/types.md), but not what reads it: that
    it indexes a runtime table of the type descriptors the debugger
    builds for the objects it keeps, one per element kind, is a
    guess.
    Tier 70 halves the second question: the key of an associative
    array is one of those descriptors and carries its number in its
    own entry whenever it is not the element's type, which leaves
    exactly one number per associative array that nothing in the file
    carries, and a dynamic array and a queue leave none.
    That the leftover belongs to an iterator is still a guess.
    What the 32 zero bits of a string or class handle record stand for,
    and whether any script or source write ever changes them, is open
    too; `set_value` on the string ends the batch script.

## What the conversion writes out

VCD first, through `github.com/filmil/go-vcd-parser`, as the checking
step: it reads the `sim.vcd` answer key, and a decoded database can be
compared against it for the bit and vector signals.

FST is the deliverable, because it holds every type in the table above.
Keep the decoder's output model separate from the VCD writer, so that
adding an FST writer is a new writer and not a rewrite.
See [fst-output.md](fst-output.md).

SQLite is the third writer, for the reader that queries rather than
looks: the same signals and changes as rows, in the schema
`go-vcd-parser` writes from a VCD.
See [sqlite-output.md](sqlite-output.md).


## Method

Build a minimal pair, mask the noise, diff, and write the answer into
the document for that area of the file before moving on.
Each answer becomes a row in the findings table plus a check in
`//pkg/wdb:wdb_test`, which asserts against `truth.json` and never
against bytes the decoder itself produced.
Where a claim can be cross-checked against `sim.vcd`, the test does that
check rather than asserting bytes.
The corpus and the pairing rules are in [corpus.md](corpus.md).

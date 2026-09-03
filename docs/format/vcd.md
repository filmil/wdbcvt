<!-- SPDX-License-Identifier: Apache-2.0 -->

# The VCD beside the database, and the cross-check against it

Every corpus case writes `sim.vcd` beside `sim.wdb`, from one run of
one design, through `open_vcd` and `log_vcd [get_objects -r /*]` in
the xsim script.
The VCD is the first guard of [../provenance.md](../provenance.md): a
public text format, written by the same simulator from the same run,
and read here by `go-vcd-parser`, which this project did not write.

`TestVCD` in `pkg/wdb/vcd_test.go` compares the two files of every
case.
Everything below is what that comparison measured on Vivado 2025.2.
Reproduce all of it with:

```sh
bazel test //pkg/wdb:wdb_test --test_filter=TestVCD
```

The test is strict where the VCD says anything.
Every `$var` must name an object of the database, and at every time the
VCD lists a value, the last value the database holds at that time must
spell the same.
Where the VCD says nothing, the test demands a reason from the rule in
`vcdOmitted`, and fails when the VCD carries an object the rule says it
leaves out.
Where the VCD spells a value wrongly, the test expects the mismatch from
the list in `vcdDeviations`, and fails when the mismatch goes away.
So the three tables below are held to the corpus, not to a reading of
the files by eye.


## What the VCD carries

All 749 cases, `//hdl/counter:sim`, `//hdl/uart:sim`, `//hdl/serv:sim`
and `//hdl/potato:sim` pass, and every object of the database is either
in its VCD or covered by one line of this table.

| Object | In the VCD |
| :--- | :--- |
| VHDL `bit`, `std_ulogic`, `std_logic` | yes, `wire 1` |
| VHDL arrays of those with any number of index dimensions: `bit_vector`, `std_logic_vector`, `signed`, `unsigned`, the `(0 to 1, 0 to 2)` array of `t18_arr_2dim` | yes, `wire N` over all elements |
| VHDL `boolean`, `integer`, `real`, `time`, `character`, a user enumeration, a record, an array of arrays such as `t2_array2d` | no |
| VHDL generic or constant, of any type, including a generate index | no |
| A signal outside the top, such as the package signal `sig_pkg.x` of `t13_pkg_log_all` | no; the script logs `/*` |
| Verilog `reg`, `wire`, a vector, `integer`, `time`, `real`, `shortreal` | yes, as `reg`, `wire`, `integer`, `time` or `real` |
| Verilog `wand`, `wor`, `tri`, `triand`, `trior`, `tri0`, `tri1`, `supply0`, `supply1` | yes, as that keyword; a `uwire` as `wire` |
| Verilog `parameter`, including a string parameter | yes, as `parameter`; an untyped one with size `0` |
| SystemVerilog `bit`, `logic`, `byte`, `int`, `longint`, an enum, a typedef | yes |
| SystemVerilog packed struct | yes, as `reg N` over the fields |
| SystemVerilog unpacked struct | yes, as `reg N` over 32 bit slots, see below |
| SystemVerilog packed union | yes, as `reg N` of its widest field, `t24_sv_union` |
| Verilog memory or any unpacked array declared without a typedef, of vectors, reals or structs | no |
| A typedef of an unpacked array, `t13_sv_tdef_ua` | yes, flattened to `reg 8` |
| SystemVerilog `string` | no, and the database has no record either |

The top is the first child of the root scope, and the packages are the
root's other children: `tb` in every corpus case, `tb_processor` in
`//hdl/potato:sim`, whose 383 VCD variables in 20 scopes all sit under
it, where the database holds 557 objects in 144 scopes.
*Found by* `//hdl/potato:sim`, whose objects `TestVCD` reported as
outside `tb` until the rule read the top's name from the root scope.

The VHDL rows repeat the tier 2 measurement in
[../format.md](../format.md), now held by the test over every case.
The Verilog rows are new; the tier 11 note that only memories and
strings are absent was right but incomplete, since an unpacked array of
anything is absent unless a typedef names it.


## How a value is spelled

The VCD writes four state values.
The test turns a decoded database value into the same spelling before
comparing, by the rules the comparison found necessary:

* An `std_ulogic` `U`, `X`, `W` or `-` becomes `x`, `L` becomes `0`,
  `H` becomes `1` and `Z` becomes `z`.
  The VCD repeats a value it has already written when the simulation
  changed the signal to another value with the same spelling:
  `t1_nine_state` writes `x!` at 10 ns for `U` to `X`.
* A vector is written with its leading zeros dropped, and a vector of
  all `x` or all `z` as one character.
  A reader extends to the declared size with `0`, or with the leftmost
  `x` or `z`, as IEEE 1800 says.
* A Verilog `integer`, `time`, `int`, `longint` or `byte` is written as
  bits, two's complement for a negative value.
  The decoder spells the same value in decimal; the test converts.
* A SystemVerilog enum is written as the bits of its value at the width
  of its base type: `DONE` of `t11_sv_enum` is `b10`, `C` of
  `t11_sv_enum4` is `b1001`.
* A `real` is written as `r1.5`; the test compares the numbers.
* A packed struct is the bits of its fields, first field on the left,
  which is the order the database holds them in.
* A packed array with two packed dimensions, `t11_sv_arr2d`, is the
  elements in index order, element 0 on the left, likewise.
* A VHDL array with two or three index dimensions, `t18_arr_2dim` and
  `t18_arr_3dim`, is one `wire` of all elements in row major order,
  `(0,0)` on the left, which is also the order of the record bytes.
* An unpacked struct puts every field in a slot of whole 32 bit words,
  so `{logic a; logic [3:0] b;}` is declared `reg 64` and `a = 1,
  b = 1010` is written `b1` followed by 28 zeros and `1010`.
  A 40 bit field takes a 64 bit slot: `t11_sv_struct40` is `reg 96`.
  The database holds the fields at their own widths; see
  [values.md](values.md).


## When a value is written

* `$dumpvars` at time 0 lists the last value each object holds at
  time 0.
  A Verilog variable that the database records as `X` and then `0` at
  time 0 appears once, as `0`.
* After that, one entry per object per time at which the simulation
  changed it, holding the last value of that time.
  Three writes to one variable in one time step, `t13_v_same_t`, are
  one VCD entry and three database records.
  Delta cycles that return a signal to its old value, `t7_delta`, are
  one VCD entry carrying the old value and two database records.
* The set of times at which an object has a VCD entry equalled the
  set of times at which it has a database record through tier 35.
  In particular the missing time 0 `X` record of a spilled Verilog
  variable, `t13_v_tr430`, is invisible here, because the `0` record at
  time 0 remains.
* Tier 36 and `//hdl/serv:sim` broke that equality: the database
  records a clocked nonblocking assignment of the value held and every
  evaluation of a shared net, see [values.md](values.md), and the VCD
  drops most of those writes and keeps a few.
  `s` of `t36_v_nb_clk_lit` has a database record at 25 ns and no VCD
  entry there.
  `rf_wen` of `//hdl/serv:sim` is `X` in the database at 0, 31000 and
  93000 ps and `x` in the VCD at 0 and 93000 ps, the same record bytes
  at 31000 and 93000.
  Which repeats the VCD keeps is open.
  Four other corpus cases have a VCD entry that repeats the value
  before it, none of them a tier 36 write: `t7_delta` and `t22_vh_ns`
  return the signal to its old value within one VCD time step, so the
  last value of the step repeats the value before the step;
  `t1_nine_state` writes `X` over `U` and both are `x` in the VCD;
  `t34_res_3drv____` writes the resolved `X` a second time at 80 ns,
  and that write the VCD keeps.
  So `TestVCD` compares the changes of value on both sides, each list
  with its repeats dropped, and leaves the count of records to
  `records` in `truth.json`.


## Names and codes

A VCD scope path is the `$scope` names joined by dots, and the test
compares it with the object path of the database after stripping
backslashes and spaces from both.
Then the two agree in every case: `\g(0)\` for a VHDL generate,
`g[0].dut` as one scope name for a Verilog generate instance, `\g.r `
for a `reg` in an `if` generate, `b` for a named block and `inc` as a
`function` scope.

Two VCD variables with one identifier code are one signal seen from two
scopes.
In every case they are objects with one handle and one offset in the
database: the port and the net of `t12_v_port_wire`, the interface
instance and its port of `t13_sv_iface`.
The converse does not hold.
A port slice, `tb.dut.a` of `t9_port_slice`, has its own VCD code but
shares the handle of `tb.x` with an offset.


## Where the VCD is wrong

Two kinds of object have a VCD value that is not the value:

| Case | Object | What the VCD writes |
| :--- | :--- | :--- |
| `t11_sv_struct_r` | `tb.s`, `{real r; logic a;}` | `reg 64`; the `r` slot at `r = 1.5` is `00zzzzzzzzzzz` and zeros, 32 bits that spell no value |
| `t28_sv_prm_time`, `t30_sv_ptm_10ps` through `t30_sv_ptm_two` | `tb.T`, `parameter T = 10ns` without a type | `parameter 64`; `b111101010000z00000000z00z000000000000000000`, the four state reading of the `float64` bytes |

The database holds `1.5` as a 64 bit float, and `truth.json` agrees.
So for a real inside an unpacked struct there is no VCD guard at all,
only the truth file.
The untyped time parameter is the same: the database holds the
`float64` of the value in the time unit, and the VCD reads the same
bytes as a 64 bit vector, so the test takes every parameter the reader
classes as a time parameter as a deviation rather than listing the
cases by name; see the tier 30 notes in [values.md](values.md).


## Reading the VCD

`go-vcd-parser` is pinned at v0.1.0.
v0.2.0 needs Go 1.27.1 and the repository's Go toolchain is 1.26.2.

v0.1.0 does not read three spellings xsim uses in `$scope` and `$var`
lines: a VHDL extended identifier `\g(0)\`, a Verilog escaped
identifier `\g.r `, and a generate instance `g[0].dut`.
Nor does it read a plain name that starts like a value: `$scope module
bufreg` of `//hdl/serv:sim` fails at the `b`, read as a binary vector,
and the top `tb(memfile="...")` fails at the parenthesis.
The test swaps every name in a `$scope` or `$var` line for `hiddenN`
before parsing and puts it back when it builds the path, in
`hideNames`.
Values, times and codes are read by the parser unchanged.

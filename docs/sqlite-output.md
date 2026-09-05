<!-- SPDX-License-Identifier: Apache-2.0 -->

# The SQLite output

`wdbcvt -in <file>.wdb -sqlite <file>.db` writes the signals of a
database and their changes into an SQLite file:

```
bazel run //cmd/wdbcvt -- \
    -in "$PWD/bazel-bin/hdl/counter/sim.wdb" \
    -sqlite "$PWD/counter.db"
sqlite3 counter.db 'SELECT Name, Size FROM Signals;'
```

The schema is the one
[github.com/filmil/go-vcd-parser](https://github.com/filmil/go-vcd-parser)
writes from a VCD, so a query written against one database reads the
other.
What this page states is the mapping: which row an object becomes, how
a record or an array is flattened, and how a value is spelled.


## The tables

```sql
CREATE TABLE Signals(
    Name TEXT PRIMARY KEY,
    Type INTEGER NOT NULL,
    Code TEXT NOT NULL,
    Size INTEGER NOT NULL);

CREATE TABLE Svalues(
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Timestamp INTEGER NOT NULL,
    Code TEXT NOT NULL,
    Value TEXT NOT NULL,
    FOREIGN KEY(Code) REFERENCES Signals(Code));

CREATE TABLE Meta(
    Key TEXT PRIMARY KEY,
    Value TEXT NOT NULL);
```

Two things differ from the upstream DDL, and both are deliberate.

The columns are `TEXT` where that package declares `STRING`.
SQLite gives a column whose declared type contains neither `TEXT` nor
`INT` numeric affinity, and a numeric column converts what looks like a
number: `00000001` is stored as the integer 1, and a 22 bit value as
the double `1.11111111111111e+21`, which is the value destroyed.
`TEXT` keeps what was written.
The table and column names, which are what a query names, do not move.

`Meta` is new.
The timestamps are counts of the file's own time unit and nothing in
the two original tables says what that unit is:

| Key | Value |
| :--- | :--- |
| `generator` | `wdbcvt` |
| `source` | `Vivado xsim waveform database` |
| `time_exponent` | the power of ten of a second, `-12` for picoseconds |
| `timescale` | the same as text, `1e-12 s` |
| `end_time` | the simulation's end time, in that unit |


## The mapping

**Names.** A row's name is the path of the object with `/` between the
steps and the root written as a leading `/`, which is the shape
`go-vcd-parser` writes: `//tb/dut/value`.
An extended identifier loses its decoration, `\g(0)\` becomes `g(0)`,
as it does in the FST output.

**Codes.** A code is invented, because a database has handles and no
VCD identifier codes: the printable ASCII characters from `!` upward,
in declaration order, and more than one character once they run out.
Two objects that share a handle share a code, which is what a VCD does
with an aliased signal: `//tb/ctl/clk` and `//tb/dut/ctl/clk` are one
code and one set of rows in `Svalues`.

**Flattening.** A record and an array of anything but bits become
another level of the name, one row per leaf, as they become a scope in
the FST output: a VHDL record `stat` is `//tb/stat/value` and
`//tb/stat/wrapped`.
An array of logic values is one row of that many bits, and an array of
characters is one row holding the string.

**Types.** `Signals.Type` is `go-vcd-parser`'s `vcd.VarKindCode`, the
number a reader of that database compares against:

| Object | Type | Size |
| :--- | ---: | :--- |
| a logic scalar or vector | 15, `wire` | the bits |
| an integer | 1, `integer` | 32 |
| a real | 3, `real` | 64 |
| a physical value, `time` included | 7, `time` | 1 |
| anything else, an enumeration or a string | 18, `string` | 1 |

**Values.** A logic value is its VCD character in lower case, and a
vector is those characters at the declared width: `00000001`, where a
VCD writes `b1` and leaves the width to the reader.
An integer, a real and an enumeration are their own text, `256`,
`0.5`, `TRUE`.
A physical value keeps its unit, `10000 ps`.
A row is written only when the value changes.

**Duplicated names.** `Name` is the primary key, and a design can
declare one path twice: `//hdl/counter` has four loop indexes named `i`
in one process, each on its own handle.
The first keeps the plain name and the rest get `#2`, `#3` and so on.
A VCD has no such trouble, because it names a signal by its code and
lets two `$var` lines agree.


## What it holds that a VCD does not

The same thing the FST output holds, for the same reason: Vivado's VCD
declares no `boolean`, `integer`, `real`, `time`, `character`, user
enumeration, record or array; see [format/vcd.md](format/vcd.md).
The counter design is 2 signals in its VCD and 19 rows here, among them

```
//tb/running     TRUE
//tb/period      10000 ps
//tb/full_scale  256
```


## What it costs

One row per leaf value change, and a row is about 70 bytes on disk:

| Design | Signal rows | Value rows | Time | File |
| :--- | ---: | ---: | ---: | ---: |
| `//hdl/counter:sim` | 19 | 1080 | 0.01 s | 78 kB |
| `//hdl/serv:sim` | 3621 | 2256661 | 21 s | 150 MB |
| `//hdl/neorv32:sim` | 67435 | 2107670 | 88 s | 157 MB |

The value rows are fewer than the database's changes, 18875466 for
neorv32, because a change that leaves a leaf holding what it held
writes no row, as it writes no FST value.

The FST of the same run is far smaller, 690 kB for neorv32, because it
compresses a signal's history as a block and this does not.
The reason to write the SQLite anyway is the query: what a row costs
buys `SELECT`.


## How it is checked

`//pkg/sqlout:sqlout_test` writes the database of `//hdl/counter:sim`
and `//hdl/uart:sim` and compares it against the answer key: it parses
the same run's `sim.vcd` with `go-vcd-parser`, converts it with that
package's own `cvt` and `db`, and requires every signal the VCD
declares to be here, at the same width, holding the same value at every
time the VCD names one.
The comparison pads a VCD value back to the declared width, since a VCD
drops the leading digits of a vector.

Those two designs are the ones whose signals are narrow enough for the
comparison: the reference database stores values in a `STRING` column,
so a value wider than about 18 bits comes back as a double and no
longer says what it held.
That is the defect the `TEXT` columns above avoid.

Every database in the repository also has a `:convert` test, which
writes both outputs and fails if either conversion does, so a break
names the design it broke on.

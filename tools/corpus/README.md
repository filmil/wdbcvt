# Corpus generators

These scripts wrote the corpus cases of tiers 57 to 66.
Each `gen_tNN.py` writes the sources, the `BUILD.bazel` and the
`truth.json` of its tier into `hdl/corpus/`, through `gen_common.py`.
They are kept so that a case can be regenerated, or a tier extended, by
editing the script instead of every file by hand.

Run one from anywhere:

```
python3 tools/corpus/gen_t60.py
```

Then register the new cases in `hdl/corpus/BUILD.bazel`, fix the SPDX
comment of VHDL files (`emit` writes a `//` header), build the cases
and fill in the truths from the dump.

`typetab.py` splits the type table of a waveform database and prints
every entry as words, for entries the reader does not parse yet:

```
python3 tools/corpus/typetab.py bazel-bin/hdl/corpus/t60_dbg_queue___/sim.wdb
```

# Corpus generators

These scripts wrote the corpus cases of tiers 57 to 67.
Each `gen_tNN.py` writes the sources, the `BUILD.bazel` and the
`truth.json` of its tier into `hdl/corpus/`, through `gen_common.py`.
They are kept so that a case can be regenerated, or a tier extended, by
editing the script instead of every file by hand.

Each is a Bazel target, so it runs on the interpreter Bazel fetches
rather than on whatever `python3` the machine has:

```
bazel run //tools/corpus:gen_t60
```

The generators write into the source tree, which is where the corpus
lives, and they reproduce it exactly: running every one of them and
then `git status` is how to check that a generator still writes the
cases it is supposed to.

Then register the new cases in `hdl/corpus/BUILD.bazel`, build them,
and fill in the truths from the dump.

`typetab.py` splits the type table of a waveform database and prints
every entry as words, for entries the reader does not parse yet:

```
bazel run //tools/corpus:typetab -- "$PWD/bazel-bin/hdl/corpus/t60_dbg_queue___/sim.wdb"
```

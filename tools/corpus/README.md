<!-- SPDX-License-Identifier: Apache-2.0 -->

# Corpus generators

These programs wrote the corpus cases of tiers 57 to 67.
Each `gen_tNN` writes the sources, the `BUILD.bazel` and the
`truth.json` of its tier into `hdl/corpus/`, through the `gencommon`
package.
They are kept so that a case can be regenerated, or a tier extended, by
editing the program instead of every file by hand.

Each is a Bazel target, in Go like the rest of the repository:

```
bazel run //tools/corpus/gen_t60
```

The generators write into the source tree, which is where the corpus
lives, and they reproduce it exactly: running every one of them and
then `git status` is how to check that a generator still writes the
cases it is supposed to.

Then register the new cases in `hdl/corpus/BUILD.bazel`, build them,
and fill in the truths from the dump.

A truth file is compared byte for byte against the checked in copy, so
`gencommon` writes JSON the way it was first written: keys in the order
they were set, two spaces of indent, and the escapes Python's `json`
module applies.
`tools/corpus/gencommon/json.go` holds that, and nothing else should
write a truth.

`typetab` splits the type table of a waveform database and prints
every entry as words, for entries the reader does not parse yet:

```
bazel run //tools/corpus/typetab -- "$PWD/bazel-bin/hdl/corpus/t60_dbg_queue___/sim.wdb"
```

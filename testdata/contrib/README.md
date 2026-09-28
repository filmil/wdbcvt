<!-- SPDX-License-Identifier: Apache-2.0 -->

# Contributed databases

This directory holds waveform databases that people outside this
repository sent in, usually with a bug report or a patch.
Each one was written by the sender's Vivado, from the sender's design.
Neither the source nor the simulation is here, so no file in this
directory has a `truth.json`.

`//pkg/wdb:contrib_test` reads every file here.
It checks that the reader opens the file and decodes every value
change of every object without an error.
It also checks that each file holds the structure it was sent to show.
That is weaker than the corpus test under `//hdl/corpus`, which
compares each value with a truth written from the source.
A contributed file shows that the reader accepts what a real
simulator wrote; it does not show that the values it reads are right.


## Layout

One directory per contribution, named after where it came from:
`gh_pr1` holds the files sent with pull request 1 on the GitHub
mirror.
The directory keeps the file names the sender gave.
A new contribution gets a new directory and a row in the table below.


## Contributions

| Directory | From | Vivado | What the files hold |
| :--- | :--- | :--- | :--- |
| `gh_pr1` | rolandking, [filmil/wdbcvt#1](https://github.com/filmil/wdbcvt/pull/1) | 2026.1 | `repro_0001_record_trailer_arp_tb.wdb` holds 13 SystemVerilog structs of origin 1, whose type table entries end in a number rather than -99. `repro_0003_partial_first_write_ip_checksum.wdb` holds `gen_level[3].stage_sum`, a `logic [18:0] [1:0]` whose first write covers 8 of the 16 bytes the reader keeps for it. Neither file holds an `event` type or a named-values entry with 16 byte values, the other two cases the pull request decodes. |

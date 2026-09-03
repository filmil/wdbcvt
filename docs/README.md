<!-- SPDX-License-Identifier: Apache-2.0 -->

# Documentation

Everything known about the Vivado `xsim` waveform database format lives
here, and it is written down as it is discovered rather than at the end.

Read [provenance.md](provenance.md) first if you intend to rely on any
of it. This is an AI-first exploration of an undocumented format, and
that document states what guards the findings and where the tools should
not be used.


## The documents

| File | What it holds |
| :--- | :--- |
| [format.md](format.md) | **The findings.** One row per confirmed fact, with the command that reproduces it, the comparison that found each, and the open questions, labelled as guesses. |
| [format/container.md](format/container.md) | The fixed header, the arena table, the trailer, the directory, the page directory, the marker and the noise mask. |
| [format/types.md](format/types.md) | The type table. |
| [format/hierarchy.md](format/hierarchy.md) | The debug section: scopes, units, declarations, files, instance records and handles. |
| [format/values.md](format/values.md) | Arenas, pages, value records, encodings and alignment. |
| [format/vcd.md](format/vcd.md) | What Vivado's VCD holds of the same run, how it spells values, and the cross-check of every case against it. |
| [corpus.md](corpus.md) | The differential corpus, and the method: minimal pairs, the noise mask, and the ground truth files. |
| [provenance.md](provenance.md) | How the knowledge was obtained, what guards it, and the limits of it. |
| [fst-output.md](fst-output.md) | The FST output format, its licence position, and the route chosen for writing it. |


## How a finding gets written down

A fact enters the findings table in [format.md](format.md) only when a
command reproduces it, and the row names that command.
Everything else stays in the open questions section, marked as a guess.
The two are kept apart on purpose: an agent inferring a format from
examples produces measurements and plausible stories, and they look
identical on the page.

Every claim is scoped to the Vivado version that produced the file.
Only 2025.2 has been observed.
A second version is a new experiment, not a confirmation.

`format.md` is the index and the findings table.
The layout itself lives in `docs/format/`, one file per area of the
container, and a new finding goes into the file for its area with the
case that found it.

#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
#
# Converts one waveform database to the outputs the tool writes and
# fails if any conversion fails. $1 is the wdbcvt binary, $2 is `all`
# or `fst`, and the rest are the files of the simulation target, of
# which the `.wdb` is the one to convert.
set -euo pipefail

tool="$1"
outputs="$2"
shift 2

wdb=""
for f in "$@"; do
  case "${f}" in
    *.wdb) wdb="${f}" ;;
  esac
done
test -n "${wdb}" || {
  echo "no .wdb among the simulation's files: $*" >&2
  exit 1
}

out="${TEST_TMPDIR:-/tmp}/out.fst"
"${tool}" -in "${wdb}" -fst "${out}"

# libfst writes a header even for an empty design, so an output of zero
# bytes means the writer produced nothing at all.
test -s "${out}" || {
  echo "FST conversion of ${wdb} produced an empty file" >&2
  exit 1
}

if [ "${outputs}" = "all" ]; then
  db="${TEST_TMPDIR:-/tmp}/out.db"
  "${tool}" -in "${wdb}" -sqlite "${db}"

  # SQLite writes its own header into an empty database, so this is the
  # same check: the file exists and is not empty.
  test -s "${db}" || {
    echo "SQLite conversion of ${wdb} produced an empty file" >&2
    exit 1
  }
fi

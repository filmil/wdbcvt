#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
#
# Converts one waveform database to FST and fails if the conversion
# fails. $1 is the wdbcvt binary; the rest are the files of the
# simulation target, of which the `.wdb` is the one to convert.
set -euo pipefail

tool="$1"
shift

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
  echo "conversion of ${wdb} produced an empty file" >&2
  exit 1
}

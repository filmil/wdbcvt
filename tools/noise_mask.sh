#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
#
# Simulate one design twice and report what differs between the two runs.
#
# This is the first experiment of the corpus method, and the one that is
# easy to skip. A waveform database usually embeds things that change
# between two runs of the same simulation: a timestamp, a host name, a
# path, a build id. Those bytes differ for reasons unrelated to the
# design. Comparing two different designs without excluding them
# identifies a run timestamp as a signal count, confidently and wrongly.
#
# Bazel is the obstacle here rather than the help: a second build of an
# unchanged target returns the cached file, so the naive "build it twice"
# compares a file with itself and reports perfect determinism. Passing a
# different WDB_NONCE changes the action key, so the simulation actually
# runs again. The variable reaches xsim's environment and nothing reads
# it, so it does not change what is simulated.
#
# Usage:
#     tools/noise_mask.sh //hdl/corpus/t1_bit_one_edge:sim [outdir]
#
# See docs/wdb-corpus.md.

set -euo pipefail

readonly target="${1:?usage: noise_mask.sh <bazel target> [outdir]}"
readonly outdir="${2:-$(mktemp -d)}"

readonly pkg="${target%:*}"
readonly name="${target##*:}"
readonly rel="${pkg#//}/${name}.wdb"

mkdir -p "${outdir}"

run() {
    local nonce="$1" dest="$2"
    echo "run ${nonce}: building ${target}" >&2
    bazelisk build --action_env="WDB_NONCE=${nonce}" "${target}" >&2
    local produced
    produced="$(bazelisk info bazel-bin)/${rel}"
    if [[ ! -s "${produced}" ]]; then
        echo "noise_mask.sh: missing or empty: ${produced}" >&2
        exit 1
    fi
    cp -f "${produced}" "${dest}"
    chmod u+w "${dest}"
}

run 1 "${outdir}/run1.wdb"
run 2 "${outdir}/run2.wdb"

echo >&2
echo "two runs of ${target}:" >&2
ls -l "${outdir}/run1.wdb" "${outdir}/run2.wdb" >&2
echo >&2

if cmp -s "${outdir}/run1.wdb" "${outdir}/run2.wdb"; then
    echo "The two runs are byte for byte identical."
    echo "The database is deterministic for this design, so no mask is needed."
    echo "That is a strong property. Record it in docs/wdb-format.md."
    exit 0
fi

echo "The two runs differ. Every offset below is noise, not design." >&2
echo >&2
bazelisk run //cmd/wdbdiff -- \
    -mask-a "${outdir}/run1.wdb" -mask-b "${outdir}/run2.wdb" \
    -a "${outdir}/run1.wdb" -b "${outdir}/run2.wdb"

echo >&2
echo "Mask files kept in ${outdir}; pass them to wdbdiff as -mask-a and -mask-b." >&2

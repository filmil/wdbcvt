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
# Bazel is the obstacle here rather than the help. A second build of an
# unchanged target returns the cached file, so the naive "build it twice"
# compares a file with itself and reports perfect determinism.
#
# `--action_env` does NOT solve this. It changes the action key only for
# actions that declare the variable in their environment, and the Vivado
# simulation action does not. Measured: two builds with different
# WDB_NONCE values both reported "1 total action" with no process run,
# and produced identical files, which looked like a deterministic format
# and was really one cached file compared with itself.
#
# What does work is removing both places the result can be cached: the
# output tree, with `bazel clean`, and the shared disk cache, with an
# empty `--disk_cache=`. Then the simulation genuinely runs again.
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
    local label="$1" dest="$2"
    echo "run ${label}: forcing a genuine re-simulation of ${target}" >&2
    bazelisk clean >/dev/null 2>&1
    bazelisk build --disk_cache= "${target}" >&2
    local produced
    produced="$(bazelisk info --disk_cache= bazel-bin)/${rel}"
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

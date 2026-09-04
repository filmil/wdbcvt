// SPDX-License-Identifier: Apache-2.0

// The C side of Read: libfst hands every value change to a callback,
// and cgo cannot pass a Go function to C. The two functions this file
// names are the exported Go ones, declared by cgo in _cgo_export.h,
// and the casts are to the pointer types the reader takes. Their
// parameters agree with those types; only the spelling differs.

#include "_cgo_export.h"

#include <fstapi.h>

typedef void (*fst_vc_cb)(void *, uint64_t, fstHandle, const unsigned char *);
typedef void (*fst_vc_varlen_cb)(void *, uint64_t, fstHandle, const unsigned char *, uint32_t);

int fstGoIterBlocks(fstReaderContext *ctx, void *user)
{
    return fstReaderIterBlocks2(ctx,
                                (fst_vc_cb)fstGoValueChange,
                                (fst_vc_varlen_cb)fstGoValueChangeVarlen,
                                user,
                                NULL);
}

#!/usr/bin/env python3
"""Split a type table into entries and print each as hex, for exploration."""
import struct, sys
def entries(b):
    i = b.index(b"Xilinx ISim TYPE FILE 001")
    n = struct.unpack_from("<I", b, i + 0x20)[0]
    size = struct.unpack_from("<I", b, i + 0x24)[0]
    p = i + 0x28
    out = []
    for k in range(n):
        ln = struct.unpack_from("<I", b, p)[0]
        body = b[p + 4:p + ln]
        out.append((p - i, body))
        p += ln
    return i, n, size, out, p - i
def show(fn):
    b = open(fn, "rb").read()
    i, n, size, es, end = entries(b)
    print("%s: %d entries, size word %#x, entries end at %#x" % (fn.split("/")[-2], n, size, end))
    for k, (off, body) in enumerate(es):
        tag = body[0]
        z = body.index(b"\0", 4)
        name = body[4:z].decode()
        rest = body[z + 1:]
        words = " ".join("%08x" % w for w in struct.unpack_from("<%dI" % (len(rest) // 4), rest, 0)) if len(rest) % 4 == 0 else rest.hex()
        print("  [%d] kind %#04x %-12s %s" % (k, tag, repr(name), words if len(rest) % 4 == 0 else "RAW " + rest.hex()))
for fn in sys.argv[1:]:
    show(fn)

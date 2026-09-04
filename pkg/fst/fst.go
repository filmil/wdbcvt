// SPDX-License-Identifier: Apache-2.0

// Package fst writes FST waveform files through libfst, the reader and
// writer GTKWave uses. The format has no specification, so the library
// is the definition of it, and this package is a thin binding: it owns
// the lifetime of the writer context and the shape of the calls, and
// leaves every byte of the file to libfst.
package fst

// #include <stdlib.h>
// #include <fstapi.h>
import "C"

import (
	"fmt"
	"unsafe"
)

// VarType is the type of a variable, as FST names them.
type VarType uint32

// The types this converter needs. FST has 30 of them.
const (
	VarWire    VarType = C.FST_VT_VCD_WIRE
	VarReg     VarType = C.FST_VT_VCD_REG
	VarInteger VarType = C.FST_VT_VCD_INTEGER
	VarReal    VarType = C.FST_VT_VCD_REAL
	VarTime    VarType = C.FST_VT_VCD_TIME
	VarString  VarType = C.FST_VT_GEN_STRING
	VarEnum    VarType = C.FST_VT_SV_ENUM
)

// Writer writes one FST file.
type Writer struct {
	ctx *C.fstWriterContext
}

// Create opens path for writing.
func Create(path string) (*Writer, error) {
	p := C.CString(path)
	defer C.free(unsafe.Pointer(p))
	ctx := C.fstWriterCreate(p, 1)
	if ctx == nil {
		return nil, fmt.Errorf("fst: cannot write %s", path)
	}
	return &Writer{ctx: ctx}, nil
}

// SetTimescale sets the file's time unit as a power of ten of a second,
// so -12 is picoseconds.
func (w *Writer) SetTimescale(exp int) { C.fstWriterSetTimescale(w.ctx, C.int(exp)) }

// PushScope enters a scope of the instance tree.
func (w *Writer) PushScope(name string) {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	C.fstWriterSetScope(w.ctx, C.FST_ST_VCD_MODULE, n, nil)
}

// PopScope leaves the scope entered last.
func (w *Writer) PopScope() { C.fstWriterSetUpscope(w.ctx) }

// Var declares a variable of bits bits in the current scope and returns
// its handle. A handle of a variable declared with the same alias is
// the same object seen from two scopes.
func (w *Writer) Var(t VarType, name string, bits int, alias uint32) uint32 {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	return uint32(C.fstWriterCreateVar(w.ctx, C.enum_fstVarType(t), C.FST_VD_IMPLICIT,
		C.uint32_t(bits), n, C.fstHandle(alias)))
}

// Time sets the time of the value changes that follow.
func (w *Writer) Time(t uint64) { C.fstWriterEmitTimeChange(w.ctx, C.uint64_t(t)) }

// Value writes a value change. The value is the variable's bits as the
// characters 0, 1, x and z, or the decimal text of a real.
func (w *Writer) Value(handle uint32, v string) {
	s := C.CString(v)
	defer C.free(unsafe.Pointer(s))
	C.fstWriterEmitValueChange(w.ctx, C.fstHandle(handle), unsafe.Pointer(s))
}

// Close finishes the file.
func (w *Writer) Close() {
	if w.ctx != nil {
		C.fstWriterClose(w.ctx)
		w.ctx = nil
	}
}

// The reader below is for tests: the converter writes FST and never
// reads it. It is here because libfst is the definition of the format,
// so reading a file back with it is the check that the writer wrote
// what it meant to.

// Var2 is a variable as the reader reports it.
type Var2 struct {
	Path   string
	Bits   int
	Handle uint32
}

// ReadVars opens path and returns its variables with their scopes.
func ReadVars(path string) ([]Var2, uint64, uint64, error) {
	p := C.CString(path)
	defer C.free(unsafe.Pointer(p))
	ctx := C.fstReaderOpen(p)
	if ctx == nil {
		return nil, 0, 0, fmt.Errorf("fst: cannot read %s", path)
	}
	defer C.fstReaderClose(ctx)
	var out []Var2
	var scope []string
	for {
		h := C.fstReaderIterateHier(ctx)
		if h == nil {
			break
		}
		switch h.htyp {
		case C.FST_HT_SCOPE:
			s := (*C.struct_fstHierScope)(unsafe.Pointer(&h.u))
			scope = append(scope, C.GoString(s.name))
		case C.FST_HT_UPSCOPE:
			if len(scope) > 0 {
				scope = scope[:len(scope)-1]
			}
		case C.FST_HT_VAR:
			v := (*C.struct_fstHierVar)(unsafe.Pointer(&h.u))
			name := C.GoString(v.name)
			full := name
			for i := len(scope) - 1; i >= 0; i-- {
				full = scope[i] + "." + full
			}
			out = append(out, Var2{Path: full, Bits: int(v.length), Handle: uint32(v.handle)})
		}
	}
	return out, uint64(C.fstReaderGetStartTime(ctx)), uint64(C.fstReaderGetEndTime(ctx)), nil
}

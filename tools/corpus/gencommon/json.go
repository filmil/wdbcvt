// SPDX-License-Identifier: Apache-2.0

package gencommon

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Obj is a JSON object that remembers the order its keys were set in.
// A truth file is compared byte for byte against the checked in copy, so
// the order of the keys is part of the output and not an accident of the
// map implementation.
type Obj struct {
	keys []string
	vals map[string]any
}

// O builds an object from alternating key and value arguments.
func O(kv ...any) *Obj {
	o := &Obj{vals: map[string]any{}}
	if len(kv)%2 != 0 {
		panic("O: odd number of arguments")
	}
	for i := 0; i < len(kv); i += 2 {
		k, ok := kv[i].(string)
		if !ok {
			panic(fmt.Sprintf("O: key %v is not a string", kv[i]))
		}
		o.Set(k, kv[i+1])
	}
	return o
}

// Set stores a value. A key that is already present keeps its place, the
// way assigning to a key of a Python dict does.
func (o *Obj) Set(k string, v any) {
	if o.vals == nil {
		o.vals = map[string]any{}
	}
	if _, seen := o.vals[k]; !seen {
		o.keys = append(o.keys, k)
	}
	o.vals[k] = v
}

// With copies the object and sets the given key and value pairs on the
// copy, which is what `dict(sig(...), **kw)` does.
func (o *Obj) With(kv ...any) *Obj {
	c := o.Clone()
	if len(kv)%2 != 0 {
		panic("With: odd number of arguments")
	}
	for i := 0; i < len(kv); i += 2 {
		c.Set(kv[i].(string), kv[i+1])
	}
	return c
}

// Clone returns a copy that can be changed without touching the original.
// Nested objects and slices of objects are copied too, because Norm
// rewrites the fields of a signal in place.
func (o *Obj) Clone() *Obj {
	c := &Obj{keys: append([]string(nil), o.keys...), vals: map[string]any{}}
	for k, v := range o.vals {
		switch t := v.(type) {
		case *Obj:
			c.vals[k] = t.Clone()
		case []*Obj:
			l := make([]*Obj, len(t))
			for i, e := range t {
				l[i] = e.Clone()
			}
			c.vals[k] = l
		default:
			c.vals[k] = v
		}
	}
	return c
}

// Update sets every key of other, in the order other holds them.
func (o *Obj) Update(other *Obj) {
	for _, k := range other.keys {
		o.Set(k, other.vals[k])
	}
}

// Keys returns the keys in the order they were set.
func (o *Obj) Keys() []string { return append([]string(nil), o.keys...) }

// Get returns a value, or nil when the key is absent.
func (o *Obj) Get(k string) any { return o.vals[k] }

// Has reports whether the key is present.
func (o *Obj) Has(k string) bool { _, ok := o.vals[k]; return ok }

// Str returns a string valued key, empty when absent.
func (o *Obj) Str(k string) string {
	s, _ := o.vals[k].(string)
	return s
}

// Int returns an int valued key, zero when absent.
func (o *Obj) Int(k string) int {
	switch t := o.vals[k].(type) {
	case int:
		return t
	case json.Number:
		var n int
		fmt.Sscan(t.String(), &n)
		return n
	}
	return 0
}

// Fields returns the fields of a signal, empty when it has none.
func (o *Obj) Fields() []*Obj {
	f, _ := o.vals["fields"].([]*Obj)
	return f
}

// WriteJSON writes the value the way `json.dump(v, f, indent=2)` does:
// two spaces per level, `": "` between a key and its value, and no space
// before a comma. The corpus was written by that function, so a byte for
// byte comparison holds the port to the same output.
func WriteJSON(w io.Writer, v any) error {
	var b strings.Builder
	writeValue(&b, v, 0)
	b.WriteString("\n")
	_, err := io.WriteString(w, b.String())
	return err
}

func writeValue(b *strings.Builder, v any, depth int) {
	pad := strings.Repeat("  ", depth+1)
	end := strings.Repeat("  ", depth)
	switch t := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		if t {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case int:
		fmt.Fprintf(b, "%d", t)
	case json.Number:
		b.WriteString(t.String())
	case string:
		writeString(b, t)
	case *Obj:
		if len(t.keys) == 0 {
			b.WriteString("{}")
			return
		}
		b.WriteString("{\n")
		for i, k := range t.keys {
			if i > 0 {
				b.WriteString(",\n")
			}
			b.WriteString(pad)
			writeString(b, k)
			b.WriteString(": ")
			writeValue(b, t.vals[k], depth+1)
		}
		b.WriteString("\n" + end + "}")
	case []*Obj:
		l := make([]any, len(t))
		for i, e := range t {
			l[i] = e
		}
		writeValue(b, l, depth)
	case []string:
		l := make([]any, len(t))
		for i, e := range t {
			l[i] = e
		}
		writeValue(b, l, depth)
	case []any:
		if len(t) == 0 {
			b.WriteString("[]")
			return
		}
		b.WriteString("[\n")
		for i, e := range t {
			if i > 0 {
				b.WriteString(",\n")
			}
			b.WriteString(pad)
			writeValue(b, e, depth+1)
		}
		b.WriteString("\n" + end + "]")
	default:
		panic(fmt.Sprintf("WriteJSON: unsupported value %T", v))
	}
}

// writeString escapes the way Python's json does with ensure_ascii, which
// is where the corpus came from: the two mandatory escapes, the short
// escapes for the control characters that have one, and \uXXXX for
// everything else outside printable ASCII.
func writeString(b *strings.Builder, s string) {
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 || r > 0x7e {
				fmt.Fprintf(b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
}

// LoadObj reads a JSON object and keeps the order of its keys, so that a
// generator can copy entries out of a truth another tier wrote and put
// them back unchanged.
func LoadObj(path string) (*Obj, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	d.UseNumber()
	t, err := d.Token()
	if err != nil {
		return nil, err
	}
	if t != json.Delim('{') {
		return nil, fmt.Errorf("%s: not a JSON object", path)
	}
	return readObj(d)
}

func readObj(d *json.Decoder) (*Obj, error) {
	o := O()
	for {
		t, err := d.Token()
		if err != nil {
			return nil, err
		}
		if t == json.Delim('}') {
			return o, nil
		}
		v, err := readValue(d)
		if err != nil {
			return nil, err
		}
		o.Set(t.(string), v)
	}
}

func readValue(d *json.Decoder) (any, error) {
	t, err := d.Token()
	if err != nil {
		return nil, err
	}
	switch t {
	case json.Delim('{'):
		return readObj(d)
	case json.Delim('['):
		var l []any
		for {
			if d.More() {
				v, err := readValue(d)
				if err != nil {
					return nil, err
				}
				l = append(l, v)
				continue
			}
			if _, err := d.Token(); err != nil {
				return nil, err
			}
			if l == nil {
				l = []any{}
			}
			return l, nil
		}
	}
	return t, nil
}

// SortStable sorts objects by an integer key, keeping equal keys in the
// order they arrived, which is what list.sort does.
func SortStable(l []*Obj, key string) {
	sort.SliceStable(l, func(i, j int) bool { return l[i].Int(key) < l[j].Int(key) })
}

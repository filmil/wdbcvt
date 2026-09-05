// SPDX-License-Identifier: Apache-2.0

// Tier 70: the numbers an associative array takes under -debug all.
package main

import (
	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
	t60 "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gent60"
)

// unlogged declares a container, which gets a declaration and a handle
// and is never logged, as tier 60 measured.
func unlogged(names ...string) t60.Opt {
	var o t60.Opt
	for _, n := range names {
		o.Signals = append(o.Signals, c.Sig("tb", n, "", 32, "logged", false))
	}
	return o
}

func main() {
	// An associative array takes two numbers with a string key and
	// three with an int key, tier 61, and in both of those cases the
	// key is either a string, which takes no number, or the element's
	// own type. These pairs separate the key from the element.
	t60.Case("t70_num_a_v_str_", "an associative array of vectors keyed by string",
		"logic [3:0] a[string];", `a["k"] = 4'd5;`, "t60_dbg_assoc___", unlogged("a"))
	t60.Case("t70_num_a_i_byte", "an associative array of int keyed by byte",
		"int a[byte];", "a[8'd3] = 5;", "t60_dbg_assoc_i_", unlogged("a"))
	t60.Case("t70_num_a_b_str_", "an associative array of byte keyed by string",
		"byte a[string];", `a["k"] = 8'd5;`, "t60_dbg_assoc___", unlogged("a"))
	t60.Case("t70_num_a_b_int_", "an associative array of byte keyed by int",
		"byte a[int];", "a[3] = 8'd5;", "t70_num_a_b_str_", unlogged("a"))
	t60.Case("t70_num_a_e_key_", "an associative array keyed by an enumeration",
		"typedef enum logic [1:0] {A, B} e_t;\n    int a[e_t];", "a[A] = 5;", "t60_dbg_assoc_i_", unlogged("a"))
	t60.Case("t70_num_a_2dim__", "an associative array of two dimensions",
		"int a[string][int];", `a["k"][3] = 5;`, "t60_dbg_assoc___", unlogged("a"))
	t60.Case("t70_num_a_in_cls", "a class with an associative array field",
		"class c_t; int a[string]; endclass\n    c_t h;", "h = new;", "t60_dbg_class___",
		t60.Held("h", "c_t"))
	// A dynamic array beside a queue, to place the dynamic array in the
	// same numbering the containers of tier 61 walk.
	t60.Case("t70_num_d_then_q", "a dynamic array then a queue",
		"int d[];\n    int q[$];", "d = new[2]; q.push_back(5);", "t61_num_a_then_q", unlogged("d", "q"))
}

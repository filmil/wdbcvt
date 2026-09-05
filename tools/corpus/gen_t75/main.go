// SPDX-License-Identifier: Apache-2.0

// Tier 75: whether the two words of an access or file type entry move
// with the type they are over.
package main

import (
	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
)

const tb = `
--! @file
--! @brief Corpus case: %(brief)s
--!
--! Axis: %(axis)s

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    %(decl)s
begin
    p: process
        %(local)s
    begin
        wait for 50 ns;
        %(body)s
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
`

// vr is a process variable of the truth.
func vr(scope, name, typ string) *c.Obj {
	return c.O("scope", scope, "name", name, "type", typ, "kind", "variable")
}

func kase(name, brief, decl, local, body, differs string, vars []*c.Obj) {
	axis := "access and file entries. " + brief +
		", to see whether the two words after the designated or element type move with that type."
	src := c.Fill(tb, "brief", brief, "axis", axis, "decl", decl, "local", local, "body", body)
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files:     []c.File{{Name: "tb.ent.vhdl", Body: src}},
		Signals:   []*c.Obj{c.Sig("tb", "s", "std_ulogic", 1)},
		Trs:       []*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")},
		Extra:     c.O("variables", vars),
		End:       100,
		NoX:       true})
}

func main() {
	// An access over a composite, where a pointer to it might be fat.
	kase("t75_acc_rec_____", "an access to a record",
		`type rec_t is record
        a : std_ulogic;
        n : integer;
    end record;
    type rec_ptr is access rec_t;`,
		"variable p : rec_ptr;",
		`p := new rec_t'('1', 5);
        deallocate(p);`, "t23_access______", []*c.Obj{vr("tb.p", "p", "rec_ptr")})

	// An access over a constrained array of forty elements.
	kase("t75_acc_arr40___", "an access to a forty element array",
		`type arr_t is array (0 to 39) of integer;
    type arr_ptr is access arr_t;`,
		"variable p : arr_ptr;",
		`p := new arr_t'(others => 5);
        deallocate(p);`, "t75_acc_rec_____", []*c.Obj{vr("tb.p", "p", "arr_ptr")})

	// An access over another access type.
	kase("t75_acc_acc_____", "an access to an access type",
		`type int_ptr is access integer;
    type ptr_ptr is access int_ptr;`,
		"variable p : ptr_ptr;",
		`p := new int_ptr'(new integer'(5));
        deallocate(p);`, "t23_access______", []*c.Obj{vr("tb.p", "p", "ptr_ptr")})

	// A file of a record, where the element is composite.
	kase("t75_fil_rec_____", "a file of a record",
		`type rec_t is record
        a : std_ulogic;
        n : integer;
    end record;
    type rec_file is file of rec_t;
    file f : rec_file;`,
		"", "", "t23_file_int____", []*c.Obj{vr("tb", "f", "rec_file")})

	// A file of a constrained array.
	kase("t75_fil_arr_____", "a file of a forty element array",
		`type arr_t is array (0 to 39) of integer;
    type arr_file is file of arr_t;
    file f : arr_file;`,
		"", "", "t75_fil_rec_____", []*c.Obj{vr("tb", "f", "arr_file")})
}

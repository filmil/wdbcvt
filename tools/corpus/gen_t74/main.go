// SPDX-License-Identifier: Apache-2.0

// Tier 74: why log_wave -recursive * skips a package.
package main

import (
	"strings"

	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
)

const pkg = `
--! @file
--! @brief Corpus child entity: a signal declared in a package.

library ieee;
    use ieee.std_logic_1164.all;

package sig_pkg is
    signal g : std_ulogic := '0';
end package;
`

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
    signal x : std_ulogic := '0';
begin
    work.sig_pkg.g <= x;

    p: process
    begin
        wait for 10 ns;
        x <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
`

const tcl = `open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
puts "SCOPE: [current_scope]"
puts "SCOPES: [get_scopes -r /*]"
puts "OBJECTS: [get_objects -r /*]"
puts "PKGOBJ: [get_objects /sig_pkg/*]"
%s
run -all
close_vcd
exit
`

// kase writes one case: the same design and the same package, logged
// by the lines given. gLogged says whether the package signal comes
// out logged, and xLogged whether the architecture's own does, which a
// script that names the package alone leaves out.
func kase(name, brief, differs string, lines []string, gLogged bool, xLogged bool) {
	axis := "the log_wave pattern. " + brief +
		", to see why the default script's log_wave -recursive * leaves a package signal unlogged."
	x := c.Sig("tb", "x", "std_ulogic", 1)
	var sigs []*c.Obj
	var trs []*c.Obj
	if xLogged {
		sigs = append(sigs, x)
		trs = append(trs, c.Tr(0, "x", "0"), c.Tr(10, "x", "1"))
	} else {
		sigs = append(sigs, x.With("logged", false))
	}
	g := c.Sig("sig_pkg", "g", "std_ulogic", 1)
	if gLogged {
		sigs = append(sigs, g)
		trs = append(trs, c.Tr(0, "sig_pkg.g", "0"), c.Tr(10, "sig_pkg.g", "1"))
	} else {
		sigs = append(sigs, g.With("logged", false))
	}
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files: []c.File{
			{Name: "sig.pkg.vhdl", Body: c.Fill(pkg)},
			{Name: "tb.ent.vhdl", Body: c.Fill(tb, "brief", brief, "axis", axis)},
		},
		Signals: sigs, Trs: trs, End: 20, NoX: true})
	c.WriteFile(name, "xsim.tcl", strings.Replace(tcl, "%s", strings.Join(lines, "\n"), 1))
	c.PatchBuild(name, func(b string) string {
		b = strings.Replace(b, "    srcs = [\n", "    # Compilation order matters to xvhdl; do not sort.\n    # do not sort\n    srcs = [\n", 1)
		return strings.Replace(b, "    ],\n)\n", "    ],\n    tcl = \"xsim.tcl\",\n)\n", 1)
	})
}

func main() {
	// The default script, which every other case runs: the pattern has
	// no leading slash.
	kase("t74_lgw_star____", "log_wave -recursive * alone", "t13_pkg_log_all_",
		[]string{"log_vcd [get_objects -r /* ]", "log_wave -recursive *"}, false, true)
	// The same pattern rooted, which is the reading under test.
	kase("t74_lgw_root____", "log_wave -recursive /*", "t74_lgw_star____",
		[]string{"log_vcd [get_objects -r /* ]", "log_wave -recursive /*"}, false, true)
	// The current scope moved to the root first, with the pattern left
	// as it was.
	kase("t74_lgw_cur_root", "current_scope / before log_wave -recursive *", "t74_lgw_star____",
		[]string{"current_scope /", "log_vcd [get_objects -r /* ]", "log_wave -recursive *"}, false, true)
	// The objects named rather than matched.
	kase("t74_lgw_objects_", "log_wave over get_objects -r /*", "t74_lgw_star____",
		[]string{"log_vcd [get_objects -r /* ]", "log_wave [get_objects -r /*]"}, false, true)
	// The objects of the package, named through the package.
	kase("t74_lgw_pkg_obj_", "log_wave over get_objects /sig_pkg/*", "t74_lgw_star____",
		// Beside the default log_wave, so that the case differs from
		// the one below in how the package is named and in nothing
		// else.
		[]string{"log_vcd [get_objects -r /* ]", "log_wave -recursive *",
			"log_wave [get_objects /sig_pkg/*]"}, true, true)
	// The package named as a scope, which tier 13 measured.
	kase("t74_lgw_pkg_name", "log_wave -recursive /sig_pkg", "t74_lgw_star____",
		[]string{"log_vcd [get_objects -r /* ]", "log_wave -recursive *", "log_wave -recursive /sig_pkg"}, true, true)
}

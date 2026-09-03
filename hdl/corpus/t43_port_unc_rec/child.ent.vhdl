-- SPDX-License-Identifier: Apache-2.0

library ieee;
    use ieee.std_logic_1164.all;

--! @file
--! @brief An entity with a port of a record with an unconstrained field.

--! The port's bravo takes its bounds from the actual.
--!
--! ```
--! time  0 ns    10 ns
--! a     actual  actual
--! ```
entity child is
    port (
        a : in work.bundle_pkg.bundle_t
    );
end entity;

architecture sim of child is
    signal s : std_ulogic := '0';
begin
    s <= a.bravo(a.bravo'right);
end architecture;

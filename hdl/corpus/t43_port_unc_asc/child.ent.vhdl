-- SPDX-License-Identifier: Apache-2.0

library ieee;
    use ieee.std_logic_1164.all;

--! @file
--! @brief An entity with an unconstrained input port.

--! The port takes its bounds from the actual.
--!
--! ```
--! time  0 ns        10 ns
--! a     actual      actual
--! s     a(a'right)  a(a'right)
--! ```
entity child is
    port (
        a : in std_ulogic_vector
    );
end entity;

architecture sim of child is
    signal s : std_ulogic := '0';
begin
    s <= a(a'right);
end architecture;

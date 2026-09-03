-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an sfixed(3 downto -4) from ieee.fixed_pkg
--!
--! Axis: type. The IEEE fixed point package: a signed fixed point vector with a negative right bound.

library ieee;
    use ieee.std_logic_1164.all;
    use ieee.fixed_pkg.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : sfixed(3 downto -4) := to_sfixed(1.5, 3, -4);
begin
    p: process
    begin
        wait for 10 ns;
        s <= to_sfixed(-2.25, 3, -4);
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

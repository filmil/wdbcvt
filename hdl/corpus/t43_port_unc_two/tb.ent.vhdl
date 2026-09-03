-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two instances with different actual widths
--!
--! Axis: ports. dut bound to eight bits and dut2 to four, to see whether the unit and the declarations repeat as they do for different generics.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal v : std_ulogic_vector(7 downto 0) := x"A5";
    signal w : std_ulogic_vector(3 downto 0) := x"5";
begin
    dut: entity work.child port map (a => v);
    dut2: entity work.child port map (a => w);
    p: process
    begin
        wait for 10 ns;
        v <= x"5A";
        w <= x"A";
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

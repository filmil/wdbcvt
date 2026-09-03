-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a clock toggling every nanosecond for 430 nanoseconds.
--!
--! Axis: a clock toggling every nanosecond for 430 nanoseconds, VHDL. The records spill into a second page.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal clk : std_ulogic := '0';
begin
    clk <= not clk after 1 ns;
    p: process
    begin
        wait for 430 ns;
        std.env.stop;
    end process;
end architecture;

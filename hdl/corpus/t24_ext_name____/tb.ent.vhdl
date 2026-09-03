-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an external name
--!
--! Axis: a <= << signal .tb.dut.s : std_ulogic >> in tb

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal a : std_ulogic := '0';
begin
    dut: entity work.child;

    a <= << signal .tb.dut.s : std_ulogic >>;

    p: process
    begin
        wait for 30 ns;
        std.env.stop;
    end process;
end architecture;

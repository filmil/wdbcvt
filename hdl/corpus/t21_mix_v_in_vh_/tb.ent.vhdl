-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a Verilog child under a VHDL testbench
--!
--! Axis: the child of t9_comp written in Verilog

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    component child is
        port (
            a : in std_ulogic
        );
    end component;
    signal x : std_ulogic := '0';
begin
    dut: child port map (a => x);

    p: process
    begin
        wait for 50 ns;
        x <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a two port Verilog child under a VHDL testbench
--!
--! Axis: a Verilog child with an input and an output port under a VHDL testbench

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    component child is
        port (
            a : in std_ulogic;
            b : out std_ulogic
        );
    end component;
    signal x : std_ulogic := '0';
    signal y : std_ulogic;
begin
    dut: child port map (a => x, b => y);

    p: process
    begin
        wait for 50 ns;
        x <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

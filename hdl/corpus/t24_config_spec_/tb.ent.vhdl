-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a configuration specification
--!
--! Axis: component child with for dut : child use entity work.child(a), against the last analysed b

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    component child is
    end component;
    for dut : child use entity work.child(a);
begin
    dut: child;

    p: process
    begin
        wait for 30 ns;
        std.env.stop;
    end process;
end architecture;

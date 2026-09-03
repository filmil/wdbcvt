-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one instance of child under an if generate, and one under a false if generate.
--!
--! Axis: if generate. A true and a false condition, to see how an if generate scope is named and whether the false one leaves anything.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    constant with_dut : boolean := true;
begin
    g: if with_dut generate
        dut: entity work.child generic map (k => 1);
    end generate;
    h: if not with_dut generate
        dut: entity work.child generic map (k => 2);
    end generate;

    p: process
    begin
        wait for 20 ns;
        std.env.stop;
    end process;
end architecture;

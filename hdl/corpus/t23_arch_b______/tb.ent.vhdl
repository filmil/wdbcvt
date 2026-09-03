-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the second architecture
--!
--! Axis: dut: entity work.child(b), where the default would pick the last analysed

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
begin
    dut: entity work.child(b);

    p: process
    begin
        wait for 30 ns;
        std.env.stop;
    end process;
end architecture;

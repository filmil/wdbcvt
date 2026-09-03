-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one bit, three levels down.
--!
--! Axis: hierarchy depth, 1 to 3. Compared against t1_hier1 it says what
--! one extra empty scope costs, with the signal itself unchanged.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
begin
    dut: entity work.mid;

    p: process
    begin
        wait for 100 ns;
        std.env.stop;
    end process;
end architecture;

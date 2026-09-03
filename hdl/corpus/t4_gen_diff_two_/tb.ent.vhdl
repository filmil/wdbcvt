-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two instances binding DIFFERENT generic values.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
begin
    dut: entity work.child generic map (k => 4);
    dut2: entity work.child generic map (k => 7);

    p: process
    begin
        wait for 20 ns;
        std.env.stop;
    end process;
end architecture;

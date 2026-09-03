-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one instance binding the generic explicitly to 7.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
begin
    dut: entity work.child generic map (k => 7);

    p: process
    begin
        wait for 20 ns;
        std.env.stop;
    end process;
end architecture;

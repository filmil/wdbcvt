-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one instance taking the generic default.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
begin
    dut: entity work.child;

    p: process
    begin
        wait for 20 ns;
        std.env.stop;
    end process;
end architecture;

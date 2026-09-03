-- SPDX-License-Identifier: Apache-2.0
--! @file
--! @brief Corpus case: no signals at all.
--!
--! The floor of the ladder. Whatever a database costs when it holds
--! nothing, it costs here.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
begin
    p: process
    begin
        wait for 100 ns;
        std.env.stop;
    end process;
end architecture;

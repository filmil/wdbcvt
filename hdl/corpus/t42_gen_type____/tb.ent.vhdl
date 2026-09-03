-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an entity with a type generic
--!
--! Axis: generic. generic (type data_t; init, next_v : data_t) mapped to integer, 5 and 7.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is

begin
    u: entity work.child generic map (data_t => integer, init => 5, next_v => 7);
    p: process
    begin
        wait for 10 ns;
        null;
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

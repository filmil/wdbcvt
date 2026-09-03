-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a type generic mapped to std_ulogic
--!
--! Axis: generic. The same child with data_t => std_ulogic, init '0' and next_v '1'.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is

begin
    u: entity work.child generic map (data_t => std_ulogic, init => '0', next_v => '1');
    p: process
    begin
        wait for 10 ns;
        null;
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

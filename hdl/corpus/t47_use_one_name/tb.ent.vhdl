-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a bit signal under a use clause naming one item
--!
--! Axis: use clause. `use ieee.std_logic_1164.std_ulogic;` over t47_use_lib_only, one name against `.all`.

library ieee;
    use ieee.std_logic_1164.std_ulogic;

entity tb is
end entity;

architecture sim of tb is
    signal s : bit := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

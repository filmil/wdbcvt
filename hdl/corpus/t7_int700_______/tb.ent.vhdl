-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an integer, 700 transitions 1 ns apart.
--!
--! Axis: record size. 700 transitions of a 4 byte value, to separate a page limit in bytes from one in records.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : integer := 0;
begin
    p: process
    begin
        for i in 1 to 700 loop
            wait for 1 ns;
            s <= 165 when s = 0 else 0;
        end loop;
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

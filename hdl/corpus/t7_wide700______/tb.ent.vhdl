-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an eight bit vector, 700 transitions 1 ns apart.
--!
--! Axis: record size. 700 transitions of a 8 byte value, to separate a page limit in bytes from one in records.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic_vector(7 downto 0) := x"00";
begin
    p: process
    begin
        for i in 1 to 700 loop
            wait for 1 ns;
            s <= x"A5" when s = x"00" else x"00";
        end loop;
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

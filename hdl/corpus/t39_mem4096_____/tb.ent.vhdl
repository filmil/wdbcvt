-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a 4096 element memory of 8 bit vectors, loaded by a loop.
--!
--! Axis: the Potato instruction memory: 32768 bytes, a loop over 3840 elements in one delta, then one element.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type memory_array is array (natural range <>) of std_ulogic_vector(7 downto 0);
    signal m : memory_array(0 to 4095);
begin
    p: process
    begin
        for i in 256 to 4095 loop
            m(i) <= x"13";
        end loop;
        wait for 10 ns;
        m(512) <= x"37";
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

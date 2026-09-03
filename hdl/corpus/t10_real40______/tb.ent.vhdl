-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an array of 40 reals, 320 bytes, one change.
--!
--! Axis: element size. 320 bytes of 8 byte elements, against the 320 bytes of std_ulogic in t10_vec320, to see whether the chunk split depends on the element.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type real_array_t is array (0 to 39) of real;
    signal s : real_array_t := (others => 0.0);
begin
    p: process
    begin
        wait for 10 ns;
        s <= (others => 1.5);
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

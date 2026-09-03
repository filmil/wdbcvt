-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an array of two dimensional arrays
--!
--! Axis: array (0 to 1) of the (0 to 1, 0 to 2) array of t18_arr_2dim

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type mat_t is array (0 to 1, 0 to 2) of std_ulogic;
    type stack_t is array (0 to 1) of mat_t;
    signal s : stack_t := (others => (others => (others => '0')));
begin
    p: process
    begin
        wait for 50 ns;
        s <= ((('1', '0', '1'), ('0', '1', '1')), (('1', '1', '0'), ('0', '0', '1')));
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

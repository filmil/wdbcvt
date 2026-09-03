-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record with a two dimensional array field
--!
--! Axis: the (0 to 1, 0 to 2) array of t18_arr_2dim as a record field beside a std_ulogic

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type mat_t is array (0 to 1, 0 to 2) of std_ulogic;
    type rec_t is record
        a : std_ulogic;
        m : mat_t;
    end record;
    signal s : rec_t := ('0', (others => (others => '0')));
begin
    p: process
    begin
        wait for 50 ns;
        s <= ('1', (('1', '0', '1'), ('0', '1', '1')));
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

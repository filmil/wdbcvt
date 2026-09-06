-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an array of four records
--!
--! Axis: the four bytes of a record. an array of four records, to see how much handle space the static value of a record adds beyond the bytes the record declares.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type rec_t is record
        n : integer;
    end record;
    type arr_t is array (0 to 3) of rec_t;
    function f(c : std_ulogic) return std_ulogic is
        variable r : arr_t := (others => (n => 0));
        variable v : integer := 0;
    begin
        v := v + r(1).n;
        return c;
    end function;
begin
    p: process
    begin
        wait for 50 ns;
        s <= f('1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

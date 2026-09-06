-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record holding two records
--!
--! Axis: the four bytes of a record. a record holding two records, to see how much handle space the static value of a record adds beyond the bytes the record declares.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type inner_t is record
        n : integer;
    end record;
    type rec_t is record
        i : inner_t;
        j : inner_t;
    end record;
    function f(c : std_ulogic) return std_ulogic is
        variable r : rec_t := (i => (n => 0), j => (n => 0));
        variable v : integer := 0;
    begin
        v := v + r.i.n;
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

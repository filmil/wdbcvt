-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record variable parameter
--!
--! Axis: a procedure with an inout record variable parameter under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type rec_t is record
        a : std_ulogic;
        n : integer;
    end record;
    procedure set(variable r : inout rec_t) is
    begin
        r.a := '1';
        r.n := r.n + 1;
    end procedure;
begin
    p: process
        variable w : rec_t := ('0', 0);
    begin
        wait for 50 ns;
        set(w);
        s <= w.a;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record signal parameter
--!
--! Axis: a procedure with an out mode record signal parameter under -debug subprogram

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
    signal r : rec_t := ('0', 0);
    procedure set(signal q : out rec_t) is
    begin
        q <= ('1', 1);
    end procedure;
begin
    p: process
    begin
        wait for 50 ns;
        set(r);
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

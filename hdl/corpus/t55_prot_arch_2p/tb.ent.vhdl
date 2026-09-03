-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two processes calling an architecture protected type
--!
--! Axis: a shared variable of an architecture protected type, its methods called from two processes under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type counter_t is protected
        procedure bump;
        impure function get return integer;
    end protected;
    type counter_t is protected body
        variable n : integer := 0;
        procedure bump is
        begin
            n := n + 1;
        end procedure;
        impure function get return integer is
        begin
            return n;
        end function;
    end protected body;
    shared variable ct : counter_t;
begin
    p: process
        variable v : integer := 0;
    begin
        wait for 50 ns;
        ct.bump;
        v := ct.get;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
    p2: process
        variable w : integer := 0;
    begin
        wait for 20 ns;
        ct.bump;
        w := ct.get;
        wait;
    end process;
end architecture;

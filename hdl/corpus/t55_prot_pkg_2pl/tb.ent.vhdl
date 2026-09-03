-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two processes calling a package protected type, the second later
--!
--! Axis: a shared variable of a package protected type, its methods called from two processes, the second declared process calling last under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    shared variable ct : work.pk.counter_t;
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
        wait for 70 ns;
        ct.bump;
        w := ct.get;
        wait;
    end process;
end architecture;

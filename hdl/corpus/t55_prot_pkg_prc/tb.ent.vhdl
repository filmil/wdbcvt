-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a package protected type the process calls
--!
--! Axis: a shared variable of a package protected type, its methods called from the process under -debug subprogram

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
end architecture;

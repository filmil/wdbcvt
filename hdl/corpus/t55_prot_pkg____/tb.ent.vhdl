-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a protected type declared in a package
--!
--! Axis: a shared variable of a protected type declared in a package under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    shared variable ct : work.pk.counter_t;
    procedure drive(signal q : out std_ulogic) is
        variable v : integer := 0;
    begin
        ct.bump;
        v := ct.get;
        q <= '1';
    end procedure;
begin
    p: process
    begin
        wait for 50 ns;
        drive(s);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

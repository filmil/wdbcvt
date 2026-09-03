-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a protected type local of a procedure
--!
--! Axis: a procedure with a local variable of a protected type under -debug subprogram

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
    procedure drive(signal q : out std_ulogic) is
        variable ct : counter_t;
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

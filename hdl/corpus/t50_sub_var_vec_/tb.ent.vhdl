-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a vector variable parameter
--!
--! Axis: a procedure with an inout vector variable parameter under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    procedure flip(variable v : inout std_ulogic_vector(3 downto 0)) is
    begin
        v := not v;
    end procedure;
begin
    p: process
        variable w : std_ulogic_vector(3 downto 0) := "0000";
    begin
        wait for 50 ns;
        flip(w);
        s <= w(0);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

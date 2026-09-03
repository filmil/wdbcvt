-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a vector signal parameter
--!
--! Axis: a procedure with an out mode vector signal parameter under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal v : std_ulogic_vector(3 downto 0) := "0000";
    procedure drive(signal q : out std_ulogic_vector(3 downto 0)) is
    begin
        q <= "0001";
    end procedure;
begin
    p: process
    begin
        wait for 50 ns;
        drive(v);
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

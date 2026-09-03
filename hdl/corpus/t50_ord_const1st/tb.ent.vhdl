-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a constant declared before the signal
--!
--! Axis: an architecture constant declared above the signal, against t5_tr1000 where it is below

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    constant i : integer := 3;
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

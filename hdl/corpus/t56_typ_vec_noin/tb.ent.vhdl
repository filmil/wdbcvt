-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a vector local without an initialiser
--!
--! Axis: a std_ulogic_vector(3 downto 0) local, not initialised under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    function f(c : std_ulogic) return std_ulogic is
        variable w : std_ulogic_vector(3 downto 0);
        variable v : integer := 0;
    begin
        v := v + 1;
        w := (others => '0');
        w(0) := '1';
        return c;
    end function;
begin
    p: process
    begin
        wait for 50 ns;
        s <= f('1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

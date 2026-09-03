-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a 32 element vector local
--!
--! Axis: w : std_ulogic_vector(31 downto 0) in the function of t23_sub_sizes

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal n : integer := 0;
    function f(c : std_ulogic; n : integer) return integer is
        variable r : real := 1.5;
        variable w : std_ulogic_vector(31 downto 0) := x"00000000";
        variable m : integer := n;
    begin
        if c = '1' then
            m := m + 1;
        end if;
        return m;
    end function;
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        n <= f(s, 3);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

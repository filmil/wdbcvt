-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a vector local of a named subtype
--!
--! Axis: a subtype of std_ulogic_vector(3 downto 0) declared in the architecture, one local of it under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    subtype vec4_t is std_ulogic_vector(3 downto 0);
    function f(c : std_ulogic) return std_ulogic is
        variable w : vec4_t := (others => '0');
        variable v : integer := 0;
    begin
        v := v + 1;
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

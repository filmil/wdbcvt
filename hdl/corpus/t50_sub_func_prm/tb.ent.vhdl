-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a function with a vector constant parameter
--!
--! Axis: a function with a constrained vector parameter and a scalar local under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    function low(a : std_ulogic_vector(3 downto 0)) return std_ulogic is
        variable r : std_ulogic;
    begin
        r := a(0);
        return r;
    end function;
begin
    p: process
    begin
        wait for 50 ns;
        s <= low("0001");
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a child instantiated through a component declaration.
--!
--! Axis: instantiation. A component declaration and default binding instead of a direct entity instantiation, to see whether the component adds a scope or a unit.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    component child is
        port (
            a : in std_ulogic
        );
    end component;
    signal x : std_ulogic := '0';
begin
    dut: child port map (a => x);

    p: process
    begin
        wait for 10 ns;
        x <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

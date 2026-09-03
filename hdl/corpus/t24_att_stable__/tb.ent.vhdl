-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: s'stable
--!
--! Axis: b <= s'stable(1 ns)

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal b : boolean := true;
begin
    b <= s'stable(1 ns);
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

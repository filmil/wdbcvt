-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two drivers of a std_logic, none active at time 0
--!
--! Axis: r <= '1' from p at 50 ns and r <= '0' from q at 70 ns, q does not assign at time 0

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal r : std_logic := 'Z';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        r <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
    q: process
    begin
        wait for 70 ns;
        r <= '0';
        wait;
    end process;
end architecture;

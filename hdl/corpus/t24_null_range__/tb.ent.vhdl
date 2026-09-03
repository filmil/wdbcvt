-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a null range vector
--!
--! Axis: signal z : std_ulogic_vector(0 downto 1), a null range beside s

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal z : std_ulogic_vector(0 downto 1);
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

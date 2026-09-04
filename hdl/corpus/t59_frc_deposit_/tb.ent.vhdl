-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a deposit on the scalar
--!
--! Axis: forcing. The script deposits 1 on the scalar with set_value before the run, on a scalar driven 1 at 10 ns and 0 at 20 ns and a vector driven 0101 and 1010 at the same times, to see what the database records of a value the script imposes.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal v : std_ulogic_vector(3 downto 0) := "0000";
begin
    p: process
    begin
        wait for 10 ns;
        s <= '1';
        v <= "0101";
        wait for 10 ns;
        s <= '0';
        v <= "1010";
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

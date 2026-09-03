-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an out port bound to a record field
--!
--! Axis: child port a : out std_ulogic, port map (a => r.b), a <= '1' inside the child

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal r : work.trio_pkg.trio_t := ('0', '0', '0');
begin
    dut: entity work.child port map (a => r.b);
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

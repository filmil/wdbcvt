-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record out port with one field written in the child
--!
--! Axis: child port p : out trio_t, p.b <= '1' inside the child, port map (p => r)

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal r : work.trio_pkg.trio_t := ('0', '0', '0');
begin
    dut: entity work.child port map (p => r);
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

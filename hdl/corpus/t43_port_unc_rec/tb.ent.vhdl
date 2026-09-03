-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a port of a record with an unconstrained field
--!
--! Axis: ports. a : in bundle_t where bravo is unconstrained, bound to a signal declared bundle_t(bravo(7 downto 0)).

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal v : work.bundle_pkg.bundle_t(bravo(7 downto 0)) := ('0', x"A5");
begin
    dut: entity work.child port map (a => v);
    p: process
    begin
        wait for 10 ns;
        v <= ('1', x"5A");
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

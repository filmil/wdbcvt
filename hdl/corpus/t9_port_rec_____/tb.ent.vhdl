-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record signal connected to a record port, the type from a package.
--!
--! Axis: package and record port. The type and a constant come from a package, to see whether a package is a scope or a unit and how a record port is bound.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : work.pair_pkg.pair_t := work.pair_pkg.zero;
begin
    dut: entity work.child port map (a => x);

    p: process
    begin
        wait for 10 ns;
        x <= (alpha => '1', bravo => '1');
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

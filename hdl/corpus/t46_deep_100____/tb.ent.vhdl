-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a chain of 100 nested instances
--!
--! Axis: scale. An entity instantiating itself 100 levels deep under an if generate, to see how a long scope path and a deep tree are stored.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is

begin
    dut: entity work.chain generic map (n => 100);

    p: process
    begin
        wait for 110 ns;
        std.env.stop;
    end process;
end architecture;

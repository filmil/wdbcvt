-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: four instances of child under two nested for generates.
--!
--! Axis: for generate. A generate inside a generate, to see how nested iteration scopes are named and numbered.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
begin
    g: for i in 0 to 1 generate
        h: for j in 0 to 1 generate
            dut: entity work.child generic map (k => 2 * i + j);
        end generate;
    end generate;

    p: process
    begin
        wait for 20 ns;
        std.env.stop;
    end process;
end architecture;

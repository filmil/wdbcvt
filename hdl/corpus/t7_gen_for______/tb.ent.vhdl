-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: three instances of child under a for generate.
--!
--! Axis: for generate. Three instances with the loop index as the generic, to see how generate scopes are named and whether the index is an object.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
begin
    g: for i in 0 to 2 generate
        dut: entity work.child generic map (k => i);
    end generate;

    p: process
    begin
        wait for 20 ns;
        std.env.stop;
    end process;
end architecture;

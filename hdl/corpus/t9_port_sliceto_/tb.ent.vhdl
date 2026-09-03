-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an input port bound to element 0 of an ascending vector.
--!
--! Axis: port association. The same element index as t9_port_slice on a `to` range, to see whether the offset counts elements from the left or index values.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_ulogic_vector(0 to 1) := "00";
begin
    dut: entity work.child port map (a => x(0));

    p: process
    begin
        wait for 10 ns;
        x <= "10";
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a package constant, then a child with a signal and a process variable.
--!
--! Axis: logged objects. The first object is a package constant with no records, to see whether the marker names the first logged object rather than object zero.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
begin
    dut: entity work.child;

    p: process
    begin
        wait for 20 ns;
        std.env.stop;
    end process;
end architecture;

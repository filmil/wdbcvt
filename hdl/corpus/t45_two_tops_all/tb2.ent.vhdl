-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a second top entity.
--!
--! Axis: elaboration. A second top, to see how two roots are written.

library ieee;
    use ieee.std_logic_1164.all;

entity tb2 is
end entity;

architecture sim of tb2 is
    signal t : std_ulogic := '0';
begin
    p: process
    begin
        wait for 7 ns;
        t <= '1';
        wait for 33 ns;
        std.env.stop;
    end process;
end architecture;

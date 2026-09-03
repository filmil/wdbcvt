-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a procedure declared in the process, driving the signal directly.
--!
--! Axis: subprograms. A procedure inside the process with no signal parameter, to separate the cost of the parameter from the cost of the procedure.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_ulogic := '0';
begin
    p: process
        procedure flip is
            variable r : std_ulogic;
        begin
            r := not x;
            x <= r;
        end procedure;
    begin
        wait for 10 ns;
        flip;
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

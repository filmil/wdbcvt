-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two top level entities.
--!
--! Axis: elaboration. xelab --top corpus.tb2 --top corpus.tb under the default script, to see whether the root scope gets two children and what log_wave -recursive * logs when there are two.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 5 ns;
        s <= '1';
        wait for 10 ns;
        s <= '0';
        wait for 10 ns;
        s <= '1';
        wait for 5 ns;
        std.env.stop;
    end process;
end architecture;

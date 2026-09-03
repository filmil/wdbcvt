-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a protected type
--!
--! Axis: shared variable c of a protected type counter with one method

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type counter is protected
        procedure bump;
    end protected;
    type counter is protected body
        variable cnt : integer := 0;
        procedure bump is
        begin
            cnt := cnt + 1;
        end procedure;
    end protected body;
    shared variable c : counter;
begin
    p: process
    begin
        wait for 50 ns;
        c.bump;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

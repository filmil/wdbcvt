-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record holding a bit, a real and an integer.
--!
--! Axis: record field alignment. A one byte field followed by an eight byte one shows where the second is placed.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type mix_t is record
        alpha : std_ulogic;
        bravo : real;
        charlie : integer;
    end record;
    signal s : mix_t := (alpha => '0', bravo => 0.0, charlie => 0);
begin
    p: process
    begin
        wait for 50 ns;
        s <= (alpha => '1', bravo => 1.5, charlie => 165);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

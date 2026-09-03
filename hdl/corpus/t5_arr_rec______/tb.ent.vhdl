-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an array of three records.
--!
--! Axis: array element stride. A record element says whether elements are padded to the record's alignment.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type pair_t is record
        alpha : std_ulogic;
        charlie : integer;
    end record;
    type pair_array_t is array (0 to 2) of pair_t;
    signal s : pair_array_t := (others => (alpha => '0', charlie => 0));
begin
    p: process
    begin
        wait for 50 ns;
        s <= ((alpha => '1', charlie => 1), (alpha => '0', charlie => 2), (alpha => '1', charlie => 3));
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

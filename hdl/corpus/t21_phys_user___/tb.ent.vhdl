-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a user physical type
--!
--! Axis: type dist_t is range 0 to 1e9 units um; mm = 1000 um; m = 1000 mm; end units

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type dist_t is range 0 to 1_000_000_000 units
        um;
        mm = 1000 um;
        m = 1000 mm;
    end units;
    signal s : dist_t := 0 um;
begin
    p: process
    begin
        wait for 50 ns;
        s <= 3 mm;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

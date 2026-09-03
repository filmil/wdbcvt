-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a bit signal under `library ieee` with no use clause
--!
--! Axis: use clause. `library ieee;` alone over t47_use_none, to see whether the library clause or the use clause costs.

library ieee;

entity tb is
end entity;

architecture sim of tb is
    signal s : bit := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

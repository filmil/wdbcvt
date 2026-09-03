-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: no library clause, no signal and no std.env.stop
--!
--! Axis: no library clause, no signal and no std.env.stop


entity tb is
end entity;

architecture sim of tb is

begin
    p: process
        variable a : integer := 0;
    begin
        a := a + 1;
        wait for 50 ns;
        wait;
    end process;
end architecture;

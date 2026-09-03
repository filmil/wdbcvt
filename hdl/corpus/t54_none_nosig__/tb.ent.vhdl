-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: no library clause and no signal
--!
--! Axis: no library clause and no signal


entity tb is
end entity;

architecture sim of tb is

begin
    p: process
        variable a : integer := 0;
    begin
        a := a + 1;
        wait for 50 ns;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a bit signal with no library or use clause at all
--!
--! Axis: use clause. No `library ieee` and no use clause; the signal is a `bit`, to see what std_logic_1164 costs.


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

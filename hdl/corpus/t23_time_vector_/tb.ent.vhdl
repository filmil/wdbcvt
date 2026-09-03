-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a time_vector signal
--!
--! Axis: signal s : time_vector(0 to 3), the predefined vector

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : time_vector(0 to 3) := (others => 0 ps);
begin
    p: process
    begin
        wait for 50 ns;
        s <= (1 ns, 2 ps, 3 us, 0 ps);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

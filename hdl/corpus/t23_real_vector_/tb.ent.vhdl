-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a real_vector signal
--!
--! Axis: signal s : real_vector(0 to 3), the predefined vector

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : real_vector(0 to 3) := (others => 0.0);
begin
    p: process
    begin
        wait for 50 ns;
        s <= (1.5, -2.0, 300.25, 0.0);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

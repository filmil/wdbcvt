-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a boolean_vector signal
--!
--! Axis: signal s : boolean_vector(0 to 3), the predefined vector

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : boolean_vector(0 to 3) := (others => false);
begin
    p: process
    begin
        wait for 50 ns;
        s <= (true, false, true, true);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

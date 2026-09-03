-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an enumeration mixing identifiers and character literals
--!
--! Axis: type mix_t is (alpha, 'b', gamma)

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type mix_t is (alpha, 'b', gamma);
    signal s : mix_t := alpha;
begin
    p: process
    begin
        wait for 50 ns;
        s <= 'b';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an enumeration of one literal
--!
--! Axis: type one_t is (only)

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type one_t is (only);
    signal s : one_t := only;
begin
    p: process
    begin
        wait for 50 ns;
        s <= only;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an enumeration of two identifiers
--!
--! Axis: type flag_t is (no, yes), the shape of BOOLEAN under another name

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type flag_t is (no, yes);
    signal s : flag_t := no;
begin
    p: process
    begin
        wait for 50 ns;
        s <= yes;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

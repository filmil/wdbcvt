-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a TEXT file object
--!
--! Axis: file f : text, declared in the architecture and never opened

library ieee;
    use ieee.std_logic_1164.all;
    use std.textio.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    file f : text;
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: both architectures instantiated
--!
--! Axis: da: entity work.child(a) and db: entity work.child(b)

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
begin
    da: entity work.child(a);
    db: entity work.child(b);

    p: process
    begin
        wait for 30 ns;
        std.env.stop;
    end process;
end architecture;

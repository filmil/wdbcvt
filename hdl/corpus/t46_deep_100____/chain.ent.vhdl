-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an entity that instantiates itself while n > 0.
--!
--! Axis: scale. An entity instantiating itself 100 levels deep under an if generate, to see how a long scope path and a deep tree are stored.

library ieee;
    use ieee.std_logic_1164.all;

entity chain is
    generic (
        --! Levels below this one.
        n : natural
    );
end entity;

architecture sim of chain is
    signal s : std_ulogic := '0';
begin
    g: if n > 0 generate
        inner: entity work.chain generic map (n => n - 1);
    end generate;

    p: process
    begin
        wait for (n + 1) * 1 ns;
        s <= '1';
        wait;
    end process;
end architecture;

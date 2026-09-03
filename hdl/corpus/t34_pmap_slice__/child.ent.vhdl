-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: a vector out port, a slice written.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
    port (
        v : out std_ulogic_vector(3 downto 0) := "0000"
    );
end entity;

architecture sim of child is
begin
    q: process
    begin
        wait for 50 ns;
        v(1 downto 0) <= "11";
        wait;
    end process;
end architecture;

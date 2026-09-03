-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: a record out port, one field written.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
    port (
        p : out work.trio_pkg.trio_t := ('0', '0', '0')
    );
end entity;

architecture sim of child is
begin
    q: process
    begin
        wait for 50 ns;
        p.b <= '1';
        wait;
    end process;
end architecture;

-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: generics of four types other than integer.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
    generic (
        kb : boolean := true;
        ks : string := "abc";
        kv : std_ulogic_vector(3 downto 0) := x"A";
        kr : real := 1.5
    );
end entity;

architecture sim of child is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 10 ns;
        s <= '1';
        wait;
    end process;
end architecture;

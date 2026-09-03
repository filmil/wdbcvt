-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an unconstrained array of an unconstrained vector
--!
--! Axis: type. arr_t is array (natural range <>) of std_ulogic_vector, the signal arr_t(0 to 1)(3 downto 0).

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type arr_t is array (natural range <>) of std_ulogic_vector;
    signal s : arr_t(0 to 1)(3 downto 0) := (others => x"0");
begin
    p: process
    begin
        wait for 10 ns;
        s <= (x"5", x"A");
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

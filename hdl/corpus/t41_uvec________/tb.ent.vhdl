-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a user unconstrained vector type over std_ulogic
--!
--! Axis: type. array (natural range <>) of std_ulogic declared in the architecture, against std_ulogic_vector.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type vec_t is array (natural range <>) of std_ulogic;
    signal s : vec_t(7 downto 0) := (others => '0');
begin
    p: process
    begin
        wait for 10 ns;
        s <= x"18";
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

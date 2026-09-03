-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: string and vector generics of the top
--!
--! Axis: constrained vector generic
--!
--! The tier 40 cases share this bench and differ in the declaration of
--! kv and in the xelab options of BUILD.bazel.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
    generic (
        ks : string := "abc";
        kv : std_ulogic_vector(3 downto 0) := x"A"
    );
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 10 ns;
        s <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

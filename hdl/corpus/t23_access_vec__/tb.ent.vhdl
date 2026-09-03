-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an access to an unconstrained array
--!
--! Axis: type vec_ptr is access std_ulogic_vector; variable p : vec_ptr

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type vec_ptr is access std_ulogic_vector;
begin
    p: process
        variable p : vec_ptr;
    begin
        wait for 50 ns;
        p := new std_ulogic_vector'(x"5a");
        s <= '1';
        deallocate(p);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

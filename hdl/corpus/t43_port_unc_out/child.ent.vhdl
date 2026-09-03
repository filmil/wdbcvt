-- SPDX-License-Identifier: Apache-2.0

library ieee;
    use ieee.std_logic_1164.all;

--! @file
--! @brief An entity with an unconstrained output port.

--! The port takes its bounds from the actual and drives all ones
--! after 10 ns.
--!
--! ```
--! time  0 ns   10 ns
--! a     0..0   1..1
--! ```
entity child is
    port (
        a : out std_ulogic_vector
    );
end entity;

architecture sim of child is
begin
    p: process
    begin
        a <= (others => '0');
        wait for 10 ns;
        a <= (others => '1');
        wait;
    end process;
end architecture;

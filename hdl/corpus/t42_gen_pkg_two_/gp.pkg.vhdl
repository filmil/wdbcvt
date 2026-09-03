-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief A generic package and an instance of it.

library ieee;
    use ieee.std_logic_1164.all;

--! A package whose word width is a generic.
package gp is
    generic (n : natural);
    subtype word_t is std_ulogic_vector(n - 1 downto 0);
end package;

--! The eight bit instance.
package gp8 is new work.gp generic map (n => 8);

--! The four bit instance.
package gp4 is new work.gp generic map (n => 4);

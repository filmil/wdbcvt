-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief A package of two null range constants, the shape of numeric_std's NAU and NAS.

library ieee;
    use ieee.std_logic_1164.all;

package nul_pkg is
    constant n1 : std_ulogic_vector(0 downto 1) := (others => '0');
    constant n2 : std_ulogic_vector(0 downto 1) := (others => '0');
end package;

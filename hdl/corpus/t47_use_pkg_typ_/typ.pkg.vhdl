-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief A package of one subtype and nothing else.

library ieee;
    use ieee.std_logic_1164.all;

package typ_pkg is
    subtype nibble_t is std_ulogic_vector(3 downto 0);
end package;

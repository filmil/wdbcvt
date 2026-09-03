-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief A plain package with one subtype.

library ieee;
    use ieee.std_logic_1164.all;

--! A package with one subtype and nothing else.
package pp is
    subtype word_t is std_ulogic_vector(7 downto 0);
end package;

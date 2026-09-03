-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief A package of four unconstrained array types, the shape of numeric_std's types.

library ieee;
    use ieee.std_logic_1164.all;

package arr_pkg is
    type a1_t is array (natural range <>) of std_ulogic;
    type a2_t is array (natural range <>) of std_ulogic;
    type a3_t is array (natural range <>) of std_ulogic;
    type a4_t is array (natural range <>) of std_ulogic;
end package;

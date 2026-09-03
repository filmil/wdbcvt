-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief A package of two constants and nothing else.

library ieee;
    use ieee.std_logic_1164.all;

package two_pkg is
    constant c1 : integer := 7;
    constant c2 : std_ulogic := '1';
end package;

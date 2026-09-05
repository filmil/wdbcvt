-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: a signal declared in a package.

library ieee;
    use ieee.std_logic_1164.all;

package sig_pkg is
    signal g : std_ulogic := '0';
end package;

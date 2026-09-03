-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: a record type and a constant.

library ieee;
    use ieee.std_logic_1164.all;

package pair_pkg is
    type pair_t is record
        alpha : std_ulogic;
        bravo : std_ulogic;
    end record;
    constant zero : pair_t := (alpha => '0', bravo => '0');
end package;

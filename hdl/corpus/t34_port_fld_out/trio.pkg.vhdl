-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus package: a record of three std_ulogic fields.

library ieee;
    use ieee.std_logic_1164.all;

package trio_pkg is
    type trio_t is record
        a : std_ulogic;
        b : std_ulogic;
        c : std_ulogic;
    end record;
end package;

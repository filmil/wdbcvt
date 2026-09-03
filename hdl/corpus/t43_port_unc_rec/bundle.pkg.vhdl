-- SPDX-License-Identifier: Apache-2.0

library ieee;
    use ieee.std_logic_1164.all;

--! @file
--! @brief A record with an unconstrained field, for a port.

--! The record type of the port.
package bundle_pkg is
    type bundle_t is record
        alpha : std_ulogic;
        bravo : std_ulogic_vector;
    end record;
end package;

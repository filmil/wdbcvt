-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a package with a procedure.

library ieee;
    use ieee.std_logic_1164.all;

--! A package holding one procedure.
package pk is
    --! Drive q with v.
    procedure drive(signal q : out std_ulogic; constant v : in std_ulogic);
end package;

package body pk is
    procedure drive(signal q : out std_ulogic; constant v : in std_ulogic) is
    begin
        q <= v;
    end procedure;
end package body;

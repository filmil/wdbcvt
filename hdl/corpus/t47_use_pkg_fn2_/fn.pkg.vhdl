-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief A package of two functions and nothing else.

library ieee;
    use ieee.std_logic_1164.all;

package fn_pkg is
    function inc(x : integer) return integer;
    function inv(x : std_ulogic) return std_ulogic;
end package;

package body fn_pkg is
    function inc(x : integer) return integer is
    begin
        return x + 1;
    end function;
    function inv(x : std_ulogic) return std_ulogic is
    begin
        return not x;
    end function;
end package body;

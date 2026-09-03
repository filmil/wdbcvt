-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief A package of two procedures and nothing else.

library ieee;
    use ieee.std_logic_1164.all;

package pr_pkg is
    procedure inc(x : inout integer);
    procedure inv(x : inout std_ulogic);
end package;

package body pr_pkg is
    procedure inc(x : inout integer) is
    begin
        x := x + 1;
    end procedure;
    procedure inv(x : inout std_ulogic) is
    begin
        x := not x;
    end procedure;
end package body;

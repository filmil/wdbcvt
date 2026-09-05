-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a package whose body declares one integer constant
--!
--! Axis: a package body in the handle space. a package whose body declares one integer constant, to see whether what a package body declares lands in the package's block or past the second region.

package pk is
    function f return integer;
end package;

package body pk is
    constant b0 : integer := 1;
    function f return integer is
    begin
        return b0;
    end function;
end package body;

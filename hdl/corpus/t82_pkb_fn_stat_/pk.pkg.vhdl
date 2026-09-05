-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a package whose function has an array local with a literal value
--!
--! Axis: a package body in the handle space. a package whose function has an array local with a literal value, to see whether what a package body declares lands in the package's block or past the second region.

package pk is
    function f return integer;
end package;

package body pk is
    type arr_t is array (0 to 3) of integer;

    function g return integer is
        variable ar : arr_t := (others => 2);
    begin
        return ar(1);
    end function;
    function f return integer is
    begin
        return g;
    end function;
end package body;

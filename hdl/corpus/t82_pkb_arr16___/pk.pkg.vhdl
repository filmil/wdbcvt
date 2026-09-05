-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a package whose body declares a constant array of sixteen integers
--!
--! Axis: a package body in the handle space. a package whose body declares a constant array of sixteen integers, to see whether what a package body declares lands in the package's block or past the second region.

package pk is
    function f return integer;
end package;

package body pk is
    type arr_t is array (0 to 15) of integer;
    constant t : arr_t := (others => 7);
    function f return integer is
    begin
        return t(0);
    end function;
end package body;

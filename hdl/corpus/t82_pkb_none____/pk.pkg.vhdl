-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a package whose body declares nothing
--!
--! Axis: a package body in the handle space. a package whose body declares nothing, to see whether what a package body declares lands in the package's block or past the second region.

package pk is
    function f return integer;
end package;

package body pk is

    function f return integer is
    begin
        return 3;
    end function;
end package body;

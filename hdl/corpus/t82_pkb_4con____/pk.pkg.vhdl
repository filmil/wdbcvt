-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a package whose body declares four integer constants
--!
--! Axis: a package body in the handle space. a package whose body declares four integer constants, to see whether what a package body declares lands in the package's block or past the second region.

package pk is
    function f return integer;
end package;

package body pk is
    constant b0 : integer := 1;
    constant b1 : integer := 2;
    constant b2 : integer := 3;
    constant b3 : integer := 4;
    function f return integer is
    begin
        return b0 + b1 + b2 + b3;
    end function;
end package body;

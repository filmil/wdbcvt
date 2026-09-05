-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a package with a deferred constant
--!
--! Axis: a package body in the handle space. a package with a deferred constant, to see whether what a package body declares lands in the package's block or past the second region.

package pk is
    constant d : integer;
    function f return integer;
end package;

package body pk is
    constant d : integer := 5;
    function f return integer is
    begin
        return d;
    end function;
end package body;

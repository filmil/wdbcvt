-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a package with one function
--!
--! Axis: where a package sits in the handle space. no package, read from the process, to see whether the package moves the generic and the process variable that come after the signals.

package pk is
    function f return integer;
end package;

package body pk is
    function f return integer is
    begin
        return 3;
    end function;
end package body;

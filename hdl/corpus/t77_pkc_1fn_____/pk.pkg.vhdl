-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one function and no constant
--!
--! Axis: package handle space. A package that declares one function and no constant, read from the process, to see what each kind of declaration costs the handle space.

package pk is
    function f return integer;
end package;

package body pk is
    function f return integer is
    begin
        return 3;
    end function;
end package body;

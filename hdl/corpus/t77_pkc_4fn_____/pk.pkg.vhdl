-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: four functions and no constant
--!
--! Axis: package handle space. A package that declares four functions and no constant, read from the process, to see what each kind of declaration costs the handle space.

package pk is
    function f return integer;
    function g return integer;
    function h return integer;
    function i return integer;
end package;

package body pk is
    function f return integer is
    begin
        return 3;
    end function;
    function g return integer is
    begin
        return 4;
    end function;
    function h return integer is
    begin
        return 5;
    end function;
    function i return integer is
    begin
        return 6;
    end function;
end package body;

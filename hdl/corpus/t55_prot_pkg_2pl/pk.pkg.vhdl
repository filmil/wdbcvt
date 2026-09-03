-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a protected type declared in a package
--!
--! Axis: a shared variable of a package protected type, its methods called from two processes, the second declared process calling last under -debug subprogram

package pk is
    type counter_t is protected
        procedure bump;
        impure function get return integer;
    end protected;
end package;

package body pk is
    type counter_t is protected body
        variable n : integer := 0;
        procedure bump is
        begin
            n := n + 1;
        end procedure;
        impure function get return integer is
        begin
            return n;
        end function;
    end protected body;
end package body;

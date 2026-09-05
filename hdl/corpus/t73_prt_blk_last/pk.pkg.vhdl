-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a block statement
--!
--! Axis: protected method scopes. A shared variable of a protected type, its methods called from the only process, with a block statement after that process, to see which scope the second pair of method scopes hangs from.

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

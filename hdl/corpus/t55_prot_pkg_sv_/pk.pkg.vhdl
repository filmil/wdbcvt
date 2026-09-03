-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a protected type declared in a package
--!
--! Axis: a shared variable of a protected type, both declared in the package, called through package subprograms under -debug subprogram

package pk is
    type counter_t is protected
        procedure bump;
        impure function get return integer;
    end protected;
    shared variable ct : counter_t;
    procedure bump;
    impure function get return integer;
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
    procedure bump is
    begin
        ct.bump;
    end procedure;
    impure function get return integer is
    begin
        return ct.get;
    end function;
end package body;

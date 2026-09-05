-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one integer constant
--!
--! Axis: package handle space. A package that declares one integer constant, read from the process, to see what each kind of declaration costs the handle space.

package pk is
    constant c0 : integer := 1;
end package;

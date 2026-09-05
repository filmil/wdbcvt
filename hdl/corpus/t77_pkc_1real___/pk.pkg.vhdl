-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one real constant
--!
--! Axis: package handle space. A package that declares one real constant, read from the process, to see what each kind of declaration costs the handle space.

package pk is
    constant r : real := 1.5;
end package;

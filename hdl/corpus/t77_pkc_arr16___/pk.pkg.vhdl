-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a constant array of sixteen integers
--!
--! Axis: package handle space. A package that declares a constant array of sixteen integers, read from the process, to see what each kind of declaration costs the handle space.

package pk is
    type arr_t is array (0 to 15) of integer;
    constant t : arr_t := (others => 7);
end package;

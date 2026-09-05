-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a package with a constant array
--!
--! Axis: where a package sits in the handle space. no package, read from the process, to see whether the package moves the generic and the process variable that come after the signals.

package pk is
    type arr_t is array (0 to 15) of integer;
    constant t : arr_t := (others => 7);
end package;

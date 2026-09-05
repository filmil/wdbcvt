-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one type and no object
--!
--! Axis: package handle space. A package that declares one type and no object, read from the process, to see what each kind of declaration costs the handle space.

package pk is
    type small_t is range 0 to 7;
end package;

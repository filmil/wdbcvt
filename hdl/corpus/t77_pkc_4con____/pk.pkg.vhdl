-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: four integer constants
--!
--! Axis: package handle space. A package that declares four integer constants, read from the process, to see what each kind of declaration costs the handle space.

package pk is
    constant c0 : integer := 1;
    constant c1 : integer := 2;
    constant c2 : integer := 3;
    constant c3 : integer := 4;
end package;

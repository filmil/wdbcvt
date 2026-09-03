-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a float32 from ieee.float_pkg
--!
--! Axis: type. The IEEE floating point package: float32 is float(8 downto -23).

library ieee;
    use ieee.std_logic_1164.all;
    use ieee.float_pkg.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : float32 := to_float(1.5);
begin
    p: process
    begin
        wait for 10 ns;
        s <= to_float(-2.25);
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

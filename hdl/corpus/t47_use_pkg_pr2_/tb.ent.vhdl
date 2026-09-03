-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the baseline with a use of a package of two procedures
--!
--! Axis: use clause. `use work.pr_pkg.all`, two procedures with a body and no object.

library ieee;
    use ieee.std_logic_1164.all;
    use work.pr_pkg.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

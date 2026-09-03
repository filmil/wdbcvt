-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: one record input port.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
    port (
        a : in work.pair_pkg.pair_t
    );
end entity;

architecture sim of child is
begin
end architecture;

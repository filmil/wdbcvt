-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief A child entity with no ports and one signal.

library ieee;
    use ieee.std_logic_1164.all;

entity solo is
end entity;

architecture sim of solo is
    signal g : std_ulogic := '0';
begin
    g <= '1' after 50 ns;
end architecture;

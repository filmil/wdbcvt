-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief The child entity, whose architecture declares one signal.

library ieee;
    use ieee.std_logic_1164.all;

entity kid is
    port (
        i : in std_ulogic
    );
end entity;

architecture sim of kid is
    signal g : std_ulogic := '0';
begin
    g <= i;
end architecture;

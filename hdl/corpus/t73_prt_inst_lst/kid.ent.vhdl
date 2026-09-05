-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief The child instantiated after the process.

library ieee;
    use ieee.std_logic_1164.all;

entity kid is
    port (
        i : in std_ulogic
    );
end entity;

architecture sim of kid is
begin
end architecture;

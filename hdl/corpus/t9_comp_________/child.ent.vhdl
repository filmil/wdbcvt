-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: one input port and nothing else.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
    port (
        a : in std_ulogic
    );
end entity;

architecture sim of child is
begin
end architecture;

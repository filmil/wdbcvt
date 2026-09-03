-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: one inout port, driven to Z inside and read.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
    port (
        a : inout std_logic
    );
end entity;

architecture sim of child is
begin
    a <= 'Z';
end architecture;

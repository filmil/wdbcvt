-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: one input port and one output port, both left open.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
    port (
        a : in std_ulogic := '1';
        q : out std_ulogic
    );
end entity;

architecture sim of child is
begin
    q <= a;
end architecture;

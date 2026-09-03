-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: an input port passed down to an inner entity's input port.

library ieee;
    use ieee.std_logic_1164.all;

entity inner is
    port (
        b : in std_ulogic
    );
end entity;

architecture sim of inner is
begin
end architecture;

library ieee;
    use ieee.std_logic_1164.all;

entity child is
    port (
        a : in std_ulogic
    );
end entity;

architecture sim of child is
begin
    u: entity work.inner port map (b => a);
end architecture;

-- SPDX-License-Identifier: Apache-2.0

library ieee;
    use ieee.std_logic_1164.all;

entity leaf is
    port (
        a : in std_ulogic;
        b : out std_ulogic
    );
end entity;

architecture rtl of leaf is
begin
    b <= a;
end architecture;

-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record of two fields.
--!
--! Axis: field count, 3 to 2, against t2_record. Isolates what one
--! record field costs.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type bundle_t is record
        alpha : std_ulogic;
        bravo : std_ulogic_vector(7 downto 0);
    end record;
    signal s : bundle_t := (alpha => '0', bravo => x"00");
begin
    p: process
    begin
        wait for 50 ns;
        s <= (alpha => '1', bravo => x"A5");
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

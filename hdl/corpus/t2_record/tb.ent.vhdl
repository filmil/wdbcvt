-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record signal with three fields.
--!
--! Axis: aggregate structure. Answers whether a record is stored as one
--! object or flattened into one object per field, and whether the field
--! names survive. The field names are distinctive so they can be found
--! in the file.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    --! One field of each of the three shapes a port record usually mixes.
    type bundle_t is record
        alpha : std_ulogic;
        bravo : std_ulogic_vector(7 downto 0);
        charlie : integer;
    end record;
    signal s : bundle_t := (alpha => '0', bravo => x"00", charlie => 0);
begin
    p: process
    begin
        wait for 50 ns;
        s <= (alpha => '1', bravo => x"A5", charlie => 165);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

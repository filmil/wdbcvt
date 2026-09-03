-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: three separate signals, matching t2_record's fields.
--!
--! Axis: aggregation. The names, the types, the values and the times are
--! identical to the three fields of the record in t2_record. The only
--! difference is that here they are three signals rather than one record.
--! The comparison says whether a record is stored as one object or is
--! flattened into one object per field.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal alpha : std_ulogic := '0';
    signal bravo : std_ulogic_vector(7 downto 0) := x"00";
    signal charlie : integer := 0;
begin
    p: process
    begin
        wait for 50 ns;
        alpha <= '1';
        bravo <= x"A5";
        charlie <= 165;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

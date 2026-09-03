-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a user-defined enumeration with one transition.
--!
--! Axis: signal type, to a type the database cannot know in advance.
--! The literal names are distinctive, so it is visible whether they are
--! stored in the file or reduced to an index.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    --! Three literals with names that appear nowhere else.
    type colour_t is (crimson, viridian, cobalt);
    signal s : colour_t := crimson;
begin
    p: process
    begin
        wait for 50 ns;
        s <= cobalt;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;

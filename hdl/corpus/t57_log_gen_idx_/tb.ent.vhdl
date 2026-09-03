-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: log_wave naming the index of one generate iteration of a design with every kind of object
--!
--! Axis: logging. log_wave names the index of one generate iteration, in a design with a scalar, a vector, a record, a constant, a shared variable, a generate with a signal, and a process with a variable and a loop, to see what the database logs.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type rec_t is record
        a : std_ulogic;
        n : integer;
    end record;
    signal s : std_ulogic := '0';
    signal v : std_ulogic_vector(3 downto 0) := "0000";
    signal r : rec_t := ('0', 0);
    constant c : integer := 3;
    shared variable sv : integer := 1;
begin
    g: for i in 0 to 1 generate
        signal gs : std_ulogic := '0';
    begin
        gs <= s;
    end generate;
    p: process
        variable w : integer := 7;
    begin
        for k in 0 to 2 loop
            w := w + k;
        end loop;
        wait for 10 ns;
        s <= '1';
        v <= "0101";
        r <= ('1', 5);
        sv := 2;
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;

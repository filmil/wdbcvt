-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief An entity with a type generic.

--! An entity whose signal type is a generic.
--!
--! ```
--! time  0 ns   10 ns
--! s     init   next_v
--! ```
entity child is
    generic (
        type data_t;
        init : data_t;
        next_v : data_t
    );
end entity;

architecture sim of child is
    signal s : data_t := init;
begin
    p: process
    begin
        wait for 10 ns;
        s <= next_v;
        wait;
    end process;
end architecture;

-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief A small first in, first out character queue.
--!
--! Public API elements:
--!
--! * @ref fifo  the queue entity

library ieee;
    use ieee.std_logic_1164.all;

use work.uart_types;

--! @brief Holds up to `depth` characters in arrival order.
--!
--! A character enters on a rising edge of `sys.clk` for which
--! `put.valid` is high and `stat.full` is low. The oldest character
--! shows on `stat.head` while `stat.empty` is low, and leaves on a
--! rising edge for which `rd.take` is high. A put and a take in one
--! cycle are both honoured.
--!
--! A typical transaction, one character in and out:
--!
--! ```
--!             __    __    __    __    __    __    __
--! clk      __/  \__/  \__/  \__/  \__/  \__/  \__/  \__
--!                _____
--! valid    _____/     \_______________________________
--!          ___________
--! empty               \_________________________/_____
--!                                      _____
--! take     ___________________________/     \_________
--! count    ___0_______X______1______________X____0____
--! ```
entity fifo is
    generic (
        --! Number of characters the queue holds.
        depth : positive := 4
    );
    port (
        --! Clock and reset.
        sys : in uart_types.sys_t;
        --! The character to enqueue.
        put : in uart_types.beat_t;
        --! The dequeue request.
        rd : in uart_types.take_t;
        --! The oldest character and the fill state.
        stat : out uart_types.fifo_stat_t
    );
end entity;

architecture rtl of fifo is
    type mem_t is array (0 to depth - 1) of uart_types.data_t;
    signal mem : mem_t := (others => (others => '0'));
    signal wr_ptr : natural range 0 to depth - 1 := 0;
    signal rd_ptr : natural range 0 to depth - 1 := 0;
    signal count : natural range 0 to depth := 0;
    signal empty : std_ulogic := '1';
    signal full : std_ulogic := '0';
begin

    stat.head <= mem(rd_ptr);
    stat.empty <= empty;
    stat.full <= full;
    stat.count <= count;

    p_fifo: process (sys.clk)
        variable putting : boolean;
        variable taking : boolean;
    begin
        if rising_edge(sys.clk) then
            if sys.reset = '1' then
                wr_ptr <= 0;
                rd_ptr <= 0;
                count <= 0;
                empty <= '1';
                full <= '0';
            else
                putting := put.valid = '1' and full = '0';
                taking := rd.take = '1' and empty = '0';
                if putting then
                    mem(wr_ptr) <= put.data;
                    if wr_ptr = depth - 1 then
                        wr_ptr <= 0;
                    else
                        wr_ptr <= wr_ptr + 1;
                    end if;
                end if;
                if taking then
                    if rd_ptr = depth - 1 then
                        rd_ptr <= 0;
                    else
                        rd_ptr <= rd_ptr + 1;
                    end if;
                end if;
                if putting and not taking then
                    count <= count + 1;
                    empty <= '0';
                    if count = depth - 1 then
                        full <= '1';
                    end if;
                elsif taking and not putting then
                    count <= count - 1;
                    full <= '0';
                    if count = 1 then
                        empty <= '1';
                    end if;
                end if;
            end if;
        end if;
    end process;

end architecture;

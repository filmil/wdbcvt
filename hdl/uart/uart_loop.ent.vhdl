-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief A transmitter looped back into a receiver feeding a queue.
--!
--! Public API elements:
--!
--! * @ref uart_loop  the loopback entity

library ieee;
    use ieee.std_logic_1164.all;

use work.uart_types;

--! @brief Sends each beat over a serial line to itself and queues it.
--!
--! Characters given on `req` go out of a @ref uart_tx, whose line is
--! wired straight into a @ref uart_rx, whose beats fill a @ref fifo.
--! The queue is read through `rd` and shows on `stat`.
--!
--! A typical transaction, at the level of characters:
--!
--! ```
--!               _____
--! req.valid ___/     \______________________________________
--!                                              _____
--! rx valid  __________________________________/     \_______
--!           __________________________________________
--! empty                                               \_____
--! ```
entity uart_loop is
    generic (
        --! Clock cycles per bit of the serial line.
        clocks_per_bit : positive := 16;
        --! Number of characters the queue holds.
        depth : positive := 4
    );
    port (
        --! Clock and reset.
        sys : in uart_types.sys_t;
        --! The character to send.
        req : in uart_types.beat_t;
        --! Ready to send.
        tx_stat : out uart_types.tx_stat_t;
        --! The dequeue request.
        rd : in uart_types.take_t;
        --! The oldest received character and the fill state.
        stat : out uart_types.fifo_stat_t
    );
end entity;

architecture rtl of uart_loop is
    signal wire : uart_types.line_t := (rxd => '1');
    signal tx_out : uart_types.tx_stat_t;
    signal rx_out : uart_types.rx_stat_t;
begin

    tx_stat <= tx_out;
    wire.rxd <= tx_out.txd;

    u_tx: entity work.uart_tx
        generic map (clocks_per_bit => clocks_per_bit)
        port map (sys => sys, req => req, stat => tx_out);

    u_rx: entity work.uart_rx
        generic map (clocks_per_bit => clocks_per_bit)
        port map (sys => sys, line => wire, stat => rx_out);

    u_fifo: entity work.fifo
        generic map (depth => depth)
        port map (sys => sys, put => rx_out.beat, rd => rd, stat => stat);

end architecture;

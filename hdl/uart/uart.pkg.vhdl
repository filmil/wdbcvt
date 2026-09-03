-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Signal bundles exchanged between the UART entities.
--!
--! Public API elements:
--!
--! * @ref uart_types.data_t     one character
--! * @ref uart_types.sys_t      clock and reset
--! * @ref uart_types.beat_t     one character with a valid strobe
--! * @ref uart_types.tx_stat_t  what a transmitter drives out
--! * @ref uart_types.line_t     the serial line into a receiver
--! * @ref uart_types.rx_stat_t  what a receiver drives out
--! * @ref uart_types.take_t     the read side request of a FIFO
--! * @ref uart_types.fifo_stat_t  what a FIFO drives out

library ieee;
    use ieee.std_logic_1164.all;

--! @brief Types shared by the UART entities and their users.
--!
--! Ports never carry loose `std_ulogic` signals. Everything related is
--! bundled into a record here.
package uart_types is

    --! Width of a character, in bits.
    constant data_width : natural := 8;

    --! One character.
    subtype data_t is std_ulogic_vector(data_width - 1 downto 0);

    --! Clock and reset, driven into every entity.
    type sys_t is record
        --! Rising edge clock.
        clk : std_ulogic;
        --! Synchronous, active high reset.
        reset : std_ulogic;
    end record;

    --! One character with its strobe.
    type beat_t is record
        --! High for the one cycle in which @ref data is to be taken.
        valid : std_ulogic;
        --! The character.
        data : data_t;
    end record;

    --! Everything driven out of a @ref uart_tx.
    type tx_stat_t is record
        --! High while the transmitter can accept a beat.
        ready : std_ulogic;
        --! The serial line, idle high.
        txd : std_ulogic;
    end record;

    --! The serial line, driven into a @ref uart_rx.
    type line_t is record
        --! The serial line, idle high.
        rxd : std_ulogic;
    end record;

    --! Everything driven out of a @ref uart_rx.
    type rx_stat_t is record
        --! The received character, valid for one cycle.
        beat : beat_t;
        --! High with @ref beat when the stop bit was low.
        frame_error : std_ulogic;
    end record;

    --! The read side request of a @ref fifo.
    type take_t is record
        --! High for one cycle to take the oldest character.
        take : std_ulogic;
    end record;

    --! Everything driven out of a @ref fifo.
    type fifo_stat_t is record
        --! The oldest character, valid while @ref empty is low.
        head : data_t;
        --! High while the FIFO holds nothing.
        empty : std_ulogic;
        --! High while the FIFO cannot accept a character.
        full : std_ulogic;
        --! The number of characters held.
        count : natural;
    end record;

end package;

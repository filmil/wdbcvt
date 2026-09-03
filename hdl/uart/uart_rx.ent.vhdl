-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief A UART receiver: 8 data bits, no parity, one stop bit.
--!
--! Public API elements:
--!
--! * @ref uart_rx  the receiver entity

library ieee;
    use ieee.std_logic_1164.all;

use work.uart_types;

--! @brief Deserialises the line, sampling each bit at its middle.
--!
--! A falling edge on `line.rxd` while idle starts a frame. The receiver
--! waits half a bit, checks the start bit is still low, then samples
--! eight data bits one bit time apart, least significant bit first, and
--! the stop bit. `stat.beat.valid` is high for the one cycle after the
--! stop bit sample, with the character in `stat.beat.data` and
--! `stat.frame_error` high when the stop bit was low.
--!
--! A typical transaction, with two clocks per bit so that it fits:
--!
--! ```
--!             __    __    __    __    __    __    __    __
--! clk      __/  \__/  \__/  \__/  \__/  \__/  \__/  \__/  \__
--!          ______             ___________
--! rxd            \___________/           \____________________
--!                start       d0          d1
--!                   ^           ^           ^  sample points
--! ```
entity uart_rx is
    generic (
        --! Clock cycles per bit.
        clocks_per_bit : positive := 16
    );
    port (
        --! Clock and reset.
        sys : in uart_types.sys_t;
        --! The serial line.
        line : in uart_types.line_t;
        --! The received character.
        stat : out uart_types.rx_stat_t
    );
end entity;

architecture rtl of uart_rx is
    type state_t is (idle, start, data, stop);
    signal state : state_t := idle;
    signal shift : uart_types.data_t := (others => '0');
    signal bit_idx : natural range 0 to uart_types.data_width - 1 := 0;
    signal tick : natural range 0 to clocks_per_bit - 1 := 0;
    signal rxd_q : std_ulogic := '1';
    signal valid : std_ulogic := '0';
    signal frame_error : std_ulogic := '0';
begin

    stat.beat.valid <= valid;
    stat.beat.data <= shift;
    stat.frame_error <= frame_error;

    p_rx: process (sys.clk)
        variable last : boolean;
        variable middle : boolean;
    begin
        if rising_edge(sys.clk) then
            rxd_q <= line.rxd;
            valid <= '0';
            if sys.reset = '1' then
                state <= idle;
                tick <= 0;
                bit_idx <= 0;
                frame_error <= '0';
            else
                last := tick = clocks_per_bit - 1;
                middle := tick = clocks_per_bit / 2;
                if last then
                    tick <= 0;
                else
                    tick <= tick + 1;
                end if;
                case state is
                    when idle =>
                        tick <= 0;
                        if rxd_q = '0' then
                            state <= start;
                        end if;
                    when start =>
                        if middle then
                            if rxd_q = '0' then
                                bit_idx <= 0;
                                state <= data;
                            else
                                state <= idle;
                            end if;
                        end if;
                    when data =>
                        if middle then
                            shift <= rxd_q & shift(shift'high downto 1);
                            if bit_idx = uart_types.data_width - 1 then
                                state <= stop;
                            else
                                bit_idx <= bit_idx + 1;
                            end if;
                        end if;
                    when stop =>
                        if middle then
                            valid <= '1';
                            frame_error <= not rxd_q;
                            state <= idle;
                        end if;
                end case;
            end if;
        end if;
    end process;

end architecture;

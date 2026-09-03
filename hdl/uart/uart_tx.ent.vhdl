-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief A UART transmitter: 8 data bits, no parity, one stop bit.
--!
--! Public API elements:
--!
--! * @ref uart_tx  the transmitter entity

library ieee;
    use ieee.std_logic_1164.all;

use work.uart_types;

--! @brief Serialises one character per beat, least significant bit first.
--!
--! A beat is taken on a rising edge of `sys.clk` for which `req.valid`
--! and `stat.ready` are both high. The line then carries a low start
--! bit, the eight data bits and a high stop bit, each `clocks_per_bit`
--! cycles long, and `stat.ready` returns high with the stop bit.
--!
--! A typical transaction, with two clocks per bit so that it fits:
--!
--! ```
--!             __    __    __    __    __    __    __    __
--! clk      __/  \__/  \__/  \__/  \__/  \__/  \__/  \__/  \__
--!                _____
--! valid    _____/     \_______________________________________
--!          ___________
--! ready               \_______________________________________
--!          _________________             ___________
--! txd                        \___________/          \_________
--!                            start       d0         d1
--! ```
entity uart_tx is
    generic (
        --! Clock cycles per bit.
        clocks_per_bit : positive := 16
    );
    port (
        --! Clock and reset.
        sys : in uart_types.sys_t;
        --! The character to send.
        req : in uart_types.beat_t;
        --! Ready and the serial line.
        stat : out uart_types.tx_stat_t
    );
end entity;

architecture rtl of uart_tx is
    type state_t is (idle, start, data, stop);
    signal state : state_t := idle;
    signal shift : uart_types.data_t := (others => '0');
    signal bit_idx : natural range 0 to uart_types.data_width - 1 := 0;
    signal tick : natural range 0 to clocks_per_bit - 1 := 0;
    signal txd : std_ulogic := '1';
begin

    stat.txd <= txd;
    stat.ready <= '1' when state = idle else '0';

    p_tx: process (sys.clk)
        variable last : boolean;
    begin
        if rising_edge(sys.clk) then
            if sys.reset = '1' then
                state <= idle;
                txd <= '1';
                tick <= 0;
                bit_idx <= 0;
            else
                last := tick = clocks_per_bit - 1;
                if last then
                    tick <= 0;
                else
                    tick <= tick + 1;
                end if;
                case state is
                    when idle =>
                        tick <= 0;
                        if req.valid = '1' then
                            shift <= req.data;
                            txd <= '0';
                            state <= start;
                        end if;
                    when start =>
                        if last then
                            txd <= shift(0);
                            shift <= '0' & shift(shift'high downto 1);
                            bit_idx <= 0;
                            state <= data;
                        end if;
                    when data =>
                        if last then
                            if bit_idx = uart_types.data_width - 1 then
                                txd <= '1';
                                state <= stop;
                            else
                                txd <= shift(0);
                                shift <= '0' & shift(shift'high downto 1);
                                bit_idx <= bit_idx + 1;
                            end if;
                        end if;
                    when stop =>
                        if last then
                            state <= idle;
                        end if;
                end case;
            end if;
        end if;
    end process;

end architecture;

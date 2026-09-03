-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Testbench: six characters through the loopback and out of the queue.

library ieee;
    use ieee.std_logic_1164.all;

use work.uart_types;

--! @brief Drives the loopback and checks what comes out of the queue.
--!
--! The clock runs at 10 ns. After a reset of three cycles six
--! characters are sent as fast as the transmitter takes them, the
--! bench waits for the queue to hold four of them, then takes all of
--! them out one per cycle and stops.
entity tb is
end entity;

architecture sim of tb is
    constant clocks_per_bit : positive := 16;
    constant depth : positive := 4;
    type text_t is array (0 to 5) of uart_types.data_t;
    constant text : text_t := (x"48", x"65", x"6c", x"6c", x"6f", x"21");

    signal sys : uart_types.sys_t := (clk => '0', reset => '1');
    signal req : uart_types.beat_t := (valid => '0', data => (others => '0'));
    signal tx_stat : uart_types.tx_stat_t;
    signal rd : uart_types.take_t := (take => '0');
    signal stat : uart_types.fifo_stat_t;
    signal done : boolean := false;
    signal errors : natural := 0;
begin

    dut: entity work.uart_loop
        generic map (clocks_per_bit => clocks_per_bit, depth => depth)
        port map (sys => sys, req => req, tx_stat => tx_stat, rd => rd, stat => stat);

    p_clk: process
    begin
        while not done loop
            sys.clk <= '0';
            wait for 5 ns;
            sys.clk <= '1';
            wait for 5 ns;
        end loop;
        wait;
    end process;

    p_send: process
    begin
        wait for 30 ns;
        wait until rising_edge(sys.clk);
        sys.reset <= '0';
        for i in text'range loop
            wait until rising_edge(sys.clk) and tx_stat.ready = '1';
            req <= (valid => '1', data => text(i));
            wait until rising_edge(sys.clk);
            req.valid <= '0';
        end loop;
        wait;
    end process;

    p_take: process
        variable got : natural := 0;
    begin
        wait until stat.count = depth;
        while got < text'length loop
            wait until rising_edge(sys.clk) and stat.empty = '0';
            if stat.head /= text(got) then
                errors <= errors + 1;
            end if;
            got := got + 1;
            rd.take <= '1';
            wait until rising_edge(sys.clk);
            rd.take <= '0';
        end loop;
        wait for 100 ns;
        done <= true;
        wait for 10 ns;
        std.env.stop;
    end process;

end architecture;

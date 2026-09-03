-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Testbench top level for @ref counter.
--!
--! Public API elements:
--!
--! * @ref tb  the simulation top level

library ieee;
    use ieee.std_logic_1164.all;
    use ieee.numeric_std.all;

use work.counter_types;

--! @brief Simulation top level.
--!
--! Drives one @ref counter through a reset, a disabled stretch, and a
--! full wrap of the counting range, checking the value after every step.
--! The run ends with `std.env.stop`, so `run -all` terminates and xsim
--! writes out the waveform database.
entity tb is
end entity;

architecture sim of tb is
    constant period : time := 10 ns;
    constant full_scale : natural := 2 ** counter_types.count_width;

    signal ctl : counter_types.ctl_t := (
        clk => '0',
        reset => '1',
        enable => '0'
    );
    signal stat : counter_types.stat_t;

    signal running : boolean := true;
begin

    ctl.clk <= (not ctl.clk) after period / 2 when running else '0';

    dut: entity work.counter
        port map (
            ctl => ctl,
            stat => stat
        );

    p_stimulus: process
    begin
        -- Hold reset for a few edges, then release it.
        ctl.reset <= '1';
        ctl.enable <= '0';
        for i in 0 to 3 loop
            wait until rising_edge(ctl.clk);
        end loop;
        wait for period / 4;
        assert stat.value = 0
            report "counter did not reset to zero"
            severity failure;
        ctl.reset <= '0';

        -- Disabled: the count must hold.
        for i in 0 to 3 loop
            wait until rising_edge(ctl.clk);
        end loop;
        wait for period / 4;
        assert stat.value = 0
            report "counter advanced while disabled"
            severity failure;

        -- Enabled: one full wrap of the counting range.
        ctl.enable <= '1';
        for i in 1 to full_scale loop
            wait until rising_edge(ctl.clk);
            wait for period / 4;
            assert stat.value = to_unsigned(i mod full_scale, counter_types.count_width)
                report "counter value is wrong at step " & integer'image(i)
                severity failure;
            assert (stat.wrapped = '1') = (i mod full_scale = 0)
                report "wrapped flag is wrong at step " & integer'image(i)
                severity failure;
        end loop;

        -- A few idle edges, so the tail of the waveform is not empty.
        ctl.enable <= '0';
        for i in 0 to 7 loop
            wait until rising_edge(ctl.clk);
        end loop;

        report "counter testbench passed" severity note;
        running <= false;
        wait for period;
        std.env.stop;
    end process;

end architecture;

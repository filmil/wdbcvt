-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief A free-running wrapping counter.
--!
--! Public API elements:
--!
--! * @ref counter  the counter entity

library ieee;
    use ieee.std_logic_1164.all;
    use ieee.numeric_std.all;

use work.counter_types;

--! @brief Counts rising clock edges while enabled, and wraps.
--!
--! The count advances by one on every rising edge of `ctl.clk` for which
--! `ctl.enable` is high. A high `ctl.reset` wins over `ctl.enable` and
--! drives the count back to zero on the next edge. `stat.wrapped` is high
--! for exactly the cycle in which the count returns to zero by wrapping;
--! a reset to zero does not raise it.
--!
--! A typical transaction, for a two-bit counter so that the wrap fits on
--! the page:
--!
--! ```
--!             __    __    __    __    __    __    __    __
--! clk      __/  \__/  \__/  \__/  \__/  \__/  \__/  \__/  \__
--!          _____
--! reset         \_____________________________________________
--!                _________________________________________
--! enable   _____/
--!          ___________ _____ _____ _____ _____ _____ _____
--! value    ___0_______X__1__X__2__X__3__X__0__X__1__X__2__
--!                                          ___
--! wrapped  _______________________________/   \____________
--! ```
entity counter is
    port (
        --! Control inputs. See @ref counter_types.ctl_t.
        ctl : in counter_types.ctl_t;
        --! Status outputs. See @ref counter_types.stat_t.
        stat : out counter_types.stat_t
    );
end entity;

architecture rtl of counter is
    signal value : counter_types.count_t := (others => '0');
    signal wrapped : std_ulogic := '0';
begin

    stat.value <= value;
    stat.wrapped <= wrapped;

    p_count: process (ctl.clk)
    begin
        if rising_edge(ctl.clk) then
            if ctl.reset = '1' then
                value <= (others => '0');
                wrapped <= '0';
            elsif ctl.enable = '1' then
                if value = (value'range => '1') then
                    value <= (others => '0');
                    wrapped <= '1';
                else
                    value <= value + 1;
                    wrapped <= '0';
                end if;
            else
                wrapped <= '0';
            end if;
        end if;
    end process;

end architecture;

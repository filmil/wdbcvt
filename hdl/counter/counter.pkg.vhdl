-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Signal bundles exchanged with the @ref counter entity.
--!
--! Public API elements:
--!
--! * @ref counter_types.count_t   the counter value type
--! * @ref counter_types.ctl_t     the control bundle driven into a counter
--! * @ref counter_types.stat_t    the status bundle driven out of a counter
--! * @ref counter_types.count_width  the counter width, in bits

library ieee;
    use ieee.std_logic_1164.all;
    use ieee.numeric_std.all;

--! @brief Types shared by the @ref counter entity and its users.
--!
--! Ports never carry loose `std_ulogic` signals. Everything related is
--! bundled into a record here, so that adding a signal to the interface
--! touches this package and not every port map in the design.
package counter_types is

    --! Width of the counter, in bits.
    constant count_width : natural := 8;

    --! The counter value.
    subtype count_t is unsigned(count_width - 1 downto 0);

    --! Everything driven into a @ref counter.
    type ctl_t is record
        --! Rising-edge clock. All other members are sampled on its
        --! rising edge.
        clk : std_ulogic;
        --! Synchronous, active high reset. Forces the count to zero.
        reset : std_ulogic;
        --! Count enable, active high. The count holds while it is low.
        enable : std_ulogic;
    end record;

    --! Everything driven out of a @ref counter.
    type stat_t is record
        --! The current count.
        value : count_t;
        --! High for the one cycle in which @ref value wraps to zero.
        wrapped : std_ulogic;
    end record;

end package;

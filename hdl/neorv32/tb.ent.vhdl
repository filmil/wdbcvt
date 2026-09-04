-- SPDX-License-Identifier: Apache-2.0

--! @brief Wrapper that stops the NEORV32 testbench.
--!
--! The project's own testbench, `neorv32_tb`, runs until the simulator
--! is told to stop, which is what GHDL's `--stop-time` does for it.
--! `run -all` under xsim has no such limit, so this entity instantiates
--! it and calls `std.env.stop` after `RUN_TIME`, which ends the whole
--! simulation and makes xsim write the waveform database out.
--!
--! ```
--!             0                        RUN_TIME
--!             |---- neorv32_tb runs ------|
--!                                         stop
--! ```
library neorv32;

entity tb is
    generic (
        --! How long the processor runs before the simulation stops.
        RUN_TIME : time := 200 us
    );
end entity tb;

architecture sim of tb is
begin

    dut : entity neorv32.neorv32_tb;

    stopper : process
    begin
        wait for RUN_TIME;
        std.env.stop;
    end process stopper;

end architecture sim;

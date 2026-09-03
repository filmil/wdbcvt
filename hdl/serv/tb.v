// SPDX-License-Identifier: Apache-2.0
//
// Bench for the SERV reference SoC, servant, running the hello_uart
// firmware that ships with SERV. It follows servant_tb from SERV's own
// bench without the vlog_tb_utils dependency: a 16 MHz clock, a reset
// that lasts one clock, a UART decoder on the GPIO pin the firmware
// bit bangs, and a timeout. The firmware ends the run itself: its
// write to the halt address makes servile_mux call $finish, so
// `run -all` returns and xsim writes the database.
//
// The decoder keeps every character it receives in `text` and counts
// in `errors` the ones that differ from the string the firmware
// prints, so the database records whether the run was right.
//
// The memory image is a parameter so that the BUILD file can point it
// at the file in the SERV archive.
//
//            ______        ______        ______
// wb_clk ___|      |______|      |______|      |___
//        ______
// wb_rst       |___________________________________
//        ___________________          _____ _____
// q                         |________|_____|_____|__ ... start, then 8 data bits
//
`default_nettype none
module tb;

   parameter memfile = "";

   // The bit banged UART of the firmware assumes 16 MHz and 57600 baud.
   localparam baud_rate = 57600;
   localparam bit_time = 1000000000 / baud_rate;

   reg wb_clk = 1'b0;
   reg wb_rst = 1'b1;

   wire q;
   wire [31:0] pc_adr;
   wire        pc_vld;

   always #31 wb_clk <= !wb_clk;
   initial #62 wb_rst <= 1'b0;

   servant_sim
     #(.memfile (memfile))
   dut
     (.wb_clk (wb_clk),
      .wb_rst (wb_rst),
      .pc_adr (pc_adr),
      .pc_vld (pc_vld),
      .q      (q));

   // What hello_uart prints, sw/hello_uart.S in the SERV archive.
   localparam [8*17-1:0] expected = "Hi, I'm Servant!\n";

   reg [7:0] text [0:31];
   reg [7:0] ch;
   integer   count = 0;
   integer   errors = 0;
   integer   i;

   initial forever begin
      @(negedge q);
      #(bit_time / 2) ch = 0;
      for (i = 0; i < 8; i = i + 1)
        #bit_time ch[i] = q;
      $write("%c", ch);
      if (count < 32)
        text[count] = ch;
      if (count >= 17 || ch != expected[8*(16-count) +: 8])
        errors = errors + 1;
      count = count + 1;
   end

   // The firmware prints 17 characters at 57600 baud, under 3 ms.
   initial begin
      #10000000;
      $display("tb: timeout");
      $finish;
   end

endmodule

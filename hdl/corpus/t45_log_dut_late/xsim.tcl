open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
run 10 ns
log_vcd [get_objects -r /tb/dut/* ]
log_wave -recursive /tb/dut
run -all
close_vcd
exit

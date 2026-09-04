open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects /tb/*]
log_wave -recursive *
run 15 ns
set_value /tb/s 0
run -all
close_vcd
exit

open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects /tb/*]
log_wave -recursive *
add_force /tb/s 1
run 15 ns
remove_forces /tb/s
run -all
close_vcd
exit

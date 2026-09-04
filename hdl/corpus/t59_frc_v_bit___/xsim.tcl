open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects /tb/*]
log_wave -recursive *
add_force {/tb/v[3]} 1
run -all
close_vcd
exit

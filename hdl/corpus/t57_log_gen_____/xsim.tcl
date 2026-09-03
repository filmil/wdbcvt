open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects /tb/g/*]
log_wave /tb/g
run -all
close_vcd
exit

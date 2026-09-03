open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects /tb/*]
log_wave /tb
run -all
close_vcd
exit

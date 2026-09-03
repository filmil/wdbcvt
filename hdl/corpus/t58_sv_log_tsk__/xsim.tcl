open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects /tb/inc/*]
log_wave /tb/inc
run -all
close_vcd
exit

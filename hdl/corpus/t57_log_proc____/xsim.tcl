open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects /tb/p/*]
log_wave /tb/p
run -all
close_vcd
exit

open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd /tb/inc/x
log_wave /tb/inc/x
run -all
close_vcd
exit

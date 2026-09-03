open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd /tb/inc/tmp
log_wave /tb/inc/tmp
run -all
close_vcd
exit

open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd /tb/m
log_wave /tb/m
run -all
close_vcd
exit

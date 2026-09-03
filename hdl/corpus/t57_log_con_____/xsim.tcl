open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd /tb/c
log_wave /tb/c
run -all
close_vcd
exit

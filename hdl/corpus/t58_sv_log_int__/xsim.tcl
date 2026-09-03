open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd /tb/i
log_wave /tb/i
run -all
close_vcd
exit

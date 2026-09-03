open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd /tb/s
log_wave /tb/s
run -all
close_vcd
exit

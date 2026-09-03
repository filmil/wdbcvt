open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd /tb/L
log_wave /tb/L
run -all
close_vcd
exit

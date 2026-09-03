open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd /tb/sv
log_wave /tb/sv
run -all
close_vcd
exit

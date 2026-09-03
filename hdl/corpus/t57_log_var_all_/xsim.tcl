open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd /tb/p/w
log_wave /tb/p/w
run -all
close_vcd
exit

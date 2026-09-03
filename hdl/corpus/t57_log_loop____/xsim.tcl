open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd /tb/p/k
log_wave /tb/p/k
run -all
close_vcd
exit

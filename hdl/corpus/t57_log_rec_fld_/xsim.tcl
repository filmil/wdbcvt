open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd /tb/r.n
log_wave /tb/r.n
run -all
close_vcd
exit

open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd /tb/P
log_wave /tb/P
run -all
close_vcd
exit

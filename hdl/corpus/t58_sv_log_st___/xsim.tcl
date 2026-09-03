open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd /tb/st
log_wave /tb/st
run -all
close_vcd
exit

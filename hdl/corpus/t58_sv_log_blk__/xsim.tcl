open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects /tb/blk/*]
log_wave /tb/blk
run -all
close_vcd
exit

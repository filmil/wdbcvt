open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd /tb/blk/bv
log_wave /tb/blk/bv
run -all
close_vcd
exit

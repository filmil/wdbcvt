open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects -r /tb/* ]
log_vcd [get_objects -r /tb2/* ]
log_wave -recursive /tb2
log_wave -recursive /tb
run -all
close_vcd
exit

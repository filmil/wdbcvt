open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects -r /* ]
log_wave -recursive *
run -all
close_vcd
exit

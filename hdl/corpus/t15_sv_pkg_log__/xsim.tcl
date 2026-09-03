open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
puts "PKG: [get_objects /p/*]"
log_vcd [get_objects -r /* ]
log_wave -recursive *
log_wave -recursive /p
run -all
close_vcd
exit

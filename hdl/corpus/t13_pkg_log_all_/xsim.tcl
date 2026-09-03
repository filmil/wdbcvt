open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
puts "SCOPES: [get_scopes -r /*]"
puts "PKG: [get_objects /sig_pkg/*]"
log_vcd [get_objects -r /* ]
log_wave -recursive *
log_wave -recursive /sig_pkg
run -all
close_vcd
exit

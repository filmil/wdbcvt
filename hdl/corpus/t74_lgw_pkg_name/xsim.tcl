open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
puts "SCOPE: [current_scope]"
puts "SCOPES: [get_scopes -r /*]"
puts "OBJECTS: [get_objects -r /*]"
puts "PKGOBJ: [get_objects /sig_pkg/*]"
log_vcd [get_objects -r /* ]
log_wave -recursive *
log_wave -recursive /sig_pkg
run -all
close_vcd
exit

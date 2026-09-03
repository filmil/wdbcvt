open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd {/tb/v[2:1]}
log_wave {/tb/v[2:1]}
run -all
close_vcd
exit

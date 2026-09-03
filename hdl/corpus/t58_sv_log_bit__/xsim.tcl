open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_wave {/tb/v[3]}
run -all
close_vcd
exit

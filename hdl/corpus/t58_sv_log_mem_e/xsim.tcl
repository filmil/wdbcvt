open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_wave {/tb/m[1]}
run -all
close_vcd
exit

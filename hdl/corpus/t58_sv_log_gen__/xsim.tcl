open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_wave {/tb/gb[1]}
run -all
close_vcd
exit

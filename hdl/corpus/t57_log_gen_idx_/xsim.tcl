open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd {/tb/\g(1)\/i}
log_wave {/tb/\g(1)\/i}
run -all
close_vcd
exit

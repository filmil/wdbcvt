open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd {/tb/\g(1)\/gs}
log_wave {/tb/\g(1)\/gs}
run -all
close_vcd
exit

open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects {/tb/\g(1)\/*}]
log_wave "/tb/\\g(1)\\"
run -all
close_vcd
exit

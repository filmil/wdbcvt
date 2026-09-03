open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_wave /tb/st.a
run -all
close_vcd
exit

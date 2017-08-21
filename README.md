# canary

Simple golang tool to check health of openswan tunnel and restart if down. We've encountered
some cases where standard dead peer detection (DPD) doesn't detect this.

This braindead tool just tries to open a TCP connection to the configured host and port, and if 
it is unable after 10 seconds then bounces the connection with the configured name.

See `sample.json` for file format.

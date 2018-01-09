# psb-rotatord
For rotation handling and pre/post run clean-ups

This application will daemonize the rotator package, which can be used via the client: psb-rotatorc. The daemon must be ran as a superuser, you can find an example Systemd service file in //link repo/examples/ <!-- psb-rotatord.service --> Installing the daemon and client is manditory for all remote backup jobs.

### Build
make build

### Install
make install
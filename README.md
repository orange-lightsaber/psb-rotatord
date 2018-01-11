# psb-rotatord
psb rotation handling

This application will daemonize the rotator package, which can be used via the client: psb-rotatorc. The daemon must be ran as a superuser, an example Systemd service file can be found in [./examples](examples). Installing the daemon and client is manditory for all remote backup jobs.

### Build
```sh
make build
```

### Install
```sh
make install
```

### Enable and start service
```sh
sudo cp ./examples/psb-rotatord.service /etc/systemd/system/psb-rotatord.service
systemctl enable psb-rotatord.service
systemctl start psb-rotatord.service
```
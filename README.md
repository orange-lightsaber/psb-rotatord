# psb-rotatord
psb rotation handling for remote backups

This application will daemonize the rotator package, which can be used via the client: psb-rotatorc. The daemon must be ran as a superuser, an example Systemd service file can be found in [./examples](examples). Installing the daemon and client is manditory/only for remote backup jobs.

## Prerequisites and setup
- Add the following line to sudoers file, this is necessary to allow Rsync to maintain file ownership during transfers. Replace "psbuser" with the username of the ssh user that receives.
    ```
    psbuser ALL= NOPASSWD:/usr/bin/rsync
    ```

### Build
```sh
make build
```

### Install
```sh
sudo make install
```

### Enable and start service
> Note: the -p flag is for defining the absolute path to the backup directory, and the default directory is '/backup', edit the example Systemd file accordingly. If the run config key *backup-directory* has a value, the path is again overridden, but individually unique to that specific backup.

```sh
sudo cp ./examples/psb-rotatord.service /etc/systemd/system/psb-rotatord.service
systemctl enable psb-rotatord.service
systemctl start psb-rotatord.service
```
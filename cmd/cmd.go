package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/marcsauter/single"
	"github.com/orange-lightsaber/psb-rotatord/rotator"
	"github.com/orange-lightsaber/psb-rotatord/sockets"
)

func Daemonize() {
	// Allow only a single instance
	instance := single.New("psb_rotator_daemon")
	instance.Lock()
	defer instance.Unlock()
	sockets.Open(func(req *sockets.Request) *sockets.Response {
		var res sockets.Response
		switch req.Request {
		case sockets.LastRun_Req:
			r, err := rotator.TimeSinceLastRun(req.RCD.Name)
			if err != nil {
				res.Error = err.Error()
			}
			res.Response = r
		case sockets.InitRun_Req:
			r, err := rotator.InitRun(req.RCD)
			if err != nil {
				res.Error = err.Error()
			}
			res.Response = r
		case sockets.Rotate_Req:
			r, err := rotator.Rotate(req.RCD)
			if err != nil {
				res.Error = err.Error()
			}
			res.Response = r
		}
		return &res
	})
}

func Exec(version string) {
	v := flag.Bool("v", false, "Print version.")
	flag.StringVar(&rotator.Config.Paths.BackupDir, "p", "/backup", "Absolute path to psb backup directory.")
	flag.Parse()
	if *v {
		fmt.Printf("psb-rotatord v%s\n", version)
		os.Exit(0)
	}
	Daemonize()
}

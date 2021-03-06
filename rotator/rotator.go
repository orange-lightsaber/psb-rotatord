package rotator

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type datesAndPaths interface {
	hasDatePath(string) bool
}

type datePaths struct {
	paths []string
	regex *regexp.Regexp
}

type Year struct {
	Duration int `toml:"yearly"`
	datePaths
}
type Month struct {
	Duration int `toml:"monthly"`
	datePaths
}
type Day struct {
	Duration int `toml:"daily"`
	datePaths
}
type Initial struct {
	Duration int `toml:"initial"`
	datePaths
}

type RunConfigData struct {
	CompatibilityKey string `toml:"compatibility-key"`
	Name             string `toml:"name"`
	BackupDir        string `toml:"backup-directory"`
	Frequency        int    `toml:"frequency"`
	RotationDelay    int    `toml:"delay"`
	Year
	Month
	Day
	Initial
	lastRun time.Time
}

func (rcd *RunConfigData) updateLastRun() {
	rcd.lastRun = time.Now().UTC()
	r := rcds[rcd.Name]
	r.lastRun = rcd.lastRun
	rcds[rcd.Name] = r
}

func (rcd *RunConfigData) getAll() (*Year, *Month, *Day, *Initial) {
	return &rcd.Year, &rcd.Month, &rcd.Day, &rcd.Initial
}

func (d datePaths) hasDatePath(path string) bool {
	return d.regex.MatchString(path)
}

func (d datePaths) getPaths() []string {
	paths := d.paths
	return paths
}

func (d *datePaths) addPath(path string) {
	path = filepath.Clean(d.regex.FindString(path))
	paths := d.paths
	paths = append(paths, path)
	d.paths = paths
}

var rcds = make(map[string]RunConfigData)

func getSnapshotPaths(d datesAndPaths, psbDir string, cwd string) ([]string, error) {
	var slc []string
	err := filepath.Walk(cwd, func(path string, f os.FileInfo, err error) error {
		if d.hasDatePath(path) {
			path = strings.Replace(path, psbDir, "", -1)
			slc = append(slc, path)
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		fmt.Errorf("error searching a directory: %s", err.Error())
	}
	return slc, err
}

func populatePathLists(psbDir string, rcd RunConfigData) (year *Year, month *Month, day *Day, initial *Initial, e error) {
	year, month, day, initial = rcd.getAll()
	year.paths, e = getSnapshotPaths(year, psbDir, psbDir)
	if e != nil {
		return
	}
	for _, wd := range year.getPaths() {
		cwd := filepath.Join(psbDir, wd)
		pList, err := getSnapshotPaths(month, psbDir, cwd)
		if err != nil {
			e = err
			return
		}
		for _, p := range pList {
			month.paths = append(month.paths, p)
		}
	}
	for _, wd := range month.getPaths() {
		cwd := filepath.Join(psbDir, wd)
		pList, err := getSnapshotPaths(day, psbDir, cwd)
		if err != nil {
			e = err
			return
		}
		for _, p := range pList {
			day.paths = append(day.paths, p)
		}
	}
	for _, wd := range day.getPaths() {
		cwd := filepath.Join(psbDir, wd)
		pList, err := getSnapshotPaths(initial, psbDir, cwd)
		if err != nil {
			e = err
			return
		}
		for _, p := range pList {
			initial.paths = append(initial.paths, p)
		}
	}
	return
}

func pathToDate(path string) (t time.Time, err error) {
	if path == "" {
		return
	}
	timeLayout := Config.DatePath.TimeLayout
	for i := 0; i < 4; i++ {
		err = nil
		t, err = time.ParseInLocation(timeLayout, path, time.UTC)
		timeLayout = filepath.Dir(timeLayout)
		if err == nil {
			break
		}
	}
	return
}

func collapse(psbDir string, path string, expDate time.Time, finalDirs map[string]string) error {
	pathDate, err := pathToDate(path)
	if err != nil {
		return err
	}
	if pathDate.Before(expDate) {
		finalPath := filepath.Dir(path)
		if finalDirs[finalPath] == "" {
			finalDirs[finalPath] = path
		}
		// Ensure latest is chosen to keep
		d, err := pathToDate(finalDirs[finalPath])
		if err != nil {
			return err
		}
		if d.Before(pathDate) {
			err := os.RemoveAll(filepath.Join(psbDir, finalDirs[finalPath]))
			if err != nil {
				return fmt.Errorf("error removing directory: %s", err.Error())
			}
			finalDirs[finalPath] = path
		}
		if path != finalDirs[finalPath] {
			err := os.RemoveAll(filepath.Join(psbDir, path))
			if err != nil {
				return fmt.Errorf("error removing directory: %s", err.Error())
			}
		}
	}
	return nil
}

func cpCmd(src string, dest string) error {
	var errmsg bytes.Buffer
	cmd := exec.Command("cp", "-al", src, dest)
	cmd.Stderr = &errmsg
	err := cmd.Run()
	if err != nil || errmsg.Len() > 0 {
		return fmt.Errorf("failed to copy: %s", errmsg.String())
	}
	return nil
}

func rotate(rcd RunConfigData) error {
	psbDir := filepath.Join(rcd.BackupDir, rcd.Name)
	year, month, day, initial, err := populatePathLists(psbDir, rcd)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	// Rotation delay
	var isDelayed bool
	if rcd.RotationDelay > rcd.Frequency {
		delay := time.Duration(rcd.RotationDelay) * time.Minute
		delayThreshold := now.Add(-delay)
		for _, p := range initial.paths {
			d, err := pathToDate(p)
			if err != nil {
				return err
			}
			if delayThreshold.Before(d) {
				isDelayed = true
			}
		}
	}
	tmp := filepath.Join(psbDir, ".tmp")
	current := filepath.Join(psbDir, Config.Paths.CurrentSnapshot)
	err = os.RemoveAll(current)
	if err != nil {
		return err
	}
	err = cpCmd(tmp, current)
	if err != nil {
		return err
	}
	if !isDelayed {
		relTimePath := now.Format(Config.DatePath.TimeLayout)
		newTimePath := filepath.Join(psbDir, relTimePath)
		if _, err := os.Stat(newTimePath); !os.IsNotExist(err) {
			return fmt.Errorf("the directory \"%s\" already exists, this probably means a rotation is running unexpectedly", newTimePath)
		}
		if err := os.MkdirAll(filepath.Dir(newTimePath), 0755); err != nil {
			return err
		}
		err := cpCmd(current, newTimePath)
		if err != nil {
			return err
		}
		initial.addPath(relTimePath)
	}
	// Update the lastRun
	rcd.updateLastRun()
	// Start rotations
	var expDate time.Time
	var finalDirs map[string]string
	// Inital rotation
	expDate = now.AddDate(0, 0, -initial.Duration)
	finalDirs = make(map[string]string)
	for _, p := range initial.datePaths.paths {
		err := collapse(psbDir, p, expDate, finalDirs)
		if err != nil {
			return err
		}
	}
	for k, _ := range finalDirs {
		day.addPath(k)
	}
	// Daily rotation
	expDate = now.AddDate(0, -day.Duration, 0)
	finalDirs = make(map[string]string)
	for _, p := range day.datePaths.paths {
		err := collapse(psbDir, p, expDate, finalDirs)
		if err != nil {
			return err
		}
	}
	for k, _ := range finalDirs {
		month.addPath(k)
	}
	// Monthly rotation
	expDate = now.AddDate(0, -month.Duration, 0)
	finalDirs = make(map[string]string)
	for _, p := range month.datePaths.paths {
		err := collapse(psbDir, p, expDate, finalDirs)
		if err != nil {
			return err
		}
	}
	for k, _ := range finalDirs {
		year.addPath(k)
	}
	// Yearly rotation
	expDate = now.AddDate(-year.Duration, 0, 0)
	finalDirs = make(map[string]string)
	for _, p := range year.datePaths.paths {
		err := collapse(psbDir, p, expDate, finalDirs)
		if err != nil {
			return err
		}
	}
	return nil
}

func rcdInit(rcd RunConfigData) RunConfigData {
	year := `\/([0-9]{4})`
	month := `\/([A-Z])\w+`
	day := `\/([0-9]|0[1-9]|[0-3][0-9])`
	time := `\/([0-9]{4})Z`
	eop := `(\z|\/\z?)`
	rcd.Year.regex = regexp.MustCompile(year + eop)
	rcd.Month.regex = regexp.MustCompile(year + month + eop)
	rcd.Day.regex = regexp.MustCompile(year + month + day + eop)
	rcd.Initial.regex = regexp.MustCompile(year + month + day + time + eop)
	if rcd.BackupDir == "" {
		rcd.BackupDir = Config.Paths.BackupDir
	}
	return rcd
}

func lastSnapshotPath(rcd RunConfigData) (string, error) {
	var latestSnapshot string
	if rcd.Name == "" {
		return latestSnapshot, errors.New("no snapshot was requested")
	}
	cwd := filepath.Join(rcd.BackupDir, rcd.Name)
	err := filepath.Walk(cwd, func(path string, f os.FileInfo, err error) error {
		if rcd.Initial.hasDatePath(path) {
			path = strings.Replace(path, cwd, "", -1)
			pd, err := pathToDate(path)
			if err != nil {
				return err
			}
			ld, err := pathToDate(latestSnapshot)
			if err != nil {
				return err
			}
			if !rcd.Initial.hasDatePath(latestSnapshot) || pd.After(ld) {
				latestSnapshot = path
			}
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error searching for snapshots: %s", err.Error())
	}
	return latestSnapshot, nil

}

func TimeSinceLastRun(rcd RunConfigData) (res string, err error) {
	if rcd.Name != rcds[rcd.Name].Name {
		rcd = rcdInit(rcd)
	} else {
		rcd = rcds[rcd.Name]
	}
	nilTime := time.Time{}
	if rcd.lastRun == nilTime {
		var latestSnapshot string
		latestSnapshot, err = lastSnapshotPath(rcd)
		if err != nil {
			return
		}
		rcd.lastRun, err = pathToDate(latestSnapshot)
		if err != nil {
			return
		}
	}
	dur := time.Now().UTC().Sub(rcd.lastRun)
	res = dur.String()
	return
}

func InitRun(rcd RunConfigData) (res string, err error) {
	if rcd.Name != rcds[rcd.Name].Name {
		rcd = rcdInit(rcd)
		rcds[rcd.Name] = rcd
	} else {
		if rcd.CompatibilityKey != rcds[rcd.Name].CompatibilityKey {
			err = errors.New("invalid compatibility key")
			return
		}
	}
	rcd = rcds[rcd.Name]
	res = filepath.Join(rcd.BackupDir, rcd.Name, ".tmp")
	return
}

func Rotate(name string) (res string, err error) {
	err = rotate(rcds[name])
	if err != nil {
		err = fmt.Errorf("error during rotation: %s", err.Error())
		return
	}
	res = rcds[name].lastRun.String()
	return
}

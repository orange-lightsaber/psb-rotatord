package rotator

type configuration struct {
	Paths    paths
	DatePath datePath
}

type paths struct {
	CurrentSnapshot string
	BackupDir       string
}

type datePath struct {
	TimeLayout string
	Regex      datePathRegex
}

type datePathRegex struct {
	Year  string
	Month string
	Day   string
	Time  string
	Eop   string
}

func (c *configuration) loadDefaults() {
	c.Paths.CurrentSnapshot = "current"
	c.Paths.BackupDir = "/backup"
	c.DatePath.TimeLayout = "/2006/January/2/1504Z"
	c.DatePath.Regex.Year = `\/([0-9]{4})`
	c.DatePath.Regex.Month = `\/([A-Z])\w+`
	c.DatePath.Regex.Day = `\/([0-9]|0[1-9]|[0-3][0-9])`
	c.DatePath.Regex.Time = `\/([0-9]{4})Z`
	c.DatePath.Regex.Eop = `(\z|\/\z?)`
}

var Config configuration

func init() {
	Config.loadDefaults()
}

package setting

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

var (
	// App settings
	AppVer  string
	AppName string
	AppURL  string
	AppPath string

	// server settings
	Protocol         string
	Domain           string
	HTTPAddr         string
	HTTPPort         string
	DisableRouterLog bool

	// Database setting
	RedisURI     string
	MongoURI     string
	ElasticHosts []string

	// Global setting objects
	Cfg      *ini.File
	ProdMode bool
	// RunUser      string
)

// execPath returns the executable path.
func execPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	return filepath.Abs(file)
}

func init() {
	log.Println("Init the app settings...")
	var err error
	if AppPath, err = execPath(); err != nil {
		log.Fatalf("Fail to get app path: %v\n", err)
	}
	AppPath = strings.Replace(AppPath, "\\", "/", -1)

	log.Println("Init the app settings OK.")
}

// NewContext init the setting
func NewContext() {
	Cfg, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, "conf/app.ini")
	if err != nil {
		log.Fatalf("Fail to parse 'conf/app.ini': %v", err)
	}
	// log.Printf("%q\n", Cfg.Section("server").Keys())

	// set AppName and AppURL
	sec := Cfg.Section("server")
	AppName = Cfg.Section("").Key("APP_NAME").MustString("Letitgo")
	AppURL = sec.Key("ROOT_URL").MustString("http://localhost:9000/")
	if AppURL[len(AppURL)-1] != '/' {
		AppURL += "/"
	}
	Domain = sec.Key("DOMAIN").MustString("localhost")
	HTTPAddr = sec.Key("HTTP_ADDR").MustString("0.0.0.0")
	HTTPPort = sec.Key("HTTP_PORT").MustString("9000")
	DisableRouterLog = sec.Key("DISABLE_ROUTER_LOG").MustBool()
}

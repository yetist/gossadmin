package main

import (
	"flag"
	"fmt"
	"github.com/gosexy/gettext"
	ss "github.com/shadowsocks/shadowsocks-go/shadowsocks"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
)

const (
	PkgName    = "gossadmin"
	PkgAuthor  = "yetist@gmail.com"
	PkgVersion = "0.1"
)

func ShowVersion() {
	ss.PrintVersion()
	fmt.Printf(gettext.Gettext("%s version: %s\n"), PkgName, PkgVersion)
}

func ShowConfig() {
	filename := "server.ini"
	sysconf := path.Join(app.dir.SiteConfig(), filename)
	userconf := path.Join(app.dir.UserConfig(), filename)
	fmt.Printf(gettext.Gettext("Config file paths:\n System: %s\n User: %s\n"), sysconf, userconf)
	if err := ioutil.WriteFile(filename, []byte(defConfig), 0644); err == nil {
		fmt.Printf(gettext.Gettext("Here is an example config file: \"%s\".\n"), filename)
	} else {
		fmt.Printf(gettext.Gettext("Here is an example:\n%s\n"), defConfig)
	}
}

func main() {
	var version bool
	var cfg bool
	gettext.BindTextdomain(PkgName, app.config.locale_root)
	gettext.Textdomain(PkgName)
	gettext.SetLocale(gettext.LC_ALL, "")

	log.SetOutput(os.Stdout)
	flag.BoolVar(&version, "version", false, gettext.Gettext("print version"))
	flag.BoolVar(&cfg, "config", false, gettext.Gettext("show about config file"))
	flag.Parse()

	if version {
		ShowVersion()
		os.Exit(0)
	}
	if cfg {
		ShowConfig()
		os.Exit(0)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	ss.SetDebug(ss.DebugLog(app.config.debug))
	ReloadServer()
	runweb()
}

package main

import (
	"github.com/Unknwon/goconfig"
	"github.com/Wessie/appdirs"
	"os"
	"path"
)

//配置项
type Config struct {
	db            string
	dbname        string
	dbpass        string
	dbtype        string
	dbuser        string
	method        string
	timeout       int
	port_min      int
	port_max      int
	bytes_max     int64
	smtp_host     string
	smtp_pass     string
	smtp_username string
	debug         bool
	log_info      bool
	log_sql       bool
	log_warn      bool
	log_error     bool
	web_env       string
	http_host     string
	http_port     int
	static_root   string
	template_root string
	locale_root   string
	all_methods   []string
}

type App struct {
	config Config
	dir    *appdirs.App
	cfg    *goconfig.ConfigFile
	log    *os.File
}

var app *App

var defConfig = `
[server]
port_min = 3001
port_max = 3050
method = aes-256-cfb
timeout = 300
bytes_max = 1024000000

[log]
debug = false
info = false
warning = false
error = false
sql = false

[database]
# database type: sqlite, mysql
dbtype = sqlite
# database name:
dbname = gossadmin.db
# database username(for mysql):
#dbuser =
# database password(for mysql):
#dbpass = 

[mail]
smtp_host =
smtp_pass =
smtp_username =

[web]
# web server env: development, production
env = production
# web server url:
domain = http://localhost
# web server host:
host = 0.0.0.0
# web server port:
port = 8080
# static directory path:
#static_root=/usr/share/gossadmin/assets
# template directory path:
#template_root=/usr/share/gossadmin/templates
# locale directory path:
#locale_root=/usr/share/locale
`

func init() {
	app = NewApp(PkgName, PkgAuthor, PkgVersion)
	app.LoadConfig("server.ini")
}

func NewApp(name, author, version string) *App {
	return &App{dir: appdirs.New(name, author, version)}
}

func (a *App) LoadConfig(cfgname string) {
	sysconf := path.Join(a.dir.SiteConfig(), cfgname)
	userconf := path.Join(a.dir.UserConfig(), cfgname)
	if IsFile(sysconf) {
		if IsFile(userconf) {
			a.cfg, _ = goconfig.LoadConfigFile(sysconf, userconf)
		} else {
			a.cfg, _ = goconfig.LoadConfigFile(sysconf)
		}
	} else {
		if IsFile(userconf) {
			a.cfg, _ = goconfig.LoadConfigFile(userconf)
		} else {
			a.cfg, _ = goconfig.LoadFromData([]byte(defConfig))
		}
	}

	a.config.port_min = a.cfg.MustInt("server", "port_min", 3000)
	a.config.port_max = a.cfg.MustInt("server", "port_max", 4000)
	a.config.method = a.cfg.MustValue("server", "method", "aes-256-cfb")
	a.config.timeout = a.cfg.MustInt("server", "timeout", 300)
	a.config.bytes_max = a.cfg.MustInt64("server", "bytes_max", 4096000000)

	a.config.debug = a.cfg.MustBool("log", "debug", true)
	a.config.log_info = a.cfg.MustBool("log", "info", true)
	a.config.log_warn = a.cfg.MustBool("log", "warning", true)
	a.config.log_error = a.cfg.MustBool("log", "error", true)
	a.config.log_sql = a.cfg.MustBool("log", "sql", true)

	a.config.dbtype = a.cfg.MustValue("database", "dbtype", "sqlite")
	a.config.dbname = a.cfg.MustValue("database", "dbname", "database.db")
	a.config.dbpass = a.cfg.MustValue("database", "dbpass", "")
	a.config.dbuser = a.cfg.MustValue("database", "dbuser", "")

	a.config.smtp_host = a.cfg.MustValue("mail", "smtp_host", "")
	a.config.smtp_host = a.cfg.MustValue("mail", "smtp_pass", "")
	a.config.smtp_username = a.cfg.MustValue("mail", "smtp_username", "")

	a.config.web_env = a.cfg.MustValue("web", "env", "production")
	a.config.http_host = a.cfg.MustValue("web", "host", "0.0.0.0")
	a.config.http_port = a.cfg.MustInt("web", "port", 8080)
	a.config.static_root = a.cfg.MustValue("web", "static_root", path.Join(getWorkDir(), "static"))
	a.config.template_root = a.cfg.MustValue("web", "template_root", path.Join(getWorkDir(), "templates"))
	a.config.locale_root = a.cfg.MustValue("web", "locale_root", path.Join(getWorkDir(), "locale"))
	a.config.all_methods = []string{"aes-128-cfb", "aes-192-cfb", "aes-256-cfb", "des-cfb", "bf-cfb", "cast5-cfb", "rc4-md5", "rc4", "table"}
}

func (a *App) Write(cfgname string) (err error) {
	userconf := path.Join(a.dir.UserConfig(), cfgname)
	return goconfig.SaveConfigFile(a.cfg, userconf)
}

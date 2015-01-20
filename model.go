package main

import (
	"errors"
	"github.com/go-xorm/xorm"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/sessionauth"
	_ "github.com/mattn/go-sqlite3"
	ss "github.com/shadowsocks/shadowsocks-go/shadowsocks"
	"net/http"
	"time"
)

var (
	x               *xorm.Engine
	ErrUserNotExist = errors.New(__("User does not exist"))
)

type User struct {
	Id            int64
	Username      string `xorm:"unique not null"`
	Fullname      string
	Email         string `xorm:"unique not null"`
	Password      string `xorm:"index not null"`
	Port          int    `xorm:"index unique not null"`
	Method        string `xorm:"default 'aes-256-cfb'"`
	IsActive      bool
	IsAdmin       bool
	Created       time.Time `xorm:"created index"`
	Updated       time.Time `xorm:"updated index"`
	UsedBytes     int64
	LimitBytes    int64
	ActiveCode    string
	authenticated bool `xorm:"-"`
}

type NetTraffic struct {
	UID     int64
	Host    string
	Direct  int
	Number  int64
	Created time.Time `xorm:"created index"`
}

func (u *User) AfterInsert() {
	ReloadServer()
}

func (u *User) AfterUpdate() {
	ReloadServer()
}

func (u *User) AfterDelete() {
	ReloadServer()
}

func (self *User) String() string {
	return self.Username
}

/* server Clienter interface */
func (self User) GetUsername() string {
	return self.Username
}

func (self User) GetId() int64 {
	return self.Id
}

func (self User) GetPassword() string {
	return self.Password
}

func (self User) GetMethod() string {
	return self.Method
}

func (self User) GetPort() int {
	return self.Port
}

func (self User) UpdateInBytes(host string, number int64) bool {
	self.UsedBytes += number
	_, _ = x.Id(self.Id).Update(self)
	net := NetTraffic{UID: self.Id, Host: host, Direct: 0, Number: number}
	affected, err := x.Insert(&net)
	if err != nil && affected == 1 {
		return true
	}
	return false
}

func (self User) UpdateOutBytes(host string, number int64) bool {
	self.UsedBytes += number
	_, _ = x.Id(self.Id).Update(self)
	net := NetTraffic{UID: self.Id, Host: host, Direct: 1, Number: number}
	affected, err := x.Insert(&net)
	if err != nil && affected == 1 {
		return true
	}
	return false
}

func (self User) FirstVisitToday() bool {
	user := &User{Id: self.Id}
	x.Get(user)
	t1 := time.Now()
	if t1.Sub(user.Updated) >= (time.Hour * 24) {
		return true
	}
	return false
}

func (self User) GetLimited() bool {
	user := &User{Id: self.Id}
	x.Get(user)
	return user.UsedBytes >= user.LimitBytes
}

func (self User) GetUsedBytes() (int64, int64) {
	user := &User{Id: self.Id}
	x.Get(user)
	return user.UsedBytes, user.LimitBytes
}

/* server Clienter interface end */

/* sessionauth.User interface */
func GenerateAnonymousUser() sessionauth.User {
	return &User{}
}

func (self *User) IsAuthenticated() bool {
	return self.authenticated
}

func (self *User) Login() {
	self.authenticated = true
}

func (self *User) Logout() {
	self.authenticated = false
}

func (self *User) UniqueId() interface{} {
	return self.Id
}

//should update self.
func (self *User) GetById(id interface{}) error {
	var err error
	var has bool
	has, err = x.Id(id.(int64)).Get(self)
	if !has {
		return ErrUserNotExist
	}
	return err
}

/* sessionauth.User interface end */

// should update self object
func (self *User) Exist() bool {
	if has, err := x.Get(self); has && err == nil {
		return true
	}
	return false
}

func (self *User) Update() error {
	if _, err := x.Id(self.Id).UseBool().Update(self); err != nil {
		return err
	}
	return nil
}

func (self *User) Delete() error {
	_, err := x.Delete(self)
	return err
}

type UserRegisterForm struct {
	Username  string `form:"username" binding:"required"`
	Email     string `form:"email" binding:"required"`
	Password  string `form:"password" binding:"required"`
	Password2 string `form:"password2" binding:"required"`
}

func (u *UserRegisterForm) Validate(errs binding.Errors, req *http.Request) binding.Errors {
	if has, _ := x.Get(&User{Username: u.Username}); has {
		errs.Add([]string{"username"}, "errs", __("User already exist"))
	}
	if has, _ := x.Get(&User{Email: u.Email}); has {
		errs.Add([]string{"email"}, "errs", __("E-mail already used"))
	}
	if len(u.Password) < 6 {
		errs.Add([]string{"password"}, "errs", __("The password length is too short(should more than 6)"))
	}
	if u.Password != u.Password2 {
		errs.Add([]string{"password"}, "errs", __("The New password does not matched"))
	}
	return errs
}

type UserLoginForm struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func (u *UserLoginForm) Validate(errs binding.Errors, req *http.Request) binding.Errors {
	if has, _ := x.Get(&User{Username: u.Username}); !has {
		errs.Add([]string{"username"}, "errs", __("User does not exist"))
	}
	return errs
}

type UserProfileForm struct {
	Fullname    string `form:"fullname"`
	Port        int    `form:"port"`
	Method      string `form:"method"`
	OldPassword string `form:"old-password"`
	Password    string `form:"password"`
	Password2   string `form:"password2"`
}

func (u *UserProfileForm) Validate(errs binding.Errors, req *http.Request) binding.Errors {
	if len(u.Fullname) == 0 {
		errs.Add([]string{"fullname"}, "errs", __("Fullname does not setup"))
	}
	if u.Port < app.config.port_min || u.Port > app.config.port_max {
		errs.Add([]string{"port"}, "errs", __("Port is invalid"))
	}
	if err := ss.CheckCipherMethod(u.Method); err != nil {
		errs.Add([]string{"method"}, "errs", err.Error())
	}
	if len(u.OldPassword) > 0 {
		if !(len(u.Password) > 0 && len(u.Password2) > 0) {
			errs.Add([]string{"password"}, "errs", __("Please input new password"))
		}
		if u.Password != u.Password2 {
			errs.Add([]string{"password"}, "errs", __("The New password does not matched"))
		}
	}
	return errs
}

type EditUserForm struct {
	Username   string `form:"username"`
	Fullname   string `form:"fullname"`
	Email      string `form:"email"`
	Port       int    `form:"port"`
	Method     string `form:"method"`
	Password   string `form:"password"`
	Password2  string `form:"password2"`
	UsedBytes  int64  `form:"used"`
	LimitBytes int64  `form:"limited"`
	IsActive   string `form:"isactive"`
	IsAdmin    string `form:"isadmin"`
}

func (u *EditUserForm) Validate(errs binding.Errors, req *http.Request) binding.Errors {
	if len(u.Fullname) == 0 {
		errs.Add([]string{"fullname"}, "errs", __("Fullname does not setup"))
	}
	if u.Port < app.config.port_min || u.Port > app.config.port_max {
		errs.Add([]string{"port"}, "errs", __("Port is invalid"))
	}
	if err := ss.CheckCipherMethod(u.Method); err != nil {
		errs.Add([]string{"method"}, "errs", err.Error())
	}
	if len(u.Password) > 0 || len(u.Password2) > 0 {
		if u.Password != u.Password2 {
			errs.Add([]string{"password"}, "errs", __("The New password does not matched"))
		}
	}
	if u.UsedBytes > u.LimitBytes {
		errs.Add([]string{"method"}, "errs", __("Net traffic values error"))
	}
	return errs
}

/*/////////////==========ok===========//////////////////*/
func (self *User) Insert() error {
	var err error
	if self.Port == 0 {
		self.Port = NextPort()
	}
	if self.LimitBytes == 0 {
		self.LimitBytes = app.config.bytes_max
	}
	if len(self.Method) == 0 {
		self.Method = app.config.method
	}
	if _, err = x.InsertOne(self); err != nil {
		return err
	}
	// Auto-set admin for user whose ID is 1.
	if self.Id == 1 {
		self.IsAdmin = true
		self.IsActive = true
		_, err = x.Id(self.Id).UseBool().Update(self)
	}
	return err
}

func (self *User) SelectAll() ([]User, error) {
	users := make([]User, 0)
	err := x.Find(&users)
	return users, err
}

func IsAdminExist() (bool, error) {
	return x.Get(&User{Id: 1})
}

func (self *User) GetVisitList(n int) ([]NetTraffic, error) {
	nets := make([]NetTraffic, 0)
	err := x.Where("u_i_d = ?", self.Id).Where("direct = ?", 1).Limit(n).Desc("created").Find(&nets)
	return nets, err
}

func NextPort() int {
	user := new(User)
	x.Cols("port").Desc("id").Limit(1).Get(user)
	if user.Port == 0 {
		return app.config.port_min
	}
	return user.Port + 1
}

func ReloadServer() {
	var k int64
	user := &User{}
	// add new user to listen...
	if users, err := user.SelectAll(); err == nil {
		for _, u := range users {
			if u.IsActive {
				passwdManager.updatePortPasswd(u)
			}
		}
	}
	// close deleted/deactive user's port
	for _, k = range passwdManager.getkeys() {
		if err := user.GetById(k); err != nil || !user.IsActive {
			passwdManager.del(k)
		}
	}
}

func init() {
	var err error
	switch app.config.dbtype {
	case "sqlite":
		x, err = xorm.NewEngine("sqlite3", app.config.dbname)
	case "mysql":
		x, err = xorm.NewEngine("mysql", app.config.dbuser+" "+app.config.dbpass+" "+app.config.dbname)
	}
	if err != nil {
		//return errors.New("尚未设定数据库连接")
		return
	}

	x.ShowDebug = app.config.debug
	x.ShowErr = app.config.log_error
	x.ShowSQL = app.config.log_sql
	x.SetMaxConns(10)

	cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 10000)
	x.SetDefaultCacher(cacher)
	//x.Logger = xorm.NewSimpleLogger(config.log)
	//return x.Sync2(new(User))
	x.Sync2(new(User), new(NetTraffic))
}

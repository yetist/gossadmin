package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessionauth"
	"github.com/martini-contrib/sessions"
	"net/http"
	"strconv"
)

func NewContext() map[string]interface{} {
	ctx := make(map[string]interface{})
	return ctx
}

func notFound(r render.Render) {
	r.HTML(404, "base/404", nil)
}

func registerView(r render.Render) {
	ctx := NewContext()
	r.HTML(200, "user/register", ctx)
}

func registerHandler(r render.Render, errs binding.Errors, reguser UserRegisterForm, req *http.Request) {
	ctx := NewContext()
	if errs.Len() > 0 {
		ctx["error"] = errs[0].Message
		r.HTML(200, "user/register", ctx)
		return
	}
	user := &User{Username: reguser.Username, Email: reguser.Email, Password: reguser.Password}
	if err := user.Insert(); err != nil {
		ctx["error"] = err.Error()
		r.HTML(200, "user/register", ctx)
		return
	}
	/* TODO: send user email to active the account */
	r.Redirect("/")
}

func loginView(r render.Render) {
	ctx := NewContext()
	r.HTML(200, "user/login", ctx)
}

func loginHandler(session sessions.Session, errs binding.Errors, loginUser UserLoginForm, r render.Render, req *http.Request) {
	ctx := NewContext()
	if errs.Len() > 0 {
		ctx["error"] = errs[0].Message
		r.HTML(200, "user/login", ctx)
		return
	}
	user := &User{Username: loginUser.Username, Password: loginUser.Password}
	if user.Exist() {
		println(user.Id, user.Username)
		err := sessionauth.AuthenticateSession(session, user)
		if err != nil {
			ctx["error"] = err.Error()
			r.HTML(500, "user/login", ctx)
			return
		}
		params := req.URL.Query()
		redirect := params.Get(sessionauth.RedirectParam)
		if len(redirect) > 0 {
			r.Redirect(redirect)
			return
		}
		r.Redirect("/")
	} else {
		r.Redirect(sessionauth.RedirectUrl)
		return
	}
}

func logoutHandler(session sessions.Session, user sessionauth.User, r render.Render) {
	sessionauth.Logout(session, user)
	r.Redirect("/")
}

func profileView(r render.Render, su sessionauth.User) {
	ctx := NewContext()
	user := &User{Id: su.UniqueId().(int64)}
	user.Exist()
	nets, _ := user.GetVisitList(10)
	ctx["user"] = user
	ctx["auth"] = user
	if !user.IsActive {
		ctx["error"] = __("Your account is not actived!")
	}
	ctx["page"] = "profile"
	ctx["traffic"] = usedTraffic(user.UsedBytes, user.LimitBytes)
	ctx["networks"] = nets
	ctx["port_min"] = app.config.port_min
	ctx["port_max"] = app.config.port_max
	methods := NewContext()
	for _, v := range app.config.all_methods {
		if user.Method == v {
			methods[v] = "selected"
		} else {
			methods[v] = " "
		}
	}
	ctx["methods"] = methods
	r.HTML(200, "user/profile", ctx)
}

func profileHandler(session sessions.Session, su sessionauth.User, errs binding.Errors, form UserProfileForm, r render.Render, req *http.Request) {
	ctx := NewContext()
	user := &User{Id: su.UniqueId().(int64)}
	user.Exist()
	nets, _ := user.GetVisitList(10)
	if !user.IsActive {
		ctx["error"] = __("Your account is not actived!")
	}
	ctx["user"] = user
	ctx["page"] = "profile"
	ctx["traffic"] = usedTraffic(user.UsedBytes, user.LimitBytes)
	ctx["networks"] = nets
	ctx["port_min"] = app.config.port_min
	ctx["port_max"] = app.config.port_max
	methods := NewContext()
	for _, v := range app.config.all_methods {
		if user.Method == v {
			methods[v] = "selected"
		} else {
			methods[v] = " "
		}
	}
	ctx["methods"] = methods
	if errs.Len() > 0 {
		ctx["error"] = errs[0].Message
		r.HTML(200, "user/profile", ctx)
		return
	}
	olduser := *user
	if len(form.OldPassword) > 0 {
		u := &User{Id: user.Id, Username: user.Username, Password: form.OldPassword}
		if !u.Exist() {
			ctx["error"] = __("user password is wrong")
			r.HTML(200, "user/profile", ctx)
			return
		} else {
			user.Password = form.Password
		}
	}
	user.Fullname = form.Fullname
	user.Port = form.Port
	user.Method = form.Method
	if err := user.Update(); err != nil {
		ctx["error"] = err.Error()
		ctx["user"] = olduser
		r.HTML(200, "user/profile", ctx)
		return
	}
	r.Redirect("/user/profile")
}

/* admin handlers below */
func adminView(r render.Render, su sessionauth.User) {
	ctx := NewContext()
	auth := &User{Id: su.UniqueId().(int64)}
	auth.Exist()
	users, err := auth.SelectAll()
	if err != nil {
		ctx["error"] = err.Error()
		r.HTML(200, "index", ctx)
		return
	}
	ctx["page"] = "dashboard"
	ctx["auth"] = auth
	ctx["users"] = users
	r.HTML(200, "admin/dashboard", ctx)
}

func editUserView(r render.Render, su sessionauth.User, params martini.Params) {
	ctx := NewContext()
	auth := &User{Id: su.UniqueId().(int64)}
	auth.Exist()
	ctx["auth"] = auth
	id, err := strconv.ParseInt(params["id"], 10, 0)
	if err != nil {
		r.Error(500)
		return
	}
	user := &User{Id: id}
	user.Exist()
	ctx["user"] = user
	ctx["port_min"] = app.config.port_min
	ctx["port_max"] = app.config.port_max
	methods := NewContext()
	for _, v := range app.config.all_methods {
		if user.Method == v {
			methods[v] = "selected"
		} else {
			methods[v] = " "
		}
	}
	ctx["methods"] = methods
	ctx["page"] = "dashboard"
	r.HTML(200, "admin/edituser", ctx)
}
func editUserHandler(su sessionauth.User, errs binding.Errors, params martini.Params, form EditUserForm, r render.Render, req *http.Request) {
	ctx := NewContext()
	auth := &User{Id: su.UniqueId().(int64)}
	auth.Exist()
	ctx["auth"] = auth
	ctx["page"] = "dashboard"
	ctx["port_min"] = app.config.port_min
	ctx["port_max"] = app.config.port_max
	id, err := strconv.ParseInt(params["id"], 10, 0)
	if err != nil {
		r.Error(500)
		return
	}
	user := &User{Id: id}
	user.Exist()
	olduser := user
	methods := NewContext()
	for _, v := range app.config.all_methods {
		if user.Method == v {
			methods[v] = "selected"
		} else {
			methods[v] = " "
		}
	}
	ctx["methods"] = methods
	if errs.Len() > 0 {
		ctx["user"] = olduser
		ctx["error"] = errs[0].Message
		r.HTML(200, "admin/edituser", ctx)
		return
	}
	user.Username = form.Username
	user.Fullname = form.Fullname
	user.Email = form.Email
	user.Port = form.Port
	user.Method = form.Method
	user.Password = form.Password
	user.UsedBytes = form.UsedBytes
	user.LimitBytes = form.LimitBytes
	user.IsActive = form.IsActive == "on"
	user.IsAdmin = form.IsAdmin == "on"
	if err := user.Update(); err != nil {
		ctx["error"] = err.Error()
		ctx["user"] = olduser
		r.HTML(200, "admin/edituser", ctx)
		return
	}
	r.Redirect("/admin/user/" + params["id"])
}

func activeUserView(r render.Render, su sessionauth.User, params martini.Params) {
	id, err := strconv.ParseInt(params["id"], 10, 0)
	if err != nil {
		r.Error(500)
		return
	}
	user := &User{Id: id}
	user.Exist()
	user.IsActive = true
	user.Update()
	r.Redirect("/admin/")
}

func deleteUserView(r render.Render, su sessionauth.User, params martini.Params) {
	id, err := strconv.ParseInt(params["id"], 10, 0)
	if err != nil {
		r.Error(500)
		return
	}
	user := &User{Id: id}
	user.Delete()
	r.Redirect("/admin/")
}

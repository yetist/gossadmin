package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessionauth"
	"github.com/martini-contrib/sessions"
	"github.com/yetist/middleware/i18n"
	"html/template"
	"strconv"
)

func runweb() {
	store := sessions.NewCookieStore([]byte("sec#@!ret123"))
	store.Options(sessions.Options{
		MaxAge: 0,
	})
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Directory: app.config.template_root,
		Funcs: []template.FuncMap{
			{
				"__": __,
				"set": func(renderArgs map[string]interface{}, key string, value interface{}) template.HTML {
					if renderArgs == nil {
						renderArgs = make(map[string]interface{})
						fmt.Println("set args is nil", key, value)
					}
					renderArgs[key] = value
					return template.HTML("")
				},
				"block": func(renderArgs map[string]interface{}, key string, value ...interface{}) template.HTML {
					if len(value) == 0 {
						return template.HTML("")
					}

					line := fmt.Sprint(value[0])
					for _, argValue := range value[1:] {
						line += "\n" + fmt.Sprint(argValue)
					}
					if renderArgs == nil {
						renderArgs = make(map[string]interface{})
						fmt.Println("block args is nil", key, value)
					}
					renderArgs[key] = template.HTML(line)
					return template.HTML("")
				},
			},
		},
	}))
	m.Use(i18n.I18n(i18n.Options{
		Domain:    PkgName,
		Directory: app.config.locale_root,
		Parameter: "lang",
	}))
	m.Use(martini.Static("static", martini.StaticOptions{
		Prefix: app.config.static_root,
	}))
	m.Use(martini.Static(app.config.static_root, martini.StaticOptions{
		Prefix: "assets",
	}))

	m.Use(sessions.Sessions("mosday", store))
	m.Use(sessionauth.SessionUser(GenerateAnonymousUser))

	m.Group("", func(r martini.Router) {
		r.Get("/", sessionauth.LoginRequired, func(r render.Render) { r.Redirect("/user/profile") })
		r.Get("/login", loginView)
		r.Post("/login", binding.Form(UserLoginForm{}), loginHandler)
		r.Get("/logout", sessionauth.LoginRequired, logoutHandler)
	})

	m.Group("/user", func(r martini.Router) {
		r.Get("/register", registerView)
		r.Post("/register", binding.Form(UserRegisterForm{}), registerHandler)

		r.Get("/profile", sessionauth.LoginRequired, profileView)
		r.Post("/profile", sessionauth.LoginRequired, binding.Form(UserProfileForm{}), profileHandler)
		//r.Get("/active", registerHandler)
	})

	m.Group("/admin", func(r martini.Router) {
		r.Get("/", adminView)
		r.Get("/user/(?P<id>[0-9]+)", editUserView)
		r.Post("/user/(?P<id>[0-9]+)", binding.Form(EditUserForm{}), editUserHandler)
		r.Get("/user/(?P<id>[0-9]+)/active", activeUserView)
		r.Get("/user/(?P<id>[0-9]+)/delete", deleteUserView)
	}, sessionauth.LoginRequired, AdminRequired)

	m.NotFound(notFound)

	martini.Env = app.config.web_env
	m.RunOnAddr(":" + strconv.Itoa(app.config.http_port))
}

func AdminRequired(r render.Render, user sessionauth.User, c martini.Context) {
	id := user.UniqueId().(int64)
	u := &User{Id: id}
	u.Exist()
	if !u.IsAdmin {
		notFound(r)
	}
}

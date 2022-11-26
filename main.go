package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"todolist.go/db"
	"todolist.go/service"
)

const port = 8000

func main() {
	// initialize DB connection
	dsn := db.DefaultDSN(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	if err := db.Connect(dsn); err != nil {
		log.Fatal(err)
	}

	// initialize Gin engine
	engine := gin.Default()
	engine.LoadHTMLGlob("views/*.html")

	// prepare session
	store := cookie.NewStore([]byte("my-secret"))
	engine.Use(sessions.Sessions("user-session", store))

	// routing
	engine.Static("/assets", "./assets")
	engine.GET("/", service.Home)
	engine.GET("/list", service.LoginCheck, service.TaskList)
	taskGroup := engine.Group("/task")
	taskGroup.Use(service.LoginCheck)
	{
		taskGroup.GET("/:id", service.UserCheck, service.ShowTask)
		taskGroup.GET("/new", service.NewTaskForm)
		taskGroup.POST("/new", service.RegisterTask)
		taskGroup.GET("/edit/:id", service.UserCheck, service.EditTaskForm)
		taskGroup.POST("/edit/:id", service.UserCheck, service.UpdateTask)
		taskGroup.GET("/delete/:id", service.UserCheck, service.DeleteTask)
	}

	// ユーザ登録
	engine.GET("/user/new", service.NewUserForm)
	engine.POST("/user/new", service.RegisterUser)

	// ログイン
	engine.GET("/login", service.ShowLoginForm)
	engine.POST("/login", service.Login)

	// ログアウト
	engine.GET("/logout", service.Logout)

	// アクセス制限
	engine.GET("/forbidden", service.Prohibited)

	//マイページ
	engine.GET("/mypage", service.LoginCheck, service.ShowMyPage)
	engine.GET("/mypage/change_username", service.ShowUsernameChanger)
	engine.POST("/mypage/change_username", service.ChangeUsername)
	engine.GET("/mypage/change_password", service.ShowPasswordChanger)
	engine.POST("/mypage/change_password", service.ChangePassword)
	engine.GET("/mypage/delete_account", service.ShowAccountDeleter)
	engine.POST("/mypage/delete_account", service.DeleteAccount)

	// start server
	engine.Run(fmt.Sprintf(":%d", port))
}

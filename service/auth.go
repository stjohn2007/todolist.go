package service

import (
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

func LoginCheck(ctx *gin.Context) {
	if sessions.Default(ctx).Get(userkey) == nil {
		ctx.Redirect(http.StatusFound, "/login")
		ctx.Abort()
	} else {
		ctx.Next()
	}
}

func UserCheck(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		ctx.Abort()
	}

	// parse taskID given as a parameter
	taskID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		ctx.Abort()
	}

	// Get task userID
	var owners []database.Owner
	err = db.Select(&owners, "SELECT user_id, task_id FROM ownership WHERE task_id = ?", taskID)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		ctx.Abort()
	}

	// Get session userID
	session_userID := sessions.Default(ctx).Get(userkey)

	// Execute taks authentication
	check := false
	for _, owner := range owners {
		if owner.UserID == session_userID {
			check = true
		}
	}
	if check {
		ctx.Next()
	} else {
		ctx.Redirect(http.StatusFound, "/forbidden")
		ctx.Abort()
	}

}

func Prohibited(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "prohibited.html", gin.H{"Title": "Access is Prohibited"})
}

package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

func NewUserForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "new_user_form.html", gin.H{"Title": "Register user"})
}

func hash(pw string) []byte {
	const salt = "todolist.go#"
	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte(pw))
	return h.Sum(nil)
}

func RegisterUser(ctx *gin.Context) {
	// フォームデータの受け取り
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	password_check := ctx.PostForm("password_check")
	switch {
	case username == "":
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Usernane is not provided", "Username": username})
		return
	case password == "":
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Password": password})
		return
	case password != password_check:
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password doesn't match confirm password"})
		return
	}

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 重複チェック
	var duplicate int
	err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	if duplicate > 0 {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username, "Password": password})
		return
	}

	// DB への保存
	result, err := db.Exec("INSERT INTO users(name, password) VALUES (?, ?)", username, hash(password))
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 保存状態の確認
	id, _ := result.LastInsertId()
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.Redirect(http.StatusFound, "/login")
}

func ShowLoginForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "Login"})
}

const userkey = "user"

func Login(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	// Get db connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// ユーザの取得
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE name = ?", username)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
		return
	}

	// パスワードの照合
	if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
		return
	}

	// セッションの保存
	session := sessions.Default(ctx)
	session.Set(userkey, user.ID)
	session.Save()

	ctx.Redirect(http.StatusFound, "/list")
}

func Logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()
	ctx.Redirect(http.StatusFound, "/")
}

func ShowMyPage(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get(userkey)

	// Get db connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	var username string
	err = db.Get(&username, "SELECT name FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	fmt.Printf("%s\n", username)

	ctx.HTML(http.StatusOK, "mypage.html", gin.H{"Title": "My Page", "Username": username})
}

func ShowUsernameChanger(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get(userkey)
	fmt.Printf("%d\n", userID)

	// Get db connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	var username string
	err = db.Get(&username, "SELECT name FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	ctx.HTML(http.StatusOK, "username_changer.html", gin.H{"Title": "Change Username", "Current_Username": username})
}

func ChangeUsername(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get(userkey)

	// Get current username
	current_username, exist := ctx.GetPostForm("current_username")
	if !exist {
		Error(http.StatusBadRequest, "No current username is given")(ctx)
		return
	}

	// Get new username
	new_username, exist := ctx.GetPostForm("new_username")
	if !exist {
		Error(http.StatusBadRequest, "No new username is given")(ctx)
		return
	}

	// Get password
	password, exist := ctx.GetPostForm("password")
	if !exist {
		Error(http.StatusBadRequest, "No password is given")(ctx)
		return
	}

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 重複チェック
	var duplicate int
	err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", new_username)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	if duplicate > 0 {
		ctx.HTML(http.StatusBadRequest, "username_changer.html", gin.H{"Title": "Change Username", "Error": "Username is already taken", "Current_Username": current_username, "New_Username": new_username})
		return
	}

	// ユーザの取得
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// パスワードの照合
	if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
		ctx.HTML(http.StatusBadRequest, "username_changer.html", gin.H{"Title": "Change Username", "Current_Username": current_username, "New_Username": new_username, "Error": "Incorrect password"})
		return
	}

	// Update data with given title and is_done on DB
	_, err = db.Exec("UPDATE users SET name=? WHERE id=?", new_username, userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	ctx.Redirect(http.StatusFound, "/mypage")
}

func ShowPasswordChanger(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get(userkey)
	fmt.Printf("%d\n", userID)

	// Get db connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	var username string
	err = db.Get(&username, "SELECT name FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	ctx.HTML(http.StatusOK, "password_changer.html", gin.H{"Title": "Change Password", "Username": username})
}

func ChangePassword(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get(userkey)

	// Get curren password
	current_password, exist := ctx.GetPostForm("current_password")
	if !exist {
		Error(http.StatusBadRequest, "No current password is given")(ctx)
		return
	}

	// Get new password
	new_password, exist := ctx.GetPostForm("new_password")
	if !exist {
		Error(http.StatusBadRequest, "No new password is given")(ctx)
		return
	}

	// Get new password check
	new_password_check, exist := ctx.GetPostForm("new_password_check")
	if !exist {
		Error(http.StatusBadRequest, "No new password check is given")(ctx)
		return
	}

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Get user
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// 新パスワードの照合
	if new_password != new_password_check {
		ctx.HTML(http.StatusBadRequest, "password_changer.html", gin.H{"Title": "Change Password", "Error": "Password mismatch", "Username": user.Name})
		return
	}

	// パスワードの照合
	if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(current_password)) {
		ctx.HTML(http.StatusBadRequest, "password_changer.html", gin.H{"Title": "Change Password", "Error": "Incorrect Password", "Username": user.Name})
		return
	} else {
		// Update data with given title and is_done on DB
		_, err = db.Exec("UPDATE users SET password=? WHERE id=?", hash(new_password), userID)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}

		ctx.Redirect(http.StatusFound, "/mypage")
	}
}

func ShowAccountDeleter(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get(userkey)
	fmt.Printf("%d\n", userID)

	// Get db connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	var username string
	err = db.Get(&username, "SELECT name FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	ctx.HTML(http.StatusOK, "account_deleter.html", gin.H{"Title": "Delete Account", "Username": username})
}

func DeleteAccount(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get(userkey)

	// Get curren username
	password, exist := ctx.GetPostForm("password")
	if !exist {
		Error(http.StatusBadRequest, "No password is given")(ctx)
		return
	}

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Get user
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// チェックボックスの確認
	_, exist = ctx.GetPostForm("confirm")
	if !exist {
		ctx.HTML(http.StatusBadRequest, "account_deleter.html", gin.H{"Title": "Change Password", "Error": "Please Check Confirmation", "Username": user.Name})
		return
	}

	// パスワードの照合
	if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
		ctx.HTML(http.StatusBadRequest, "account_deleter.html", gin.H{"Title": "Change Password", "Error": "Incorrect Password", "Username": user.Name})
		return
	}

	// アカウント削除開始（物理削除）
	// ownershipからuserIDが関連するレコードを削除
	tx := db.MustBegin()
	_, err = tx.Exec("DELETE FROM ownership WHERE user_id=?", userID)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 誰も所有していないタスクを削除
	_, err = tx.Exec("DELETE FROM tasks WHERE NOT exists (SELECT * FROM ownership WHERE tasks.id = ownership.task_id)")
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// usersからuserを削除
	_, err = tx.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	tx.Commit()
	ctx.Redirect(http.StatusFound, "/logout")
}

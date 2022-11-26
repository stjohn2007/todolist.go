package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

// TaskList renders list of tasks in DB
func TaskList(ctx *gin.Context) {
	// Get User ID
	userID := sessions.Default(ctx).Get("user")

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Get query parameter
	kw := ctx.Query("kw")
	is_done := ctx.Query("is_done")

	// Get tasks in DB
	var tasks []database.Task
	query := "SELECT id, title, created_at, is_done FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ?"
	kw_query := ""
	done_query := ""

	switch {
	case is_done == "true":
		done_query = " AND is_done = " + "1"
	case is_done == "false":
		done_query = " AND is_done = " + "0"
	default:
		done_query = ""
	}

	switch {
	case kw != "":
		kw_query = " AND title LIKE ?"
		err = db.Select(&tasks, query+done_query+kw_query, userID, "%"+kw+"%")
	default:
		kw_query = ""
		err = db.Select(&tasks, query+done_query+kw_query, userID)
	}
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks})
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// parse ID given as a parameter
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get a task with given ID
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Render task
	ctx.HTML(http.StatusOK, "task.html", task) // Modify it!!
}

func NewTaskForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration"})
}

func RegisterTask(ctx *gin.Context) {
	// Get user ID
	userID := sessions.Default(ctx).Get("user")

	// Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist {
		Error(http.StatusBadRequest, "No title is given")(ctx)
		return
	}

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx := db.MustBegin()
	result, err := tx.Exec("INSERT INTO tasks (title) VALUES (?)", title)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	_, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", userID, taskID)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx.Commit()
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/task/%d", taskID))
}

func EditTaskForm(ctx *gin.Context) {
	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Get target task
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Render edit form
	ctx.HTML(http.StatusOK, "form_edit_task.html", gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task})
}

func UpdateTask(ctx *gin.Context) {
	// Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist {
		Error(http.StatusBadRequest, "No title is given")(ctx)
		return
	}

	// Get task done flag
	done, exist := ctx.GetPostForm("is_done")
	if !exist {
		Error(http.StatusBadRequest, "No is_done is given")(ctx)
		return
	}

	is_done, _ := strconv.ParseBool(done)

	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Update data with given title and is_done on DB
	_, err = db.Exec("UPDATE tasks SET title=?, is_done=? WHERE id=?", title, is_done, id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Render status
	path := fmt.Sprintf("/task/%d", id)
	ctx.Redirect(http.StatusFound, path)
}

func DeleteTask(ctx *gin.Context) {
	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Delete the task from DB
	tx := db.MustBegin()
	_, err = tx.Exec("DELETE FROM tasks WHERE id=?", id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	_, err = tx.Exec("DELETE FROM ownership WHERE task_id=?", id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx.Commit()
	// Redirect to /list
	ctx.Redirect(http.StatusFound, "/list")
}

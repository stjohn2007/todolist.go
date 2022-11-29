package service

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	tag := ctx.Query("tag")
	is_done := ctx.Query("is_done")
	deadline := ctx.Query("deadline")
	priority := ctx.Query("priority")
	order := ctx.Query("order")
	rev := ctx.Query("rev")

	// Get current datetime
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	current_time := time.Now().In(jst).Format("2006-01-02 15:04:05")

	// Get tasks in DB
	var tasks []database.Task
	query := "SELECT id, title, tag, deadline, priority, is_done FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ?"
	kw_query := ""
	tag_query := ""
	done_query := ""
	deadline_query := ""
	priority_query := ""
	order_query := ""

	switch {
	case tag != "":
		tag_query = " AND tag = \"" + tag + "\""
	default:
		tag_query = ""
	}

	switch {
	case is_done == "true":
		done_query = " AND is_done = 1"
	case is_done == "false":
		done_query = " AND is_done = 0"
	default:
		done_query = ""
	}

	switch {
	case deadline == "past":
		deadline_query = " AND deadline < '" + current_time + "'"
	case deadline == "yet":
		deadline_query = " AND deadline > '" + current_time + "'"
	default:
		deadline_query = ""
	}

	switch {
	case priority == "true":
		priority_query = " AND priority = 1"
	case priority == "false":
		priority_query = " AND priority = 0"
	default:
		priority_query = ""
	}

	switch {
	case order != "":
		order_query = " ORDER BY " + order + " " + rev
	default:
		order_query = ""
	}

	switch {
	case kw != "":
		kw_query = " AND title LIKE ?"
		err = db.Select(&tasks, query+tag_query+done_query+deadline_query+priority_query+kw_query+order_query, userID, "%"+kw+"%")
	default:
		kw_query = ""
		err = db.Select(&tasks, query+tag_query+done_query+deadline_query+priority_query+kw_query+order_query, userID)
		fmt.Println(query + tag_query + done_query + deadline_query + priority_query + kw_query + order_query)
	}
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Get tag list
	var tags []string
	err = db.Select(&tags, "SELECT tag FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	if order == "" {
		order = "id"
	}
	if rev == "" {
		rev = "ASC"
	}
	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks, "Kw": kw, "Tags": tags, "Is_done": is_done, "Priority": priority, "Deadline": deadline, "Order": order, "Rev": rev})
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

	// Get owner users of the task
	var users []string
	err = db.Select(&users, "SELECT name FROM users INNER JOIN ownership ON user_id = id WHERE task_id=?", task.ID)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Render task
	ctx.HTML(http.StatusOK, "task.html", gin.H{"Task": task, "Users": users})
	fmt.Println(users)
}

func NewTaskForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration", "Tag": "その他"})
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

	// Get task tag
	tag, _ := ctx.GetPostForm("tag")

	// Get task deadline
	deadline, exist := ctx.GetPostForm("deadline")
	if !exist {
		Error(http.StatusBadRequest, "No deadline is given")(ctx)
		return
	}

	// Get task priority
	priority, exist := ctx.GetPostForm("priority")
	if !exist {
		Error(http.StatusBadRequest, "No priority is given")(ctx)
		return
	}
	priority_b, _ := strconv.ParseBool(priority)
	fmt.Println(priority_b)

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Register task
	tx := db.MustBegin()
	result, err := tx.Exec("INSERT INTO tasks (title, tag, deadline, priority) VALUES (?, ?, ?, ?)", title, tag, deadline, priority_b)
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

	// Register own ownership
	_, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", userID, taskID)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Register other ownership
	owner_name := ""
	owner_form_idx := 0
	var ownerID string
	for {
		owner_name, exist = ctx.GetPostForm(fmt.Sprintf("owner_%d", owner_form_idx))
		owner_form_idx = owner_form_idx + 1
		if !exist {
			break
		}
		if owner_name == "" {
			continue
		}
		err = db.Get(&ownerID, "SELECT id FROM users WHERE name=?", owner_name)
		if err != nil {
			tx.Rollback()
			ctx.HTML(http.StatusBadRequest, "form_new_task.html", gin.H{"Title": "Task registration", "TaskTitle": title, "Deadline": deadline, "Priority": priority, "Error": "user: " + owner_name + " does not exist."})
			return
		}
		_, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", ownerID, taskID)
		if err != nil {
			tx.Rollback()
			ctx.HTML(http.StatusBadRequest, "form_new_task.html", gin.H{"Title": "Task registration", "TaskTitle": title, "Deadline": deadline, "Priority": priority, "Error": "user: " + owner_name + " is already owner."})
			return
		}
	}

	tx.Commit()
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/task/%d", taskID))
}

func EditTaskForm(ctx *gin.Context) {
	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	fmt.Println(id)
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

	// Get task tag
	tag, _ := ctx.GetPostForm("tag")

	// Get task deadline
	deadline, exist := ctx.GetPostForm("deadline")
	if !exist {
		Error(http.StatusBadRequest, "No deadline is given")(ctx)
		return
	}

	// Get task priority
	priority, exist := ctx.GetPostForm("priority")
	if !exist {
		Error(http.StatusBadRequest, "No priority is given")(ctx)
		return
	}
	priority_b, _ := strconv.ParseBool(priority)

	// Get task done flag
	done, exist := ctx.GetPostForm("is_done")
	if !exist {
		Error(http.StatusBadRequest, "No is_done is given")(ctx)
		return
	}

	is_done, _ := strconv.ParseBool(done)

	// ID の取得
	taskID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	var task database.Task
	task.ID = uint64(taskID)
	task.Title = title
	deadline_time, _ := time.Parse("2006-01-02T15:04", deadline)
	task.Deadline = deadline_time.Add(time.Hour * -9)
	task.Priority = priority_b
	task.IsDone = is_done

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Update data with given title and is_done on DB
	tx := db.MustBegin()
	_, err = tx.Exec("UPDATE tasks SET title=?, tag=?, deadline=?, priority=?, is_done=? WHERE id=?", title, tag, deadline, priority_b, is_done, taskID)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Register other ownership
	owner_name := ""
	owner_form_idx := 0
	var ownerID string
	for {
		owner_name, exist = ctx.GetPostForm(fmt.Sprintf("owner_%d", owner_form_idx))
		owner_form_idx = owner_form_idx + 1
		if !exist {
			break
		}
		if owner_name == "" {
			continue
		}
		err = db.Get(&ownerID, "SELECT id FROM users WHERE name=?", owner_name)
		if err != nil {
			tx.Rollback()
			ctx.HTML(http.StatusBadRequest, "form_edit_task.html", gin.H{"Title": "Task registration", "Task": task, "Error": "user: " + owner_name + " does not exist."})
			return
		}
		_, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", ownerID, taskID)
		if err != nil {
			tx.Rollback()
			ctx.HTML(http.StatusBadRequest, "form_edit_task.html", gin.H{"Title": "Task registration", "Task": task, "Error": "user: " + owner_name + " is already owner."})
			return
		}
	}

	tx.Commit()

	// Render status
	path := fmt.Sprintf("/task/%d", taskID)
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

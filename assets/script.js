/* placeholder file for JavaScript */
const confirm_task_delete = (id) => {
    if(window.confirm(`Task ${id} を削除します．よろしいですか？`)) {
        location.href = `/task/delete/${id}`;
    }
}
 
const confirm_task_update = (id) => {
    if(window.confirm(`Task ${id} を更新します．よろしいですか？`)) {
        location.href = `/task/edit/${id}`;
    }
}
 
const confirm_username_update = () => {
    if(window.confirm(`ユーザ名を更新します．よろしいですか？`)) {
        location.href = `/mypage`;
    }
}

const confirm_password_update = () => {
    if(window.confirm(`パスワードを更新します．よろしいですか？`)) {
        location.href = `/mypage`;
    }
}

const confirm_delete_account = () => {
    if(window.confirm(`アカウントを削除します．よろしいですか？`)) {
        location.href = `/`;
    }
}

// task_listのtaskの表示を制御
function generate_task() {
    var tasks = document.getElementsByClassName("task");
    const today = new Date()
    const youbi_array = ["(日)", "(月)", "(火)", "(水)", "(木)", "(金)", "(土)"]
    for (var i = 0; i < tasks.length; i++) {
        // 優先度による表示切り替え
        var priority = tasks[i].getElementsByTagName('td')[4];
        if (priority.textContent == "高"){
            tasks[i].style.fontWeight = "bold"
        }

        // 締め切りによる表示切り替え
        var deadline = tasks[i].getElementsByTagName('td')[3];
        const deadline_date = str2date(deadline.textContent);

        var compare_date = new Date(today.getFullYear(), today.getMonth(), today.getDate(), today.getHours(), today.getMinutes())
        var compare_deadline = new Date(deadline_date.getFullYear(), deadline_date.getMonth(), deadline_date.getDate(), deadline_date.getHours(), deadline_date.getMinutes())

        console.log(compare_date)
        console.log(compare_deadline)

        // 締切過ぎ
        if (+compare_date > +compare_deadline){
            deadline.textContent = `${deadline_date.getMonth() + 1}月${deadline_date.getDate()}日${youbi_array[deadline_date.getDay()]}（締切過ぎ）`
            if (tasks[i].getElementsByTagName('td')[4].textContent == "未完了"){
                deadline.style.color = "red"
            }
            continue
        }

        compare_date = new Date(today.getFullYear(), today.getMonth(), today.getDate())
        compare_deadline = new Date(deadline_date.getFullYear(), deadline_date.getMonth(), deadline_date.getDate())

        // 1日以内
        if (+compare_date == +compare_deadline){
            deadline.textContent = `${deadline_date.getMonth() + 1}月${deadline_date.getDate()}日${youbi_array[deadline_date.getDay()]}（本日中）`
            continue
        }

        // 1日後
        compare_date.setDate(compare_date.getDate() + 1)
        if (+compare_date == +compare_deadline){
            deadline.textContent = `${deadline_date.getMonth() + 1}月${deadline_date.getDate()}日${youbi_array[deadline_date.getDay()]}（1日後）`
            continue
        }

        // 2日後
        compare_date.setDate(compare_date.getDate() + 1)
        if (+compare_date == +compare_deadline){
            deadline.textContent = `${deadline_date.getMonth() + 1}月${deadline_date.getDate()}日${youbi_array[deadline_date.getDay()]}（2日後）`
            continue
        }

        deadline.textContent = `${deadline_date.getMonth() + 1}月${deadline_date.getDate()}日${youbi_array[deadline_date.getDay()]}`
    }
}

// 締め切りをプリセットする
function preset_deadline(date){
    if (date === null){
        var preset_date = new Date()
        preset_date.setDate(preset_date.getDate() + 7);
    }
    else{
        var preset_date = new Date(date)
    }
    const year = preset_date.getFullYear().toString().padStart(4, '0');
    const month = (preset_date.getMonth() + 1).toString().padStart(2, '0');
    const day = preset_date.getDate().toString().padStart(2, '0');
    const hour = preset_date.getHours().toString().padStart(2, '0');
    const minutes = preset_date.getMinutes().toString().padStart(2, '0');
    document.getElementById('cal').value = `${year}-${month}-${day} ${hour}:${minutes}`
}

// タスクページの表示制御
function display_task() {
    var deadline = document.getElementsByTagName("dd")[1]
    var deadline_date = str2date(deadline.textContent)
    deadline.textContent = `${deadline_date.getFullYear()}年${deadline_date.getMonth() + 1}月${deadline_date.getDate()}日${deadline_date.getHours().toString().padStart(2, '0')}時${deadline_date.getMinutes().toString().padStart(2, '0')}分`

    var created = document.getElementsByTagName("dd")[2]
    var created_date = str2date(created.textContent)
    created.textContent = `${created_date.getFullYear()}年${created_date.getMonth() + 1}月${created_date.getDate()}日${created_date.getHours().toString().padStart(2, '0')}時${created_date.getMinutes().toString().padStart(2, '0')}分`
}

// str型の日付をDate型に変換して返す
function str2date(date) {
    const year = date.split(' ')[0].split('-')[0];
    const month = date.split(' ')[0].split('-')[1];
    const day = date.split(' ')[0].split('-')[2];
    const hour = date.split(' ')[1].split(':')[0];
    const minute = date.split(' ')[1].split(':')[1];
    const second = date.split(' ')[1].split(':')[2];
    const new_date = new Date(year, month - 1, day, hour, minute, second)
    return new_date
}

// タスク管理者を追加するフォームを増やす
function add_owner_form() {
    const owner_form_group = document.getElementById("owner_form")
    var owner_length = 0
    try{
        owner_length = owner_form_group.getElementsByTagName("input").length
    } catch (error) {
        owner_length = 0
    }
    const owner_form = document.createElement("input")
    owner_form.type = "text"
    owner_form.name = "owner_" + owner_length.toString()
    owner_form_group.appendChild(owner_form)
    const br = document.createElement("br")
    owner_form_group.appendChild(br)
}


// パスワードの表示設定
function pushHideButton() {
    var password = document.getElementById("password")
    var eye = document.getElementById("buttonEye")
    if (password.type === "text"){
        password.type = "password"
        eye.className = "fa fa-eye"
    } else {
        password.type = "text"
        eye.className = "fa fa-eye-slash"
    }
}
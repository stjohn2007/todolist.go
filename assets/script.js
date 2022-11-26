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
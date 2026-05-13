package test_sample

import (
	"net/http"
)

// @desc 获取用户列表
// @param page int - 页码 [required] [query]
// @param page_size int - 每页数量 [query]
// @return 200 - 成功
// @return 401 - 未授权
// @example-request
// {
//   "page": 1,
//   "page_size": 20
// }
// @example-response
// {
//   "code": 0,
//   "data": [{"id": 1, "name": "test"}]
// }
func GetUsers(w http.ResponseWriter, r *http.Request) {
}

// @desc 创建用户
// @param user object - 用户信息 [required] [body]
// @return 200 - 创建成功
// @return 400 - 参数错误
// @example-request
// {"name": "new_user", "email": "test@example.com"}
// @example-response
// {"code": 0, "message": "success"}
func CreateUser(w http.ResponseWriter, r *http.Request) {
}

// @desc 获取用户详情
// @param id int - 用户ID [required] [path]
// @return 200 - 成功
// @return 404 - 用户不存在
func GetUserDetail(w http.ResponseWriter, r *http.Request) {
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
}

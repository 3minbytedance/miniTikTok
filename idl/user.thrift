namespace go user

struct User {
    1: i64 id, // 用户id
    2: string name, // 用户名称
    3: i32 follow_count, // 关注总数
    4: i32 follower_count, // 粉丝总数
    5: bool is_follow, // true-已关注，false-未关注
    6: string avatar, // 用户头像
    7: string background_image, // 用户个人顶部大图
    8: string signature, // 个人简介
    9: string total_favorited, // 获赞数量
    10: i32 work_count, // 作品数量
    11: i32 favorite_count, // 点赞数量
}

struct UserRegisterRequest {
    1: string username, // 注册用户名，最长32个字符
    2: string password, // 密码，最长32个字符
}

struct UserRegisterResponse {
    1: i32 status_code, // 状态码，0-成功，其他值-失败
    2: string status_msg, // 返回状态描述
    3: i64 user_id, // 用户id
    4: string token, // 用户鉴权token
}

struct UserLoginRequest {
    1: string username, // 注册用户名，最长32个字符
    2: string password, // 密码，最长32个字符
}

struct UserLoginResponse {
    1: i32 status_code, // 状态码，0-成功，其他值-失败
    2: string status_msg, // 返回状态描述
    3: i64 user_id, // 用户id
    4: string token, // 用户鉴权token
}

struct UserInfoByIdRequest {
    1: i64 actor_id, // 调用者id
    2: i64 user_id,  // 查询该id的信息
}

struct UserInfoByIdResponse {
    1: i32 status_code, // 状态码，0-成功，其他值-失败
    2: string status_msg, // 返回状态描述
    3: User user, // 用户信息
}

struct UserInfoByNameRequest {
    1: string username, // 用户名
}

struct UserInfoByNameResponse {
    1: i32 status_code, // 状态码，0-成功，其他值-失败
    2: string status_msg, // 返回状态描述
    3: i64 user_id, // 用户信息
    4: string password,
}

struct UserExistsRequest {
    1: string username,
}

struct UserExistsResponse {
    1: bool exist,
}

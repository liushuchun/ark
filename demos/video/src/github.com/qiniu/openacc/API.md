openacc api 协议
---------

# 常规接口

## 授权

### 用户名/密码登录

请求包：

```
POST /v1/user/token
Content-Type: application/x-www-form-urlencoded

grant_type=password&
client_id=<ClientAppId>&
device_id=<DeviceId>&
username=<UserName>& (或者email=<Email> 或者 uid=<UserId> 三选一）
password=<Password>&
scope=<Scope>
```

返回包：

```
200 OK
Content-Type: application/json

{
  access_token: <AccessToken>
  token_type: <TokenType> #目前只有bearer
  expires_in: <ExpireSconds>
  refresh_token: <RefreshToken>
}
```

### RefreshToken续约

请求包：

```
POST /v1/user/token
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token&
client_id=<ClientAppId>&
device_id=<DeviceId>&
refresh_token=<RefreshToken>&
scope=<Scope>
```

返回包同 "用户名/密码登录"。

## RefreshToken清除

请求包：

```
POST /v1/user/logout
Content-Type: application/x-www-form-urlencoded

refresh_token=<RefreshToken>
```

返回包：

```
200 OK
```

# 管理员接口

## 创建账号

请求包：

```
POST /v1/user/new
Content-Type: application/x-www-form-urlencoded
Authorization: Bearer <AdminToken>

username=<UserName>&
password=<Password>&
email=<Email>&
email_status=<EmailStatus>&
utype=<Utype>
```

* `<Password>`: 用户的新密码。
* `<Email>`: 用户的登录邮箱。
* `<EmailStatus>`: 其中 1：表示用户邮箱已经验证； 0x80: 表示用户邮箱未验证。
* `<Status>`: 其中 1: 表示用户正常；0x8: 表示用户被冻结（disabled）。
* `<Utype>`: 用户帐号类型。对于管理员（Admin），他能够修改非管理员用户；对于超级管理员，他能够修改所有用户。

返回包：

```
200 OK
Content-Type: application/json

{
  "id": <UserId>
}
```

## 更新帐号

请求包：

```
POST /v1/user/update
Authorization: Bearer <AdminToken>

username=<UserName>& (或者email=<Email> 或者 uid=<UserId> 三选一）
password=<Password>&
new_email=<NewEmail>&
email_status=<EmailStatus>&
status=<Status>&
utype=<Utype>
```

* `<Password>`: 如果非空，表示是用户的新密码。
* `<NewEmail>`: 如果非空，表示是用户新的登录邮箱。
* `<EmailStatus>`: 其中 1：表示用户邮箱已经验证； 0x80: 表示用户邮箱未验证。
* `<Status>`: 其中 1: 表示用户正常；0x8: 表示用户被冻结（disabled）。
* `<Utype>`: 如果非空，表示是用户新的帐号类型。

返回包：

```
200 OK
```

## 查询用户

请求包：

```
GET /v1/user/info?username=<UserName> (或者email=<Email> 或者 uid=<UserId> 三选一）
Authorization: Bearer <AdminToken>
```

返回包：

```
200 OK
Content-Type: application/json

{
  "id": <UserId>,
  "username": <UserName>,
  "email": <Email>,
  "email_status": <EmailStatus>,
  "status": <Status>,
  "utype": <Utype>
}
```

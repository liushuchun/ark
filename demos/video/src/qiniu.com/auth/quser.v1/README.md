# QUser Authorization

```
<Method> <PathWithRawQuery>
<Header1>: <Value1>
<Header2>: <Value2>
...
Host: <Host>
...
Content-Type: <ContentType>
...
Authorization: QUser <QUserAK>:<QUserSign>
...

<Body>
```

其中 `<QUserAK>` 是这样的形式：

```
<AK>/<Expiry>/uid=<EndUserId>&...
```

对于上面这样一个请求，我们构造如下这个待签名的 `<Data>`：

```
<Method> <PathWithRawQuery>
Host: <Host>
Content-Type: <ContentType>
Authorization: QUser <QUserAK>

[<Body>] #这里的 <Body> 只有在 <ContentType> 存在且不为 application/octet-stream 时才签进去。
```

有了 `<Data>`，就可以计算对应的 `<QUserSign>`，如下：

```
<QUserSK> = urlsafe_base64( hmac_sha1(<SK>, <QUserAK>) )
<QUserSign> = urlsafe_base64( hmac_sha1(<QUserSK>, <Data>) )
```

# QAdmin Authorization

```
<Method> <PathWithRawQuery>
<Header1>: <Value1>
<Header2>: <Value2>
...
Host: <Host>
...
Content-Type: <ContentType>
...
Authorization: QAdmin <SuInfo>:<QUserAK>:<QUserSign>
...

<Body>
```

对于上面这样一个请求，我们构造如下这个待签名的 `<Data>`：

```
<Method> <PathWithRawQuery>
Host: <Host>
Content-Type: <ContentType>
Authorization: QAdmin <SuInfo>:<QUserAK>

[<Body>] #这里的 <Body> 只有在 <ContentType> 存在且不为 application/octet-stream 时才签进去。
```

有了 `<Data>`，就可以计算对应的 `<QUserSign>`，如下：

```
<QUserSK> = urlsafe_base64( hmac_sha1(<SK>, <SuInfo>:<QUserAK>) )
<QUserSign> = urlsafe_base64( hmac_sha1(<QUserSK>, <Data>) )
```

其中，`SuInfo`表示`[uid]/[appid]`，当请求方无法获得appid时，设置appid为0，表示appName是default对应的那个appid

# 最佳实践

## AK/SK 更换

* 服务端应该定期更换 AK/SK
* 发现 AK/SK 泄露应该主动立刻更换 AK/SK

为了支持 AK/SK 的立刻更换，服务端在 AK 不存在的时候返回一个特殊 Code（4xx）。客户端收到该 Code 需要申请新的 QUserAK/QUserSK。

## QUserAK/QUserSK

* 服务端支持声明某个 QUserAK 进入黑名单（发现某个 QUserAK 异常）
* QUserAK/QUserSK 的 Expiry 时间，不要超过 1 个月
* 通过 QUserAK/QUserSK 不应该可以获取新的 QUserAK2/QUserSK2，否则将导致黑名单机制失效（有人会通过获取的 QUserAK/QUserSK 对生成一堆 QUserAK2/QUserSK2 备用）

# Gateway Port Watcher 使用说明

`external` 包提供了 Gateway 端口配置监听 SDK，用于监听 `/etc/casaos/gateway.ini` 中 HTTP/HTTPS 端口和启用状态变化，并通过 callback 通知调用方。

## 监听内容

SDK 会读取以下字段：

```ini
[gateway]
enabled = true
port    = 80

[ssl]
enabled = true
port    = 443
```

说明：

- `[gateway].port` 对应 HTTP 端口。
- `[gateway].enabled` 对应 HTTP 是否启用；如果配置文件中没有该字段，默认按 `true` 处理，兼容现有配置。
- `[ssl].port` 对应 HTTPS 端口。
- `[ssl].enabled` 对应 HTTPS 是否启用。

## 数据结构

```go
type GatewayPortConfig struct {
    HTTPEnabled  bool `json:"http_enabled"`
    HTTPPort     int  `json:"http_port"`
    HTTPSPort    int  `json:"https_port"`
    HTTPSEnabled bool `json:"https_enabled"`
}
```

## 基本用法

```go
package main

import (
    "context"
    "log"

    "github.com/IceWhaleTech/CasaOS-Common/external"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    err := external.ListenGatewayPortChanges(ctx, func(config external.GatewayPortConfig) {
        log.Printf(
            "gateway port changed: http_enabled=%t http_port=%d https_enabled=%t https_port=%d",
            config.HTTPEnabled,
            config.HTTPPort,
            config.HTTPSEnabled,
            config.HTTPSPort,
        )
    })
    if err != nil {
        log.Fatalf("listen gateway port changes: %v", err)
    }
}
```

`ListenGatewayPortChanges` 是阻塞调用，会一直运行到 `ctx` 被取消。

SDK 启动时会读取一次当前配置作为基线，但不会触发 callback。只有后续解析出的 `GatewayPortConfig` 和上一次有效配置不同时，才会通知调用方。

## 自定义配置文件路径

默认监听 `/etc/casaos/gateway.ini`。测试或特殊环境可以指定路径：

```go
err := external.ListenGatewayPortChanges(ctx, callback,
    external.WithGatewayPortConfigPath("/path/to/gateway.ini"),
)
```

## 自定义轮询间隔

SDK 会优先使用 `fsnotify` 监听文件变化。在 Linux 上，`fsnotify` 底层使用 inotify。若监听初始化失败或监听过程中出错，会自动切换为定时轮询文件的修改时间和大小。

默认轮询间隔为 `1s`，可以调整：

```go
err := external.ListenGatewayPortChanges(ctx, callback,
    external.WithGatewayPortPollInterval(3*time.Second),
)
```

## 只读取当前配置

如果只需要读取一次当前配置，不需要持续监听：

```go
config, err := external.ReadGatewayPortConfig("")
if err != nil {
    return err
}

// config.HTTPEnabled
// config.HTTPPort
// config.HTTPSEnabled
// config.HTTPSPort
```

传空字符串时会读取默认路径 `/etc/casaos/gateway.ini`。

## Callback 行为

callback 类型：

```go
type GatewayPortChangeCallback func(config GatewayPortConfig)
```

触发规则：

- 监听启动时，SDK 会读取当前配置作为基线，不触发 callback。
- 文件变化后，解析出的 `GatewayPortConfig` 与上一次有效配置不同时触发 callback。
- 如果文件变化但 HTTP/HTTPS 端口和启用状态没有变化，不触发 callback。
- 配置文件临时不可读、内容不完整或端口非法时，不触发 callback；监听会继续运行，等待下一次有效变化。

## 关闭监听

通过取消 context 停止监听：

```go
ctx, cancel := context.WithCancel(context.Background())

go func() {
    _ = external.ListenGatewayPortChanges(ctx, callback)
}()

// 需要退出时
cancel()
```

## 错误处理建议

`ListenGatewayPortChanges` 在以下情况下会直接返回错误：

- `ctx` 为 `nil`。
- callback 为 `nil`。

监听文件失败不会直接返回错误，SDK 会自动进入轮询兜底逻辑。

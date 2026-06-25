# Scream

Scream 是一个基于 **Actor 模型** 的分布式运行时框架（仍在积极开发中）。它将业务逻辑封装为 Actor，通过消息驱动执行，并支持单节点内通信与跨节点 gRPC 路由。

---

## 核心概念

### Node 与 System（sys）

框架中最常用的入口是 **Node**（接口 `core.INode`），具体由 `process` 结构体实现（`core/node/process_impl.go`）。

**命名说明**：

| 术语 | 位置 | 含义 |
|------|------|------|
| **Node** | `core/node.go` | 对外抽象：`INode`、`NodeParm`、`NodeWith*` |
| **process** | `core/node/process_impl.go` | `INode` 的内部实现 |
| **Node（地址簿）** | `core/addressbook` | 集群中的节点、负载权重（Redis） |

```
┌─────────────────────────────────────┐
│  process (implement INode)          │
│  ├── NodeParm                       │
│  │   ID / IP / Port / Weight        │
│  │   Loader / Factory               │
│  └── sys (ISystem)                  │
│      ├── Actor                      │
│      ├── AddressBook (Redis)        │
│      ├── gRPC Client / Acceptor     │
│      └── Pub/Sub                    │
└─────────────────────────────────────┘
```

| 层级 | 类型 | 职责 |
|------|------|------|
| **Node** | `core.INode` | 节点生命周期：配置、`Init()`、监听信号、`WaitClose()` |
| **System** | `core.ISystem` | Actor 运行时：`Register`、`Call`、`Send`、地址簿、消息路由 |

典型用法：

```go
nod := node.BuildProcessWithOption(
    core.NodeWithID("node-1"),
    core.NodeWithLoader(loader),
    core.NodeWithFactory(factory),
    core.NodeWithPort(8080),
)

nod.Init()                                      // 节点级初始化
nod.System().Loader("myActor").Register(ctx)    // 通过 sys 注册 Actor
nod.System().Call("id", "type", "event", msg)   // 同步调用
nod.System().Send("id", "type", "event", msg)   // 异步投递
nod.WaitClose()                                 // 优雅退出
```

**关系总结**：`INode` 是对外接口，`process` 是其具体实现，`sys` 是 process 内部的运行时引擎；业务代码通过 `nod.System()` 访问运行时能力。

---

### Actor

Actor 是框架的基本执行单元：

- 每个 Actor 有唯一 **ID** 和 **Type**
- 通过 **Event** 注册消息处理器（`OnEvent`）
- 拥有独立的消息队列，**同一 Actor 内消息串行处理**
- 不同 Actor 之间可并行

```go
func (a *myActor) Init(ctx context.Context) {
    a.OnEvent("ping", func(ctx core.ActorContext) core.IChain {
        return &actor.DefaultChain{
            Handler: func(w *msg.Wrapper) error {
                // 处理消息
                return nil
            },
        }
    })
}
```

---

### Call 与 Send

| 方法 | 语义 | 是否阻塞调用方 | 是否有响应 |
|------|------|----------------|------------|
| **Call** | 同步 RPC | 是（等待处理完成或超时，默认 5s） | 有，写入 `msg.Wrapper` |
| **Send** | 异步 fire-and-forget | 否（消息入队即返回） | 无 |

**Send 适用场景**：触发耗时操作、发送通知、不关心返回结果。

**Call 适用场景**：需要返回值、或必须等待处理完成。

路由符号（`def/system_def.go`）：

| 符号 | 含义 |
|------|------|
| `?` (`SymbolWildcard`) | 路由到该类型的任意 Actor（优先低权重节点） |
| `~` (`SymbolLocalFirst`) | 优先本节点，否则随机选其他节点 |
| `#` (`SymbolGroup`) | 路由到一组 Actor（仅 Send） |
| `*` (`SymbolAll`) | 广播到该类型所有 Actor（仅 Send） |

---

### AddressBook（地址簿）

分布式场景下，Actor 的注册信息存储在 **Redis** 中，由 `core/addressbook` 管理：

- 记录 Actor ID → 节点地址（IP、Port）的映射
- 维护每个节点的 **total_weight**（已注册 Actor 权重之和）
- 支持 **动态选址**：`GetLowWeightNodeForActor` 选择负载最低的节点

动态 Actor 创建流程：

```
Picker() → MockDynamicPicker → 选低权重节点 → MockDynamicRegister → Register
```

---

## 项目结构

```
scream/
├── core/                   # 框架核心
│   ├── node.go             # INode 接口、NodeParm、NodeWith* 配置
│   ├── node/               # Node 实现与 System
│   │   ├── process_impl.go # process 结构体，实现 INode
│   │   ├── system_impl.go  # NormalSystem，实现 ISystem
│   │   └── system_acceptor.go
│   ├── actor/              # Actor 运行时、消息链、定时器
│   └── addressbook/        # Redis 地址簿
├── router/                 # 消息路由、gRPC proto
├── lib/                    # 基础设施
│   ├── grpc/               # gRPC 客户端/服务端
│   ├── pubsub/             # 发布订阅
│   ├── span/               # 链路追踪
│   └── tracer/             # Jaeger 集成
├── 3rd/                    # 第三方封装
│   ├── log/                # Zap 日志
│   ├── redis/              # Redis 客户端
│   └── etcd/               # Etcd 客户端
├── def/                    # 全局常量与定义
├── tests/                  # 集成测试
│   └── mock/               # Mock Actor 与 Loader
├── config/                 # 配置
└── utils/                  # 工具函数
```

---

## 快速开始

### 环境要求

- Go 1.24+
- Redis（测试使用 [miniredis](https://github.com/alicebob/miniredis) 内存模拟，无需外部依赖）

### 运行测试

```bash
go test ./tests/... -v
```

测试入口 `tests/main_test.go` 的 `TestMain` 会完成：

1. 初始化日志
2. 构建 Mock Actor Factory / Loader
3. 启动 miniredis 并配置 Redis 客户端

### 编写一个简单测试

```go
func TestExample(t *testing.T) {
    nod := node.BuildProcessWithOption(
        core.NodeWithID("test-1"),
        core.NodeWithLoader(loader),   // 来自 TestMain 的全局变量
        core.NodeWithFactory(factory),
    )

    _, err := nod.System().Loader("mocka").WithID("mocka").Register(context.TODO())
    assert.NoError(t, err)

    nod.Init()
    defer func() {
        wg := sync.WaitGroup{}
        nod.System().Exit(&wg)
        wg.Wait()
    }()

    m := msg.NewBuilder(context.TODO()).Build()
    err = nod.System().Call("mocka", "mocka", "ping", m)
    assert.NoError(t, err)
}
```

---

## 日志配置

日志基于 [Uber Zap](https://github.com/uber-go/zap)，封装在 `3rd/log`。

### 打印调用位置（文件名 + 行号）

```go
logger, err := log.NewDefaultLogger(
    log.WithCallerSkip(2), // 通过 InfoF/WarnF 等封装调用时需设为 2
)
```

| 参数 | 说明 |
|------|------|
| `WithCaller(true)` | 开启调用位置（`NewDefaultLogger` 对 Text 日志默认已开启） |
| `WithCallerFunc(true)` | 同时打印函数名，如 `main.go:42 Foo()` |
| `WithCallerSkip(n)` | 跳过 n 层调用栈，修正封装函数导致的定位偏移 |

---

## 设计目标

1. **Actor 模型**：业务逻辑封装为 Actor，消息驱动、串行处理
2. **分布式路由**：通过 Redis 地址簿 + gRPC 实现跨节点 Call/Send
3. **动态负载均衡**：按节点权重动态分配 Actor 实例
4. **可观测性**：集成 Jaeger 链路追踪
5. **同玩家串行、多玩家并行**：单个 Actor 消息队列保证串行；不同 Actor 并行

---

## 测试说明

| 测试文件 | 验证内容 |
|----------|----------|
| `call_test.go` | 同步 Call、链式调用、TCC |
| `send_test.go` | Send 异步语义（不阻塞调用方） |
| `addressbook_test.go` | 多节点动态 Picker 与权重分布 |
| `pubsub_test.go` | 发布订阅 |
| `timer_test.go` | Actor 定时器 |
| `reenter_test.go` | 重入异步调用 |

---

## 注意事项

- `BuildProcessWithOption` 内部使用全局变量 `pcs`，同一进程内仅维护一个 Node 实例；多节点测试需各自独立进程或使用不同 Port。
- `Send` 只保证消息投递成功，不等待 handler 执行完成；测试清理（`Exit`）与业务处理是独立的两个阶段。
- 框架仍在开发中，API 与行为可能变更。

---

## License

待定

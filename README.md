# Go 语言构建的简单分布式系统
面向学习目的的简单分布式系统示例

## 常见的典型分布式模型
1. Hub & Spoke (中心辐射模型) 
2. Peer to Peer (点对点模型)
3. Message Queue (消息队列模型)

但实际应用中，分布式系统往往是上述模型的混合体。

## 项目结构
本项目服务注册、健康检查、配置管理等做集中式管理，web 服务做点对点模式。
```.
├── 服务注册
│   ├── 服务注册
│   ├── 健康检查
│   └── 配置管理
├── 用户门户
│   ├── Api 网关
│   └── Web 服务
├── 日志服务
│   └── 集中式日志收集
└── 业务服务
    ├── 业务逻辑
    └── 数据持久化
```

使用 Go 语言构建，HTTP进行数据通信，JSON作为数据交换格式。

## 功能模块
1. 创建服务注册和取消服务
2. 日志服务

> 注: 文档的顺序和代码实现的顺序可能不一致.

### 服务注册中心
这里注册中心本质就是维护在内存中的一个slice, 提供锁和添加、删除方法, 调用http服务添加服务到slice中. slice中存储的是自定义的Service结构体, 包含服务的名称和URL.
1. 定义Service结构体, 包含服务名称和URL.
2. 定义存储服务的`registry`结构体, 包含服务列表和读写锁, 提供添加服务的方法.
3. 定义`RegistryService`结构体, 实现`http.Handle`接口, 处理服务注册的HTTP请求.
4. 在`cmd/registerservice/main.go`中配置并启动注册服务的HTTP服务器.

> 1. 服务注册中心的启动和其他服务的启动使用不同的方式进行配置。 注册中心是实现`http.Handle`接口来处理HTTP请求, 其他服务是使用`http.HandleFunc`处理HTTP请求. 本质一样, HandleFunc基于Handle的.
> 2. golang中的变量声明的两种方式: `:=` 和 `var`。 `:=`是短变量声明方式, 只能在函数内部使用; `var`用于显式声明变量, 可以在任何地方使用. ~~在函数外采用短变量声明服务存储的结构体reg导致报错~~

### 启动服务
1. 启动服务的公共功能独立到services包中. 提供`Start`函数启动HTTP服务.
2. 每个服务都需要单独启动, 然后注册到服务注册中心. 创建`cmd`目录存放各个服务的启动代码.

#### 使用默认的 HTTP 实例
`"net/http"`包会导出三个默认实例:
+ `http.DefaultServeMux` : 默认的多路复用器, 用于注册路由和处理请求.
+ `http.DefaultClient` : 默认的HTTP客户端实例, 用于发送HTTP请求.
+ `http.DefaultTransport` : 默认的HTTP传输实现, 被`http.DefaultClient`使用.

所以引入包后可以直接使用方法, 例如:
```go 
import "net/http"
http.HandleFunc("/", handler) // 使用默认的多路复用器注册路由
http.ListenAndServe(":8080", nil) // 使用默认的多路复用器启动HTTP服务器

http.Get("http://example.com") // 使用默认的HTTP客户端发送GET请求
```

所以可以将`http.HandleFunc`和`http.ListenAndServe`分开定义和使用.

### 服务注册
项目中的服务使用统一的方式进行启动, 只需要在启动后将其添加到注册中心即可.  
在`registry.client`中封装了注册服务的方法`RegisterService`, 通过HTTP POST请求将服务信息发送到注册中心. 这样在`service.Start`函数中, 启动服务后调用`registryclient.RegisterService`即可玩成注册服务的功能.

> HTTP POST请求的body参数要求是io.Reader接口类型, 该类型要求实现的Read方法能从字节流中读取数据. 使用`bytes.NewBuffer`创建一个可读可写的buffer, 然后使用`json.NewEncoder`将结构体编码为JSON格式并写入buffer中, 最后将buffer作为body参数传递给HTTP请求.


#### 使用 context 进行协程间通信和管理
每个服务都会独立启动, 使用`context`包管理服务的生命周期.  
项目中使用了`context.WithCancel`创建可取消的上下文并返回, 在对应逻辑中调用`cancel`函数取消上下文.  
启动程序通过监听返回的`cancelContext`: `<-cancelContext.Done()` 来等待取消信号, 然后优雅关闭服务.


### 日志服务

#### 创建日志服务
1. 封装log服务, 自定义log所使用的io.Writer接口的实现: `fileLog`
2. 封装初始化log的方法, 指定日志写入的路径
3. 封装对外提供的HTTP Handler, 让每个服务管理自己的Web服务注册. 只处理POST请求的服务.

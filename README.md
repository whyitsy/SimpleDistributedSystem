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
│   └── 服务发现 
├── 日志服务
│   └── 集中式日志
└── 业务服务
```

使用 Go 语言构建，HTTP进行数据通信，JSON作为数据交换格式。

## 组成模块

> 注: 文档的顺序和代码实现的顺序可能不一致.

### 注册中心
注册中心的功能职责：将web服务添加到中心节点, 其他服务就可以从该节点获取依赖的服务信息和变动信息.
1. 服务注册：一个线程访问安全的slice, 用于存储注册的服务信息(服务名称和url). 提供注册和取消注册的HTTP接口.
2. 服务发现：注册服务时, 根据服务的依赖信息, 遍历slice找到存在的依赖服务, 通过回调该注册服务的更新接口将依赖服务的信息发送给注册服务.
3. 依赖变更通知：当依赖服务上线或下线时, 需要通知通知对应的服务. 
4. 健康检查：定期调用已注册服务的健康检查接口, 如果服务不可用则将其从注册列表中删除, 并通知依赖该服务的其他服务. 
所有服务都需要提供两个接口:`update`和`healthcheck`, 分别用于接收依赖服务的更新信息和健康检查.  

服务存储：收到依赖的服务信息后, 通过`provider模式`维护依赖服务的信息. 需要时从provider中获取，可以在获取时做负载均衡.

#### 注册中心实现
本质就是维护在内存中的一个slice, 提供线程安全的添加、删除方法以及对应的接口. 存储的是自定义的Service结构体, 包含服务的名称和URL.
1. 定义存储slice的结构体, `slice`和`RWMutex`. 
2. 定义添加和删除服务请求的参数类型, 包含`ServiceName`、`ServiceURL`、`[]RequiredServices`、`ServiceUpdateURL`和`HeartbeatUrl`.
3. 定义注册和取消注册的方法并作为对应的api接口处理函数.
   1. 定义更新依赖服务请求的参数类型, 包含`Added`和`Removed`两个slice.
   2. 将服务信息添加到注册列表中, 然后遍历依赖的服务列表, 查找已注册的服务并将其信息通过POST请求发送到该次注册服务的`update`接口.
   3. 有服务变动时, 就需要遍历已注册的服务, 查找依赖该服务的服务, 并将变动的信息通过POST请求发送到对应服务的`update`接口.
4. 健康检查: 每隔一段时间遍历已注册的服务, 通过HTTP GET请求调用服务的`healthcheck`接口:
   1. 使用goroutine并发调用每个服务的健康检查接口, 尝试3次.
   2. 出现失败则将该服务从注册列表中删除, 并通知依赖该服务的其他服务.
   3. 如果重试次数之内又恢复正常, 则重新添加到注册列表中, 并通知依赖该服务的其他服务.

> 1. 服务注册中心的启动和其他服务的启动使用不同的方式进行配置。 注册中心是实现`http.Handle`接口来处理HTTP请求, 其他服务是使用`http.HandleFunc`处理HTTP请求. 本质一样, HandleFunc基于Handle的.
> 2. golang中的变量声明的两种方式: `:=` 和 `var`。 `:=`是短变量声明方式, 只能在函数内部使用; `var`用于显式声明变量, 可以在任何地方使用. ~~在函数外采用短变量声明服务存储的结构体reg导致报错~~

用到了`sync.WaitGroup`来等待所有的健康检查goroutine完成.`
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

### 服务注册与取消
项目中的服务使用统一的方式进行启动, 只需要在启动后将其添加到注册中心即可.  
注册：在`registry.client`中封装了注册服务的方法`RegisterService`, 通过HTTP POST请求将服务信息发送到注册中心. 这样在`service.Start`函数中, 启动服务后调用`registryclient.RegisterService`即可玩成注册服务的功能.  
取消：取消和注册类似, 在结束之前使用HTTP DELETE请求将服务信息发送到注册中心, 注册中心收到请求后从服务列表中删除对应的服务.
> HTTP POST请求的body参数要求是io.Reader接口类型, 该类型要求实现的Read方法能从字节流中读取数据. 使用`bytes.NewBuffer`创建一个可读可写的buffer, 然后使用`json.NewEncoder`将结构体编码为JSON格式并写入buffer中, 最后将buffer作为body参数传递给HTTP请求.


#### 使用 context 进行协程间通信和管理
每个服务都会独立启动, 使用`context`包管理服务的生命周期.  
项目中使用了`context.WithCancel`创建可取消的上下文并返回, 在对应逻辑中调用`cancel`函数取消上下文.  
启动程序通过监听返回的`cancelContext`: `<-cancelContext.Done()` 来等待取消信号, 然后优雅关闭服务.

### 服务发现
1. 扩展基础的服务注册时使用的结构体, 提供依赖的服务列表和服务更新URL:
    ```go
    type RegistrationEntry struct {
        ServiceName      ServiceName
        ServiceURL       string
        RequiredServices []ServiceName // 依赖的服务, 在注册时请求这些服务的URL
        ServiceUpdateURL string        // 服务自身配置的更新URL, 供注册中心调用
    }
    ```
2. 在注册自身后, 注册中心会将依赖信息POST到`ServiceURL+ServiceUpdateURL`这个地址, 携带更新的服务列表patch.
    ```go
    type patchEntry struct {
        Name ServiceName
        URL  string
    }
    
    type patch struct {
        Added   []patchEntry
        Removed []patchEntry
    }
    ```
3. 服务收到patch后, 需要通过`providers`维护自己的依赖服务列表, 提供更新和获取`provider`的方法. 这个provider就是提供服务的URL.



### 日志服务

#### 创建日志服务
1. 封装log服务, 自定义log所使用的io.Writer接口的实现: `fileLog`
2. 封装初始化log的方法, 指定日志写入的路径
3. 封装对外提供的HTTP Handler, 让每个服务管理自己的Web服务注册. 只处理POST请求的服务.

### 业务服务
实现一个业务服务, 主要用于处理多个服务之间的调用和依赖关系.  
主要是实现了三个path:
1. /students  GET 获取所有学生
2. /students/{id}  GET 获取单个学生的信息
3. /students/{id}/grades  POST 添加学生成绩

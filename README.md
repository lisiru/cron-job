**此项目仅供学习go 和redis使用**

参考有赞延迟队列使用go 实现定时任务

#### 需要用到延时队列的场景

- 当订单一直处于未支付状态时，如何及时的关闭订单，并退还库存？
- 如何定期检查处于退款状态的订单是否已经退款成功？
- 新创建店铺，N天内没有上传商品，系统如何知道该信息，并发送激活短信？等等

我们原始的做法，就是每个业务定时去扫表，把符合执行要求的数据进行处理更新。现在借鉴有赞的延迟队列设计，使用go设计一个延迟队列实现定时任务的效果

大概的设计

![img](https://tech.youzan.com/content/images/2016/03/delay-queue.png)

#### 延迟队列的结构和数据组成

- 一个key-value 的结构来存放所有任务的具体信息，这里使用redis 的hash数据结构进行保存Ti
- 一个bucket 按执行时间排序存放任务id，并且形成一个有序集合，这里用到的数据结构为redis 的sort set ,score为任务的执行绝对时间
- Timmer 使用go 的定时器实现,分别使用goroutine扫描Delay bucket，目前采用的是集中存储机制，在多实例部署时Timer程序可能会并发执行，导致job被重复放入ready queue。为了解决这个问题，我们使用了redis的setnx命令实现了简单的分布式锁，以保证每个bucket每次只有一个timer thread来扫描。 和消费Ready Queue
- Ready Queue 为可以执行的任务，里面存放的数据为任务id,任务id各个业务方生成并保证唯一。使用的数据结构为redis 的list。

#### 消息的结构

```go
type Job struct {
	JobId        string 
	DelaySeconds int64 
	TtrSeconds   int64 
	Body         string 
	IsLoop       bool   
	NotifyUrl    string 
	Stat         int  
}
```

- JobId ：任务id，唯一，由业务方生成
- DelaySeconds 延迟执行的时间，服务方会转为决定执行时间进行保存
- TtrSeconds 超时时间，也可以理解为循环任务的距离下次的执行时间
- Body 通知业务方执行逻辑的json数据
- IsLoop 是否为循环任务
- NotifyUrl 回调通知业务方的数据
- Stat 任务的状态

#### 任务的状态组成和转换

```go
const (
	JOB_STAT_DELAY int = iota + 1
	JOB_STAT_READY
	JOB_STAT_RESERVED
	JOB_STAT_DELETED
)
```

- DELAY延迟的状态，等待时钟周期
- READY 可执行的状态
- RESERVED 被消费，但还没得到业务方的响应
- DELETED 已被消费完成或者删除

![Job State Flow](https://tech.youzan.com/content/images/2016/03/job-state.png)

#### 通信协议和目前提供的接口

目前使用go 搭建了rpc服务，对外提供rpc的接口，可以直接通过rpc来操作添加任务。目前提供的接口如下

```go
Add(context.Context, *AddRequest) (*AddReply, error)
Del(context.Context, *DelRequest) (*DelReply, error)
Update(context.Context, *UpdateRequest) (*UpdateReply, error)
Finish(context.Context, *FinishRequest) (*FinishReply, error)
```

具体的请求参数可以拉取proto文件生成。

![img](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/preview/boxcnMoPBj3ajcigiPId2SUy7Ke?mount_point=explorer&preview_type=11&version=7071184447114084380)



#### 待实现的地方

- web页面管理任务，暂停，删除。

- 改用rabbitmq通知业务方

- 任务的更新操作

- 消息的持久化。

- 日志的记录分析。

  






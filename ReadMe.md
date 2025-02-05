# 基于golang的任务调度系统框架

## 介绍
这是一个基于golang实现的任务调度系统，旨在实现持久化的任务调度管理、任务执行过程和结果记录。


## 示例代码

``` golang
	tasker := tasker.Default()
    // 注册任务处理方法
	tasker.register("handler1", func handler(c taskContext) error)
    tasker.register("handler2", func handler(c taskContext) error)
    // 添加任务
    // 任务id为task1
    // 任务处理方法为handler1
    // 任务参数为param1为json字符串
    tasker.addTask("task1", "handler1", "param1")
    tasker.addTask("task2", "handler2", "param2")

    // 触发执行任务task1
    tasker.trigger("task1")

    // 添加任务调度
    // 任务id为task1
    // 调度计划为cron1
    tasker.addTaskSchedule("task1", "cron1")

    // 启动计划任务调度器
	tasker.run()
```
## 系统设计
系统分为两部分：任务调度和任务执行。

### 任务调度功能
任务分为两种：
1. 单次执行任务：用户直接触发执行
2. 计划任务：用户定义调度计划，系统根据计划调度任务执行
计划执行任务是在单次执行任务的基础上，添加了调度计划，系统根据计划调度任务执行。

``` golang
Task{
    // 任务id
	TaskId      string
    // 任务处理方法
	TaskHandler string
    // 任务描述
	Desc        string
    // 任务参数
	Param       json.RawMessage
}

TaskScheme{
    // 调度计划id
	SchemeId string
    // 任务id
	TaskId   string
    // 调度计划
	Cron     string
}
```

### 任务执行
任务执行时，需要提供任务执行上下文，用于记录任务执行过程中的事件、结果等信息。
context 由tasker提供，用户只需要在任务执行方法中使用context提供的方法记录事件、结果等信息。
``` golang
// 任务执行上下文
TaskContext{
	event()
	failed()
	success()
	bind()
}

... 

func handler(c taskContext) error{
    // 记录任务执行事件
	c.event("info", "test task start")
    // 执行失败
	c.failed(err)
    // 执行成功
	c.success(result)
    // 获取任务参数,直接解析json为对象
	obj,err := c.bind()
}
```

任务事件在任务执行过程中产生，记录任务执行过程（时间、参数、事件、结果）。用户可以使用context提供的event()方法记录事件。
```go
// 任务执行事件
TaskEvent{
	EventId
	ExecId
	Time
    // Debug/Normal/Warning/Error
	Level
	Message
}
```

任务执行记录和事件记录是分开存储的，任务执行记录用于记录任务执行过程，事件记录用于记录任务执行过程中的事件。一次任务执行会产生一个任务执行记录，一个任务执行记录会产生多个事件记录。
``` golang
// 任务执行开始时产生，结束时更新结果和时间
type TaskExec struct {
	ExecId    string
	TaskId    string
	StartTime time.Time
	EndTime   time.Time
	Success   bool
	Result    json.RawMessage
}
```

事件清理/任务执行清理

一个任务只保留100（默认）条执行记录，产生新的exec时，删除100条之前的exec和event记录

一个任务只保留1小时（默认）之前的执行记录，产生新的exec时，删除1小时之前的exec和event记录

↑ 在Storage Deiver中实现，所以如果是用户自定义的Storage Deiver，需要自己实现清理功能。

## 持久化
任务调度系统需要持久化任务、任务调度计划、任务执行记录、任务事件记录。所以需要一个持久化驱动，用于实现任务、任务调度计划、任务执行记录、任务事件记录的存储和读取。

目前支持内存存储(仅用于测试，重启后数据丢失，不建议用于生产环境)和mysql存储，用户可以根据需要选择合适的存储方式。

如有需要，可以实现自己的存储驱动：驱动需要实现Storage接口。
``` golang
type Storage interface {
	AddTask(task *Task) error
	RemoveTask(taskId string) error
	ListTasks() ([]*Task, error)

	AddTaskScheme(scheme *TaskScheme) error
	RemoveTaskScheme(schemeId string) error
	ListTaskSchemes() ([]*TaskScheme, error)

	AddTaskExec(exec *TaskExec) error
	UpdateTaskExec(exec *TaskExec) error

	AddTaskEvent(event *TaskEvent) error
	ListTaskEvents(execId string) ([]*TaskEvent, error)
}
```


## Q&A
Q： 我们是如何实现任务定制化的？

A： 让用户自己负责定义任务参数和任务执行方法；所以任务运行前需要先注册任务类型和对应的处理方法。
``` golang
tasker.register("handler1", func handler(c taskContext) error)
```

Q： 任务执行时，如何获取任务参数？

A： 任务执行时，使用context提供的bind()方法获取任务参数，bind()方法会直接解析json为对象。
``` golang
obj,err := c.bind()
```

Q： 任务执行时，如何记录任务执行结果？

A： 任务执行时，使用context提供的success()和failed()方法记录任务执行结果。
``` golang
c.success(result)
c.failed(err)
```

Q： 任务执行时，如何记录任务执行事件？

A： 任务执行时，使用context提供的event()方法记录任务执行事件。
``` golang
c.event("info", "test task start")
```

# 项目支持
- cursor
- gorm

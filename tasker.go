package tasker

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	cronv3 "github.com/robfig/cron/v3"
)

type Tasker struct {
	Storage Storage

	// [taskId]TaskHandler
	handlers map[string]TaskHandler
	// [taskId]Task
	tasks map[string]*Task
	// [schemeId]TaskScheme
	schemes map[string]*TaskScheme

	cron *cronv3.Cron
}

func Default() *Tasker {
	return &Tasker{
		Storage:  NewMemoryStorage(),
		handlers: make(map[string]TaskHandler),
		tasks:    make(map[string]*Task),
		schemes:  make(map[string]*TaskScheme),
		cron:     cronv3.New(cronv3.WithSeconds()),
	}
}

func (t *Tasker) SetStorage(storage Storage) {
	t.Storage = storage
}

func (t *Tasker) Register(taskType string, handler TaskHandler) {
	t.handlers[taskType] = handler
}

func (t *Tasker) AddTask(taskId, taskType, desc, param string) {
	// 判断taskType是否已注册
	if _, ok := t.handlers[taskType]; !ok {
		log.Println("taskType not found", taskType)
		return
	}
	// 判断taskId是否已存在
	if _, ok := t.tasks[taskId]; ok {
		log.Println("taskId already exists", taskId)
		return
	}

	task := &Task{
		TaskId:      taskId,
		TaskHandler: taskType,
		Desc:        desc,
		Param:       json.RawMessage(param),
	}
	// 添加task
	err := t.Storage.AddTask(task)
	if err != nil {
		log.Println("add task failed", err)
		return
	}
	t.tasks[taskId] = task
}

// RemoveTask 删除任务
func (t *Tasker) RemoveTask(taskId string) error {
	err := t.Storage.RemoveTask(taskId)
	if err != nil {
		log.Println("remove task failed", err)
		return err
	}

	delete(t.tasks, taskId)
	return nil
}

// AddTaskScheme 添加任务执行计划
func (t *Tasker) AddTaskScheme(schemeId, taskId, cronStr string) error {
	// 判断taskId是否已存在
	_, ok := t.tasks[taskId]
	if !ok {
		log.Println("taskId not found", taskId)
		return fmt.Errorf("taskId not found")
	}
	// 判断schemeId是否已存在
	if _, ok := t.schemes[schemeId]; ok {
		log.Println("schemeId already exists", schemeId)
		return fmt.Errorf("schemeId already exists")
	}

	scheme := &TaskScheme{
		SchemeId: schemeId,
		TaskId:   taskId,
		Cron:     cronStr,
	}
	// 添加scheme
	err := t.Storage.AddTaskScheme(scheme)
	if err != nil {
		log.Println("add task scheme failed", err)
		return fmt.Errorf("add task scheme failed")
	}

	// 添加cron执行函数
	cronId, err := t.cron.AddFunc(cronStr, func() {
		_, err := t.RunTask(taskId)
		if err != nil {
			log.Println("run task failed", err)
		}
	})
	if err != nil {
		log.Println("add cron failed", err)
		return fmt.Errorf("add cron failed")
	}
	scheme.cronId = int(cronId)
	// 添加scheme
	t.schemes[schemeId] = scheme
	return nil
}

func (t *Tasker) RemoveTaskScheme(schemeId string) error {
	// 删除cron
	t.cron.Remove(cron.EntryID(t.schemes[schemeId].cronId))

	// 删除scheme
	err := t.Storage.RemoveTaskScheme(schemeId)
	if err != nil {
		log.Println("remove task scheme failed", err)
		return fmt.Errorf("remove task scheme failed")
	}
	delete(t.schemes, schemeId)
	return nil
}

// Run 启动Tasker
func (t *Tasker) Run() error {
	// 获取任务，记载至tasks
	var err error
	taskList, err := t.Storage.ListTasks()
	if err != nil {
		return err
	}
	for _, task := range taskList {
		t.tasks[task.TaskId] = task
	}
	schemes, err := t.Storage.ListTaskSchemes()
	if err != nil {
		return err
	}
	for _, scheme := range schemes {
		t.schemes[scheme.SchemeId] = scheme
	}

	// cron执行
	for _, scheme := range schemes {
		_, ok := t.tasks[scheme.TaskId]
		if !ok {
			log.Println("task not found", scheme.TaskId)
			continue
		}
		// 添加cron执行函数
		cronId, err := t.cron.AddFunc(scheme.Cron, func() {
			_, err := t.RunTask(scheme.TaskId)
			if err != nil {
				log.Println("run task failed", err)
			}
		})
		if err != nil {
			log.Println("add cron failed", err)
			continue
		}
		// 记录cronId，用于后续删除
		scheme.cronId = int(cronId)
		t.schemes[scheme.SchemeId] = scheme
	}

	return nil
}

// RunTask 执行任务
func (t *Tasker) RunTask(taskId string) (string, error) {
	// 获取任务
	task, ok := t.tasks[taskId]
	if !ok {
		return "", fmt.Errorf("task not found")
	}
	// 添加任务执行记录
	exec := &TaskExec{
		ExecId:    uuid.New().String(),
		TaskId:    taskId,
		StartTime: time.Now(),
	}
	err := t.Storage.AddTaskExec(exec)
	if err != nil {
		return "", err
	}

	// NewTaskContext并执行任务
	handler := t.handlers[task.TaskHandler]
	context := t.NewTaskContext(exec)
	go func(handler TaskHandler, context *TaskContext) {
		err := handler(context)
		if context.taskExec.EndTime.IsZero() {
			context.complete(err == nil, nil)
		}
	}(handler, context)

	return exec.ExecId, nil
}

func (t *Tasker) Events(execId string) ([]*TaskEvent, error) {
	return t.Storage.ListTaskEvents(execId)
}

// TaskContext 任务执行上下文
type TaskContext struct {
	task     *Task
	taskExec *TaskExec
	Storage  Storage
}

// NewTaskContext 创建任务执行上下文
func (t *Tasker) NewTaskContext(exec *TaskExec) *TaskContext {
	task, ok := t.tasks[exec.TaskId]
	if !ok {
		log.Println("task not found", exec.TaskId)
		return nil
	}

	return &TaskContext{task: task, taskExec: exec, Storage: t.Storage}
}

// Event 记录任务执行事件
func (t *TaskContext) Event(eventType string, message string) {
	t.Storage.AddTaskEvent(&TaskEvent{
		ExecId:  t.taskExec.ExecId,
		Time:    time.Now(),
		Level:   eventType,
		Message: message,
	})
}

// BindJSON 绑定任务参数
func (c *TaskContext) BindJSON(obj interface{}) error {
	return json.Unmarshal([]byte(c.task.Param), obj)
}

// Complete 完成任务执行
func (c *TaskContext) complete(success bool, result interface{}) error {
	c.taskExec.Success = success
	c.taskExec.EndTime = time.Now()
	if result != nil {
		data, err := json.Marshal(result)
		if err != nil {
			return err
		}
		c.taskExec.Result = data
	}
	return c.Storage.UpdateTaskExec(c.taskExec)
}

// Success 完成任务执行
func (c *TaskContext) Success(result interface{}) error {
	return c.complete(true, result)
}

// Fail 完成任务执行
func (c *TaskContext) Fail(result interface{}) error {
	return c.complete(false, result)
}

// TaskHandler 任务处理方法
type TaskHandler func(c *TaskContext) error

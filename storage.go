/*
 * @Author: pangqc5
 * @Date: 2025-01-26 10:44:49
 * @LastEditors: pangqc5
 * @LastEditTime: 2025-02-05 16:59:24
 * @FilePath: \tasker\storage.go
 * @Description:
 *
 * Copyright (c) 2021-2031. All rights reserved.
 */
package tasker

import (
	"sort"
	"sync"
)

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

type MemoryStorage struct {
	mu        sync.Mutex
	taskMap   map[string]*Task
	schemeMap map[string]*TaskScheme
	execMap   map[string]*TaskExec
	eventMap  map[string][]*TaskEvent
}

func NewMemoryStorage() Storage {
	return &MemoryStorage{
		taskMap:   make(map[string]*Task),
		schemeMap: make(map[string]*TaskScheme),
		execMap:   make(map[string]*TaskExec),
		eventMap:  make(map[string][]*TaskEvent),
	}
}

func (m *MemoryStorage) AddTask(task *Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.taskMap[task.TaskId] = task
	return nil
}

func (m *MemoryStorage) RemoveTask(taskId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 删除scheme
	for _, scheme := range m.schemeMap {
		if scheme.TaskId == taskId {
			delete(m.schemeMap, scheme.SchemeId)
		}
	}
	// 删除TaskEvent  TaskExec
	for _, exec := range m.execMap {
		if exec.TaskId == taskId {
			delete(m.execMap, exec.ExecId)
			delete(m.eventMap, exec.ExecId)
		}
	}
	delete(m.taskMap, taskId)
	return nil
}

func (m *MemoryStorage) ListTasks() ([]*Task, error) {
	taskList := make([]*Task, 0, len(m.taskMap))
	for _, task := range m.taskMap {
		taskList = append(taskList, task)
	}
	return taskList, nil
}

func (m *MemoryStorage) AddTaskScheme(scheme *TaskScheme) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.schemeMap[scheme.SchemeId] = scheme
	return nil
}

func (m *MemoryStorage) RemoveTaskScheme(schemeId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.schemeMap, schemeId)
	return nil
}

func (m *MemoryStorage) ListTaskSchemes() ([]*TaskScheme, error) {
	schemeList := make([]*TaskScheme, 0, len(m.schemeMap))
	for _, scheme := range m.schemeMap {
		schemeList = append(schemeList, scheme)
	}
	return schemeList, nil
}

func (m *MemoryStorage) GetTaskScheme(schemeId string) (*TaskScheme, error) {
	return m.schemeMap[schemeId], nil
}

func (m *MemoryStorage) AddTaskExec(exec *TaskExec) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.execMap[exec.ExecId] = exec

	// 只保留最新的100条任务执行记录
	go func() {
		if len(m.execMap) > 100 {
			// 先按时间升序排序
			execs := make([]*TaskExec, 0, len(m.execMap))
			for _, exec := range m.execMap {
				execs = append(execs, exec)
			}
			sort.Slice(execs, func(i, j int) bool {
				return execs[i].StartTime.Before(execs[j].StartTime)
			})
			// 删除前100条
			for _, exec := range execs[:100] {
				delete(m.execMap, exec.ExecId)
			}
		}
	}()
	return nil
}

func (m *MemoryStorage) UpdateTaskExec(exec *TaskExec) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.execMap[exec.ExecId] = exec
	return nil
}

func (m *MemoryStorage) AddTaskEvent(event *TaskEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	execId := event.ExecId
	events := m.eventMap[execId]
	events = append(events, event)
	// 只保留最新的100条任务事件
	if len(events) > 100 {
		events = events[len(events)-100:]
	}

	m.eventMap[execId] = events
	return nil
}

func (m *MemoryStorage) ListTaskEvents(eventId string) ([]*TaskEvent, error) {
	return m.eventMap[eventId], nil
}

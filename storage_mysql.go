/*
 * @Author: pangqc5
 * @Date: 2025-02-05 16:02:12
 * @LastEditors: pangqc5
 * @LastEditTime: 2025-02-05 16:52:43
 * @FilePath: \tasker\storage_mysql.go
 * @Description:
 *
 * Copyright (c) 2021-2031. All rights reserved.
 */
package tasker

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MysqlStorage struct {
	db *gorm.DB
}

func NewMysqlStorage(dsn string) (*MysqlStorage, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移表结构
	db.AutoMigrate(&Task{}, &TaskScheme{}, &TaskExec{}, &TaskEvent{})

	return &MysqlStorage{db: db}, nil
}

// 实现Storage接口方法
func (m *MysqlStorage) AddTask(task *Task) error {
	return m.db.Create(task).Error
}

func (m *MysqlStorage) RemoveTask(taskId string) error {
	tx := m.db.Begin()
	// 级联删除关联数据
	tx.Delete(&TaskScheme{}, "task_id = ?", taskId)
	tx.Delete(&TaskEvent{}, "exec_id IN (SELECT exec_id FROM task_execs WHERE task_id = ?)", taskId)
	tx.Delete(&TaskExec{}, "task_id = ?", taskId)
	tx.Delete(&Task{}, "task_id = ?", taskId)
	return tx.Commit().Error
}

func (m *MysqlStorage) ListTasks() ([]*Task, error) {
	var tasks []*Task
	if err := m.db.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (m *MysqlStorage) AddTaskScheme(scheme *TaskScheme) error {
	return m.db.Create(scheme).Error
}

func (m *MysqlStorage) RemoveTaskScheme(schemeId string) error {
	return m.db.Delete(&TaskScheme{}, "scheme_id = ?", schemeId).Error
}

func (m *MysqlStorage) ListTaskSchemes() ([]*TaskScheme, error) {
	var schemes []*TaskScheme
	if err := m.db.Find(&schemes).Error; err != nil {
		return nil, err
	}
	return schemes, nil
}

func (m *MysqlStorage) AddTaskExec(exec *TaskExec) error {
	err := m.db.Create(exec).Error
	if err != nil {
		return err
	}
	// 只保留最新的100条任务执行记录
	go func() {
		err := m.db.Delete(&TaskExec{}, "exec_id = ?", exec.ExecId).Order("start_time ASC").Limit(100).Error
		if err != nil {
			log.Println("clean task exec error:", err)
		}
	}()
	return nil
}

func (m *MysqlStorage) UpdateTaskExec(exec *TaskExec) error {
	return m.db.Save(exec).Error
}

func (m *MysqlStorage) ListTaskExecs() ([]*TaskExec, error) {
	var execs []*TaskExec
	if err := m.db.Find(&execs).Error; err != nil {
		return nil, err
	}
	return execs, nil
}

func (m *MysqlStorage) AddTaskEvent(event *TaskEvent) error {
	err := m.db.Create(event).Error
	if err != nil {
		return err
	}
	// 只保留最新的100条任务事件
	go func() {
		// 按时间升序并删除前100条
		err := m.db.Delete(&TaskEvent{}, "exec_id = ?", event.ExecId).Order("time ASC").Limit(100).Error
		if err != nil {
			log.Println("clean task event error:", err)
		}
	}()

	//

	return nil
}

func (m *MysqlStorage) ListTaskEvents() ([]*TaskEvent, error) {
	var events []*TaskEvent
	if err := m.db.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

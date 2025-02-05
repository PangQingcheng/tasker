package tasker

import (
	"encoding/json"
	"time"
)

type Task struct {
	TaskId      string
	TaskHandler string
	Desc        string
	Param       json.RawMessage
}

type TaskScheme struct {
	SchemeId string
	TaskId   string
	Cron     string

	cronId int
}

type TaskExec struct {
	ExecId    string
	TaskId    string
	StartTime time.Time
	EndTime   time.Time
	Success   bool
	Result    json.RawMessage
}

type TaskEvent struct {
	ExecId string
	Time   time.Time
	// Debug/Normal/Warning/Error
	Level   string
	Message string
}

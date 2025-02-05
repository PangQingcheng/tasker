package tasker

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestTasker(t *testing.T) {
	tasker := Default()
	tasker.SetStorage(NewMemoryStorage())
	tasker.Register("TestTask", func(c *TaskContext) error {
		c.Event("info", "test task start")
		fmt.Println("hello, world!")
		c.Event("info", "test task end")
		return c.Success(nil)
	})

	tasker.AddTask("task1", "TestTask", "this is a test task", "{}")

	execId, err := tasker.RunTask("task1")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	events, err := tasker.Events(execId)
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))

	if len(events) != 2 {
		t.Fatal("events length not equal 2")
	}
}

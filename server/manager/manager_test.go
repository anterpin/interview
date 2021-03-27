package manager

import (
	"testing"
	"time"
)

func TestSequential(t *testing.T) {
	tt := []struct {
		name         string
		command      string
		failingStage int
	}{
		{"empty program", "", 1},
		{"easy program with args", "ls -la", 2},
		{"loop program", "watch date", 0},
		{"no existing program", "jadfadf", 1},
	}

	manager := NewManager()
	manager.AddUser(1)

	const userid = 1
	for _, tc := range tt {
		// start test
		processId, err := manager.Start(tc.command, userid)
		if err != nil {
			if tc.failingStage != 1 {
				t.Errorf("test %s shouldn't fail at stage 0", tc.name)
			}
			continue
		}
		// sleep to allow the process to run
		time.Sleep(time.Millisecond * 100)

		// list test
		_ = manager.List(userid)

		// log test
		_, err = manager.Log(processId, userid)
		if err != nil {
			t.Errorf("test %s shouldn't fail at logging", tc.name)
			continue
		}

		// status test
		state, err := manager.Status(processId, userid)
		if err != nil {
			t.Errorf("test %s shouldn't fail at status", tc.name)
			continue
		}
		if state == nil && tc.failingStage != 0 {
			t.Errorf("test %s shouldn't be active", tc.name)
			continue
		}
		if state != nil && tc.failingStage == 0 {
			t.Errorf("test %s shouldn't be terminated", tc.name)
			continue
		}

		// stop test
		err = manager.Stop(processId, userid)
		if err != nil && tc.failingStage == 0 {
			t.Errorf("test %s should be killed", tc.name)
			continue
		}

		if err == nil && tc.failingStage != 0 {
			t.Errorf("test %s should raise an error on kill", tc.name)
			continue
		}
	}

}

func TestEdgeCases(t *testing.T) {
	manager := NewManager()
	const userid = 1
	manager.AddUser(userid)
	t.Run("unvalid program id", func(t *testing.T) {
		_, err := manager.Status("cadfljkl", userid)
		if err == nil {
			t.Fatalf("%s failed", t.Name())
		}
	})

	t.Run("unknown program id", func(t *testing.T) {
		_, err := manager.Status("95bf5b81-74bc-47e7-8622-e2aace3e866f", userid)
		if err == nil {
			t.Fatalf("%s failed", t.Name())
		}
	})

	t.Run("handling new userid", func(t *testing.T) {
		_, err := manager.Start("echo hello", 2)
		if err != nil {
			t.Fatalf("%s failed", t.Name())
		}
	})

}

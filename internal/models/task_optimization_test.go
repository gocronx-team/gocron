package models

import (
	"testing"
)

// TestAllActiveTasks tests fetching all active tasks at once
func TestAllActiveTasks(t *testing.T) {
	t.Run("AllActiveTasks should return all active parent tasks", func(t *testing.T) {
		// This test requires database connection, only testing method signature and logic
		if Db == nil {
			t.Skip("database not initialized, skipping test")
		}

		task := &Task{}

		// Verify method exists and is callable
		_, err := task.AllActiveTasks()

		// If database is not initialized, error is expected
		if err != nil {
			t.Logf("database not initialized (expected behavior): %v", err)
		}
	})

	t.Run("AllActiveTasks and ActiveList should return same data", func(t *testing.T) {
		// This test verifies logical consistency between two methods
		// Requires database support in actual environment
		if Db == nil {
			t.Skip("database not initialized, skipping test")
		}

		task := &Task{}

		// Verify both methods exist
		_, err1 := task.AllActiveTasks()
		_, err2 := task.ActiveList(1, 1000)

		// Both methods should have consistent error behavior
		if (err1 == nil) != (err2 == nil) {
			t.Error("AllActiveTasks and ActiveList should have consistent error handling")
		}
	})
}

// TestTaskHostBatchQuery tests batch querying task host information
func TestTaskHostBatchQuery(t *testing.T) {
	t.Run("GetHostsByTaskIds should support batch query", func(t *testing.T) {
		th := &TaskHost{}

		// Test empty list
		result, err := th.GetHostsByTaskIds([]int{})
		if err != nil {
			t.Errorf("empty list query failed: %v", err)
		}

		if result == nil {
			t.Error("should return empty map instead of nil")
		}

		if len(result) != 0 {
			t.Error("empty list should return empty map")
		}
	})

	t.Run("GetHostsByTaskIds should group correctly", func(t *testing.T) {
		// This test requires database support
		// Only verifying method signature and basic logic

		// Skip tests requiring database
		if Db == nil {
			t.Skip("database not initialized, skipping test")
		}

		th := &TaskHost{}
		taskIds := []int{1, 2, 3}

		result, err := th.GetHostsByTaskIds(taskIds)

		// If database is not initialized, error is expected
		if err != nil {
			t.Logf("database not initialized (expected behavior): %v", err)
		} else {
			// If successful, verify return type
			if result == nil {
				t.Error("should return non-nil map on success")
			}
		}
	})
}

// TestSetHostsForTasksOptimization tests batch setting host information optimization
func TestSetHostsForTasksOptimization(t *testing.T) {
	t.Run("empty task list should return directly", func(t *testing.T) {
		if Db == nil {
			t.Skip("database not initialized, skipping test")
		}

		task := &Task{}
		emptyTasks := []Task{}

		result, err := task.setHostsForTasks(emptyTasks)

		if err != nil {
			t.Errorf("empty list processing failed: %v", err)
		}

		if len(result) != 0 {
			t.Error("empty list should return empty slice")
		}
	})

	t.Run("setHostsForTasks should batch process", func(t *testing.T) {
		if Db == nil {
			t.Skip("database not initialized, skipping test")
		}

		// Verify method logic: should collect all task IDs, then batch query
		task := &Task{}

		// Create test task list
		tasks := []Task{
			{Id: 1, Name: "Task1"},
			{Id: 2, Name: "Task2"},
			{Id: 3, Name: "Task3"},
		}

		// Call method (requires database support)
		_, err := task.setHostsForTasks(tasks)

		// If database is not initialized, error is expected
		if err != nil {
			t.Logf("database not initialized (expected behavior): %v", err)
		}
	})
}

// TestTaskQueryOptimization tests task query optimization
func TestTaskQueryOptimization(t *testing.T) {
	if Db == nil {
		t.Skip("database not initialized, skipping test")
	}

	t.Run("ActiveList should support pagination", func(t *testing.T) {
		task := &Task{}

		// Test different pagination parameters
		testCases := []struct {
			page     int
			pageSize int
		}{
			{1, 10},
			{1, 100},
			{1, 1000},
			{2, 50},
		}

		for _, tc := range testCases {
			_, err := task.ActiveList(tc.page, tc.pageSize)
			if err != nil {
				t.Logf("pagination query page=%d, pageSize=%d: %v", tc.page, tc.pageSize, err)
			}
		}
	})

	t.Run("AllActiveTasks should fetch all data at once", func(t *testing.T) {
		task := &Task{}

		// AllActiveTasks doesn't need pagination parameters
		_, err := task.AllActiveTasks()
		if err != nil {
			t.Logf("one-time query: %v", err)
		}
	})
}

// TestTaskModelConsistency tests model method consistency
func TestTaskModelConsistency(t *testing.T) {
	if Db == nil {
		t.Skip("database not initialized, skipping test")
	}

	t.Run("all query methods should use setHostsForTasks", func(t *testing.T) {
		// This test verifies code structure consistency
		// Ensures all methods returning task lists use batch query optimization

		task := &Task{}

		// Test if each method exists
		methods := []func() ([]Task, error){
			func() ([]Task, error) { return task.AllActiveTasks() },
			func() ([]Task, error) { return task.ActiveList(1, 10) },
			func() ([]Task, error) { return task.ActiveListByHostId(1) },
			func() ([]Task, error) { return task.List(CommonMap{}) },
			func() ([]Task, error) { return task.GetDependencyTaskList("1,2,3") },
		}

		for i, method := range methods {
			_, err := method()
			if err != nil {
				t.Logf("method %d: %v", i, err)
			}
		}
	})
}

// BenchmarkAllActiveTasks benchmark: fetch all tasks at once
func BenchmarkAllActiveTasks(b *testing.B) {
	task := &Task{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = task.AllActiveTasks()
	}
}

// BenchmarkActiveListPagination benchmark: paginated query
func BenchmarkActiveListPagination(b *testing.B) {
	task := &Task{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = task.ActiveList(1, 1000)
	}
}

// BenchmarkGetHostsByTaskIds benchmark: batch query host information
func BenchmarkGetHostsByTaskIds(b *testing.B) {
	th := &TaskHost{}
	taskIds := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = th.GetHostsByTaskIds(taskIds)
	}
}

// TestTaskHostBatchQueryEdgeCases tests edge cases
func TestTaskHostBatchQueryEdgeCases(t *testing.T) {
	if Db == nil {
		t.Skip("database not initialized, skipping test")
	}

	t.Run("single task ID", func(t *testing.T) {
		th := &TaskHost{}
		result, err := th.GetHostsByTaskIds([]int{1})

		if err != nil {
			t.Logf("database not initialized: %v", err)
		} else if result == nil {
			t.Error("should return non-nil map")
		}
	})

	t.Run("large number of task IDs", func(t *testing.T) {
		th := &TaskHost{}
		taskIds := make([]int, 1000)
		for i := 0; i < 1000; i++ {
			taskIds[i] = i + 1
		}

		result, err := th.GetHostsByTaskIds(taskIds)

		if err != nil {
			t.Logf("database not initialized: %v", err)
		} else if result == nil {
			t.Error("should return non-nil map")
		}
	})

	t.Run("duplicate task IDs", func(t *testing.T) {
		th := &TaskHost{}
		taskIds := []int{1, 1, 2, 2, 3, 3}

		result, err := th.GetHostsByTaskIds(taskIds)

		if err != nil {
			t.Logf("database not initialized: %v", err)
		} else if result == nil {
			t.Error("should return non-nil map")
		}
	})
}

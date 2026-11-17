package service

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gocronx-team/gocron/internal/models"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// TestSingleInstanceControl 测试单实例运行控制
func TestSingleInstanceControl(t *testing.T) {
	t.Run("Multi=0时阻止并发执行", func(t *testing.T) {
		instance := &Instance{}
		taskId := 100

		// 第一次检查，应该不存在
		if instance.has(taskId) {
			t.Error("任务不应该在运行中")
		}

		// 添加任务
		instance.add(taskId)

		// 第二次检查，应该存在
		if !instance.has(taskId) {
			t.Error("任务应该在运行中")
		}

		// 完成任务
		instance.done(taskId)

		// 第三次检查，应该不存在
		if instance.has(taskId) {
			t.Error("任务不应该在运行中")
		}
	})

	t.Run("并发场景下的单实例控制", func(t *testing.T) {
		instance := &Instance{}
		taskId := 200
		var wg sync.WaitGroup
		executionCount := 0
		var mu sync.Mutex

		// 模拟10个并发请求
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// 使用互斥锁保护检查和添加操作的原子性
				mu.Lock()
				if !instance.has(taskId) {
					instance.add(taskId)
					executionCount++
					mu.Unlock()

					// 模拟任务执行
					time.Sleep(10 * time.Millisecond)

					instance.done(taskId)
				} else {
					mu.Unlock()
				}
			}()
		}

		wg.Wait()

		// 只有第一个请求应该执行
		if executionCount != 1 {
			t.Errorf("期望只有1次执行，实际执行了%d次", executionCount)
		}
	})

	t.Run("不同任务ID互不影响", func(t *testing.T) {
		instance := &Instance{}

		instance.add(1)
		instance.add(2)
		instance.add(3)

		if !instance.has(1) || !instance.has(2) || !instance.has(3) {
			t.Error("所有任务都应该在运行中")
		}

		instance.done(2)

		if !instance.has(1) || instance.has(2) || !instance.has(3) {
			t.Error("只有任务2应该被移除")
		}
	})
}

// TestBeforeExecJobSingleInstance 测试beforeExecJob中的单实例逻辑
func TestBeforeExecJobSingleInstance(t *testing.T) {
	t.Run("Multi=0且任务已运行时应该取消", func(t *testing.T) {
		// 重置runInstance
		runInstance = Instance{}

		task := models.Task{
			Id:    1,
			Name:  "测试任务",
			Multi: 0, // 不允许并发
		}

		// 模拟任务已在运行
		runInstance.add(task.Id)

		// 验证任务在运行中
		if !runInstance.has(task.Id) {
			t.Error("任务应该在运行中")
		}

		// 模拟beforeExecJob的逻辑：如果Multi=0且任务已运行，应该取消
		shouldCancel := task.Multi == 0 && runInstance.has(task.Id)
		if !shouldCancel {
			t.Error("任务已在运行，应该取消本次执行")
		}

		// 清理
		runInstance.done(task.Id)
	})

	t.Run("Multi=1时允许并发执行", func(t *testing.T) {
		// 重置runInstance
		runInstance = Instance{}

		task := models.Task{
			Id:    2,
			Name:  "测试任务",
			Multi: 1, // 允许并发
		}

		// 模拟任务已在运行
		runInstance.add(task.Id)

		// Multi=1时，不应该取消
		shouldCancel := task.Multi == 0 && runInstance.has(task.Id)
		if shouldCancel {
			t.Error("Multi=1时应该允许并发执行")
		}

		// 清理
		runInstance.done(task.Id)
	})
}

// TestCreateJobSingleInstanceLogic 测试createJob中的单实例逻辑
func TestCreateJobSingleInstanceLogic(t *testing.T) {
	t.Run("Multi=0时应该添加和移除实例标记", func(t *testing.T) {
		// 重置runInstance
		runInstance = Instance{}

		taskId := 100

		// 模拟createJob中的逻辑
		if !runInstance.has(taskId) {
			// Multi=0时，添加实例标记
			runInstance.add(taskId)

			// 验证已添加
			if !runInstance.has(taskId) {
				t.Error("应该已添加实例标记")
			}

			// 模拟任务执行完成
			runInstance.done(taskId)

			// 验证已移除
			if runInstance.has(taskId) {
				t.Error("应该已移除实例标记")
			}
		}
	})

	t.Run("Multi=1时不应该添加实例标记", func(t *testing.T) {
		// 重置runInstance
		runInstance = Instance{}

		taskId := 200
		multi := 1

		// Multi=1时，不添加实例标记
		if multi == 0 {
			runInstance.add(taskId)
		}

		// 验证未添加
		if runInstance.has(taskId) {
			t.Error("Multi=1时不应该添加实例标记")
		}
	})
}

// TestInstanceThreadSafety 测试Instance的线程安全性
func TestInstanceThreadSafety(t *testing.T) {
	instance := &Instance{}
	var wg sync.WaitGroup

	// 并发添加和删除
	for i := 0; i < 100; i++ {
		wg.Add(3)

		taskId := i

		// 添加
		go func(id int) {
			defer wg.Done()
			instance.add(id)
		}(taskId)

		// 检查
		go func(id int) {
			defer wg.Done()
			_ = instance.has(id)
		}(taskId)

		// 删除
		go func(id int) {
			defer wg.Done()
			time.Sleep(1 * time.Millisecond)
			instance.done(id)
		}(taskId)
	}

	wg.Wait()

	// 测试通过表示没有发生竞态条件
	t.Log("线程安全测试通过")
}

// TestSingleInstanceRealScenario 测试真实场景
func TestSingleInstanceRealScenario(t *testing.T) {
	t.Run("定时任务触发时上次未完成", func(t *testing.T) {
		runInstance = Instance{}

		task := models.Task{
			Id:    1,
			Name:  "慢速任务",
			Multi: 0,
		}

		var executionCount int
		var mu sync.Mutex

		// 模拟第一次执行（耗时较长）
		go func() {
			if task.Multi == 0 && runInstance.has(task.Id) {
				return
			}

			if task.Multi == 0 {
				runInstance.add(task.Id)
				defer runInstance.done(task.Id)
			}

			mu.Lock()
			executionCount++
			mu.Unlock()

			time.Sleep(50 * time.Millisecond)
		}()

		// 等待第一次执行开始
		time.Sleep(5 * time.Millisecond)

		// 模拟第二次触发（应该被阻止）
		if task.Multi == 0 && runInstance.has(task.Id) {
			t.Log("第二次执行被正确阻止")
		} else {
			if task.Multi == 0 {
				runInstance.add(task.Id)
				defer runInstance.done(task.Id)
			}

			mu.Lock()
			executionCount++
			mu.Unlock()
		}

		// 等待第一次执行完成
		time.Sleep(60 * time.Millisecond)

		mu.Lock()
		count := executionCount
		mu.Unlock()

		if count != 1 {
			t.Errorf("期望执行1次，实际执行%d次", count)
		}
	})
}

func TestCreateJobSkipsRunWhileInstanceRunning(t *testing.T) {
	originalDB, originalPrefix := models.Db, models.TablePrefix
	t.Cleanup(func() { models.Db, models.TablePrefix = originalDB, originalPrefix })
	db, err := gorm.Open(gormlite.Open(filepath.Join(t.TempDir(), "single.db")), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	models.Db, models.TablePrefix = db, ""
	if err := new(models.Migration).Install("single"); err != nil {
		t.Fatal(err)
	}
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits.Add(1) }))
	defer srv.Close()
	task := models.Task{Name: "single", Level: models.TaskLevelParent, Spec: "* * * * * *",
		Protocol: models.TaskHTTP, Command: srv.URL, HttpMethod: models.TaskHTTPMethodGet,
		DependencyStatus: models.TaskDependencyStatusWeak, Status: models.Enabled, Multi: 0, Timeout: 10}
	if err := db.Create(&task).Error; err != nil {
		t.Fatal(err)
	}

	originalQueue := concurrencyQueue
	concurrencyQueue = ConcurrencyQueue{queue: make(chan struct{}, 1)}
	t.Cleanup(func() { concurrencyQueue = originalQueue })

	// An earlier run of the task is still going
	runInstance.add(task.Id)
	defer runInstance.done(task.Id)

	createJob(task)()

	if got := hits.Load(); got != 0 {
		t.Fatalf("command ran %d times while the task was running, want 0", got)
	}
	if !runInstance.has(task.Id) {
		t.Fatal("running mark of the earlier run was cleared")
	}
	var log models.TaskLog
	if err := db.First(&log).Error; err != nil {
		t.Fatal(err)
	}
	if log.Status != models.Cancel {
		t.Fatalf("log status = %d, want Cancel", log.Status)
	}
}

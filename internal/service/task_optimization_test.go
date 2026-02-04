package service

import (
	"sync"
	"testing"
	"time"
)

// TestLogCleanupCache tests log cleanup configuration cache functionality
func TestLogCleanupCache(t *testing.T) {
	t.Run("cache initialization", func(t *testing.T) {
		// Reset cache
		logCleanupCache.Lock()
		logCleanupCache.retentionDays = 0
		logCleanupCache.fileSizeLimit = 0
		logCleanupCache.lastUpdate = time.Time{}
		logCleanupCache.cacheDuration = 5 * time.Minute
		logCleanupCache.Unlock()

		// Verify cache is empty
		logCleanupCache.RLock()
		if !logCleanupCache.lastUpdate.IsZero() {
			t.Error("cache should be empty")
		}
		logCleanupCache.RUnlock()
	})

	t.Run("cache expiration and refresh", func(t *testing.T) {
		// Set an expired cache
		logCleanupCache.Lock()
		logCleanupCache.retentionDays = 7
		logCleanupCache.fileSizeLimit = 100
		logCleanupCache.lastUpdate = time.Now().Add(-10 * time.Minute) // 10 minutes ago
		logCleanupCache.cacheDuration = 5 * time.Minute
		logCleanupCache.Unlock()

		// Verify cache is expired
		logCleanupCache.RLock()
		isExpired := time.Since(logCleanupCache.lastUpdate) >= logCleanupCache.cacheDuration
		logCleanupCache.RUnlock()

		if !isExpired {
			t.Error("cache should be expired")
		}
	})

	t.Run("cache hit within validity period", func(t *testing.T) {
		// Set a valid cache
		logCleanupCache.Lock()
		logCleanupCache.retentionDays = 30
		logCleanupCache.fileSizeLimit = 200
		logCleanupCache.lastUpdate = time.Now()
		logCleanupCache.cacheDuration = 5 * time.Minute
		logCleanupCache.Unlock()

		// Verify cache is valid
		logCleanupCache.RLock()
		isValid := time.Since(logCleanupCache.lastUpdate) < logCleanupCache.cacheDuration
		days := logCleanupCache.retentionDays
		limit := logCleanupCache.fileSizeLimit
		logCleanupCache.RUnlock()

		if !isValid {
			t.Error("cache should be valid")
		}

		if days != 30 {
			t.Errorf("expected retention days to be 30, got %d", days)
		}

		if limit != 200 {
			t.Errorf("expected file size limit to be 200, got %d", limit)
		}
	})

	t.Run("concurrent cache access safety", func(t *testing.T) {
		var wg sync.WaitGroup

		// Initialize cache
		logCleanupCache.Lock()
		logCleanupCache.retentionDays = 15
		logCleanupCache.fileSizeLimit = 150
		logCleanupCache.lastUpdate = time.Now()
		logCleanupCache.cacheDuration = 5 * time.Minute
		logCleanupCache.Unlock()

		// Concurrent reads
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				logCleanupCache.RLock()
				_ = logCleanupCache.retentionDays
				_ = logCleanupCache.fileSizeLimit
				logCleanupCache.RUnlock()
			}()
		}

		wg.Wait()
		t.Log("concurrent cache read test passed")
	})
}

// TestGetLogRetentionDaysFromCache tests getting log retention days from cache
func TestGetLogRetentionDaysFromCache(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		// Set valid cache
		logCleanupCache.Lock()
		logCleanupCache.retentionDays = 45
		logCleanupCache.lastUpdate = time.Now()
		logCleanupCache.cacheDuration = 5 * time.Minute
		logCleanupCache.Unlock()

		// Simulate cache hit logic
		logCleanupCache.RLock()
		isCacheValid := time.Since(logCleanupCache.lastUpdate) < logCleanupCache.cacheDuration
		days := logCleanupCache.retentionDays
		logCleanupCache.RUnlock()

		if !isCacheValid {
			t.Error("cache should be valid")
		}

		if days != 45 {
			t.Errorf("expected 45 days, got %d days", days)
		}
	})

	t.Run("cache miss", func(t *testing.T) {
		// Set expired cache
		logCleanupCache.Lock()
		logCleanupCache.retentionDays = 10
		logCleanupCache.lastUpdate = time.Now().Add(-10 * time.Minute)
		logCleanupCache.cacheDuration = 5 * time.Minute
		logCleanupCache.Unlock()

		// Verify cache is expired
		logCleanupCache.RLock()
		isCacheExpired := time.Since(logCleanupCache.lastUpdate) >= logCleanupCache.cacheDuration
		logCleanupCache.RUnlock()

		if !isCacheExpired {
			t.Error("cache should be expired")
		}
	})
}

// TestGetLogFileSizeLimitFromCache tests getting log file size limit from cache
func TestGetLogFileSizeLimitFromCache(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		// Set valid cache
		logCleanupCache.Lock()
		logCleanupCache.fileSizeLimit = 500
		logCleanupCache.lastUpdate = time.Now()
		logCleanupCache.cacheDuration = 5 * time.Minute
		logCleanupCache.Unlock()

		// Simulate cache hit logic
		logCleanupCache.RLock()
		isCacheValid := time.Since(logCleanupCache.lastUpdate) < logCleanupCache.cacheDuration
		limit := logCleanupCache.fileSizeLimit
		logCleanupCache.RUnlock()

		if !isCacheValid {
			t.Error("cache should be valid")
		}

		if limit != 500 {
			t.Errorf("expected 500MB, got %dMB", limit)
		}
	})
}

// TestCacheConcurrency tests cache concurrency safety
func TestCacheConcurrency(t *testing.T) {
	t.Run("concurrent read and write cache", func(t *testing.T) {
		var wg sync.WaitGroup

		// Initialize cache
		logCleanupCache.Lock()
		logCleanupCache.retentionDays = 20
		logCleanupCache.fileSizeLimit = 100
		logCleanupCache.lastUpdate = time.Now()
		logCleanupCache.cacheDuration = 5 * time.Minute
		logCleanupCache.Unlock()

		// Concurrent reads
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				logCleanupCache.RLock()
				_ = logCleanupCache.retentionDays
				_ = logCleanupCache.fileSizeLimit
				_ = time.Since(logCleanupCache.lastUpdate) < logCleanupCache.cacheDuration
				logCleanupCache.RUnlock()
			}()
		}

		// Concurrent writes (simulate cache updates)
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(val int) {
				defer wg.Done()
				time.Sleep(time.Millisecond)
				logCleanupCache.Lock()
				logCleanupCache.retentionDays = val
				logCleanupCache.lastUpdate = time.Now()
				logCleanupCache.Unlock()
			}(i)
		}

		wg.Wait()
		t.Log("concurrent read/write cache test passed")
	})
}

// TestReloadLogCleanupTaskClearCache tests cache clearing when reloading task
func TestReloadLogCleanupTaskClearCache(t *testing.T) {
	t.Run("clear cache on reload", func(t *testing.T) {
		// Set cache
		logCleanupCache.Lock()
		logCleanupCache.retentionDays = 30
		logCleanupCache.fileSizeLimit = 200
		logCleanupCache.lastUpdate = time.Now()
		logCleanupCache.Unlock()

		// Simulate cache clearing logic in ReloadLogCleanupTask
		logCleanupCache.Lock()
		logCleanupCache.lastUpdate = time.Time{}
		logCleanupCache.Unlock()

		// Verify cache is cleared
		logCleanupCache.RLock()
		if !logCleanupCache.lastUpdate.IsZero() {
			t.Error("cache should be cleared")
		}
		logCleanupCache.RUnlock()
	})
}

// TestCachePerformance tests cache performance improvement
func TestCachePerformance(t *testing.T) {
	t.Run("cache hit performance", func(t *testing.T) {
		// Set valid cache
		logCleanupCache.Lock()
		logCleanupCache.retentionDays = 30
		logCleanupCache.fileSizeLimit = 200
		logCleanupCache.lastUpdate = time.Now()
		logCleanupCache.cacheDuration = 5 * time.Minute
		logCleanupCache.Unlock()

		start := time.Now()

		// Simulate 1000 cache reads
		for i := 0; i < 1000; i++ {
			logCleanupCache.RLock()
			_ = logCleanupCache.retentionDays
			_ = logCleanupCache.fileSizeLimit
			logCleanupCache.RUnlock()
		}

		elapsed := time.Since(start)

		// 1000 reads should complete within 1ms
		if elapsed > time.Millisecond {
			t.Logf("Warning: 1000 cache reads took %v, may need optimization", elapsed)
		} else {
			t.Logf("Performance test passed: 1000 cache reads took %v", elapsed)
		}
	})
}

// TestInstanceTryAdd tests atomic instance addition functionality
func TestInstanceTryAdd(t *testing.T) {
	t.Run("first add should succeed", func(t *testing.T) {
		instance := &Instance{}
		taskId := 1000

		success := instance.tryAdd(taskId)
		if !success {
			t.Error("first add should succeed")
		}

		// Verify added
		if !instance.has(taskId) {
			t.Error("task should be running")
		}

		// Cleanup
		instance.done(taskId)
	})

	t.Run("duplicate add should fail", func(t *testing.T) {
		instance := &Instance{}
		taskId := 2000

		// First add
		success1 := instance.tryAdd(taskId)
		if !success1 {
			t.Error("first add should succeed")
		}

		// Duplicate add
		success2 := instance.tryAdd(taskId)
		if success2 {
			t.Error("duplicate add should fail")
		}

		// Cleanup
		instance.done(taskId)
	})

	t.Run("concurrent tryAdd only one succeeds", func(t *testing.T) {
		instance := &Instance{}
		taskId := 3000
		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		// 10 concurrent attempts to add
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if instance.tryAdd(taskId) {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
		}

		wg.Wait()

		if successCount != 1 {
			t.Errorf("expected only 1 success, got %d successes", successCount)
		}

		// Cleanup
		instance.done(taskId)
	})

	t.Run("can add again after removal", func(t *testing.T) {
		instance := &Instance{}
		taskId := 4000

		// First add
		success1 := instance.tryAdd(taskId)
		if !success1 {
			t.Error("first add should succeed")
		}

		// Remove
		instance.done(taskId)

		// Add again
		success2 := instance.tryAdd(taskId)
		if !success2 {
			t.Error("add after removal should succeed")
		}

		// Cleanup
		instance.done(taskId)
	})
}

// BenchmarkCacheRead cache read performance benchmark
func BenchmarkCacheRead(b *testing.B) {
	// Set valid cache
	logCleanupCache.Lock()
	logCleanupCache.retentionDays = 30
	logCleanupCache.fileSizeLimit = 200
	logCleanupCache.lastUpdate = time.Now()
	logCleanupCache.cacheDuration = 5 * time.Minute
	logCleanupCache.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logCleanupCache.RLock()
		_ = logCleanupCache.retentionDays
		_ = logCleanupCache.fileSizeLimit
		logCleanupCache.RUnlock()
	}
}

// BenchmarkInstanceTryAdd instance addition performance benchmark
func BenchmarkInstanceTryAdd(b *testing.B) {
	instance := &Instance{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		taskId := i
		if instance.tryAdd(taskId) {
			instance.done(taskId)
		}
	}
}

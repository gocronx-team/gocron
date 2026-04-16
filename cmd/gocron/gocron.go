// Command gocron

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/app"
	"github.com/gocronx-team/gocron/internal/modules/leader"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/setting"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers"
	"github.com/gocronx-team/gocron/internal/service"
	"github.com/urfave/cli/v2"
)

var (
	AppVersion           = "1.6.0"
	BuildDate, GitCommit string

	// leaderElection 全局选举实例，用于 graceful shutdown 时释放锁
	leaderElection *leader.Election
)

// Default port for web server
const DefaultPort = 5920

// Graceful shutdown timeout
const shutdownTimeout = 30 * time.Second

func main() {
	cliApp := cli.NewApp()
	cliApp.Name = "gocron"
	cliApp.Usage = "gocron service"
	cliApp.Version, _ = utils.FormatAppVersion(AppVersion, GitCommit, BuildDate)
	cliApp.Commands = getCommands()
	cliApp.Flags = append(cliApp.Flags, []cli.Flag{}...)

	// Auto-append "web" command when double-clicking on Windows
	if len(os.Args) == 1 && utils.IsWindows() {
		os.Args = append(os.Args, "web")
	}

	err := cliApp.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}

// getCommands
func getCommands() []*cli.Command {
	command := &cli.Command{
		Name:   "web",
		Usage:  "run web server",
		Action: runWeb,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Value: "0.0.0.0",
				Usage: "bind host",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   DefaultPort,
				Usage:   "bind port",
			},
			&cli.StringFlag{
				Name:    "env",
				Aliases: []string{"e"},
				Value:   "prod",
				Usage:   "runtime environment, dev|test|prod",
			},
		},
	}

	return []*cli.Command{command}
}

func runWeb(ctx *cli.Context) error {
	// Set runtime environment
	setEnvironment(ctx)
	fmt.Printf("Starting gocron web server...\n")
	// Initialize application
	app.InitEnv(AppVersion)
	fmt.Printf("Application initialized\n")
	// Initialize modules: DB, scheduled tasks, etc.
	initModule()
	fmt.Printf("Modules initialized\n")

	r := gin.Default()
	// Register middleware
	routers.RegisterMiddleware(r)
	// Register routes
	routers.Register(r)

	host := parseHost(ctx)
	port := parsePort(ctx)
	addr := fmt.Sprintf("%s:%d", host, port)

	// Use http.Server to support graceful shutdown
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Start HTTP server in a goroutine
	go func() {
		fmt.Printf("Server listening on %s\n", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal, blocks the main goroutine
	waitForShutdown(srv)

	return nil
}

func initModule() {
	if !app.Installed {
		return
	}

	config, err := setting.Read(app.AppConfig)
	if err != nil {
		logger.Fatal("Failed to read application config", err)
	}
	app.Setting = config

	// Initialize DB
	models.Db = models.CreateDb()

	// Version upgrade
	upgradeIfNeed()

	// Auto-create missing tables
	ensureTables()

	// Repair missing settings records
	if err := models.RepairSettings(); err != nil {
		logger.Error("Failed to repair settings records", err)
	}

	// Initialize scheduler infrastructure
	service.ServiceTask.Initialize()

	// SQLite: single-node only, skip leader election
	if models.Db.Dialector.Name() == "sqlite" {
		logger.Info("SQLite detected, skipping leader election (single-node mode)")
		service.ServiceTask.StartScheduler()
		return
	}

	// Ensure scheduler_lock table exists
	if !models.Db.Migrator().HasTable(&models.SchedulerLock{}) {
		logger.Info("scheduler_lock table not found, creating...")
		if err := models.Db.AutoMigrate(&models.SchedulerLock{}); err != nil {
			logger.Error("Failed to create scheduler_lock table", err)
		}
	}

	// Start leader election — scheduler only runs on the leader node
	leaderElection = leader.New(
		models.Db,
		func() { service.ServiceTask.StartScheduler() }, // onElected
		func() { service.ServiceTask.StopScheduler() },  // onEvicted
	)
	leaderElection.Start()
}

// parsePort parses the port from CLI flags
func parsePort(ctx *cli.Context) int {
	port := DefaultPort
	if ctx.IsSet("port") {
		port = ctx.Int("port")
	}
	if port <= 0 || port >= 65535 {
		port = DefaultPort
	}

	return port
}

func parseHost(ctx *cli.Context) string {
	if ctx.IsSet("host") {
		return ctx.String("host")
	}

	return "0.0.0.0"
}

func setEnvironment(ctx *cli.Context) {
	env := "prod"
	if ctx.IsSet("env") {
		env = ctx.String("env")
	}

	switch env {
	case "test":
		gin.SetMode(gin.TestMode)
	case "dev":
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
}

// waitForShutdown waits for OS signals and performs graceful shutdown
func waitForShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for {
		s := <-quit
		logger.Info("Received signal -- ", s)
		switch s {
		case syscall.SIGHUP:
			logger.Info("Received terminal disconnect signal, ignoring")
			continue
		case syscall.SIGINT, syscall.SIGTERM:
			// Proceed to graceful shutdown
		}
		break
	}

	logger.Info("Shutting down gracefully, press Ctrl+C again to force exit...")

	// Allow forced exit: immediately exit on receiving another signal
	go func() {
		forceQuit := make(chan os.Signal, 1)
		signal.Notify(forceQuit, syscall.SIGINT, syscall.SIGTERM)
		<-forceQuit
		logger.Warn("Forced shutdown")
		os.Exit(1)
	}()

	// Step 1: Stop HTTP server, reject new requests, wait for in-flight requests to complete
	logger.Info("Step 1/3: Stopping HTTP server (waiting for in-flight requests)...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("HTTP server shutdown error: %v", err)
	} else {
		logger.Info("HTTP server stopped successfully")
	}

	// Step 2: Stop leader election and release lock
	if app.Installed {
		if leaderElection != nil {
			logger.Info("Step 2/4: Stopping leader election (releasing lock)...")
			leaderElection.Stop()
			logger.Info("Leader election stopped")
		}

		// Step 3: Stop scheduled task scheduler, wait for running tasks to complete
		logger.Info("Step 3/4: Stopping scheduled task scheduler (waiting for running tasks)...")
		service.ServiceTask.WaitAndExit()
		logger.Info("Scheduled task scheduler stopped")

		// Step 4: Close database connections
		logger.Info("Step 4/4: Closing database connections...")
		closeDatabase()
		logger.Info("Database connections closed")
	}

	logger.Info("Graceful shutdown completed")
	logger.Close()
}

// closeDatabase closes the database connection pool
func closeDatabase() {
	if models.Db == nil {
		return
	}
	sqlDB, err := models.Db.DB()
	if err != nil {
		logger.Errorf("Failed to get database connection for closing: %v", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		logger.Errorf("Failed to close database connection: %v", err)
	}
}

// upgradeIfNeed checks if the app needs upgrading when version file exists and version < app.VersionId
func upgradeIfNeed() {
	currentVersionId := app.GetCurrentVersionId()
	// No version file found
	if currentVersionId == 0 {
		return
	}
	if currentVersionId >= app.VersionId {
		return
	}

	migration := new(models.Migration)
	logger.Infof("Starting version upgrade, current version: %d", currentVersionId)

	migration.Upgrade(currentVersionId)
	app.UpdateVersionFile()

	logger.Infof("Upgraded to latest version: %d", app.VersionId)
}

// ensureTables ensures all required tables exist
func ensureTables() {
	if !models.Db.Migrator().HasTable(&models.AgentToken{}) {
		logger.Info("agent_token table not found, creating...")
		if err := models.Db.AutoMigrate(&models.AgentToken{}); err != nil {
			logger.Error("Failed to create agent_token table", err)
		} else {
			logger.Info("agent_token table created successfully")
		}
	}

	if !models.Db.Migrator().HasTable(&models.AuditLog{}) {
		logger.Info("audit_log table not found, creating...")
		if err := models.Db.AutoMigrate(&models.AuditLog{}); err != nil {
			logger.Error("Failed to create audit_log table", err)
		} else {
			logger.Info("audit_log table created successfully")
		}
	}

	// 始终 AutoMigrate 新表，确保字段同步（AutoMigrate 幂等，只加列不删列）
	if err := models.Db.AutoMigrate(&models.TaskScriptVersion{}); err != nil {
		logger.Error("Failed to migrate task_script_version table", err)
	}
	if err := models.Db.AutoMigrate(&models.TaskTemplate{}); err != nil {
		logger.Error("Failed to migrate task_template table", err)
	}
}

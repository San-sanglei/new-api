package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/oauth"
	perfmetrics "github.com/QuantumNous/new-api/pkg/perf_metrics"
	"github.com/QuantumNous/new-api/relay"
	"github.com/QuantumNous/new-api/router"
	"github.com/QuantumNous/new-api/service"
	_ "github.com/QuantumNous/new-api/setting/performance_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "net/http/pprof"
)

//go:embed web/default/dist
var buildFS embed.FS

//go:embed web/default/dist/index.html
var indexPage []byte

//go:embed web/classic/dist
var classicBuildFS embed.FS

//go:embed web/classic/dist/index.html
var classicIndexPage []byte

func main() {
	startTime := time.Now()

	// P1 修复：初始化应用生命周期 context，所有后台任务通过此 ctx 接收退出信号
	common.InitLifecycle()
	appCtx := common.LifecycleContext()

	err := InitResources()
	if err != nil {
		common.FatalLog("failed to initialize resources: " + err.Error())
		return
	}

	common.SysLog("Took " + common.Version + " started")
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	if common.DebugEnabled {
		common.SysLog("running in debug mode")
	}

	if common.RedisEnabled {
		// for compatibility with old versions
		common.MemoryCacheEnabled = true
	}
	if common.MemoryCacheEnabled {
		common.SysLog("memory cache enabled")
		common.SysLog(fmt.Sprintf("sync frequency: %d seconds", common.SyncFrequency))

		// Add panic recovery and retry for InitChannelCache
		func() {
			defer func() {
				if r := recover(); r != nil {
					common.SysLog(fmt.Sprintf("InitChannelCache panic: %v, retrying once", r))
					// Retry once
					// P4-5: FixAbility 现在区分 DB 级错误 (fixErr) 与缓存刷新错误 (cacheErr)。
					// - fixErr != nil: DB 写入失败，启动阶段致命，FatalLog
					// - cacheErr != nil: abilities 已写入但缓存未刷新，非致命（SyncChannelCache 会定期同步）
					_, _, cacheErr, fixErr := model.FixAbility()
					if fixErr != nil {
						common.FatalLog(fmt.Sprintf("InitChannelCache failed: %s", fixErr.Error()))
					} else if cacheErr != nil {
						common.SysError(fmt.Sprintf(
							"InitChannelCache retry: cache refresh failed but abilities table rebuilt: cache_refresh_required=true, err=%v",
							cacheErr,
						))
					}
				}
			}()
			if err := model.InitChannelCache(); err != nil {
				common.SysError(fmt.Sprintf("InitChannelCache failed: %v", err))
			}
		}()

		// P1 修复：通过 GoSafeWithContext 启动，支持 panic recovery 和 ctx 退出
		common.GoSafeWithContext(func(ctx context.Context) {
			model.SyncChannelCache(ctx, common.SyncFrequency)
		})
	}

	// 热更新配置
	common.GoSafeWithContext(func(ctx context.Context) {
		model.SyncOptions(ctx, common.SyncFrequency)
	})

	// 数据看板
	common.GoSafeWithContext(func(ctx context.Context) {
		model.UpdateQuotaData(ctx)
	})

	if os.Getenv("CHANNEL_UPDATE_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("CHANNEL_UPDATE_FREQUENCY"))
		if err != nil {
			common.FatalLog("failed to parse CHANNEL_UPDATE_FREQUENCY: " + err.Error())
		}
		common.GoSafeWithContext(func(ctx context.Context) {
			controller.AutomaticallyUpdateChannels(ctx, frequency)
		})
	}

	common.GoSafeWithContext(func(ctx context.Context) {
		controller.AutomaticallyTestChannels(ctx)
	})

	// Codex credential auto-refresh check every 10 minutes, refresh when expires within 1 day
	service.StartCodexCredentialAutoRefreshTask(appCtx)

	// Subscription quota reset task (daily/weekly/monthly/custom)
	service.StartSubscriptionQuotaResetTask(appCtx)

	// Wire task polling adaptor factory (breaks service -> relay import cycle)
	service.GetTaskAdaptorFunc = func(platform constant.TaskPlatform) service.TaskPollingAdaptor {
		a := relay.GetTaskAdaptor(platform)
		if a == nil {
			return nil
		}
		return a
	}

	// Channel upstream model update check task
	controller.StartChannelUpstreamModelUpdateTask(appCtx)

	if common.IsMasterNode && constant.UpdateTask {
		common.GoSafeWithContext(func(ctx context.Context) {
			controller.UpdateMidjourneyTaskBulk(ctx)
		})
		common.GoSafeWithContext(func(ctx context.Context) {
			controller.UpdateTaskBulk(ctx)
		})
	}
	if os.Getenv("BATCH_UPDATE_ENABLED") == "true" {
		common.BatchUpdateEnabled = true
		common.SysLog("batch update enabled with interval " + strconv.Itoa(common.BatchUpdateInterval) + "s")
		model.InitBatchUpdater(appCtx)
	}

	if os.Getenv("ENABLE_PPROF") == "true" {
		common.GoSafe(func() {
			// pprof 仅绑定本地回环，禁止公网访问
			pprofSrv := &http.Server{Addr: "127.0.0.1:8005", Handler: nil}
			go func() {
				if err := pprofSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					common.SysError("pprof server error: " + err.Error())
				}
			}()
			<-appCtx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = pprofSrv.Shutdown(shutdownCtx)
			common.SysLog("pprof server stopped")
		})
		common.GoSafeWithContext(func(ctx context.Context) {
			common.Monitor(ctx)
		})
		common.SysLog("pprof enabled")
	}

	err = common.StartPyroScope()
	if err != nil {
		common.SysError(fmt.Sprintf("start pyroscope error : %v", err))
	}

	// Initialize HTTP server
	server := gin.New()
	server.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		common.SysLog(fmt.Sprintf("panic detected: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("Panic detected, error: %v. Please submit a issue here: https://github.com/Calcium-Ion/new-api", err),
				"type":    "new_api_panic",
			},
		})
	}))
	// This will cause SSE not to work!!!
	//server.Use(gzip.Gzip(gzip.DefaultCompression))
	server.Use(middleware.RequestId())
	server.Use(middleware.PoweredBy())
	server.Use(middleware.I18n())
	server.Use(middleware.StatsMiddleware())
	middleware.SetUpLogger(server)
	// Initialize session store
	store := cookie.NewStore([]byte(common.SessionSecret))
	// COOKIE_SECURE 环境变量控制，生产环境应设为 true（HTTPS）
	// 未设置时默认 true（安全优先），开发环境设为 false
	cookieSecure := os.Getenv("COOKIE_SECURE") != "false"
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   2592000, // 30 days
		HttpOnly: true,
		Secure:   cookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
	server.Use(sessions.Sessions("session", store))

	InjectUmamiAnalytics()
	InjectGoogleAnalytics()

	// 设置路由
	router.SetRouter(server, router.ThemeAssets{
		DefaultBuildFS:   buildFS,
		DefaultIndexPage: indexPage,
		ClassicBuildFS:   classicBuildFS,
		ClassicIndexPage: classicIndexPage,
	})
	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}

	// Log startup success message
	common.LogStartupSuccess(startTime, port)

	// P1 修复：使用 http.Server + Shutdown 实现优雅关闭
	// 原 server.Run() 会阻塞且无法响应 SIGTERM/SIGINT，导致部署重启超时被强杀
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: server,
	}

	// 在 goroutine 中启动 HTTP server，主 goroutine 用于监听信号
	go func() {
		common.SysLog("HTTP server listening on :" + port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			common.FatalLog("failed to start HTTP server: " + err.Error())
		}
	}()

	// P1 修复：监听 SIGTERM/SIGINT 信号，触发优雅关闭
	// 退出流程：收到信号 → 停止接收新请求 → 等待正在处理请求完成 → 关闭 HTTP/DB/Redis/后台goroutine
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	common.SysLog(fmt.Sprintf("received signal %v, shutting down gracefully...", sig))

	// 1. 取消生命周期 context，通知所有后台 goroutine 退出
	common.ShutdownLifecycle()

	// 2. 停止接收新请求并等待正在处理的请求完成（最多 30 秒）
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		common.SysError("HTTP server shutdown error: " + err.Error())
	} else {
		common.SysLog("HTTP server stopped gracefully")
	}

	// 3. 刷盘 BatchUpdate 内存队列，避免 quota/token/channel 增量丢失
	// P3 修复：必须在 CloseDB 之前执行，否则 batchUpdate 无法落库。
	model.FlushBatchUpdate()

	// 4. 关闭 Redis 连接
	if common.RedisEnabled && common.RDB != nil {
		if err := common.RDB.Close(); err != nil {
			common.SysError("Redis close error: " + err.Error())
		} else {
			common.SysLog("Redis connection closed")
		}
	}

	// 5. 关闭数据库连接
	if err := model.CloseDB(); err != nil {
		common.FatalLog("failed to close database: " + err.Error())
	} else {
		common.SysLog("database connection closed")
	}

	common.SysLog("Took shutdown complete")
}

func InjectUmamiAnalytics() {
	analyticsInjectBuilder := &strings.Builder{}
	if os.Getenv("UMAMI_WEBSITE_ID") != "" {
		umamiSiteID := os.Getenv("UMAMI_WEBSITE_ID")
		umamiScriptURL := os.Getenv("UMAMI_SCRIPT_URL")
		if umamiScriptURL == "" {
			umamiScriptURL = "https://analytics.umami.is/script.js"
		}
		analyticsInjectBuilder.WriteString("<script defer src=\"")
		analyticsInjectBuilder.WriteString(umamiScriptURL)
		analyticsInjectBuilder.WriteString("\" data-website-id=\"")
		analyticsInjectBuilder.WriteString(umamiSiteID)
		analyticsInjectBuilder.WriteString("\"></script>")
	}
	analyticsInjectBuilder.WriteString("<!--Umami QuantumNous-->\n")
	analyticsInject := []byte(analyticsInjectBuilder.String())
	placeholder := []byte("<!--umami-->\n")
	indexPage = bytes.ReplaceAll(indexPage, placeholder, analyticsInject)
	classicIndexPage = bytes.ReplaceAll(classicIndexPage, placeholder, analyticsInject)
}

func InjectGoogleAnalytics() {
	analyticsInjectBuilder := &strings.Builder{}
	if os.Getenv("GOOGLE_ANALYTICS_ID") != "" {
		gaID := os.Getenv("GOOGLE_ANALYTICS_ID")
		// Google Analytics 4 (gtag.js)
		analyticsInjectBuilder.WriteString("<script async src=\"https://www.googletagmanager.com/gtag/js?id=")
		analyticsInjectBuilder.WriteString(gaID)
		analyticsInjectBuilder.WriteString("\"></script>")
		analyticsInjectBuilder.WriteString("<script>")
		analyticsInjectBuilder.WriteString("window.dataLayer = window.dataLayer || [];")
		analyticsInjectBuilder.WriteString("function gtag(){dataLayer.push(arguments);}")
		analyticsInjectBuilder.WriteString("gtag('js', new Date());")
		analyticsInjectBuilder.WriteString("gtag('config', '")
		analyticsInjectBuilder.WriteString(gaID)
		analyticsInjectBuilder.WriteString("');")
		analyticsInjectBuilder.WriteString("</script>")
	}
	analyticsInjectBuilder.WriteString("<!--Google Analytics QuantumNous-->\n")
	analyticsInject := []byte(analyticsInjectBuilder.String())
	placeholder := []byte("<!--Google Analytics-->\n")
	indexPage = bytes.ReplaceAll(indexPage, placeholder, analyticsInject)
	classicIndexPage = bytes.ReplaceAll(classicIndexPage, placeholder, analyticsInject)
}

func InitResources() error {
	// Initialize resources here if needed
	// This is a placeholder function for future resource initialization
	err := godotenv.Load(".env")
	if err != nil {
		if common.DebugEnabled {
			common.SysLog("No .env file found, using default environment variables. If needed, please create a .env file and set the relevant variables.")
		}
	}

	// 加载环境变量
	common.InitEnv()

	logger.SetupLogger()

	// Initialize model settings
	ratio_setting.InitRatioSettings()

	service.InitHttpClient()

	service.InitTokenEncoders()

	// Initialize SQL Database
	err = model.InitDB()
	if err != nil {
		common.FatalLog("failed to initialize database: " + err.Error())
		return err
	}

	model.CheckSetup()

	// Initialize options, should after model.InitDB()
	model.InitOptionMap()

	// 清理旧的磁盘缓存文件
	common.CleanupOldCacheFiles()

	// 初始化模型
	model.GetPricing()

	// Initialize SQL Database
	err = model.InitLogDB()
	if err != nil {
		return err
	}

	// Initialize Redis
	err = common.InitRedisClient()
	if err != nil {
		return err
	}

	perfmetrics.Init()

	// 启动系统监控
	common.StartSystemMonitor(common.LifecycleContext())

	// Initialize i18n
	err = i18n.Init()
	if err != nil {
		common.SysError("failed to initialize i18n: " + err.Error())
		// Don't return error, i18n is not critical
	} else {
		common.SysLog("i18n initialized with languages: " + strings.Join(i18n.SupportedLanguages(), ", "))
	}
	// Register user language loader for lazy loading
	i18n.SetUserLangLoader(model.GetUserLanguage)

	// Load custom OAuth providers from database
	err = oauth.LoadCustomProviders()
	if err != nil {
		common.SysError("failed to load custom OAuth providers: " + err.Error())
		// Don't return error, custom OAuth is not critical
	}

	return nil
}

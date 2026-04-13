package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ssnakerss/modmas/config"
	authPkg "github.com/Ssnakerss/modmas/internal/auth"
	"github.com/Ssnakerss/modmas/internal/ddl"
	fieldPkg "github.com/Ssnakerss/modmas/internal/field"
	"github.com/Ssnakerss/modmas/internal/logger"
	"github.com/Ssnakerss/modmas/internal/middleware"
	permissionPkg "github.com/Ssnakerss/modmas/internal/permission"
	rowPkg "github.com/Ssnakerss/modmas/internal/row"
	spreadsheetPkg "github.com/Ssnakerss/modmas/internal/spreadsheet"
	workspacePkg "github.com/Ssnakerss/modmas/internal/workspace"
	"github.com/Ssnakerss/modmas/pkg/jwt"
	"github.com/Ssnakerss/modmas/pkg/postgres"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {

	cfg := config.Load()
	l := logger.Setup(cfg.Env, os.Stdout)
	l.Log.Info("server startin config")
	l.Log.Info(fmt.Sprintf("env: %v", cfg))

	// ─── База данных ──────────────────────────────────────────────────────────
	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.DB.DSN())
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer pool.Close()

	l.Log.Info("connected to database")

	// ─── Инфраструктура ───────────────────────────────────────────────────────
	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL)
	ddlExec := ddl.NewExecutor(pool)

	// ─── Repositories ─────────────────────────────────────────────────────────
	authRepo := authPkg.NewRepository(pool)
	wsRepo := workspacePkg.NewRepository(pool)
	spreadsheetRepo := spreadsheetPkg.NewRepository(pool)
	fieldRepo := fieldPkg.NewRepository(pool)
	rowRepo := rowPkg.NewRepository(pool)
	permRepo := permissionPkg.NewRepository(pool)

	// ─── Services ─────────────────────────────────────────────────────────────

	// field.Service использует интерфейс SpreadsheetGetter — передаём spreadsheetRepo
	fieldService := fieldPkg.NewService(
		fieldRepo,
		spreadsheetRepo, // реализует field.SpreadsheetGetter
		ddlExec,
	)

	// spreadsheet.Service использует интерфейсы FieldRepository, WorkspaceGetter и PermissionRepository
	spreadsheetSvc := spreadsheetPkg.NewService(
		spreadsheetRepo,
		fieldRepo, // реализует spreadsheet.FieldRepository
		wsRepo,    // реализует spreadsheet.WorkspaceGetter
		ddlExec,
		permRepo, // реализует spreadsheet.PermissionRepository
	)

	// workspace.Service
	wsService := workspacePkg.NewService(wsRepo, ddlExec)

	// auth.Service
	authService := authPkg.NewService(authRepo, jwtManager)

	// permission.Enforcer и permission.Service
	enforcer := permissionPkg.NewEnforcer(pool, permRepo, fieldRepo)

	permService := permissionPkg.NewService(
		permRepo,
		wsRepo,
		enforcer,
	)

	// row.Service
	rowService := rowPkg.NewService(
		rowRepo,
		spreadsheetRepo,
		fieldRepo,
		enforcer,
	)

	// ─── Handlers ─────────────────────────────────────────────────────────────
	authHandler := authPkg.NewHandler(authService, l.Log)
	wsHandler := workspacePkg.NewHandler(wsService, l.Log)
	spreadsheetHandler := spreadsheetPkg.NewHandler(spreadsheetSvc, l.Log)
	fieldHandler := fieldPkg.NewHandler(fieldService, l.Log)
	rowHandler := rowPkg.NewHandler(rowService, l.Log)
	permHandler := permissionPkg.NewHandler(permService, l.Log)

	// ─── Router ───────────────────────────────────────────────────────────────
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.Timeout(30 * time.Second))
	r.Use(middleware.CORS())

	// ─── Public routes ────────────────────────────────────────────────────────
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	// ─── Protected routes ─────────────────────────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtManager))

		// Auth
		r.Get("/auth/me", authHandler.Me)

		// ── Workspaces ────────────────────────────────────────────────────────
		r.Get("/workspaces", wsHandler.List)
		r.Post("/workspaces", wsHandler.Create)
		r.Get("/workspaces/{id}", wsHandler.Get)

		// ── Spreadsheets ──────────────────────────────────────────────────────
		r.Post("/spreadsheets", spreadsheetHandler.Create)
		r.Get("/workspaces/{workspaceId}/spreadsheets", spreadsheetHandler.ListByWorkspace)
		r.Get("/spreadsheets/{id}", spreadsheetHandler.Get)
		r.Put("/spreadsheets/{id}", spreadsheetHandler.Update)
		r.Delete("/spreadsheets/{id}", spreadsheetHandler.Delete)

		// ── Fields ────────────────────────────────────────────────────────────
		r.Post("/spreadsheets/{id}/fields", fieldHandler.Create)
		r.Get("/spreadsheets/{id}/fields", fieldHandler.ListBySpreadsheet)
		r.Put("/fields/{fieldId}", fieldHandler.Update)
		r.Delete("/fields/{fieldId}", fieldHandler.Delete)

		// ── Rows ──────────────────────────────────────────────────────────────
		r.Post("/spreadsheets/{id}/rows/query", rowHandler.Query)
		r.Post("/spreadsheets/{id}/rows", rowHandler.Create)
		r.Patch("/spreadsheets/{id}/rows/{rowId}", rowHandler.Update)
		r.Delete("/spreadsheets/{id}/rows/{rowId}", rowHandler.Delete)
		r.Delete("/spreadsheets/{id}/rows", rowHandler.BulkDelete)

		// ── Permissions — spreadsheet level ───────────────────────────────────
		r.Get("/spreadsheets/{id}/permissions", permHandler.GetSpreadsheetAccess)
		r.Put("/spreadsheets/{id}/permissions", permHandler.UpsertSpreadsheetAccess)
		r.Delete("/spreadsheets/{id}/permissions/{principalId}", permHandler.RemoveSpreadsheetAccess)

		// ── Permissions — field level ─────────────────────────────────────────
		r.Get("/spreadsheets/{id}/field-permissions", permHandler.GetFieldAccess)
		r.Put("/spreadsheets/{id}/field-permissions/{fieldId}", permHandler.UpsertFieldAccess)

		// ── Permissions — row rules ────────────────────────────────────────────
		r.Get("/spreadsheets/{id}/row-rules", permHandler.GetRowRules)
		r.Put("/spreadsheets/{id}/row-rules", permHandler.UpsertRowRule)
		r.Delete("/spreadsheets/{id}/row-rules/{ruleId}", permHandler.DeleteRowRule)

		// ── Permissions — my permissions ──────────────────────────────────────
		r.Get("/spreadsheets/{id}/my-permissions", permHandler.GetMyPermissions)
	})

	// ─── HTTP Server ──────────────────────────────────────────────────────────
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ─── Graceful shutdown ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		l.Log.Info("server started", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-quit
	l.Log.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	l.Log.Info("server stopped")
}

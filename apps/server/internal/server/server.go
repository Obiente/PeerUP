package server

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	monitorv1connect "github.com/Obiente/Uppe/apps/server/gen/monitor/v1/monitorv1connect"
	networkv1connect "github.com/Obiente/Uppe/apps/server/gen/network/v1/networkv1connect"
	resultv1connect "github.com/Obiente/Uppe/apps/server/gen/result/v1/resultv1connect"
	settingsv1connect "github.com/Obiente/Uppe/apps/server/gen/settings/v1/settingsv1connect"
	statuspagev1connect "github.com/Obiente/Uppe/apps/server/gen/statuspage/v1/statuspagev1connect"
	"github.com/Obiente/Uppe/apps/server/internal/config"
	monitorservice "github.com/Obiente/Uppe/apps/server/internal/connect/monitor"
	networkservice "github.com/Obiente/Uppe/apps/server/internal/connect/network"
	resultservice "github.com/Obiente/Uppe/apps/server/internal/connect/result"
	settingsservice "github.com/Obiente/Uppe/apps/server/internal/connect/settings"
	statuspageservice "github.com/Obiente/Uppe/apps/server/internal/connect/statuspage"
	"github.com/Obiente/Uppe/apps/server/internal/db"
)

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
	database   db.Database
}

// CORS middleware to allow requests from the frontend
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from localhost during development
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms")
		w.Header().Set("Access-Control-Expose-Headers", "Connect-Protocol-Version")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func New(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	// Initialize database
	database, err := db.NewDatabase(&cfg.Database)
	if err != nil {
		return nil, err
	}

	// Test database connection
	if err := database.Ping(context.Background()); err != nil {
		return nil, err
	}

	// Create HTTP mux
	mux := http.NewServeMux()

	// Create service handlers with database
	monitorService := monitorservice.NewMonitorService(logger, database)
	networkService := networkservice.NewNetworkService(logger, database)
	resultService := resultservice.NewResultService(logger, database)
	settingsService := settingsservice.NewSettingsService(logger, database)
	statusPageService := statuspageservice.NewStatusPageService(logger, database)

	// Register ConnectRPC handlers
	monitorPath, monitorHandler := monitorv1connect.NewMonitorServiceHandler(monitorService)
	networkPath, networkHandler := networkv1connect.NewNetworkServiceHandler(networkService)
	resultPath, resultHandler := resultv1connect.NewResultServiceHandler(resultService)
	settingsPath, settingsHandler := settingsv1connect.NewSettingsServiceHandler(settingsService)
	statusPagePath, statusPageHandler := statuspagev1connect.NewStatusPageServiceHandler(statusPageService)

	mux.Handle(monitorPath, monitorHandler)
	mux.Handle(networkPath, networkHandler)
	mux.Handle(resultPath, resultHandler)
	mux.Handle(settingsPath, settingsHandler)
	mux.Handle(statusPagePath, statusPageHandler)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create HTTP server with h2c (HTTP/2 Cleartext) for gRPC
	httpServer := &http.Server{
		Addr:         cfg.ServerAddress(),
		Handler:      corsMiddleware(h2c.NewHandler(mux, &http2.Server{})),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
		database:   database,
	}, nil
}

func (s *Server) Start() error {
	s.logger.Info("Server starting",
		zap.String("address", s.httpServer.Addr),
	)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Server shutting down")
	if err := s.database.Close(); err != nil {
		s.logger.Error("Failed to close database", zap.Error(err))
	}
	return s.httpServer.Shutdown(ctx)
}

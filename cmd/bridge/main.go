package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	bmwcardata "github.com/tjamet/bmw-cardata"
	"github.com/tjamet/bmw-cardata/cardataapi"

	"github.com/wkulhane/bmw-loxone-bridge/internal/bridge"
	"github.com/wkulhane/bmw-loxone-bridge/internal/config"
	"github.com/wkulhane/bmw-loxone-bridge/internal/handler"
	"github.com/wkulhane/bmw-loxone-bridge/internal/store"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--healthcheck" {
		healthcheck()
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.LogLevel)

	os.MkdirAll(cfg.DataDir, 0700)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	sessionStore, err := bmwcardata.NewFileSessionStore(cfg.SessionPath())
	if err != nil {
		logger.Error("failed to create session store", "error", err)
		os.Exit(1)
	}

	auth, err := bmwcardata.NewAuthenticator(
		bmwcardata.WithClientID(cfg.ClientID),
		bmwcardata.WithSessionStore(sessionStore),
		bmwcardata.WithPromptURI(func(verificationURI, userCode string) {
			fmt.Println("=== BMW CarData Authentication ===")
			fmt.Printf("Open:  %s\n", verificationURI)
			fmt.Printf("Code:  %s\n", userCode)
			fmt.Println("Waiting for authentication...")
		}),
	)
	if err != nil {
		logger.Error("failed to create authenticator", "error", err)
		os.Exit(1)
	}

	logger.Info("authenticating with BMW CarData...")
	_, err = auth.GetSession(ctx)
	if err != nil {
		logger.Error("authentication failed", "error", err)
		os.Exit(1)
	}
	logger.Info("authentication successful")

	client, err := bmwcardata.NewClient(bmwcardata.WithAuthenticator(auth))
	if err != nil {
		logger.Error("failed to create client", "error", err)
		os.Exit(1)
	}

	dataStore := store.New()
	b := bridge.New(dataStore, logger)

	containerID, err := ensureContainer(ctx, client, logger)
	if err != nil {
		logger.Warn("failed to set up container for REST API", "error", err)
	} else {
		fetchTelematicData(ctx, client, b, cfg.VIN, containerID, logger)
	}

	var mqttConnected atomic.Bool
	mqttConnected.Store(true)

	sub, err := client.Subscribe(ctx, cfg.VIN, b.HandleMessage)
	if err != nil {
		logger.Error("failed to subscribe to vehicle stream", "error", err)
		os.Exit(1)
	}
	logger.Info("subscribed to vehicle stream", "vin", cfg.VIN, "subscription", sub.ID)

	client.StartEventStream()
	logger.Info("MQTT event stream started")

	if containerID != "" && cfg.RefreshInterval > 0 {
		initialInterval := cfg.RefreshInterval
		if isCharging(dataStore) {
			initialInterval = cfg.ActiveRefreshInterval
			logger.Info("charging detected, using fast refresh", "interval", initialInterval.String())
		}
		go func() {
			interval := initialInterval
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					fetchTelematicData(ctx, client, b, cfg.VIN, containerID, logger)
					newInterval := cfg.RefreshInterval
					if isCharging(dataStore) {
						newInterval = cfg.ActiveRefreshInterval
					}
					if newInterval != interval {
						interval = newInterval
						ticker.Reset(interval)
						logger.Info("refresh interval adjusted", "interval", interval.String(), "charging", interval == cfg.ActiveRefreshInterval)
					}
				}
			}
		}()
		logger.Info("periodic REST refresh enabled",
			"idle_interval", cfg.RefreshInterval.String(),
			"charging_interval", cfg.ActiveRefreshInterval.String(),
		)
	}

	h := handler.New(dataStore, func() bool { return mqttConnected.Load() })
	if containerID != "" {
		h.RefreshFunc = func() {
			fetchTelematicData(ctx, client, b, cfg.VIN, containerID, logger)
		}
	}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("HTTP server starting", "addr", cfg.ListenAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
			cancel()
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case <-client.Done():
		logger.Warn("MQTT client disconnected unexpectedly")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	client.StopEventStream()
	server.Shutdown(shutdownCtx)
	logger.Info("shutdown complete")
}

const containerName = "bmw-loxone-bridge"

func ensureContainer(ctx context.Context, client *bmwcardata.Client, logger *slog.Logger) (string, error) {
	list, err := client.ListContainers(ctx)
	if err != nil {
		return "", fmt.Errorf("list containers: %w", err)
	}
	if list != nil && list.Containers != nil {
		for _, c := range *list.Containers {
			if c.Name != nil && *c.Name == containerName && c.ContainerId != nil {
				if c.State != nil && *c.State == cardataapi.ContainerDtoStateACTIVE {
					logger.Info("using existing container", "id", *c.ContainerId)
					return *c.ContainerId, nil
				}
			}
		}
	}

	logger.Info("creating new container for REST API data fetch")
	descriptors := bmwcardata.FindDescriptors(
		bmwcardata.MatchAll(
			bmwcardata.MatchBrand(bmwcardata.BrandBMW),
			bmwcardata.MatchVehicleType(bmwcardata.VehicleTypePHEV),
		),
	)
	resp, err := client.CreateContainer(ctx, containerName, "Loxone bridge data", descriptors)
	if err == nil && resp != nil && resp.ContainerId != nil {
		logger.Info("created container", "id", *resp.ContainerId)
		return *resp.ContainerId, nil
	}
	if err != nil {
		logger.Debug("CreateContainer returned error (may be false negative)", "error", err)
	}

	// List again to find the container (may have been created despite the error)
	list, err = client.ListContainers(ctx)
	if err != nil {
		return "", fmt.Errorf("list containers after create: %w", err)
	}
	if list != nil && list.Containers != nil {
		for _, c := range *list.Containers {
			if c.Name != nil && *c.Name == containerName && c.ContainerId != nil {
				if c.State != nil && *c.State == cardataapi.ContainerDtoStateACTIVE {
					logger.Info("found container after creation", "id", *c.ContainerId)
					return *c.ContainerId, nil
				}
			}
		}
	}
	return "", fmt.Errorf("could not find or create container")
}

func fetchTelematicData(ctx context.Context, client *bmwcardata.Client, b *bridge.Bridge, vin, containerID string, logger *slog.Logger) {
	logger.Info("fetching telematic data via REST API")
	resp, err := client.GetTelematicData(ctx, vin, containerID)
	if err != nil {
		logger.Warn("failed to fetch telematic data", "error", err)
		return
	}
	if resp.TelematicData == nil {
		logger.Warn("telematic data response was empty")
		return
	}
	b.HandleTelematicData(*resp.TelematicData)
	logger.Info("telematic data loaded", "data_points", len(*resp.TelematicData))
}

func isCharging(s *store.Store) bool {
	dp, ok := s.Get("charging_status")
	if !ok {
		return false
	}
	return strings.Contains(strings.ToUpper(dp.Value), "CHARGING") &&
		!strings.Contains(strings.ToUpper(dp.Value), "NOT") &&
		!strings.Contains(strings.ToUpper(dp.Value), "ENDED")
}

func setupLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}

func healthcheck() {
	resp, err := http.Get("http://localhost:8400/api/health")
	if err != nil {
		fmt.Fprintf(os.Stderr, "health check failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "health check returned status %d\n", resp.StatusCode)
		os.Exit(1)
	}
	os.Exit(0)
}

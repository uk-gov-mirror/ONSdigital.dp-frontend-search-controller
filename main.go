package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/renderer"
	search "github.com/ONSdigital/dp-api-clients-go/site-search"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"

	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/routes"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

const serviceName = "dp-frontend-search-controller"

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

func main() {
	log.Namespace = serviceName
	cfg, err := config.Get()
	ctx := context.Background()
	if err != nil {
		log.Event(ctx, "unable to retrieve service configuration", log.Error(err))
		os.Exit(1)
	}

	log.Event(ctx, "got service configuration", log.Data{"config": cfg})

	versionInfo, err := health.NewVersionInfo(
		BuildTime,
		GitCommit,
		Version,
	)

	r := mux.NewRouter()

	clients := routes.Clients{
		Renderer: renderer.New(cfg.RendererURL),
		Search:   search.NewClient(cfg.SearchAPIURL),
	}

	healthcheck := health.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	if err = registerCheckers(ctx, &healthcheck, clients); err != nil {
		os.Exit(1)
	}
	routes.Setup(ctx, r, cfg, healthcheck, clients)

	healthcheck.Start(ctx)

	s := server.New(cfg.BindAddr, r)
	s.HandleOSSignals = false

	log.Event(ctx, "Starting server", log.Data{"config": cfg})

	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Event(ctx, "failed to start http listen and serve", log.Error(err))
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	log.Event(ctx, "shutting service down gracefully")
	defer cancel()
	if err := s.Server.Shutdown(ctx); err != nil {
		log.Event(ctx, "failed to shutdown http server", log.Error(err))
	}
}

func registerCheckers(ctx context.Context, h *health.HealthCheck, c routes.Clients) (err error) {
	if err = h.AddCheck("frontend renderer", c.Renderer.Checker); err != nil {
		log.Event(ctx, "failed to add frontend renderer checker", log.Error(err))
	}

	if err = h.AddCheck("Search API", c.Search.Checker); err != nil {
		log.Event(ctx, "failed to add search API checker", log.Error(err))
	}
	return
}

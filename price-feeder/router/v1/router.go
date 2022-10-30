package v1

import (
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/CosmosContracts/juno/price-feeder/config"
	"github.com/CosmosContracts/juno/price-feeder/pkg/httputil"
	"github.com/CosmosContracts/juno/price-feeder/router/middleware"
)

const (
	APIPathPrefix = "/api/v1"
)

// Router defines a router wrapper used for registering v1 API routes.
type Router struct {
	logger  zerolog.Logger
	cfg     config.Config
	oracle  Oracle
	metrics Metrics
}

func New(logger zerolog.Logger, cfg config.Config, oracle Oracle, metrics Metrics) *Router {
	return &Router{
		logger:  logger.With().Str("module", "router").Logger(),
		cfg:     cfg,
		oracle:  oracle,
		metrics: metrics,
	}
}

// RegisterRoutes register v1 API routes on the provided sub-router.
func (r *Router) RegisterRoutes(rtr *mux.Router, prefix string) {
	v1Router := rtr.PathPrefix(prefix).Subrouter()

	// build middleware chain
	mChain := middleware.Build(r.logger, r.cfg)

	// handle all preflight request
	v1Router.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for _, origin := range r.cfg.Server.AllowedOrigins {
			if origin == req.Header.Get("Origin") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With",
		)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.WriteHeader(http.StatusOK)
	})

	v1Router.Handle(
		"/healthz",
		mChain.ThenFunc(r.healthzHandler()),
	).Methods(httputil.MethodGET)

	v1Router.Handle(
		"/prices",
		mChain.ThenFunc(r.pricesHandler()),
	).Methods(httputil.MethodGET)

	v1Router.Handle(
		"/prices/providers/tvwap",
		mChain.ThenFunc(r.candlePricesHandler()),
	).Methods(httputil.MethodGET)

	v1Router.Handle(
		"/prices/providers/vwap",
		mChain.ThenFunc(r.tickerPricesHandler()),
	).Methods(httputil.MethodGET)

	if r.cfg.Telemetry.Enabled {
		v1Router.Handle(
			"/metrics",
			mChain.ThenFunc(r.metricsHandler()),
		).Methods(httputil.MethodGET)
	}
}

func (r *Router) healthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		resp := HealthZResponse{
			Status: StatusAvailable,
		}

		resp.Oracle.LastSync = r.oracle.GetLastPriceSyncTimestamp().Format(time.RFC3339)

		httputil.RespondWithJSON(w, http.StatusOK, resp)
	}
}

func (r *Router) pricesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		resp := PricesResponse{
			Prices: r.oracle.GetPrices(),
		}

		httputil.RespondWithJSON(w, http.StatusOK, resp)
	}
}

func (r *Router) candlePricesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		resp := PricesPerProviderResponse{
			Prices: r.oracle.GetTvwapPrices(),
		}

		httputil.RespondWithJSON(w, http.StatusOK, resp)
	}
}

func (r *Router) tickerPricesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		resp := PricesPerProviderResponse{
			Prices: r.oracle.GetVwapPrices(),
		}

		httputil.RespondWithJSON(w, http.StatusOK, resp)
	}
}

func (r *Router) metricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		format := strings.TrimSpace(req.FormValue("format"))

		gr, err := r.metrics.Gather(format)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to gather metrics: %s", err))
			return
		}

		w.Header().Set("Content-Type", gr.ContentType)
		if t, err := template.New("metrics").Parse(string(gr.Metrics)); err == nil {
			// unchecked err, too late for bad response
			_ = t.ExecuteTemplate(w, "metrics", nil)
		} else {
			writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to parse metrics: %s", err))
		}
	}
}

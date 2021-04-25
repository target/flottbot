package core

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/target/flottbot/models"
)

var promRouter *mux.Router

var (
	botResponseCollector = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flottbot_ruleCount",
			Help: "Total No. of bot rules triggered",
		},
		[]string{"rulename"},
	)
)

// Prommetric creates a local Prometheus server to rule metrics
func Prommetric(input string, bot *models.Bot) {
	if bot.Metrics {
		if input == "init" {
			// init router
			promRouter = mux.NewRouter()

			// metrics health check handler
			promHealthHandle := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					bot.Log.Error().Msgf("Prometheus Server: invalid method %s", r.Method)
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				bot.Log.Info().Msg("Prometheus Server: health check hit!")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}
			promRouter.HandleFunc("/metrics_health", promHealthHandle).Methods("GET")

			// metrics handler
			prometheus.MustRegister(botResponseCollector)
			promRouter.Handle("/metrics", promhttp.Handler())

			// start prometheus server
			go http.ListenAndServe(":8080", promRouter)
			bot.Log.Info().Msg("Prometheus Server: serving metrics at /metrics")
		} else {
			botResponseCollector.With(prometheus.Labels{"rulename": input}).Inc()
		}
	}
}

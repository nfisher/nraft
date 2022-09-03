package server

import (
	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"log"
	"net/http"
)

func Mux(r *Raft) http.Handler {
	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		log.Fatalf("Failed to register server views for HTTP metrics: %v", err)
	}

	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "raft_rpc",
	})
	if err != nil {
		panic(err)
	}

	view.RegisterExporter(pe)

	mux := http.NewServeMux()
	mux.HandleFunc("/request_vote", requestVote(r))
	mux.HandleFunc("/append_entries", appendEntries(r))
	mux.Handle("/metrics", pe)
	och := &ochttp.Handler{
		Handler: mux, // The handler you'd have used originally
	}
	return och
}

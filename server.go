package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

func routes() http.Handler {
	router := chi.NewRouter()

	fs := http.FileServer(http.Dir("./static"))
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/summary", homeHandler)

	router.Route("/sessions", func(r chi.Router) {
		r.Get("/", endSessionHandler)
		r.Get("/{activity_name}", startSessionHandler)
	})

	return router

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// render the home page
	homepageBytes, err := renderHomepage()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(homepageBytes)
}

func activityChartHandler(w http.ResponseWriter, r *http.Request) {
	// parses the year from the query paramater
	// computes the chart data for that year
	// writes the rendered activity_chart.html to w
}

func startSessionHandler(w http.ResponseWriter, r *http.Request) {
	activity := chi.URLParam(r, "activity_name")
	err := startSession(activity)
	if err != nil {
		if err.Error() == ErrStartSession {
			log.Println(ErrStartSession)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	err = updateTemplateData(false)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func endSessionHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sessionAction := query.Get("action")
	switch sessionAction {
	case "end":
		_, err := endCurrentActiveSession()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// recompute and update the todays sessions data
		err = updateTemplateData(true)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func serve() error {
	srv := &http.Server{
		Addr:    ":4000",
		Handler: routes(),
	}
	shutdownErr := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		log.Printf("caught signal %s", s.String())
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		shutdownErr <- srv.Shutdown(ctx)
	}()

	log.Printf("starting server addr: %s\n", srv.Addr)
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdownErr
	if err != nil {
		return err
	}
	return nil
}

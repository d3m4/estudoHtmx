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
)

const listenAddr = ":5002"

func main() {
	db, err := openDB()
	if err != nil {
		log.Fatalf("erro abrindo banco: %v", err)
	}
	defer db.Close()

	srv, err := newServer(db)
	if err != nil {
		log.Fatalf("erro inicializando server: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", srv.handleIndex)
	mux.HandleFunc("GET /expenses", srv.handleListExpenses)
	mux.HandleFunc("GET /expenses/form", srv.handleFormFragment)
	mux.HandleFunc("GET /expenses/total-context", srv.handleTotalContext)
	mux.HandleFunc("POST /expenses", srv.handleSaveExpense)
	mux.HandleFunc("POST /expenses/{id}/zerar", srv.handleZerarExpense)

	httpServer := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// graceful shutdown via SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("estudohtmx-go ouvindo em http://localhost%s", listenAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("erro do servidor: %v", err)
		}
	}()

	<-stop
	log.Println("desligando...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("erro no shutdown: %v", err)
	}
}

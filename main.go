package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/TechBowl-japan/go-stations/db"
	"github.com/TechBowl-japan/go-stations/handler/router"
)

func main() {
	err := realMain()
	if err != nil {
		log.Fatalln("main: failed to exit successfully, err =", err)
	}
}

func realMain() error {
	//graceful shutdownのためのcontext
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	// config values
	const (
		defaultPort   = ":8080"
		defaultDBPath = ".sqlite3/todo.db"
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	// set time zone
	var err error
	time.Local, err = time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return err
	}

	// set up sqlite3
	todoDB, err := db.NewDB(dbPath)
	if err != nil {
		return err
	}
	defer todoDB.Close()

	// NOTE: 新しいエンドポイントの登録はrouter.NewRouterの内部で行うようにする
	mux := router.NewRouter(todoDB)

	// TODO: サーバーをlistenする
	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	log.Printf("Server is listening on port %v", port)
	select {
	case <-ctx.Done():
		log.Print("shutdown by signal")
	case err := <-errCh:
		return err
	}
	// ここでタイムアウトの時間設定のコンテキストを作る
	contextStop, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Print("Shutting down gracefully")
	//コンテキスト情報を使いシャットダウン
	if err := srv.Shutdown(contextStop); err != nil {
		log.Print("Shutdown(): ", err)
		return err
	}
	wg.Wait()
	return nil
}

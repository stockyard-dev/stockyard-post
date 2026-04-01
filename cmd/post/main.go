// Stockyard Post — Form backend.
// Collect form submissions, fire webhooks, redirect after submit. Self-hosted.
// Single binary, embedded SQLite, zero external dependencies.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/stockyard-dev/stockyard-post/internal/license"
	"github.com/stockyard-dev/stockyard-post/internal/server"
	"github.com/stockyard-dev/stockyard-post/internal/store"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Printf("post %s\n", version)
		os.Exit(0)
	}
	if len(os.Args) > 1 && (os.Args[1] == "--health" || os.Args[1] == "health") {
		fmt.Println("ok")
		os.Exit(0)
	}

	log.SetFlags(log.Ltime | log.Lshortfile)

	retentionDays := 30
	if r := os.Getenv("RETENTION_DAYS"); r != "" {
		if n, err := strconv.Atoi(r); err == nil && n > 0 {
			retentionDays = n
		}
	}

	port := 8830
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	licenseKey := os.Getenv("POST_LICENSE_KEY")
	licInfo, licErr := license.Validate(licenseKey, "post")
	if licenseKey != "" && licErr != nil {
		log.Printf("[license] WARNING: %v — running in free tier", licErr)
		licInfo = nil
	}
	limits := server.LimitsFor(licInfo)
	if licInfo != nil && licInfo.IsPro() {
		log.Printf("  License:   Pro (%s)", licInfo.CustomerID)
	} else {
		log.Printf("  License:   Free tier (set POST_LICENSE_KEY to unlock Pro)")
	}

	if limits.RetentionDays > retentionDays {
		retentionDays = limits.RetentionDays
	}

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	log.Printf("")
	log.Printf("  Stockyard Post %s", version)
	log.Printf("  Submit:         http://localhost:%d/f/{form_id}", port)
	log.Printf("  API:            http://localhost:%d/api/forms", port)
	log.Printf("  Retention:      %d days", retentionDays)
	log.Printf("  Dashboard:      http://localhost:%d/ui", port)
	log.Printf("")

	go func() {
		for {
			time.Sleep(6 * time.Hour)
			n, err := db.Cleanup(retentionDays)
			if err != nil {
				log.Printf("[cleanup] error: %v", err)
			} else if n > 0 {
				log.Printf("[cleanup] deleted %d submissions older than %d days", n, retentionDays)
			}
		}
	}()

	srv := server.New(db, port, limits)
	if err := srv.Start(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

package main

import (
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/jparrill/michishirube/internal/data"
	"github.com/jparrill/michishirube/internal/ui"
)

func main() {
	// Initialize the application
	a := app.New()

	// Set application icon
	icon, err := fyne.LoadResourceFromPath("assets/michishirube-logo.png")
	if err != nil {
		log.Printf("Warning: Failed to load application icon: %v", err)
	} else {
		a.SetIcon(icon)
	}

	w := a.NewWindow("Michishirube - Link Organizer")

	// Set a reasonable default size
	w.Resize(fyne.NewSize(900, 600))

	// Initialize the database
	db, err := data.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Check if we should load fixtures
	if shouldLoadFixtures() {
		log.Println("Loading fixtures requested via environment variable")
		if err := data.LoadFixtures(db); err != nil {
			log.Printf("Warning: Failed to load fixtures: %v", err)
		}
	}

	// Create the UI
	ui := ui.NewUI(w, db)

	// Set the window content and show
	w.SetContent(ui.Content())
	w.ShowAndRun()
}

// shouldLoadFixtures checks if the LOAD_FIXTURES environment variable is set to true
func shouldLoadFixtures() bool {
	loadFixtures := os.Getenv("LOAD_FIXTURES")
	return strings.ToLower(loadFixtures) == "true"
}

package main

import (
	"log"

	"github.com/bonefabric/htop/internal/ui"
)

func main() {
	dashboard, err := ui.NewDashboard()
	if err != nil {
		log.Fatalf("Failed to create dashboard: %v", err)
	}

	if err := dashboard.Run(); err != nil {
		log.Fatalf("Error running dashboard: %v", err)
	}
} 
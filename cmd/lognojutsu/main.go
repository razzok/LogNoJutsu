package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"lognojutsu/internal/server"
)

const banner = `
 _                    _   _       _       _
| |    ___   __ _  _ | \ | | ___ | |_   _| |_ ___ _   _
| |   / _ \ / _` + "`" + ` ||  _||  \| |/ _ \| | | | | __/ __| | | |
| |__| (_) | (_| || |  | |\  | (_) | | |_| | |_\__ \ |_| |
|_____\___/ \__, ||_|  |_| \_|\___/| |\__,_|\__|___/\__,_|
            |___/                  |_|
  SIEM Validation & ATT&CK Simulation Tool  v0.1.0
`

func main() {
	host := flag.String("host", "127.0.0.1", "Bind address (use 0.0.0.0 for network access)")
	port := flag.Int("port", 8080, "HTTP port")
	password := flag.String("password", "", "Optional UI password (empty = no auth)")
	flag.Parse()

	fmt.Println(banner)

	cfg := server.Config{
		Host:     *host,
		Port:     *port,
		Password: *password,
	}

	log.Printf("Starting LogNoJutsu on http://%s:%d", cfg.Host, cfg.Port)
	if cfg.Password != "" {
		log.Printf("Password protection enabled")
	}
	if cfg.Host == "0.0.0.0" {
		log.Printf("WARNING: UI accessible from network — ensure firewall rules are in place")
	}

	if err := server.Start(cfg); err != nil {
		log.Fatalf("Server error: %v", err)
		os.Exit(1)
	}
}

package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/RoaycL/emby-scraper-plus/config"
	"github.com/RoaycL/emby-scraper-plus/internal/router"
	"github.com/RoaycL/emby-scraper-plus/internal/server"
)

func main() {
	port := flag.Int("port", config.GlobalConfig.Port, "Port to listen on")
	bind := flag.String("bind", config.GlobalConfig.Host, "Bind address")
	token := flag.String("token", config.GlobalConfig.Token, "Access token for API authentication")
	proxy := flag.String("proxy", config.GlobalConfig.Proxy, "HTTP/SOCKS5 outbound proxy address")
	flag.Parse()

	config.GlobalConfig.Port = *port
	if *bind != "" {
		config.GlobalConfig.Host = *bind
	}
	if *token != "" {
		config.GlobalConfig.Token = *token
	}
	if *proxy != "" {
		config.GlobalConfig.Proxy = *proxy
	}

	banner := `
  ______           _             _____                                 _____  _             
 |  ____|         | |           / ____|                               |  __ \| |            
 | |__   _ __ ___ | |__  _   _ | (___   ___ _ __ __ _ _ __   ___ _ __ | |__) | |_   _ ___  
 |  __| | '_ ` + "`" + ` _ \| '_ \| | | | \___ \ / __| '__/ _` + "`" + ` | '_ \ / _ \ '__||  ___/| | | | / __| 
 | |____| | | | | | |_) | |_| | ____) | (__| | | (_| | |_) |  __/ |   | |    | | |_| \__ \ 
 |______|_| |_| |_|_.__/ \__, ||_____/ \___|_|  \__,_| .__/ \___|_|   |_|    |_|\__,_|___/ 
                          __/ |                      | |                                    
                         |___/                       |_|                                    
`
	fmt.Println(banner)
	log.Printf("[Emby Scraper Plus] Starting server on http://%s:%d ...\n", config.GlobalConfig.Host, config.GlobalConfig.Port)
	if config.GlobalConfig.Token != "" {
		log.Printf("[Emby Scraper Plus] Token authentication enabled\n")
	}
	if config.GlobalConfig.Proxy != "" {
		log.Printf("[Emby Scraper Plus] Outbound proxy configured: %s\n", config.GlobalConfig.Proxy)
	}

	engine := router.NewEngine()
	if err := server.StartServer(engine); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}

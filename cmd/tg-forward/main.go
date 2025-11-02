package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gabrielmelo/tg-forward/internal/api"
	"github.com/gabrielmelo/tg-forward/internal/config"
	"github.com/gabrielmelo/tg-forward/internal/matcher"
	"github.com/gabrielmelo/tg-forward/internal/rules"
	"github.com/gabrielmelo/tg-forward/internal/telegram"
	"github.com/gotd/td/tg"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	log.Println("Starting Telegram Message Forwarder...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	apiPort := cfg.API.Port
	if apiPort == "" {
		apiPort = "8080"
	}

	ctx := context.Background()

	db, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDB.URI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	if err := db.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	rulesRepo, err := rules.NewRepository(
		db,
		cfg.MongoDB.Database,
		"rules",
	)
	if err != nil {
		log.Fatalf("Failed to initialize rules repository: %v", err)
	}
	defer rulesRepo.Close()

	patterns, err := rulesRepo.GetPatterns()
	if err != nil {
		log.Fatalf("Failed to get patterns: %v", err)
	}

	m, err := matcher.New(patterns)
	if err != nil {
		log.Fatalf("Failed to initialize matcher: %v", err)
	}

	rulesService := rules.NewService(rulesRepo, m)

	bot, err := telegram.NewBot(
		cfg.Telegram.Bot.Token,
		cfg.Telegram.Bot.TargetChatID,
		cfg.Telegram.Bot.TargetUsername,
	)
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	apiServer := api.NewServer(rulesService, apiPort, cfg.API.Token)

	var mu sync.RWMutex
	messageHandler := func(ctx context.Context, msg *tg.Message) error {
		text := extractMessageText(msg)
		if text == "" {
			return nil
		}

		mu.RLock()
		currentMatcher := apiServer.GetMatcher()
		mu.RUnlock()

		if currentMatcher.Match(text) {
			log.Printf("Message matched pattern, forwarding: %s", text)
			if err := bot.ForwardMessage(text); err != nil {
				log.Printf("Failed to forward message: %v", err)
				return err
			}
		}

		return nil
	}

	client := telegram.NewClient(
		cfg.Telegram.User.AppID,
		cfg.Telegram.User.AppHash,
		cfg.Telegram.User.Phone,
		messageHandler,
		bot.GetBotID(),
		cfg.Telegram.User.Session,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting API server...")
		if err := apiServer.Start(); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting user client...")
		if err := client.Run(ctx, cfg.Telegram.User.AppID, cfg.Telegram.User.AppHash); err != nil {
			log.Printf("Client error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down API server: %v", err)
	}

	wg.Wait()
	log.Println("Shutdown complete")
}

func extractMessageText(msg *tg.Message) string {
	if msg == nil {
		return ""
	}
	return msg.Message
}

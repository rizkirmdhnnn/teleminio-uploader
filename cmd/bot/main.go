package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	"github.com/pkg/errors"
	"github.com/rizkirmdhnnn/teleminio-uploader/internal/client"
	"github.com/rizkirmdhnnn/teleminio-uploader/internal/config"
	"github.com/rizkirmdhnnn/teleminio-uploader/internal/handler"
	store "github.com/rizkirmdhnnn/teleminio-uploader/internal/storage"
	"github.com/rizkirmdhnnn/teleminio-uploader/internal/utils"
)

func run(ctx context.Context) error {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize storage
	s, err := store.NewStorage(cfg.Phone)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize Minio
	minio, err := store.NewMinio(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize MinIO: %w", err)
	}

	// Initialize logger
	logger := config.LoadLogger(s.SessionDir)

	// Initialize Telegram client
	clientSetup, err := client.NewClient(cfg.AppID, cfg.AppHash, s, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize Telegram client: %w", err)
	}

	// Initialize media downloader
	mediaDir := filepath.Join(s.SessionDir, "media")
	downloader := utils.NewMediaDownloader(clientSetup.API, mediaDir)

	// Initialize message handler
	messageHandler := handler.NewMessageHandler(downloader, minio, s.PeerDB, cfg.UserTarget)

	// Handle new messages
	clientSetup.Dispatcher.OnNewMessage(messageHandler.HandleNewMessage)

	// Run the client
	return clientSetup.Waiter.Run(ctx, func(ctx context.Context) error {
		return start(ctx, cfg, clientSetup)
	})
}

func start(ctx context.Context, cfg config.Config, clientSetup *client.Setup) error {
	flow := auth.NewFlow(examples.Terminal{PhoneNumber: cfg.Phone}, auth.SendCodeOptions{})
	err := clientSetup.Client.Run(ctx, func(ctx context.Context) error {
		// Authenticate if necessary
		if err := clientSetup.Client.Auth().IfNecessary(ctx, flow); err != nil {
			return errors.Wrap(err, "auth")
		}

		// Get self info
		self, err := clientSetup.Client.Self(ctx)
		if err != nil {
			return errors.Wrap(err, "get self")
		}

		// Display user info
		name := self.FirstName
		if self.Username != "" {
			name = fmt.Sprintf("%s (@%s)", name, self.Username)
		}
		fmt.Println("Current user:", name)

		// // Fill peer storage
		// fmt.Println("Filling peer storage from dialogs to cache entities")
		// collector := storage.CollectPeers(s.PeerDB)
		// if err := collector.Dialogs(ctx, query.GetDialogs(clientSetup.API).Iter()); err != nil {
		// 	return errors.Wrap(err, "collect peers")
		// }
		// fmt.Println("Filled")

		// Start listening for updates
		fmt.Println("Listening for updates. Interrupt (Ctrl+C) to stop.")
		return clientSetup.UpdatesManager.Run(ctx, clientSetup.API, self.ID, updates.AuthOptions{
			IsBot: self.Bot,
			OnStart: func(ctx context.Context) {
				fmt.Println("Update recovery initialized and started, listening for events")
			},
		})
	})

	if err != nil {
		return errors.Wrap(err, "run client")
	}
	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		if errors.Is(err, context.Canceled) && ctx.Err() == context.Canceled {
			fmt.Println("Application stopped")
			os.Exit(0)
		}
		fmt.Println("Application error:", err)
		os.Exit(1)
	} else {
		fmt.Println("Application stopped")
		os.Exit(0)
	}
}

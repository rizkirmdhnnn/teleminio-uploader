package client

import (
	"context"
	"strconv"
	"time"

	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/contrib/storage"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	store "github.com/rizkirmdhnnn/teleminio-uploader/internal/storage"
)

// Setup contains the initialized Telegram client and related components
type Setup struct {
	Client         *telegram.Client
	API            *tg.Client
	Dispatcher     *tg.UpdateDispatcher
	UpdatesManager *updates.Manager
	Resolver       peer.Resolver
	Waiter         *floodwait.Waiter
	Sender         *message.Sender
}

// NewClient initializes a new Telegram client with all necessary components
func NewClient(appID string, appHash string, storageSetup *store.Setup, logger *zap.Logger) (*Setup, error) {
	// Create update dispatcher
	dispatcher := tg.NewUpdateDispatcher()

	// Set up update handler with peer storage
	updateHandler := storage.UpdateHook(dispatcher, storageSetup.PeerDB)

	// Set up updates manager for recovery
	updatesManager := updates.New(updates.Config{
		Handler: updateHandler,
		Logger:  logger.Named("updates.recovery"),
		Storage: storageSetup.StateStorage,
	})

	// Set up flood wait handler
	waiter := floodwait.NewWaiter().WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		logger.Warn("Flood wait", zap.Duration("wait", wait.Duration))
	})

	// Configure client options
	options := telegram.Options{
		Logger:         logger,
		SessionStorage: storageSetup.SessionStorage,
		UpdateHandler:  updatesManager,
		Middlewares: []telegram.Middleware{
			waiter,
			ratelimit.New(rate.Every(time.Millisecond*100), 5),
		},
	}

	// Convert appID to int
	appIDInt, err := strconv.Atoi(appID)
	if err != nil {
		return nil, err
	}

	// Create the client
	client := telegram.NewClient(appIDInt, appHash, options)

	// Get API client
	api := client.API()

	// Create resolver with peer storage
	resolver := storage.NewResolverCache(peer.Plain(api), storageSetup.PeerDB)

	// Create sender
	sender := message.NewSender(api)

	return &Setup{
		Client:         client,
		API:            api,
		Dispatcher:     &dispatcher,
		UpdatesManager: updatesManager,
		Resolver:       resolver,
		Waiter:         waiter,
		Sender:         sender,
	}, nil
}

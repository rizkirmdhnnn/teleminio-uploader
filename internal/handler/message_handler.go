package handler

import (
	"context"
	"fmt"
	"os"

	"slices"

	"github.com/gotd/contrib/storage"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	store "github.com/rizkirmdhnnn/teleminio-uploader/internal/storage"
	"github.com/rizkirmdhnnn/teleminio-uploader/internal/utils"
)

// MessageHandler handles incoming messages
type MessageHandler struct {
	Downloader *utils.MediaDownloader
	Minio      *store.MinioClient
	PeerDB     storage.PeerStorage
	UserTarget []string
	Sender     *message.Sender
	workerPool chan struct{}
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(downloader *utils.MediaDownloader, minio *store.MinioClient, peerDB storage.PeerStorage, sender *message.Sender, userTarget []string) *MessageHandler {
	return &MessageHandler{
		Downloader: downloader,
		Minio:      minio,
		PeerDB:     peerDB,
		UserTarget: userTarget,
		Sender:     sender,
		workerPool: make(chan struct{}, 15), // Allow 5 concurrent operations
	}
}

// HandleNewMessage processes new messages
func (h *MessageHandler) HandleNewMessage(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
	msg, ok := u.Message.(*tg.Message)
	if !ok {
		return nil
	}

	// Find peer information
	p, err := storage.FindPeer(ctx, h.PeerDB, msg.GetPeerID())
	if err != nil {
		return fmt.Errorf("find peer: %w", err)
	}

	// Check if user is in target list if target list is not empty
	if len(h.UserTarget) > 0 {
		isTarget := slices.Contains(h.UserTarget, p.User.Username)
		if !isTarget {
			return nil
		}
	}

	// Print message with formatted output
	fmt.Printf("Message from %s: %s\n", p.User.Username, msg.Message)

	// Process media if present
	if msg.Media != nil {
		// Acquire a worker from the pool
		h.workerPool <- struct{}{}

		// Process media concurrently
		go func() {
			defer func() {
				// Release the worker back to the pool
				<-h.workerPool
			}()

			err := h.handleMedia(ctx, msg.Media, p.User.Username)
			if err != nil {
				fmt.Printf("Error processing media from %s: %v\n", p.User.Username, err)
			}
		}()
	}

	return nil
}

// handleMedia processes media in messages
func (h *MessageHandler) handleMedia(ctx context.Context, media tg.MessageMediaClass, username string) error {
	fmt.Printf("Message contains media from %s\n", username)
	// Download the media
	path, ext, err := h.Downloader.DownloadMedia(ctx, media, username)
	if err != nil {
		return fmt.Errorf("download media: %w", err)
	}

	// Get file info for upload
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("get file info: %w", err)
	}

	// Open the file for reading
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// Upload to MinIO
	objectName := fmt.Sprintf("%s/%s/%s", username, ext, fileInfo.Name())
	url, err := h.Minio.UploadFile(ctx, objectName, file, fileInfo.Size(), ext)
	if err != nil {
		return fmt.Errorf("upload file: %w", err)
	}

	fmt.Printf("File uploaded to %s\n", url)
	h.Sender.Self().Text(ctx, fmt.Sprintf("File uploaded to %s", url))
	return nil
}

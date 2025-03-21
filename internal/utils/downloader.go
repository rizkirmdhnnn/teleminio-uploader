package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

// MediaDownloader handles downloading of media files
type MediaDownloader struct {
	MediaDir string
	API      *tg.Client
}

// NewMediaDownloader creates a new media downloader
func NewMediaDownloader(api *tg.Client, mediaDir string) *MediaDownloader {
	return &MediaDownloader{
		MediaDir: mediaDir,
		API:      api,
	}
}

// EnsureMediaDir ensures the media directory exists
func (m *MediaDownloader) EnsureMediaDir() error {
	return os.MkdirAll(m.MediaDir, 0755)
}

// DownloadPhoto downloads a photo from a message
func (m *MediaDownloader) DownloadPhoto(ctx context.Context, photo *tg.Photo, username string) (string, error) {
	mediaTypeDir := "photo"
	targetDir := filepath.Join(m.MediaDir, username, mediaTypeDir)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory structure: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	fileName := filepath.Join(targetDir, fmt.Sprintf("photo_%s.jpg", timestamp))

	// Get the largest photo size
	var largest *tg.PhotoSize
	for _, size := range photo.Sizes {
		if photoSize, ok := size.(*tg.PhotoSize); ok {
			if largest == nil || (photoSize.W*photoSize.H > largest.W*largest.H) {
				largest = photoSize
			}
		}
	}

	if largest != nil {
		loc := &tg.InputPhotoFileLocation{
			ID:            photo.ID,
			AccessHash:    photo.AccessHash,
			FileReference: photo.FileReference,
			ThumbSize:     largest.Type,
		}
		d := downloader.NewDownloader()
		_, err := d.Download(m.API, loc).ToPath(ctx, fileName)
		if err != nil {
			return "", fmt.Errorf("failed to download photo: %w", err)
		}
		return fileName, nil
	}

	return "", fmt.Errorf("no suitable photo size found")
}

// DownloadDocument downloads a document from a message
func (m *MediaDownloader) DownloadDocument(ctx context.Context, doc *tg.Document, username string) (string, error) {
	// Determine media type (video or document)
	mediaTypeDir := "document"
	for _, attr := range doc.Attributes {
		if _, ok := attr.(*tg.DocumentAttributeVideo); ok {
			mediaTypeDir = "video"
			break
		}
	}

	// Create directory structure: mediaDir/username/mediaType/YYYYMMDD/
	currentDate := time.Now().Format("20060102")
	targetDir := filepath.Join(m.MediaDir, username, mediaTypeDir, currentDate)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory structure: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	fileName := filepath.Join(targetDir, fmt.Sprintf("doc_%s", timestamp))

	// Get original filename if available
	for _, attr := range doc.Attributes {
		if fileAttr, ok := attr.(*tg.DocumentAttributeFilename); ok {
			fileName = filepath.Join(targetDir, fileAttr.FileName)
			break
		}
	}

	loc := doc.AsInputDocumentFileLocation()
	d := downloader.NewDownloader()
	_, err := d.Download(m.API, loc).ToPath(ctx, fileName)
	if err != nil {
		return "", fmt.Errorf("failed to download document: %w", err)
	}
	return fileName, nil
}

// DownloadMedia downloads media from a message
// return the path to the downloaded file, file extension and an error
func (m *MediaDownloader) DownloadMedia(ctx context.Context, media tg.MessageMediaClass, username string) (string, string, error) {
	switch med := media.(type) {
	case *tg.MessageMediaPhoto:
		photo := med.Photo.(*tg.Photo)
		file, err := m.DownloadPhoto(ctx, photo, username)
		return file, filepath.Ext(file), err
	case *tg.MessageMediaDocument:
		doc := med.Document.(*tg.Document)
		file, err := m.DownloadDocument(ctx, doc, username)
		return file, filepath.Ext(file), err
	default:
		return "", "", fmt.Errorf("unsupported media type")
	}
}

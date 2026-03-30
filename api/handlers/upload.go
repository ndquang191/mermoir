package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"memoir/api/models"
	"memoir/api/queue"
)

type UploadHandler struct {
	db          *pgxpool.Pool
	pub         *queue.Publisher
	storagePath string
}

func NewUploadHandler(db *pgxpool.Pool, pub *queue.Publisher, storagePath string) *UploadHandler {
	return &UploadHandler{db: db, pub: pub, storagePath: storagePath}
}

func (h *UploadHandler) UploadPhoto(c *gin.Context) {
	entryID := c.Param("id")

	// Verify entry exists
	var exists bool
	err := h.db.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM entries WHERE id = $1)`, entryID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}

	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "photo file is required"})
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}

	photoID := uuid.New().String()
	rawDir := filepath.Join(h.storagePath, "raw")
	if err := os.MkdirAll(rawDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create storage directory"})
		return
	}

	rawPath := filepath.Join(rawDir, fmt.Sprintf("%s%s", photoID, ext))
	dst, err := os.Create(rawPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	defer dst.Close()

	buf := make([]byte, 32*1024)
	for {
		n, readErr := file.Read(buf)
		if n > 0 {
			if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write file"})
				return
			}
		}
		if readErr != nil {
			break
		}
	}

	var photo models.Photo
	err = h.db.QueryRow(context.Background(),
		`INSERT INTO photos (id, entry_id, raw_path, status) VALUES ($1, $2, $3, 'pending') RETURNING id, entry_id, raw_path, thumb_path, status`,
		photoID, entryID, rawPath,
	).Scan(&photo.ID, &photo.EntryID, &photo.RawPath, &photo.ThumbPath, &photo.Status)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.pub.PublishImageJob(photo.ID, photo.RawPath); err != nil {
		// Non-fatal: photo is saved, worker will retry or it can be requeued
		c.JSON(http.StatusCreated, gin.H{
			"photo":   photo,
			"warning": "photo saved but failed to queue for processing",
		})
		return
	}

	c.JSON(http.StatusCreated, photo)
}

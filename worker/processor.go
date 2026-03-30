package main

import (
	"context"
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/jackc/pgx/v5/pgxpool"
)

func runWorker(ctx context.Context, id int, jobs <-chan ImageJob, db *pgxpool.Pool, storagePath string) {
	log.Printf("worker %d started", id)
	for {
		select {
		case <-ctx.Done():
			log.Printf("worker %d stopping", id)
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			log.Printf("worker %d processing photo %s", id, job.PhotoID)
			if err := processImage(ctx, job, db, storagePath); err != nil {
				log.Printf("worker %d failed to process photo %s: %v", id, job.PhotoID, err)
				updatePhotoStatus(ctx, db, job.PhotoID, "failed", "")
			}
		}
	}
}

func processImage(ctx context.Context, job ImageJob, db *pgxpool.Pool, storagePath string) error {
	src, err := imaging.Open(job.RawPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	// Generate thumbnail (300px wide, maintain aspect ratio)
	thumbDir := filepath.Join(storagePath, "thumb")
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		return fmt.Errorf("failed to create thumb dir: %w", err)
	}
	thumb := imaging.Resize(src, 300, 0, imaging.Lanczos)
	thumbPath := filepath.Join(thumbDir, job.PhotoID+".jpg")
	thumbFile, err := os.Create(thumbPath)
	if err != nil {
		return fmt.Errorf("failed to create thumb file: %w", err)
	}
	defer thumbFile.Close()
	if err := jpeg.Encode(thumbFile, thumb, &jpeg.Options{Quality: 85}); err != nil {
		return fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	// Save compressed full image
	processedDir := filepath.Join(storagePath, "processed")
	if err := os.MkdirAll(processedDir, 0755); err != nil {
		return fmt.Errorf("failed to create processed dir: %w", err)
	}
	processedPath := filepath.Join(processedDir, job.PhotoID+".jpg")
	processedFile, err := os.Create(processedPath)
	if err != nil {
		return fmt.Errorf("failed to create processed file: %w", err)
	}
	defer processedFile.Close()

	// Resize to max 1920px wide if larger
	processed := src
	if src.Bounds().Dx() > 1920 {
		processed = imaging.Resize(src, 1920, 0, imaging.Lanczos)
	}
	if err := jpeg.Encode(processedFile, processed, &jpeg.Options{Quality: 85}); err != nil {
		return fmt.Errorf("failed to encode processed image: %w", err)
	}

	updatePhotoStatus(ctx, db, job.PhotoID, "ready", thumbPath)
	return nil
}

func updatePhotoStatus(ctx context.Context, db *pgxpool.Pool, photoID, status, thumbPath string) {
	var err error
	if thumbPath != "" {
		_, err = db.Exec(ctx,
			`UPDATE photos SET status = $1, thumb_path = $2 WHERE id = $3`,
			status, thumbPath, photoID)
	} else {
		_, err = db.Exec(ctx,
			`UPDATE photos SET status = $1 WHERE id = $2`,
			status, photoID)
	}
	if err != nil {
		log.Printf("failed to update photo status for %s: %v", photoID, err)
	}
}

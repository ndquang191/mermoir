package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"memoir/api/models"
)

type EntryHandler struct {
	db *pgxpool.Pool
}

func NewEntryHandler(db *pgxpool.Pool) *EntryHandler {
	return &EntryHandler{db: db}
}

func (h *EntryHandler) GetEntries(c *gin.Context) {
	ctx := context.Background()

	rows, err := h.db.Query(ctx,
		`SELECT id, date::text, story, created_at FROM entries ORDER BY date DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	entries := []models.Entry{}
	entryMap := map[string]*models.Entry{}
	var order []string

	for rows.Next() {
		var e models.Entry
		if err := rows.Scan(&e.ID, &e.Date, &e.Story, &e.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		e.Photos = []models.Photo{}
		entryMap[e.ID] = &e
		order = append(order, e.ID)
	}

	photoRows, err := h.db.Query(ctx,
		`SELECT id, entry_id, raw_path, thumb_path, status FROM photos WHERE entry_id = ANY($1)`,
		order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer photoRows.Close()

	for photoRows.Next() {
		var p models.Photo
		if err := photoRows.Scan(&p.ID, &p.EntryID, &p.RawPath, &p.ThumbPath, &p.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if entry, ok := entryMap[p.EntryID]; ok {
			entry.Photos = append(entry.Photos, p)
		}
	}

	for _, id := range order {
		entries = append(entries, *entryMap[id])
	}

	c.JSON(http.StatusOK, entries)
}

func (h *EntryHandler) CreateEntry(c *gin.Context) {
	var req struct {
		Date  string `json:"date" binding:"required"`
		Story string `json:"story"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var entry models.Entry
	err := h.db.QueryRow(context.Background(),
		`INSERT INTO entries (date, story) VALUES ($1, $2) RETURNING id, date::text, story, created_at`,
		req.Date, req.Story,
	).Scan(&entry.ID, &entry.Date, &entry.Story, &entry.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	entry.Photos = []models.Photo{}
	c.JSON(http.StatusCreated, entry)
}

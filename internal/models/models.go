package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Images struct {
	ID             int         `json:"id"`
	UserID         int         `json:"user_id"`
	OriginalURL    string      `json:"original_url"`
	TransformedURL string      `json:"transformed_url"`
	Metadata       interface{} `json:"metadata"` // or use map[string]any if you want it typed
	CreatedAt      time.Time   `json:"created_at"`
}

type TransformationRequest struct {
	Resize    *ResizeOptions    `json:"resize,omitempty"`
	Crop      *CropOptions      `json:"crop,omitempty"`
	Rotate    *float64          `json:"rotate"` //  Using a pointer lets you check if the rotation is requested (nil vs. 0).
	Watermark *WatermarkOptions `json:"watermark,omitempty"`
	Flip      bool              `json:"flip,omitempty"`
	Mirror    bool              `json:"mirror,omitempty"`
	Compress  *int              `json:"compress,omitempty"` //  1-100 Detect if the field was actually provided or not.
	//  Avoid assuming 0 means “no compression,” which would be invalid or misleading.
	Format string `json:"format,omitempty"` // "jpeg", "png", etc.
	Filter string `json:"filter,omitempty"` // "grayscale", "sepia", etc.
}

type ResizeOptions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type CropOptions struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type WatermarkOptions struct {
	X int `json:"x"`
	Y int `json:"y"`
}
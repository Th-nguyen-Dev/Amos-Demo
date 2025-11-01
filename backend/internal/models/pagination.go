package models

// CursorParams represents cursor pagination parameters
type CursorParams struct {
	Limit     int    `form:"limit" validate:"min=1,max=100"`
	Cursor    string `form:"cursor"`
	Direction string `form:"direction" validate:"omitempty,oneof=next prev"`
	Search    string `form:"search" validate:"omitempty,max=200"`
}

// CursorPagination represents cursor pagination metadata
type CursorPagination struct {
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasNext    bool   `json:"has_next"`
	HasPrev    bool   `json:"has_prev"`
}

// NewCursorParams creates default cursor params
func NewCursorParams() CursorParams {
	return CursorParams{
		Limit:     50,
		Direction: "next",
	}
}

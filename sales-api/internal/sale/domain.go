package sale

import "time"

// Sale represents a system user with metadata for auditing and versioning.
type Sale struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Amount    float32   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
}

type metadata struct {
	Quantity    int     `json:"quantity"`
	Approved    int     `json:"approved"`
	Rejected    int     `json:"rejected"`
	Pending     int     `json:"pending"`
	TotalAmount float32 `json:"total_amount"`
}

type informe struct {
	Metadata metadata `json:"metadata"`
	Results  []Sale   `json:"results"`
}

// UpdateFields represents the optional fields for updating a Sale.
// A nil pointer means “no change” for that field.
type UpdateFieldsSale struct {
	Status string `json:"status"`
} //ver * de UpdateFieldsUser

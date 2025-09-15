package server

type Asset struct {
	ID        string           `json:"id"`
	OwnerID   string           `json:"owner_id,omitzero"`
	OwnerType string           `json:"owner_typ,omitzero"`
	Kind      string           `json:"kind,omitzero"`
	Path      string           `json:"path"`
	IsPrimary bool             `json:"is_primary"`
	Position  int              `json:"position,omitzero"`
	Colors    map[int][4]uint8 `json:"colors"`
	CreatedAt string           `json:"created_at,omitzero"`
	UpdatedAt string           `json:"updated_at,omitzero"`
	DeletedAt *string          `json:"deleted_at,omitempty"`
}

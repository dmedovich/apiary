// Package nullable exercises nullable (*T) and enum schema generation at
// every nesting depth apiary supports: a bare pointer/enum field, a
// pointer/enum as a slice element, and a pointer/enum as a map value.
// It intentionally uses the typed-handler style (no gin needed) since the
// bug it guards against is in schema generation, not annotation
// resolution.
package nullable

// apiary:operation GET /api/v1/widgets/{id}
// summary: Get widget
// tags: widgets
func GetWidget() (Widget, error) { return Widget{}, nil }

type Status string

const (
	StatusActive  Status = "active"
	StatusRetired Status = "retired"
)

type Address struct {
	City string `json:"city"`
}

type Widget struct {
	// Bare pointer / bare enum -- the case that always worked.
	Nickname *string `json:"nickname" doc:"Optional display name"`
	Status   Status  `json:"status"`

	// Pointer/enum as a slice element.
	Tags         []*string  `json:"tags"`
	AltAddresses []*Address `json:"alt_addresses"`
	PastStatuses []Status   `json:"past_statuses"`

	// Pointer/enum as a map value.
	Metadata     map[string]*string `json:"metadata"`
	StatusByYear map[string]Status  `json:"status_by_year"`

	// Pointer to the composite type itself, not its element.
	ShipTo *Address `json:"ship_to"`
}

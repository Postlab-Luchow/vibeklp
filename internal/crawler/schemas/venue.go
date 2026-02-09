package schemas

// VenueSchema returns JSON schema for venue extraction
func VenueSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":   "venue_extraction",
		"strict": true,
		"schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Venue name from h1 tag",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Subtitle or short description",
				},
				"address": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"street": map[string]interface{}{
							"type":        "string",
							"description": "Street name and number",
						},
						"postalCode": map[string]interface{}{
							"type":        "string",
							"description": "5-digit German postal code",
						},
						"city": map[string]interface{}{
							"type":        "string",
							"description": "City name, may include OT (Orsteil)",
						},
					},
					"required":             []string{"street", "postalCode", "city"},
					"additionalProperties": false,
				},
				"contact": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"phone": map[string]interface{}{
							"type":        "string",
							"description": "Phone number without 'Fon' prefix",
						},
						"email": map[string]interface{}{
							"type":        "string",
							"description": "Email address, decode JavaScript if encoded",
						},
						"website": map[string]interface{}{
							"type":        "string",
							"description": "Full website URL with http/https",
						},
					},
					"required":             []string{"phone", "email", "website"},
					"additionalProperties": false,
				},
				"bikeRoute": map[string]interface{}{
					"type":        "string",
					"description": "Bike route number mentioned in 'Fahrradtour: X'",
				},
			},
			"required":             []string{"name", "description", "address", "contact", "bikeRoute"},
			"additionalProperties": false,
		},
	}
}

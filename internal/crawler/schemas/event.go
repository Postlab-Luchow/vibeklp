package schemas

// EventSchema returns JSON schema for event extraction
func EventSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":   "event_extraction",
		"strict": true,
		"schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Event title from <b> tag",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Description from <em> tag",
				},
				"artist": map[string]interface{}{
					"type":        "string",
					"description": "Artist or organizer name from first <p> tag",
				},
				"dates": map[string]interface{}{
					"type":        "array",
					"description": "ALL dates for this event - some events have multiple dates",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"date": map[string]interface{}{
								"type":        "string",
								"description": "Date in YYYY-MM-DD format. Year is 2026",
							},
							"startTime": map[string]interface{}{
								"type":        "string",
								"description": "Start time in HH:MM format (24-hour)",
							},
							"endTime": map[string]interface{}{
								"type":        "string",
								"description": "End time in HH:MM format if mentioned, otherwise empty string",
							},
						},
						"required":             []string{"date", "startTime", "endTime"},
						"additionalProperties": false,
					},
				},
				"admission": map[string]interface{}{
					"type":        "string",
					"description": "Admission info from text in parentheses like (Hutkasse) or (Eintritt frei)",
				},
			},
			"required":             []string{"title", "description", "artist", "dates", "admission"},
			"additionalProperties": false,
		},
	}
}

// EventsBatchSchema returns schema for batch event extraction
func EventsBatchSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":   "events_batch",
		"strict": true,
		"schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"events": map[string]interface{}{
					"type":        "array",
					"description": "List of extracted events",
					"items":       EventSchema()["schema"],
				},
			},
			"required":             []string{"events"},
			"additionalProperties": false,
		},
	}
}

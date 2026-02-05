package schemas

// ExhibitionSchema returns JSON schema for exhibition extraction
func ExhibitionSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":   "exhibition_extraction",
		"strict": true,
		"schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Exhibition title from <b> tag",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Description from <em> tag",
				},
				"artist": map[string]interface{}{
					"type":        "string",
					"description": "Artist name from first <p> tag",
				},
			},
			"required":             []string{"title"},
			"additionalProperties": false,
		},
	}
}

// ExhibitionsBatchSchema returns schema for batch exhibition extraction
func ExhibitionsBatchSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":   "exhibitions_batch",
		"strict": true,
		"schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"exhibitions": map[string]interface{}{
					"type":        "array",
					"description": "List of extracted exhibitions",
					"items":       ExhibitionSchema()["schema"],
				},
			},
			"required":             []string{"exhibitions"},
			"additionalProperties": false,
		},
	}
}

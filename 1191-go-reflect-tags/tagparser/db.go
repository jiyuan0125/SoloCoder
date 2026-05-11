package tagparser

func ParseDBTag(tag string) *DBTag {
	if tag == "" {
		return nil
	}

	parts := splitTagOptions(tag)
	if len(parts) == 0 {
		return nil
	}

	result := &DBTag{
		Column: parts[0],
	}

	for _, opt := range parts[1:] {
		switch opt {
		case "pk", "primary_key":
			result.IndexType = "pk"
		case "unique":
			result.IndexType = "unique"
		case "index":
			result.IndexType = "index"
		}
	}

	return result
}

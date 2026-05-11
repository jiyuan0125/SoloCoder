package parser

func Search(doc *PackageDoc, query string, includeUnexported bool) []Searchable {
	var results []Searchable

	for _, item := range doc.AllDeclarations() {
		if !includeUnexported && !item.Exported {
			continue
		}
		if item.Matches(query) {
			results = append(results, item)
		}
	}

	return results
}

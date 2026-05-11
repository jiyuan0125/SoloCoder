package jsonpath

func Query(data interface{}, path string) ([]interface{}, error) {
	matches, err := Evaluate(data, path)
	if err != nil {
		return nil, err
	}
	values := make([]interface{}, len(matches))
	for i, m := range matches {
		values[i] = m.Value
	}
	return values, nil
}

func QueryWithPaths(data interface{}, path string) ([]Match, error) {
	return Evaluate(data, path)
}

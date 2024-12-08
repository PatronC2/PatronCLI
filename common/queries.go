package common

import "strings"

func QueryFields(items []map[string]interface{}, query string) []map[string]interface{} {
	fields := strings.Split(query, ",")
	result := []map[string]interface{}{}

	for _, item := range items {
		filteredItem := map[string]interface{}{}
		for _, field := range fields {
			field = strings.TrimSpace(field) // Ensure no extra spaces
			if value, exists := item[field]; exists {
				filteredItem[field] = value
			}
		}
		result = append(result, filteredItem)
	}

	return result
}

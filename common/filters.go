package common

import (
	"fmt"
	"strings"
)

func FilterItemsWithTags(items []map[string]interface{}, filter string) []map[string]interface{} {
	if filter == "" {
		return items
	}

	criteria := parseFilter(filter)
	if len(criteria) == 0 {
		fmt.Println("Invalid filter syntax. Use 'key=value' or 'tags.key=[value1,value2]'")
		return nil
	}

	var filteredItems []map[string]interface{}
	for _, item := range items {
		matchesAll := true

		for key, values := range criteria {
			if strings.HasPrefix(key, "tags.") {
				tagKey := strings.TrimPrefix(key, "tags.")
				tags, ok := item["tags"].([]interface{})
				if !ok {
					matchesAll = false
					break
				}

				tagMatched := false
				for _, tag := range tags {
					tagMap, ok := tag.(map[string]interface{})
					if !ok {
						continue
					}
					if fmt.Sprintf("%v", tagMap["key"]) == tagKey {
						tagValue := fmt.Sprintf("%v", tagMap["value"])
						if contains(values, tagValue) {
							tagMatched = true
							break
						}
					}
				}
				if !tagMatched {
					matchesAll = false
					break
				}
			} else {
				fieldValue := fmt.Sprintf("%v", item[key])
				if !contains(values, fieldValue) {
					matchesAll = false
					break
				}
			}
		}

		if matchesAll {
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems
}

func parseFilter(filter string) map[string][]string {
	criteria := make(map[string][]string)

	parts := splitFilter(filter, ',')

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			fmt.Printf("Invalid filter segment: %s\n", part)
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
			listValues := strings.Trim(value, "[]")
			values := strings.Split(listValues, ",")
			for i := range values {
				values[i] = strings.TrimSpace(values[i])
			}
			criteria[key] = values
		} else {
			criteria[key] = []string{value}
		}
	}

	return criteria
}

func splitFilter(input string, sep rune) []string {
	var parts []string
	var buffer []rune
	var insideBrackets bool

	for _, char := range input {
		if char == '[' {
			insideBrackets = true
		} else if char == ']' {
			insideBrackets = false
		}

		if char == sep && !insideBrackets {
			parts = append(parts, string(buffer))
			buffer = []rune{}
		} else {
			buffer = append(buffer, char)
		}
	}

	if len(buffer) > 0 {
		parts = append(parts, string(buffer))
	}

	return parts
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

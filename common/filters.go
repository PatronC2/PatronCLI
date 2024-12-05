package common

import "fmt"

func FilterItems(items []interface{}, filterKey, filterValue string) []interface{} {
	if filterKey == "" || filterValue == "" {
		// No filtering criteria provided, return all items
		return items
	}

	var filteredItems []interface{}
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if fmt.Sprintf("%v", itemMap[filterKey]) == filterValue {
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems
}

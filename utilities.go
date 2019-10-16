package main

func inArray(array []string, keyword string) bool {
	for _, value := range array {
		if value == keyword {
			return true
		}
	}
	return false
}

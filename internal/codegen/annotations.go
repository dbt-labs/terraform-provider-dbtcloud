package codegen

import (
	"strings"
	"unicode"
)

// fieldNameToLabel converts a Go field name to a human-readable label.
//
// Examples:
//
//	"Name" → "Name"
//	"ID" → "ID"
//	"AccountID" → "Account ID"
//	"ProjectId" → "Project ID"
//	"Dbt_Version" → "dbt Version"
//	"Use_Custom_Branch" → "Use Custom Branch"
func fieldNameToLabel(name string) string {
	// First handle underscore-separated names.
	if strings.Contains(name, "_") {
		parts := strings.Split(name, "_")
		for i, p := range parts {
			if p == "" {
				continue
			}
			parts[i] = capitalizeWord(p)
		}
		return strings.Join(parts, " ")
	}

	// camelCase / PascalCase splitting
	var words []string
	runes := []rune(name)
	start := 0

	for i := 1; i < len(runes); i++ {
		if unicode.IsLower(runes[i-1]) && unicode.IsUpper(runes[i]) {
			words = append(words, string(runes[start:i]))
			start = i
		} else if i+1 < len(runes) && unicode.IsUpper(runes[i-1]) && unicode.IsUpper(runes[i]) && unicode.IsLower(runes[i+1]) {
			words = append(words, string(runes[start:i]))
			start = i
		}
	}
	words = append(words, string(runes[start:]))

	for i, w := range words {
		words[i] = capitalizeWord(w)
	}

	return strings.Join(words, " ")
}

// capitalizeWord handles special casing for common abbreviations and brand names.
func capitalizeWord(w string) string {
	lower := strings.ToLower(w)
	switch lower {
	case "dbt":
		return "dbt"
	}
	upper := strings.ToUpper(w)
	switch upper {
	case "ID", "URL", "HTTP", "API", "SQL", "JSON", "XML", "HTML", "CSS", "JS":
		return upper
	}
	// Capitalize first letter only.
	if len(w) == 0 {
		return w
	}
	runes := []rune(w)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// autoFormat returns the auto-derived format string for a given field name and type.
func autoFormat(fieldName, goType string) string {
	if fieldName == "State" && (goType == "int" || goType == "*int") {
		return "state"
	}
	if goType == "bool" {
		return "bool"
	}
	if goType == "[]string" {
		return "join"
	}
	return ""
}

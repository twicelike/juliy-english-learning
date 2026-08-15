package translation

import "strings"

type Language struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

var SupportedLanguages = []Language{
	{Code: "en", Name: "English"},
	{Code: "vi", Name: "Vietnamese"},
	{Code: "de", Name: "German"},
	{Code: "ja", Name: "Japanese"},
	{Code: "ko", Name: "Korean"},
	{Code: "fr", Name: "French"},
	{Code: "es", Name: "Spanish"},
	{Code: "zh", Name: "Chinese"},
}

func IsSupportedLanguage(code string) bool {
	code = strings.ToLower(code)
	for _, l := range SupportedLanguages {
		if l.Code == code {
			return true
		}
	}
	return false
}

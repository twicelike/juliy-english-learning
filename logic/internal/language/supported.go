package language

type Language struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

var SupportedLanguages = []Language{
	{Code: "en", Name: "English"},
	{Code: "vi", Name: "Vietnamese"},
	{Code: "es", Name: "Spanish"},
	{Code: "fr", Name: "French"},
	{Code: "de", Name: "German"},
	{Code: "ja", Name: "Japanese"},
	{Code: "ko", Name: "Korean"},
	{Code: "zh", Name: "Chinese"},
}

// IsSupported checks if a given language code is supported
func IsSupported(code string) bool {
	for _, lang := range SupportedLanguages {
		if lang.Code == code {
			return true
		}
	}
	return false
}

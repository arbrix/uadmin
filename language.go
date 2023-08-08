package uadmin

import "fmt"

// Language !
type Language struct {
	Model
	EnglishName    string `uadmin:"required;read_only;filter;search"`
	Name           string `uadmin:"required;read_only;filter;search"`
	Flag           string `uadmin:"image;list_exclude"`
	Code           string `uadmin:"filter;read_only;list_exclude"`
	RTL            bool   `uadmin:"list_exclude"`
	Default        bool   `uadmin:"help:Set as the default language;list_exclude"`
	Active         bool   `uadmin:"help:To show this in available languages;filter"`
	AvailableInGui bool   `uadmin:"help:The App is available in this language;read_only"`
}

// String !
func (l Language) String() string {
	return l.Code
}

// Save !
func (l *Language) Save() {
	if l.Default {
		Update([]Language{}, "default", false, "`default` = ?", true)
		DefaultLang = *l
	}
	Save(l)
	tempActiveLangs := []Language{}
	Filter(&tempActiveLangs, "`active` = ?", true)
	ActiveLangs = tempActiveLangs

	tanslationList := []translation{}
	for i := range ActiveLangs {
		tanslationList = append(tanslationList, translation{
			Active:  ActiveLangs[i].Active,
			Default: ActiveLangs[i].Default,
			Code:    ActiveLangs[i].Code,
			Name:    fmt.Sprintf("%s (%s)", ActiveLangs[i].Name, ActiveLangs[i].EnglishName),
		})
	}

	for modelName := range Schema {
		for i := range Schema[modelName].Fields {
			if Schema[modelName].Fields[i].Type == cMULTILINGUAL || Schema[modelName].Fields[i].Type == cHTML_MULTILINGUAL {
				Schema[modelName].Fields[i].Translations = tanslationList
			}
		}
	}
}

// GetDefaultLanguage returns the default language
func GetDefaultLanguage() Language {
	return DefaultLang
}

// GetActiveLanguages returns a list of active langages
func GetActiveLanguages() []Language {
	return ActiveLangs
}

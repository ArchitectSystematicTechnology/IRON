package langs

import "fmt"

// GetLangHelper returns a LangHelper for the passed in language
func GetLangHelper(lang string) (LangHelper, error) {
	switch lang {
	case "go":
		return &GoLangHelper{}, nil
	}
	return nil, fmt.Errorf("No language helper found for %v", lang)
}

type LangHelper interface {
	Entrypoint() (string, error)
	HasPreBuild() bool
	PreBuild() error
	AfterBuild() error
}

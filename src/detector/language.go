package detector

import (
	"os"
	"path/filepath"

	"github.com/go-enry/go-enry/v2"
)

type LanguageDetector struct{}

func New() *LanguageDetector {
	return &LanguageDetector{}
}

func (d *LanguageDetector) DetectLanguage(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	language := enry.GetLanguage(filepath.Base(path), content)
	if language == "" {
		return "Unknown", nil
	}
	return language, nil
}

func (d *LanguageDetector) DetectByExtension(filename string) string {
	languages := enry.GetLanguagesByExtension(filename, nil, nil)
	if len(languages) > 0 {
		return languages[0]
	}
	return "Unknown"
}

func (d *LanguageDetector) DetectByFilename(filename string) string {
	languages := enry.GetLanguagesByFilename(filename, nil, nil)
	if len(languages) > 0 {
		return languages[0]
	}
	return "Unknown"
}

func (d *LanguageDetector) IsBinary(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return enry.IsBinary(content)
}

func (d *LanguageDetector) IsVendor(path string) bool {
	return enry.IsVendor(path)
}

func (d *LanguageDetector) IsGenerated(path string, content []byte) bool {
	return enry.IsGenerated(path, content)
}

func (d *LanguageDetector) IsTest(path string) bool {
	return enry.IsTest(path)
}

func (d *LanguageDetector) IsDocumentation(path string) bool {
	return enry.IsDocumentation(path)
}

func (d *LanguageDetector) ShouldAnalyze(path string) bool {
	if d.IsVendor(path) {
		return false
	}
	if d.IsBinary(path) {
		return false
	}
	if d.IsDocumentation(path) {
		return false
	}
	return true
}

func DetectProjectLanguages(rootPath string) (map[string]int, error) {
	detector := New()
	languages := make(map[string]int)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if detector.IsVendor(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if !detector.ShouldAnalyze(path) {
			return nil
		}

		lang, err := detector.DetectLanguage(path)
		if err != nil {
			return nil
		}

		if lang != "Unknown" && lang != "" {
			languages[lang]++
		}

		return nil
	})

	return languages, err
}

package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

var defaultExcludeDirs = map[string]bool{
	"node_modules":   true,
	"vendor":         true,
	".git":           true,
	"__pycache__":    true,
	".venv":          true,
	"venv":           true,
	".env":           true,
	"env":            true,
	".tox":           true,
	".pytest_cache":  true,
	".mypy_cache":    true,
	".ruff_cache":    true,
	"dist":           true,
	"build":          true,
	".next":          true,
	".nuxt":          true,
	".output":        true,
	".cache":         true,
	".tmp":           true,
	".temp":          true,
	"coverage":       true,
	".nyc_output":    true,
	".parcel-cache":  true,
	".turbo":         true,
	".vercel":        true,
	".netlify":       true,
	"target":         true,
	".gradle":        true,
	".idea":          true,
	".vscode":        true,
	".vs":            true,
	".DS_Store":      true,
	"Thumbs.db":      true,
	".svn":           true,
	".hg":            true,
}

var defaultExcludeExtensions = map[string]bool{
	".pyc":     true,
	".pyo":     true,
	".so":      true,
	".dll":     true,
	".dylib":   true,
	".class":   true,
	".jar":     true,
	".war":     true,
	".ear":     true,
	".o":       true,
	".a":       true,
	".lib":     true,
	".exe":     true,
	".bin":     true,
	".jpg":     true,
	".jpeg":    true,
	".png":     true,
	".gif":     true,
	".bmp":     true,
	".ico":     true,
	".svg":     true,
	".webp":    true,
	".mp3":     true,
	".mp4":     true,
	".wav":     true,
	".avi":     true,
	".mov":     true,
	".webm":    true,
	".flv":     true,
	".woff":    true,
	".woff2":   true,
	".ttf":     true,
	".eot":     true,
	".otf":     true,
	".pdf":     true,
	".zip":     true,
	".tar":     true,
	".gz":      true,
	".rar":     true,
	".7z":      true,
	".lock":    true,
	".min.js":  true,
	".min.css": true,
	".map":     true,
}

var defaultExcludeFiles = map[string]bool{
	"package-lock.json": true,
	"yarn.lock":         true,
	"pnpm-lock.yaml":    true,
	"poetry.lock":       true,
	"Pipfile.lock":      true,
	"composer.lock":     true,
	"Gemfile.lock":      true,
	"Cargo.lock":        true,
	"go.sum":            true,
	".gitignore":        true,
	".gitattributes":    true,
	".editorconfig":     true,
	".prettierrc":       true,
	".eslintrc":         true,
	".eslintignore":     true,
	"tsconfig.tsbuildinfo": true,
}

func ScanDirectory(root string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	gitignorePath := filepath.Join(absRoot, ".gitignore")
	var gi *ignore.GitIgnore
	if _, err := os.Stat(gitignorePath); err == nil {
		gi, _ = ignore.CompileIgnoreFile(gitignorePath)
	}

	var paths []string

	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			return nil
		}

		if relPath == "." {
			return nil
		}

		name := d.Name()

		if d.IsDir() {
			if defaultExcludeDirs[name] {
				return filepath.SkipDir
			}
			if gi != nil && gi.MatchesPath(relPath+"/") {
				return filepath.SkipDir
			}
			paths = append(paths, relPath+"/")
			return nil
		}

		if defaultExcludeFiles[name] {
			return nil
		}

		ext := filepath.Ext(name)
		if defaultExcludeExtensions[ext] {
			return nil
		}

		if strings.HasSuffix(name, ".min.js") || strings.HasSuffix(name, ".min.css") {
			return nil
		}

		if gi != nil && gi.MatchesPath(relPath) {
			return nil
		}

		paths = append(paths, relPath)
		return nil
	})

	if err != nil {
		return "", err
	}

	sort.Strings(paths)

	return strings.Join(paths, "\n"), nil
}

func FindReadme(root string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	readmeNames := []string{
		"README.md",
		"README.MD",
		"readme.md",
		"README",
		"README.txt",
		"README.rst",
		"Readme.md",
	}

	for _, name := range readmeNames {
		path := filepath.Join(absRoot, name)
		content, err := os.ReadFile(path)
		if err == nil {
			return string(content), nil
		}
	}

	return "", os.ErrNotExist
}

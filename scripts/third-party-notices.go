// Command third-party-notices generates the checked-in inventory and license texts for
// production Go and npm dependencies. Run it from backend/ so `go list` sees the backend module.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type component struct {
	ecosystem string
	name      string
	version   string
	license   string
	url       string
	texts     []licenseText
}

type licenseText struct {
	name string
	body []byte
}

type goPackage struct {
	Module *goModule
}

type goModule struct {
	Path    string
	Version string
	Dir     string
	Main    bool
	Replace *goModule
}

type npmLock struct {
	Packages map[string]struct {
		Version string `json:"version"`
		License string `json:"license"`
		Dev     bool   `json:"dev"`
	} `json:"packages"`
}

var check = flag.Bool("check", false, "verify generated files without changing them")

func main() {
	flag.Parse()
	root, err := repositoryRoot()
	must(err)
	components, err := collect(root)
	must(err)
	files, err := render(components)
	must(err)
	if *check {
		must(checkFiles(root, files))
		fmt.Println("third-party notices are current")
		return
	}
	must(writeFiles(root, files))
	fmt.Printf("generated notices for %d production dependencies\n", len(components))
}

func repositoryRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if filepath.Base(cwd) == "backend" {
		return filepath.Dir(cwd), nil
	}
	if _, err := os.Stat(filepath.Join(cwd, "backend", "go.mod")); err == nil {
		return cwd, nil
	}
	return "", errors.New("run from the repository root or backend/")
}

func collect(root string) ([]component, error) {
	goComponents, err := collectGo(filepath.Join(root, "backend"))
	if err != nil {
		return nil, err
	}
	npmComponents, err := collectNPM(root)
	if err != nil {
		return nil, err
	}
	components := append(goComponents, npmComponents...)
	sort.Slice(components, func(i, j int) bool {
		if components[i].ecosystem != components[j].ecosystem {
			return components[i].ecosystem < components[j].ecosystem
		}
		return components[i].name < components[j].name
	})
	return components, nil
}

func collectGo(backend string) ([]component, error) {
	cmd := exec.Command("go", "list", "-deps", "-json", "./...")
	cmd.Dir = backend
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("go list: %w: %s", err, bytes.TrimSpace(out))
	}
	decoder := json.NewDecoder(bytes.NewReader(out))
	modules := map[string]*goModule{}
	for {
		var pkg goPackage
		err := decoder.Decode(&pkg)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode go list: %w", err)
		}
		module := pkg.Module
		if module == nil || module.Main {
			continue
		}
		if module.Replace != nil {
			module = module.Replace
		}
		if module.Version == "" || module.Dir == "" {
			return nil, fmt.Errorf("Go module %s has no locked version or local directory", module.Path)
		}
		modules[module.Path+"@"+module.Version] = module
	}
	keys := sortedKeys(modules)
	result := make([]component, 0, len(keys))
	for _, key := range keys {
		module := modules[key]
		texts, err := readLicenseTexts(module.Dir)
		if err != nil {
			return nil, fmt.Errorf("%s@%s: %w", module.Path, module.Version, err)
		}
		license, err := identifyLicense(texts)
		if err != nil {
			return nil, fmt.Errorf("%s@%s: %w", module.Path, module.Version, err)
		}
		result = append(result, component{
			ecosystem: "Go", name: module.Path, version: module.Version, license: license,
			url: "https://pkg.go.dev/" + module.Path + "@" + module.Version, texts: texts,
		})
	}
	return result, nil
}

func collectNPM(root string) ([]component, error) {
	data, err := os.ReadFile(filepath.Join(root, "frontend", "package-lock.json"))
	if err != nil {
		return nil, err
	}
	var lock npmLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, err
	}
	paths := sortedKeys(lock.Packages)
	var result []component
	for _, packagePath := range paths {
		entry := lock.Packages[packagePath]
		if entry.Dev || !strings.HasPrefix(packagePath, "node_modules/") {
			continue
		}
		name := packagePath[strings.LastIndex(packagePath, "node_modules/")+len("node_modules/"):]
		if entry.Version == "" || entry.License == "" {
			return nil, fmt.Errorf("npm package %s has no locked version or license", name)
		}
		if !compatibleSPDX(entry.License) {
			return nil, fmt.Errorf("npm package %s@%s has unknown or incompatible license %q", name, entry.Version, entry.License)
		}
		texts, err := readLicenseTexts(filepath.Join(root, "frontend", packagePath))
		if err != nil {
			return nil, fmt.Errorf("%s@%s: %w", name, entry.Version, err)
		}
		result = append(result, component{
			ecosystem: "npm", name: name, version: entry.Version, license: entry.License,
			url: "https://www.npmjs.com/package/" + name + "/v/" + entry.Version, texts: texts,
		})
	}
	return result, nil
}

func readLicenseTexts(dir string) ([]licenseText, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var texts []licenseText
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		lower := strings.ToLower(entry.Name())
		if !strings.HasPrefix(lower, "license") && !strings.HasPrefix(lower, "copying") && !strings.HasPrefix(lower, "notice") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		texts = append(texts, licenseText{name: entry.Name(), body: bytes.TrimSpace(body)})
	}
	sort.Slice(texts, func(i, j int) bool { return texts[i].name < texts[j].name })
	if len(texts) == 0 {
		return nil, errors.New("no license or notice file was published")
	}
	return texts, nil
}

func identifyLicense(texts []licenseText) (string, error) {
	combined := ""
	for _, text := range texts {
		combined += "\n" + strings.ToLower(string(text.body))
	}
	switch {
	case strings.Contains(combined, "apache license") && strings.Contains(combined, "version 2.0"):
		return "Apache-2.0", nil
	case strings.Contains(combined, "permission is hereby granted, free of charge"):
		return "MIT", nil
	case strings.Contains(combined, "redistribution and use in source and binary forms") && strings.Contains(combined, "neither the name"):
		return "BSD-3-Clause", nil
	case strings.Contains(combined, "redistribution and use in source and binary forms"):
		return "BSD-2-Clause", nil
	case strings.Contains(combined, "permission to use, copy, modify, and/or distribute"):
		return "ISC", nil
	default:
		return "", errors.New("unknown or incompatible license text")
	}
}

func compatibleSPDX(value string) bool {
	switch value {
	case "0BSD", "Apache-2.0", "BSD-2-Clause", "BSD-3-Clause", "ISC", "MIT":
		return true
	default:
		return false
	}
}

func render(components []component) (map[string][]byte, error) {
	files := map[string][]byte{}
	var index strings.Builder
	index.WriteString("# Third-party notices\n\n")
	index.WriteString("This file is generated by `scripts/third-party-notices.go`. Do not edit it manually.\n\n")
	index.WriteString("It inventories production dependencies distributed with Catalina Support. Development-only and test-only dependencies, and externally referenced container images, are excluded.\n\n")
	index.WriteString("| Ecosystem | Component | Version | License | Source | License text |\n")
	index.WriteString("| --- | --- | --- | --- | --- | --- |\n")
	for _, item := range components {
		filename := licenseFilename(item)
		fmt.Fprintf(&index, "| %s | `%s` | `%s` | `%s` | [source](%s) | [text](third_party_licenses/%s) |\n",
			item.ecosystem, item.name, item.version, item.license, item.url, filename)
		var body strings.Builder
		fmt.Fprintf(&body, "%s %s\nSource: %s\nSPDX license: %s\n", item.name, item.version, item.url, item.license)
		for _, text := range item.texts {
			fmt.Fprintf(&body, "\n--- %s ---\n\n%s\n", text.name, text.body)
		}
		files[filepath.Join("third_party_licenses", filename)] = []byte(body.String())
	}
	files["THIRD_PARTY_NOTICES.md"] = []byte(index.String())
	return files, nil
}

func licenseFilename(item component) string {
	replacer := strings.NewReplacer("/", "--", "@", "-at-", "\\", "--")
	return strings.ToLower(item.ecosystem) + "--" + replacer.Replace(item.name) + "--" + item.version + ".txt"
}

func checkFiles(root string, expected map[string][]byte) error {
	actual, err := generatedFiles(root)
	if err != nil {
		return err
	}
	for name, body := range expected {
		current, ok := actual[name]
		if !ok {
			return fmt.Errorf("missing generated file %s; run the generator", name)
		}
		if !bytes.Equal(current, body) {
			return fmt.Errorf("generated file %s is stale; run the generator", name)
		}
		delete(actual, name)
	}
	for name := range actual {
		return fmt.Errorf("obsolete generated file %s; run the generator", name)
	}
	return nil
}

func generatedFiles(root string) (map[string][]byte, error) {
	result := map[string][]byte{}
	index := filepath.Join(root, "THIRD_PARTY_NOTICES.md")
	if body, err := os.ReadFile(index); err == nil {
		result["THIRD_PARTY_NOTICES.md"] = body
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	dir := filepath.Join(root, "third_party_licenses")
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		result[rel] = body
		return nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return result, nil
	}
	return result, err
}

func writeFiles(root string, files map[string][]byte) error {
	dir := filepath.Join(root, "third_party_licenses")
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for name, body := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, body, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package tools

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func ExtractPackage(pkg ToolPackage, extractedRoot string) (ToolPackage, error) {
	if pkg.Manifest == nil {
		return pkg, fmt.Errorf("tool manifest is not loaded")
	}
	if !pkg.Validation.Valid {
		return pkg, fmt.Errorf("cannot extract invalid package")
	}

	targetDir := filepath.Join(extractedRoot, sanitizeFileName(pkg.ID), sanitizeFileName(pkg.Version))
	if err := os.RemoveAll(targetDir); err != nil {
		return pkg, err
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return pkg, err
	}

	reader, err := zip.OpenReader(pkg.Location)
	if err != nil {
		return pkg, err
	}
	defer reader.Close()

	for _, file := range reader.File {
		name := normalizeZipName(file.Name)
		if name == "" {
			continue
		}
		if err := validatePortablePackagePath(name); err != nil {
			return pkg, err
		}

		destination := filepath.Join(targetDir, filepath.FromSlash(name))
		if err := ensureInside(targetDir, destination); err != nil {
			return pkg, err
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destination, 0o755); err != nil {
				return pkg, err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return pkg, err
		}

		in, err := file.Open()
		if err != nil {
			return pkg, err
		}

		out, err := os.Create(destination)
		if err != nil {
			in.Close()
			return pkg, err
		}

		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		in.Close()
		if copyErr != nil {
			return pkg, copyErr
		}
		if closeErr != nil {
			return pkg, closeErr
		}
	}

	pkg.ExtractedPath = targetDir
	return pkg, nil
}

func ResolveExtractedPackagePath(extractedRoot string, toolID string, version string, relativePath string) (string, error) {
	if err := validatePortablePackagePath(relativePath); err != nil {
		return "", err
	}
	root := filepath.Join(extractedRoot, sanitizeFileName(toolID), sanitizeFileName(version))
	target := filepath.Join(root, filepath.FromSlash(normalizeZipName(relativePath)))
	if err := ensureInside(root, target); err != nil {
		return "", err
	}
	return target, nil
}

func RemoveExtractedPackage(extractedRoot string, toolID string) error {
	target := filepath.Join(extractedRoot, sanitizeFileName(toolID))
	return os.RemoveAll(target)
}

func ensureInside(root string, target string) error {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return err
	}
	if rel == "." {
		return nil
	}
	if rel == ".." || len(rel) >= 3 && rel[:3] == ".."+string(filepath.Separator) {
		return fmt.Errorf("path escapes extracted package directory")
	}
	return nil
}

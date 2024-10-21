package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileComparer struct {
	SourceDir      string
	TargetDir      string
	IgnorePatterns []string
}

func NewFileComparer(sourceDir, targetDir string, ignorePatterns []string) *FileComparer {
	return &FileComparer{
		SourceDir:      sourceDir,
		TargetDir:      targetDir,
		IgnorePatterns: ignorePatterns,
	}
}

func (fc *FileComparer) GenerateComparisonScript(outputFile string) error {
	sourceFiles, err := fc.getFiles(fc.SourceDir)
	if err != nil {
		return err
	}

	commands := []string{"#!/usr/bin/env bash\n"}

	for _, sourceFile := range sourceFiles {
		relPath, err := filepath.Rel(fc.SourceDir, sourceFile)
		if err != nil {
			return err
		}

		if fc.shouldIgnore(relPath) {
			continue
		}

		targetFile := filepath.Join(fc.TargetDir, relPath)

		if fc.fileExists(targetFile) {
			// Convert both sourceFile and targetFile to absolute paths
			absSourceFile, err := filepath.Abs(sourceFile)
			if err != nil {
				return err
			}
			absTargetFile, err := filepath.Abs(targetFile)
			if err != nil {
				return err
			}

			command := fmt.Sprintf("diff --unified --ignore-all-space %q %q", absSourceFile, absTargetFile)
			commands = append(commands, command)
		}
	}

	script := strings.Join(commands, "\n")
	script += "\n"
	return os.WriteFile(outputFile, []byte(script), 0o755)
}

func (fc *FileComparer) getFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (fc *FileComparer) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (fc *FileComparer) shouldIgnore(path string) bool {
	for _, pattern := range fc.IgnorePatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}


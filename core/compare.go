package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
)

type FileComparer struct {
	SourceDir      string
	TargetDir      string
	IgnorePatterns []string
	logger         logr.Logger
}

func NewFileComparer(
	ctx context.Context,
	sourceDir, targetDir string,
	ignorePatterns []string,
) *FileComparer {
	return &FileComparer{
		SourceDir:      sourceDir,
		TargetDir:      targetDir,
		IgnorePatterns: ignorePatterns,
		logger:         logr.FromContextOrDiscard(ctx),
	}
}

func (fc *FileComparer) GenerateComparisonScript(writer io.Writer) error {
	sourceFiles, err := fc.getFiles(fc.SourceDir)
	if err != nil {
		return fmt.Errorf("error getting source files: %w", err)
	}

	if _, err := io.WriteString(writer, "#!/usr/bin/env bash\n\n"); err != nil {
		return fmt.Errorf("error writing script header: %w", err)
	}

	for _, sourceFile := range sourceFiles {
		if err := fc.writeComparisonCommand(writer, sourceFile); err != nil {
			return fmt.Errorf("error writing comparison command: %w", err)
		}
	}

	return nil
}

func (fc *FileComparer) writeComparisonCommand(writer io.Writer, sourceFile string) error {
	relPath, err := filepath.Rel(fc.SourceDir, sourceFile)
	if err != nil {
		return fmt.Errorf("error getting relative path: %w", err)
	}

	if fc.shouldIgnore(relPath) {
		return nil
	}

	targetFile := filepath.Join(fc.TargetDir, relPath)

	if !fc.fileExists(targetFile) {
		return nil
	}

	absSourceFile, err := filepath.Abs(sourceFile)
	if err != nil {
		return fmt.Errorf("error getting absolute source path: %w", err)
	}

	absTargetFile, err := filepath.Abs(targetFile)
	if err != nil {
		return fmt.Errorf("error getting absolute target path: %w", err)
	}

	command := fmt.Sprintf(
		"diff --unified --ignore-all-space %q %q\n",
		absSourceFile,
		absTargetFile,
	)
	if _, err := io.WriteString(writer, command); err != nil {
		return fmt.Errorf("error writing diff command: %w", err)
	}

	return nil
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

func RunComparison(
	ctx context.Context,
	sourceDir, targetDir, outputFile string,
	ignorePatterns []string,
) error {
	logger := logr.FromContextOrDiscard(ctx)

	if sourceDir == "" || targetDir == "" {
		return fmt.Errorf("please provide both source and target directories")
	}

	comparer := NewFileComparer(ctx, sourceDir, targetDir, ignorePatterns)

	var writer io.Writer
	if outputFile == "-" {
		writer = os.Stdout
	} else {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("error creating output file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	if err := comparer.GenerateComparisonScript(writer); err != nil {
		return fmt.Errorf("error generating comparison script: %w", err)
	}

	if outputFile != "-" {
		logger.Info("Comparison script generated", "file", outputFile)
	}

	return nil
}

package cmd

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/giantswarm/frontmatter-validator/pkg/config"
	"github.com/giantswarm/frontmatter-validator/pkg/output"
	"github.com/giantswarm/frontmatter-validator/pkg/validator"
)

var (
	validationMode string
	outputFormat   string
	targetPath     string
	configPath     string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "frontmatter-validator",
	Short: "Validate Hugo frontmatter in Markdown files",
	Long: `An opinionated validator for HUGO frontmatter, designed for Giant Swarm requirements.

Frontmatter is metadata enclosed in Markdown files, like page title, description, and more.

The validator scans a target path recursively for Markdown files and validates their frontmatter
against a configurable set of rules, creating GitHub Actions run annotations for problems found.`,
	RunE: runValidation,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVar(&validationMode, "validation", "all", "Which validation to run. Use 'last-reviewed' or 'all'")
	rootCmd.Flags().StringVar(&outputFormat, "output", "stdout", "Output format: 'json' or 'stdout'")
	rootCmd.Flags().StringVar(&targetPath, "path", ".", "Target path to scan for Markdown files")
	rootCmd.Flags().StringVar(&configPath, "config", "./frontmatter-validator.yaml", "Path to configuration file")
}

func runValidation(cmd *cobra.Command, args []string) error {
	// Validate the validation mode flag
	if err := validateValidationMode(validationMode); err != nil {
		return err
	}

	// Load configuration
	configManager, err := config.NewManager(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create validator with configuration
	v := validator.NewWithConfig(configManager)
	formatter := output.New()
	results := make(map[string]validator.ValidationResult)

	// Get list of files to process
	filePaths, err := getFilesToProcess()
	if err != nil {
		return fmt.Errorf("failed to get files to process: %w", err)
	}

	// Process each file
	for _, filePath := range filePaths {
		if !fileExists(filePath) {
			continue
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not read file %s: %v\n", filePath, err)
			continue
		}

		result := v.ValidateFile(string(content), filePath, validationMode)
		if len(result.Checks) > 0 {
			results[filePath] = result
		}
	}

	// Output results
	switch outputFormat {
	case "json":
		formatter.PrintJSON(results)
	default:
		formatter.PrintStdout(results)
	}

	// Create annotations for GitHub Actions if running in CI
	if os.Getenv("GITHUB_ACTIONS") != "" {
		if err := formatter.DumpAnnotations(results); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not create annotations file: %v\n", err)
		}
	}

	return nil
}

// getFilesToProcess returns the list of files to validate
func getFilesToProcess() ([]string, error) {
	var filePaths []string

	// Check if we have input from stdin
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && strings.HasSuffix(line, ".md") {
				filePaths = append(filePaths, line)
				fmt.Printf("Adding to files checked: %s\n", line)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	} else {
		// Walk the target path
		err := filepath.WalkDir(targetPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && strings.HasSuffix(path, ".md") {
				filePaths = append(filePaths, path)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return filePaths, nil
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// validateValidationMode validates the validation mode flag value
func validateValidationMode(mode string) error {
	switch mode {
	case validator.ValidateAll, validator.ValidateLastReviewDate:
		return nil
	default:
		return fmt.Errorf("invalid validation mode '%s'. Valid options are: 'all', 'last-reviewed'", mode)
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zakaria-chahboun/cute"
)

var (
	rootCmd = &cobra.Command{
		Use:   "tarjem",
		Short: "CLI tool for translation management",
	}

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize translations.yaml files",
		Run: func(cmd *cobra.Command, args []string) {
			force, _ := cmd.Flags().GetBool("force")
			// Forced
			if force {
				err := createInitTranslationFile(DEFAULT_TRANSLATIONS_FILE_PATH, DEFAULT_TRANSLATIONS_FILE_DATA)
				cute.Check("Error", err)
				cute.Println(DEFAULT_TRANSLATIONS_FILE_PATH, "was created successfully.")
				os.Exit(1)
			}
			// Non-Forced: Check if the file exists
			if _, err := os.Stat(DEFAULT_TRANSLATIONS_FILE_PATH); !os.IsNotExist(err) {
				cute.Println(DEFAULT_TRANSLATIONS_FILE_PATH, "already exists. Use --force to overwrite.")
			} else {
				err := createInitTranslationFile(DEFAULT_TRANSLATIONS_FILE_PATH, DEFAULT_TRANSLATIONS_FILE_DATA)
				cute.Check("Error", err)
				cute.Println(DEFAULT_TRANSLATIONS_FILE_PATH, "was created successfully.")
			}
			os.Exit(1)
		},
	}

	exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export generated Go files",
		Run: func(cmd *cobra.Command, args []string) {
			lang, _ := cmd.Flags().GetString("lang")
			pkg, _ := cmd.Flags().GetString("package")

			if lang == "" {
				cute.Check("Error", fmt.Errorf("language must be specified using --lang"))
				cmd.Help()
				os.Exit(1)
			}

			// translations.yaml exists?
			_, err := os.Stat(DEFAULT_TRANSLATIONS_FILE_PATH)
			if err != nil {
				alert()
				os.Exit(1)
			}

			// Export logic
			if pkg != "" {
				EXPORTED_PACKAGE_NAME = pkg
			}
			ExportForProgrammingLanguage(lang)
			os.Exit(1)
		},
	}

	clearCmd = &cobra.Command{
		Use:   "clear",
		Short: "Remove the exported translations.go file",
		Run: func(cmd *cobra.Command, args []string) {
			// Check if the file exists
			if _, err := os.Stat(EXPORTED_TRANSLATIONS_FILE); os.IsNotExist(err) {
				cute.Println("No exported files to remove.")
				os.Exit(1)
			}
			// Remove the file
			err := os.Remove(EXPORTED_TRANSLATIONS_FILE)
			if err != nil {
				cute.Check("Error removing file", err)
				os.Exit(1)
			}
			cute.Println("Successfully removed", EXPORTED_TRANSLATIONS_FILE)
			os.Exit(1)
		},
	}

	helpCmd = &cobra.Command{
		Use:   "help",
		Short: "Show help information",
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.Help()
			os.Exit(1)
		},
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			cute.Println("tarjem version", version)
			os.Exit(1)
		},
	}
)

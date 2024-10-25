package cmd

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/gkwa/creepymatrix/core"
	"github.com/gkwa/creepymatrix/internal/logger"
)

var (
	cfgFile        string
	verbose        int
	logFormat      string
	cliLogger      logr.Logger
	sourceDir      string
	targetDir      string
	outputFile     string
	ignorePatterns []string
)

var rootCmd = &cobra.Command{
	Use:   "creepymatrix",
	Short: "A brief description of your application",
	Long:  `A longer description that spans multiple lines and likely contains examples and usage of using your application.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cliLogger = logger.NewConsoleLogger(verbose, logFormat == "json")
		cmd.SetContext(logr.NewContext(cmd.Context(), cliLogger))
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return core.RunComparison(cmd.Context(), sourceDir, targetDir, outputFile, ignorePatterns)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().
		StringVar(&cfgFile, "config", "", "config file (default is $HOME/.creepymatrix.yaml)")
	rootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "increase verbosity")
	rootCmd.PersistentFlags().
		StringVar(&logFormat, "log-format", "", "json or text (default is text)")

	rootCmd.Flags().StringVar(&sourceDir, "source", "", "Source project directory")
	rootCmd.Flags().StringVar(&targetDir, "target", "", "Target project directory")
	rootCmd.Flags().
		StringVar(&outputFile, "output", "-", "Output bash script file (use '-' for stdout)")
	rootCmd.Flags().StringArrayVar(&ignorePatterns, "ignore", []string{
		".git/",
		".nearwait.",
		".timestamps/",
		"go.mod",
		"go.sum",
		".gitignore",
		"make_txtar.sh",
		"node_modules",
		"README.md",
	}, "Patterns to ignore (can be used multiple times)")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		fmt.Printf("Error binding persistent flags: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		fmt.Printf("Error binding flags: %v\n", err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".creepymatrix")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

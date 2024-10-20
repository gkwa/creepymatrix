package cmd

import (
	"context"
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
		if cliLogger.IsZero() {
			cliLogger = logger.NewConsoleLogger(verbose, logFormat == "json")
		}

		ctx := logr.NewContext(context.Background(), cliLogger)
		cmd.SetContext(ctx)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if sourceDir == "" || targetDir == "" {
			return fmt.Errorf("please provide both source and target directories")
		}

		comparer := core.NewFileComparer(sourceDir, targetDir, ignorePatterns)
		err := comparer.GenerateComparisonScript(outputFile)
		if err != nil {
			return fmt.Errorf("error generating comparison script: %v", err)
		}

		fmt.Printf("Comparison script generated: %s\n", outputFile)
		return nil
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
	rootCmd.Flags().StringVar(&outputFile, "output", "compare_files.sh", "Output bash script file")
	rootCmd.Flags().StringArrayVar(&ignorePatterns, "ignore", []string{
		".git",
		"go.mod",
		"go.sum",
		"make_txtar.sh",
		"node_modules",
		"README.md",
	}, "Patterns to ignore (can be used multiple times)")

	if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		fmt.Printf("Error binding verbose flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("log-format", rootCmd.PersistentFlags().Lookup("log-format")); err != nil {
		fmt.Printf("Error binding log-format flag: %v\n", err)
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

	logFormat = viper.GetString("log-format")
	verbose = viper.GetInt("verbose")
}

func LoggerFrom(ctx context.Context, keysAndValues ...interface{}) logr.Logger {
	if cliLogger.IsZero() {
		cliLogger = logger.NewConsoleLogger(verbose, logFormat == "json")
	}
	newLogger := cliLogger
	if ctx != nil {
		if l, err := logr.FromContext(ctx); err == nil {
			newLogger = l
		}
	}
	return newLogger.WithValues(keysAndValues...)
}

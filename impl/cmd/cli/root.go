package cli

import (
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.diddht/diddht.json)")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Error("failed to execute root command")
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "diddht",
	Short: "diddht is a command line tool for interacting with resolvable sovereign keys.",
	Long:  `diddht is a command line tool for interacting with resolvable sovereign keys on the mainline DHT.`,
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			logrus.WithError(err).Error("failed to find home directory")
			os.Exit(1)
		}

		// Search config in home directory with name ".diddht" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".diddht")
	}

	// read in environment variables that match
	viper.AutomaticEnv()

	// if a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logrus.Info("Using config file:", viper.ConfigFileUsed())
	}
}

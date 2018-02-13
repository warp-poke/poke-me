package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// Catalyst init - define command line arguments.
func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file to use")
	RootCmd.Flags().IntP("log-level", "v", 4, "Log level (from 1 to 5)")
	RootCmd.Flags().StringP("listen", "l", "127.0.0.1:9100", "listen address")

	viper.BindPFlags(RootCmd.Flags())
}

// Load config - initialize defaults and read config.
func initConfig() {

	// Default

	// Bind environment variables
	viper.SetEnvPrefix("poke-me")
	viper.AutomaticEnv()

	// Set config search path
	viper.AddConfigPath("/etc/poke-me/")
	viper.AddConfigPath("$HOME/.poke-me")
	viper.AddConfigPath(".")

	// Load config
	viper.SetConfigName("config")
	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Debug("No config file found")
		} else {
			log.Panicf("Fatal error in config file: %v \n", err)
		}
	}

	// Load user defined config
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		err := viper.ReadInConfig()
		if err != nil {
			log.Panicf("Fatal error in config file: %v \n", err)
		}
	}

	log.SetLevel(log.AllLevels[viper.GetInt("log-level")])
}

// RootCmd launch the aggregator agent.
var RootCmd = &cobra.Command{
	Use:   "poke-me",
	Short: "Poke-me Warp10 runner deploy hook",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Poke-me starting")
	},
}

package cmd

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/warp-poke/poke-me/core"
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
	viper.SetDefault("listen", "127.0.0.1:8080")
	viper.SetDefault("secrets", map[string]string{})

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

		sshKey := viper.GetString("cloner.ssh.key")
		if sshKey == "" {
			log.Fatal("Cannot do clone actions without SSH KEY")
		}

		gitURL := viper.GetString("cloner.git.url")
		if gitURL == "" {
			log.Fatal("Cannot clone repository without URI")
		}

		clonePath := viper.GetString("cloner.path")
		if clonePath == "" {
			log.Fatal("Cannot clone repository without path")
		}

		gitSecret := viper.GetString("cloner.git.secret")
		if gitSecret == "" {
			log.Fatal("Cannot clone repository without git secret")
		}

		zkServers := viper.GetStringSlice("zk.servers")
		if len(zkServers) == 0 {
			log.Fatal("Cannot connect to ZK without servers")
		}

		zk, err := core.NewZK(zkServers, time.Second*10)
		if err != nil {
			log.Panic(err)
		}

		znode, err := zk.ZNode("/poke-me/commit-id")
		if err != nil {
			log.Panic(err)
		}

		/*for v := range znode.Values {
			log.Info(string(v))
			time.Sleep(time.Duration(rand.Int63n(1500)) * time.Millisecond)
			time.Sleep(1000 * time.Millisecond)

			znode.Update([]byte(time.Now().Format(time.RFC3339)))
		}*/

		c, err := core.NewCloner(sshKey, gitURL, clonePath)
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			for v := range znode.Values {
				if len(v) == 0 {
					continue
				}

				if err := c.Clone(string(v), viper.GetStringMapString("secrets"), false); err != nil {
					log.WithError(err).Error("Failed to clone")
				}
			}
		}()

		http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
			body, err := github.ValidatePayload(r, []byte(gitSecret))
			if err != nil {
				log.WithError(err).Warn("Bad payload")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
			}

			var hook core.WebHookPayload
			if err := json.Unmarshal(body, &hook); err != nil {
				log.WithError(err).Warn("Failed to unmarshal")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
			}
			// Send update on ZK
			sha := hook.HeadCommit.ID
			if sha == nil {
				log.Error("Failed to get commit sha")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to get commit sha"))
			}

			if err := znode.Update([]byte(*sha)); err != nil {
				log.WithError(err).Warn("Failed to set ZK value")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
		})

		log.Fatal(http.ListenAndServe(viper.GetString("listen"), nil))
	},
}

package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

const ProjectName = "supraworker"

type Config struct {
	ClientId        string        `mapstructure:"clientId"`
	JobsAPI         ApiOperations `mapstructure:"jobs"`
	LogsAPI         ApiOperations `mapstructure:"logs"`
	CallAPIDelaySec int           `mapstructure:"api_delay_sec"`
}

type ApiOperations struct {
	Get    UrlConf `mapstructure:"get"`    // defines how to get item
	Lock   UrlConf `mapstructure:"lock"`   // defines how to lock item
	Update UrlConf `mapstructure:"update"` // defines how to update item
	Unlock UrlConf `mapstructure:"unlock"` // defines how to unlock item
	Finish UrlConf `mapstructure:"finish"` // defines how to finish item
}

type UrlConf struct {
	Url             string            `mapstructure:"url"`
	Method          string            `mapstructure:"method"`
	Headers         map[string]string `mapstructure:"headers"`
	PreservedFields map[string]string `mapstructure:"preservedfields"`
	Params          map[string]string `mapstructure:"params"`
}

var (
	CfgFile string
	C       Config = Config{
		CallAPIDelaySec: int(2),

		// JobsAPI: UrlConf{
		//     Method: "GET",
		//     Headers: []RequestHeader{
		//         RequestHeader{
		//             Key: "Content-type",
		//             Value: "application/json",
		//         },
		//     },
		// },
	}
	log = logrus.WithFields(logrus.Fields{"package": "config"})
)

// Init configuration
func init() {

	// configCMD.PersistentFlags().StringVar(&CfgFile, "config", "", "config file (default is $HOME/supraworker.yaml)")
	// viper.SetDefault("license", "apache")
	// configCMD.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	// viper.Set("Verbose", true)
	cobra.OnInitialize(initConfig)
	// rootCmd.AddCommand(configCMD)

}

// var configCMD = &cobra.Command{
// 	Use:   "config",
// 	Run: func(command *cobra.Command, args []string) {
//         log.Debug("viper config file:", viper.ConfigFileUsed())
// 	},
// }

func initConfig() {
	// Don't forget to read config either from CfgFile or from home directory!
	// log.Info("logrus")
	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
	} else {

		lProjectName := strings.ToLower(ProjectName)
		log.Debug("Searching for config with project", ProjectName)
		viper.AddConfigPath(".")
		viper.AddConfigPath("..")
		viper.AddConfigPath("$HOME/")
		viper.AddConfigPath(fmt.Sprintf("$HOME/.%s/", lProjectName))
		viper.AddConfigPath("/etc/")
		viper.AddConfigPath(fmt.Sprintf("/etc/%s/", lProjectName))
		if conf := os.Getenv(fmt.Sprintf("%s_CFG", strings.ToUpper(ProjectName))); conf != "" {
			viper.SetConfigName(conf)
		} else {
			viper.SetConfigType("yaml")
			viper.SetConfigName(lProjectName)
		}
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatal("Can't read config:", err)
	}
	err := viper.Unmarshal(&C)
	if err != nil {
		logrus.Fatal(fmt.Sprintf("unable to decode into struct, %v", err))

	}
	log.Debug(viper.ConfigFileUsed())

}

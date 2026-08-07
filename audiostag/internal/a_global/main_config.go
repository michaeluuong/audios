package a_global

import (
	"os"

	"github.com/michaeluuong/audios/audiostag/internal/pkg/config"
)

const (
	mainEnvVar         = "CONFIG_FILE"
	mainConfigFilename = "audiostag_cfg.json"
	mainProgName       = "audios"
)

// MainConfig represents all available primary configurations.
type MainConfig struct {
	Add_source              bool `json:"add_source"` // true adds filename/method name to the log line
	Dev                     bool `json:"dev"`        // true alters behaviors convenient for development
	Log_func                bool `json:"log_func"`   // true prints log lines is json format
	Log_json                bool `json:"log_json"`   // true runs the ReplacAttr function (e.g. Add_source)
	config.DefaultConfigger      // Default implemention of Configger
}

func (m *MainConfig) SetupConfig() error {
	if err := m.DefaultConfigger.DefaultSetupConfig(m); err != nil {
		return err

	}

	return nil

}

func NewMainConfig() *MainConfig {
	progName := os.Getenv("PROGNAME")
	if progName == "" {
		progName = mainProgName

	}

	newMainConfig := &MainConfig{
		DefaultConfigger: config.DefaultConfigger{
			DefaultConfigFilename: mainConfigFilename,
			DirName:               progName,
			EnvVar:                mainEnvVar,
			//IsMandatory:           true,
		},
	}
	newMainConfig.SetupConfig()

	return newMainConfig

}

func NewMainConfigMan() *MainConfig {
	newMainConfig := NewMainConfig()
	config.GetConfigManInstance().AddConfig(newMainConfig)

	return newMainConfig

}

// ToMainConfig return a config.Configger object as type *MainConfig.
//   - configger is the Configger object to type assert to *MainConfig
//
// Return the configger object as *MainConfig or a NewMainConfig() if configger was nil.
func ToMainConfig(configger config.Configger) *MainConfig {
	m, ok := configger.(*MainConfig)
	if !ok {
		return NewMainConfig()

	}

	return m

}

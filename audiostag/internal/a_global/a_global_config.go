// Package global performs operations whose results are available globally
package a_global

import (
	"os"

	"github.com/michaeluuong/audios/audiostag/internal/pkg/config"
)

var _ = os.Setenv("PROGNAME", "audiostag")
var ConfigMan *config.ConfigMan = config.GetConfigManInstance()

func init() {
	config.NewSlogConfigMan()
	AudiostagConfigMan()
	//NewMainConfigMan()

}

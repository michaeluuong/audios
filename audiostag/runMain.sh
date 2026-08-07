## export GOFLAGS="-ldflags=-s"
## export GOTMPDIR=$(mktemp -d -t gotmpdir)
## export GOTMPDIR=$(mktemp -d /tmp/gotmpdir)
## export GOCACHE="$HOME/.cache/go-build2"
## mkdir -p $GOCACHE
//export GOTMPDIR=$HOME/go-tmp
//mkdir -p $HOME/go-tmp

FILE="$1"; shift
## CONFIG_FILE=`pwd`/config/audiostag_cfg.json DEV=true LOG_FUNC=true  go run cmd/main.go cmd/main_config.go
## DEV=true CONFIG_FILE=`pwd`/config/audiostag_cfg.json LOG_FUNC=true OUT_FILE=/Users/irenepiechota-wong/DEV/GO/audios/audiostag/log/audios.log  go run cmd/main.go cmd/main_config.go "$@"
## DEV=true CONFIG_FILE=`pwd`/config/audiostag_cfg.json LOG_FUNC=true OUT_FILE=/Users/irenepiechota-wong/DEV/GO/audios/audiostag/log/audios.log  go run -x -ldflags="-s -w" cmd/main.go cmd/main_config.go
## DEV=true CONFIG_FILE=`pwd`/config/audiostag_cfg.json LOG_FUNC=true OUT_FILE=/Users/irenepiechota-wong/DEV/GO/audios/audiostag/log/audios.log  go run cmd/main.go "$@"
## go run cmd/main.go --dir "/Users/irenepiechota-wong/DEV/GO/audios/audiostag/audiofiles" --file "$FILE" "$@"

##rm audiostag.log; go run -ldflags="-w -s" cmd/main.go --file "$FILE" "$@"
rm audiostag.log; go run -x cmd/main.go --file "$FILE" "$@"

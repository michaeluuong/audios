TAR_FILE="$1"
DEV=true go run -x cmd/main.go cmd/main_config.go --extract  -d"/Users/irenepiechota-wong/DEV/GO/audios/audiostag/audiofiles" -f"$TAR_FILE"

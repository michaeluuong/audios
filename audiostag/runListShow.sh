TAR_File="$1"
go run cmd/main.go --dir "/Users/irenepiechota-wong/DEV/GO/audios/audiostag/audiofiles" --file "$TAR_FILE" --list --show-archive-tags "$@"

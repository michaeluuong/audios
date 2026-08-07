TAR_FILE="$1"; shift


## DEV=true go run -ldflags="-w -s" cmd/main.go --album -d"/Users/irenepiechota-wong/DEV/GO/audios/audiostag/audiofiles" -f"$TAR_FILE"
## DEV=true go run -ldflags="-w -s" cmd/main.go --album -d"/Users/irenepiechota-wong/DEV/GO/audios/audiostag/audiofiles" -f"Marcus_King-Darling_Blue__No_Room_For_Blue-WEB-2026-QUAVER.tar" --tag "Artist=The Marcus King Band|year=2005|rool=rain" --case "title"
## DEV=true go run -ldflags="-w -s" cmd/main.go --album -d"/Users/irenepiechota-wong/DEV/GO/audios/audiostag/audiofiles" -f"VA-Riverside_Jazz_Keynote_Recordings_From_One_Of_Jazzs_Greatest_Label_1953-1964-2006-NOiR.tar"  --tag "Artist=Riverside Jazz"
## rm audiostag.log; DEV=true go run -ldflags="-w -s" cmd/main.go --album -d"/Users/irenepiechota-wong/DEV/GO/audios/audiostag/audiofiles" -f"$TAR_FILE" "$@"
## rm audiostag.log; DEV=true go run -ldflags="-w -s" cmd/main.go  -d"/Users/irenepiechota-wong/DEV/GO/audios/audiostag/audiofiles/Kacey Musgraves/(2026) Middle of Nowhere" "$@"

rm audiostag.log; DEV=true go run -ldflags="-w -s" cmd/main.go --album -f"$TAR_FILE" "$@"

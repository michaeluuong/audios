# audios 
Provides popularly unused applications for audio and bit of visuals. 
## audiostag
The sole audios provision that aims to exist as a universal tagging automaton, but relegated to insurecting manual taggage.
Built on a foundation of OS independence but birthed for the mac; both quality and assurance has been aborted for non Darwin operating systems.  
While it is possible to modify tags on single files (with the proper regular expressions and action choices) this was meant to tag full albums.

## go build -o audiostag ./cmd/...
You can name it whatever you want, but if consistency is important you'll also need to code out the annoying little dependencies that I've neglected to account for.

This can be used as a command line interface but it's use was always predicated on being a right-click action
- macOS Quick Action
- Windows context menu (not added to regular programming)

## macOS Quick Action Shortcuts
audiostag provides the ability to create two Shortcuts but instructions are provided for manual creation.
<img width="637" height="457" alt="image" src="https://github.com/user-attachments/assets/10958236-b770-48eb-9c6d-3667019b9682" />

In an unfortunate outcome of wasted time investment, I could not figure out how to programmatically add items to the Finder Quick Actions menu.
The Shortcut will either need to be manually edited to check "Finder" under "Use as Quick Action" in the Details tab of the "Shortcut Details" section, or 
use audiostag to add the same Shortcut twice (replacing the original) which magically checks the Finder box (at least for me it does).
The created Shortcut will point to the location that you've placed the audiostag binary.

<img width="862" height="610" alt="image" src="https://github.com/user-attachments/assets/62837c82-7bb9-4dd7-8dc3-72ec55770a3f" />

<img width="259" height="294" alt="image" src="https://github.com/user-attachments/assets/5f7eab4b-89ee-44b4-9a83-faf36b4a2b07" />

1. The primary action of audiostag is a show-tags Shortcut which does all the mp3 tag viewing and modifying
   - audiostag --create-show-tags-shortcut
       - This will open a window where you will need to click on "Add Shortcut"
   <img width="532" height="740" alt="image" src="https://github.com/user-attachments/assets/bf479379-d03f-4bc0-b263-3cf95da12c68" />

   - If you chose to add the Shortcut twice you need to click on Replace the second time
   <img width="258" height="290" alt="image" src="https://github.com/user-attachments/assets/a1cd0b02-c345-4b33-b33c-88deb01273a2" />


3. The secondary action is the extract-archive Shortcut which will extract/uncompress archives and compressed files. 
    - audiostag --create-extract-archive-shortcut
        - Extract files from .tar files
        - Extract split .rar files from a .tar file and automatically uncompress the .rar split archive
    <img width="532" height="740" alt="image" src="https://github.com/user-attachments/assets/123ad927-dadd-405b-a414-b00f414ed4bb" />
 

# Manual Shortcut Creation
1. Open the Shortcuts app
2. Click on + at the top of the window to create a new shortcut
3. Name the shortcut at the top where it says Title (e.g. show-tags)
4. Click on the "Shortcut Details" icon (an i in a circle to the right)
    - check "Use as Quick Action"
    - check Finder
    - uncheck "Services Menu" if you don't want it in the right-click Services menu
5. In the Receive box
    - click to the right of Receive and unselect everything that's chosen except for files and folders
    - "input from" should remain "Quick Actions"
    - "If there's no input:"
      - click on "Continue" and choose "Ask For" and then choose "Files"
    - click "Show More"
        - choose "Shortcut Input" for "Type"
    <img width="247" height="343" alt="image" src="https://github.com/user-attachments/assets/0dda53fe-bb35-42a1-9735-d398c7dc2dd5" />
    <img width="862" height="512" alt="image" src="https://github.com/user-attachments/assets/3a088d14-f28f-4e42-a06d-3e61324d554e" />

6. Go back to "Action Library" by clicking the .+ folder icon to the right
    - search for "Run Shell Script" and drag it into your workflow (i.e. right under the Receive box)
7. In "Run Shell Script"
    - add the full path to the audiostag binary which you can find in /audios/audiostag
      - e.g. ~/audiostag --path ""
      - follow ~/audiostag with the --path flag and two double quotes then right-click in the middle of the two quotes
        - expand "Insert Variable" and choose "Shortcut Input"
        - click on "Shortcut Input"
          - choose "File" for "Type"
          - choose "File Path" for "Get"
      - add the action and GUI flags to the end
          - --show-tags-chooser --out window
          - --extract --out window
   <img width="333" height="271" alt="image" src="https://github.com/user-attachments/assets/9b3b35e6-63cf-4a22-a983-a4abea994424" />
   
    - choose sh for "Shell"
    - choose Input for "Input"
    - choose "as arguments" for "Pass Input"
  <img width="862" height="578" alt="image" src="https://github.com/user-attachments/assets/a48f9cd9-0624-4747-acb8-a859826195fe" />

 
You should now be able to right-click on a .tar file that contains an MP3 album or movie and choose your shortcut in the "Quick Actions" menu. 
Upon first use you will have allow audiostag to run in Privacy settings where I would choose "Always Allow"
<img width="512" height="520" alt="image" src="https://github.com/user-attachments/assets/17debd66-6ed6-49a7-9d89-a8a4507a7824" />

<img width="1796" height="792" alt="image" src="https://github.com/user-attachments/assets/26590047-fbfc-44fd-854b-001f97039d15" />


# How to use audiostag
- The goal is to modify all the tracks in an album with minimal intervention.
- You can start with an archive (.tar) file or directory that contains an "album".
- Configurable options live in ~/Library/Application Support/audiostag/config/audiostag_cfg.json
```
{
  "cli": {
    "starting_directory": "!/audios/audiostag/audiofiles"
  },
  "archive": {
    "exclude_file_regex": "\\.(nfo|sfv|m3u)$|.*proof.*\\.(jpg|png)", // do not extract these files from archives
    "extract_concurrent": 1,                                         // extract this many archives concurrently
    "trash_archive": false,                                          // automatically delete archive file after extraction
    "extract_rename": {
      ".gif": "cover",                                               // rename files with this key to this value
      ".jpg": "cover",
      ".png": "cover",
      "exclude": ".*proof.*"                                         // exclude this file from extraction
    }
  },
  "mp3": {
    "artist_to_title": "with",                                 // if there are multiple comma-separated artist names in the artist tag move them to the title field after with e.g. (with B.I.G.)
    "cover_sleep_time": 0,                                     // sleep this amount of time after hitting the musicbrainz webservice for album art (they have a rate limit)
    "cover_size": "large",                                     // use this sized image from coverartarchive.com (T250, Small, T500, T1200, Large, Front)
    "remove_genre": true,                                      // always remove the genre tag      
    "remove_disc_1_of_1": true,                                // if the "Disc Number" field has 1/1 remove it
    "remove_bonus_track": true,                                // remove the words (Bonus Track) from the title because WTF is that
    "featuring_fix": "feat",                                   // change anything in the title that resembles featuring to this e.g. feat Marlon Brando
    "featuring_paren": true,                                   // surround the featuring_fix phrase with parentheses
    "file_replace_spaces": false,                              // replace spaces in the filename with underscores
    "file_rename_exp": "%02s{track number}. %{title}",         // rename the file to this expression
    "album_folder_exp": "(%{year}) %{album}",                  // rename the album folder to this expression
    "artist_folder_exp": "%{artist}",                          // rename the artist folder to this expression
    "multi_disc_album_add": "",                                // the default is CD e.g. Nevermind (CD 1)
    "playlist_exp": "%{artist} - %{album}",                    // create the playlist with this expression
    "various_artists_playlist_exp": "%{album}",                // if the album is a various artist album use this expression for the playlist
    "cover_filename_no_ext": "cover",                          // if we find cover art name it with this name
    "total_tracks": false,                                     // if true set Track Number tag to track number/total track number
    "tag_replacements": "",                                    // always make these replacements (regex1|replacement string1~regex2|replecement string2~...)
    "show_tags": true,                                         // show tags in a window (i.e. instead of text in a terminal)
    "max_file_limit": 0                                        // stop processing if the folder has this many files (0 is unlimited)
  }
}
```
- Most tag changes can be done any time except for the Album name which needs to be done up front
    - To properly change the Album name and have all the directories created correctly is to use a pre-action which changes the tag up front
        - Pre Condition
        - Pre Replace
- If the artist names in the Artist tag field are mostly different (i.e. 80%) the album will be considered a various artist album and will move the artist to the title (i.e. {Title} - {Artist})
  and the Artist field will copy the name of the album
- You can click on table fields to enter them into the text fields for each action
- If there is a picture file in the archive or folder it will be added to the picture field of the track
   - If it's an archive you can exclude the picture from being extracted and choosing the "Exclude File" action (e.g. Exclude File .*\.jpg)
- Characters within regular expressions that shouldn't be considered part of the expression should be escaped with a single backslash

# audiostag actions
- Album Folder (album-folder-name)
   - use this "folder name" to name the album folder (i.e. directory immediately enclosing album)
<img width="1183" height="464" alt="image" src="https://github.com/user-attachments/assets/1e4b5c92-a0b6-44b2-89d3-91c149335e03" />

- Cover Album (--cover-album)
   - search musicbrainz with this album name (instead of whats in the album tag)
- Cover Artist (--cover-artist)
   - search musicbrainz with this artist name (instead of whats in the artist tag)
- Dummy Files (--dummy-files)
   - creates empty mp3 files, copies the tags, and archives
- Exclude File (--extract-exclude)
   - do not extract this file or file(s) matching this regular expression
      - characters in the regular expression that match regular expression characters need to be escaped e.g. ".*\.jpg"
<img width="1066" height="81" alt="image" src="https://github.com/user-attachments/assets/8e266f10-14a2-45a4-87b8-b91739c70688" />

- Keep Tag (--keep-tag)
   - leave this tag in the file because audiostag automatically removes non-required tags 
- Playlist Name (--playlist-name)
   - name the playlist file 
- Precondition (--precondition)
   - at the beginning of the process, if the "source tag" matches this "regex", set the "destination tag" to the "match value" otherwise set it to the "else value"
   - the Album tag can be set here
   - hides the only non-intuitive/secret expression in audiostag
<img width="1193" height="82" alt="image" src="https://github.com/user-attachments/assets/a0e4dc8f-858c-48d5-998d-f476a37a35d2" />

- Prereplace
   - at the beginning of the process, if the "tag" matches the "regex" set it to the "replace value"
   - the Album tag can be set here
<img width="1195" height="85" alt="image" src="https://github.com/user-attachments/assets/0bf0a83e-cb99-4151-abdc-d7c6ef076e1b" /> 
   
- Single Artist (--single-artist)
   - treat this album as a single artist album (i.e. if audiostag has incorrectly determined that the album is a "Various Artist" album, correct it)
- Tag (--tag)
   - set the "source tag" to TRCK and provide a range (e.g. 1-10) and set the "destination tag" to the "match value" else "else value" (helpful to set the disc number; despite an inherent underlying quality of annoyance)
   - You can also use this to set the cover art (i.e. Picture frame) (musicbrainz doesn't always work out)
      - set tag to Picture by clicking on a Picture/APIC cell
      - set value to a link that points to a picture (.jpg, .png, .gif) or a local image file of the same types
<img width="1193" height="83" alt="image" src="https://github.com/user-attachments/assets/2d3b3fe0-4b24-4c1a-8771-1502e01921cf" />
            
- Various Artists (--various)
   - treat this album as a "Various Artists" album (i.e. if audiostag has incorrectly determined that the album is a "Single Artist" album, correct it)


# audiostag CLI flags
- --precondition "source tag=regex=destination tag=expression to place in matching tags=expression to place in all other tags~Source Tag2..."
   - --precondition "Track Number=1-10=Disc Number=1/2=2/2"
   - also see audiostag options for Precondition
- --prereplace "tag=regex=replace value\~tag2=regex2=replace value2\~..."
   - --prereplace "TALB= \\(Deluxe Edition\\)=" (you can also use the tag name "Album")
   - also see audiostag options for Prereplace
- --replace|-r "tag=regexp=replacement value\~tag2=regexp2=replacement value2\~..."
   - -r "Title= \\(Deluxe [^)]*\\)="
   - also see audiostag options for Replace
- --tag|-t "tag=tag value\~tag2=tag value2\~..."
   - --tag "Genre=Technical Priapism"
   - also see audiostag options for Tag
- --keep-tag|-k "tag field1\~tag field2\~..."
   - --keep-tag "TPOS\~TCOM\~TPUB"
   - also see audiostag options for Keep Tag

- --dir|-d "absolute path"
   - look for files in this directory (absolute path)
- --extract-exclude "regex"
   - --extract-exclude ".*\\.png"
   - also see audiostag options for Exclude File
- --file|-f "filename"
   - look for this filename
- --out|-o "filename|window"
   - specify a filename to write tags to
      - --out album_tags.txt
   - window to write data to a nice little table
      - --out window
- --path|-p "absolute path"
   - the full path to the album we want to tag
  
- --album-folder-name "Album Folder Name"
   - --album-folder-name "Ibiza Hallucinations Volume 69"
   - also see audiostag options for Album Folder
- --case "CaseType"
   - change the case of Title, Artist & Album
      - FoldCase to fold to lower case
      - LowerCase to change all letters to lowercase
      - SentenceCase to capitalize the first letter of each line
      - TitleCase to capitalize the first letter of each word
      - UpperCase to capitalize all letters
   - --case "TitleCase"
- --cover-album "album name"
   - --cover-album "Mellin Collie and Sadness Ad Infinitum"
   - also see audiostag options for Cover Album
- --cover-artist "artist name"
   - --cover-artist "Smashed Winter Squash of the Orange Varietal"
   - also see audiostag options for Cover Artist
- --cover "Artist|Album"
   - use instead of separate --cover-album and --cover-artist to get cover art from musicbrainz/coverartarchive
- --cover-source|-c "URL|Path to local file"
   - instead of from a local file or calling out to musicbrainz, use the cover art from URL or path to local file
- --multi-disc
   - process this album as though it has multiple discs
- --playlist-name "name of playlist"
   - see audiostag options for Playlist Name
- --single-artist
   - see audiostag options for Single Artist
- --various
   - see audiostag options for Various Artists

- --album
   - extract files from archive (i.e. .tar file)
   - tag the extracted files with default values
- --dummy-files
   - copy all files in archive or folder to empty .mp3 files and copy tags
- --extract|-e
   - extract files from archive
   - if the archive contains split .rar files the .rar files will also be extracted
- --list|-l
   - list all tags in file to the terminal
- --no-artist-split
   - do not separate comma-separated artists values
- --no-directory-rename
   - do not rename any directories (this is especially helpful when you know audiostag is going to get it wrong)
- --playlist
   - create a playlist file for all files in the specified folder
- --print-tags
   - display tags 
- --set-default-tags|-s
   - sets all tags in directory to default values
- --show-archive-tags
   - display tags from files in specified archive file
- --show-tags
   - display tags from files in specified directory
- --show-tags-chooser
   - display tags in archive or folder

- --create-extract-archive-shortcut
   - creates a macOS Shortcut that triggers archive extraction in the "Quick Actions" menu of a Finder right-click
- --create-show-tags-shortcut
   - creates a macOS Shortcut that triggers the display of tags within a folder or archive from the "Quick Actions" menu of a Finder right-click 

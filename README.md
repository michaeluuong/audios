# audios provides popularly unused applications for audio. 
## audiostag is the current sole audios provision, aiming to exist just shy of a universal tagging automaton.
I've written this with the hopes of being OS independent but really just needed something for macOS and have never tested  
this elsewhere.

## go build -o audiostag cmd/main.go
You can name it whatever you want but nasty little logs and dependencies rely on it's given title.

You can solely use the command line interface but it's use was always predicated on being a right-click action 
- macOS Quick Action
- Windows context menu

## macOS Quick Action instructions
1. Open the Shortcuts app
2. Click on + at the top of the window to create a new shortcut
3. Name the shortcut at the top where it says Title (e.g. show-archive-tags)
4. Click on the "Shortcut Details" icon (an i in a circle to the right)
    - check "Use as Quick Action"
    - check Finder
    - uncheck "Services Menu"
5. In the Receive box
    - click to the right of Receive and unselect everything that's chosen except for files and folders
    - input from should remain "Quick Actions"
    - "If there's no input:"
      - click on "Continue" and choose "Ask For" and then choose "Files"
6. Go back to "Action Library" by clicking the .+ folder icon to the right
    - search for "Run Shell Script" and drag it into your workflow (i.e. right under the Receive box)
7. In "Run Shell Script"
    - add the full path to the runShowArchiveTags.sh script which is wherever the project was built
      - e.g. ~/audios/audiostag/SCRIPTS/runShowArchiveTags.sh
      - follow the script with two double quotes and right-click in the middle of the two quotes
        - expand "Insert Variable" and choose "Shortcut Input"
        - click on "Shortcut Input"
          - choose "File" for "Type"
          - choose "File Path" for "Get"
    - choose sh for "Shell"
    - choose Input for "Input"
    - choose "as arguments" for "Pass Input"
  
You should now be able to right-click on a .tar file and choose your shortcut in the "Quick Actions" menu. 

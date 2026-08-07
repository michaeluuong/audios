# 1. Define variables
NAME="show-tags"
SERVICEDIR="$HOME/Library/Services"
WORKFLOWPATH="$SERVICEDIR/$NAME.workflow"
CONTENTS="$WORKFLOWPATH/Contents"
DOCUMENT="$CONTENTS/document.wflow"

# 2. Make directory structure
mkdir -p "$CONTENTS"

# 3. Create the Automator XML document
cat << 'EOF' > "$DOCUMENT"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://apple.com">
<plist version="1.0">
<dict>
	<key>AMApplicationBuild</key>
	<string>521.1</string>
	<key>AMApplicationVersion</key>
	<string>2.10</string>
	<key>AMDocumentVersion</key>
	<string>2</string>
	<key>actions</key>
	<array>
		<dict>
			<key>action</key>
			<dict>
				<key>AMAccepts</key>
				<dict>
					<key>Container</key>
					<string>List</string>
					<key>Optional</key>
					<true/>
					<key>Types</key>
					<array>
						<string>com.apple.cocoa.path</string>
					</array>
				</dict>
				<key>AMActionVersion</key>
				<string>2.1.1</string>
				<key>AMApplication</key>
				<array>
					<string>Automator</string>
				</array>
				<key>AMParameterProperties</key>
				<dict>
					<key>COMMAND_STRING</key>
					<dict/>
					<key>CheckedForUserDefaultShell</key>
					<dict/>
					<key>inputMethod</key>
					<dict/>
					<key>shell</key>
					<dict/>
					<key>source</key>
					<dict/>
				</dict>
				<key>AMProvides</key>
				<dict>
					<key>Container</key>
					<string>List</string>
					<key>Types</key>
					<array>
						<string>com.apple.cocoa.string</string>
					</array>
				</dict>
				<key>ActionBundlePath</key>
				<string>/System/Library/Automator/Run Shell Script.action</string>
				<key>ActionName</key>
				<string>Run Shell Script</string>
				<key>ActionParameters</key>
				<dict>
					<key>COMMAND_STRING</key>
					<string>for f in "$@"
do
    # Put your shell script logic here
    echo "Processing: $f"
done</string>
					<key>CheckedForUserDefaultShell</key>
					<true/>
					<key>inputMethod</key>
					<integer>1</integer>
					<key>shell</key>
					<string>/bin/zsh</string>
					<key>source</key>
					<string></string>
				</dict>
				<key>BundleIdentifier</key>
				<string>com.apple.RunShellScript</string>
				<key>CFBundleVersion</key>
				<string>2.1.1</string>
				<key>CanShowSelectedItemsWhenRun</key>
				<false/>
				<key>CanShowWhenRun</key>
				<true/>
				<key>Category</key>
				<string>AMCategoryUtilities</string>
				<key>DefaultName</key>
				<string>Run Shell Script</string>
				<key>Disabled</key>
				<false/>
				<key>NameOfAction</key>
				<string>Run Shell Script</string>
				<key>Parameters</key>
				<dict>
					<key>COMMAND_STRING</key>
					<string>for f in "$@"
do
    echo "Processing: $f"
done</string>
					<key>CheckedForUserDefaultShell</key>
					<true/>
					<key>inputMethod</key>
					<integer>1</integer>
					<key>shell</key>
					<string>/bin/zsh</string>
					<key>source</key>
					<string></string>
				</dict>
				<key>ApplicationBundleID</key>
				<string>com.apple.finder</string>
				<key>ApplicationBundleIDsByPath</key>
				<dict>
					<key>com.apple.finder</key>
					<string>/System/Library/CoreServices/Finder.app</string>
				</dict>
				<key>AssociatedApplications</key>
				<array>
					<string>com.apple.finder</string>
				</array>
			</dict>
		</dict>
	</array>
	<key>connectors</key>
	<dict/>
	<key>workflowMetaData</key>
	<dict>
		<key>applicationBundleID</key>
		<string>com.apple.finder</string>
		<key>applicationBundleIDsByPath</key>
		<dict>
			<key>com.apple.finder</key>
			<string>/System/Library/CoreServices/Finder.app</string>
		</dict>
		<key>associatedApplications</key>
		<array>
			<string>com.apple.finder</string>
		</array>
		<key>inputTypeIdentifier</key>
		<string>com.apple.automator.fileSystemObject</string>
		<key>outputTypeIdentifier</key>
		<string>com.apple.automator.nothing</string>
		<key>presentationMode</key>
		<string>15</string>
		<key>processesInput</key>
		<integer>0</integer>
		<key>serviceApplicationBundleID</key>
		<string>com.apple.finder</string>
		<key>serviceApplicationPath</key>
		<string>/System/Library/CoreServices/Finder.app</string>
		<key>serviceInputTypeIdentifier</key>
		<string>com.apple.automator.fileSystemObject</string>
		<key>serviceOutputTypeIdentifier</key>
		<string>com.apple.automator.nothing</string>
		<key>serviceProcessesInput</key>
		<integer>0</integer>
		<key>systemSettingsVersion</key>
		<string>2</string>
		<key>targetApplication</key>
		<string>com.apple.finder</string>
		<key>targetPath</key>
		<string>/System/Library/CoreServices/Finder.app`</string>
	</dict>
</dict>
</plist>
EOF

# 4. Create Info.plist for the bundle
cat << EOF > "$CONTENTS/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://apple.com">
<plist version="1.0">
<dict>
	<key>CFBundleIdentifier</key>
	<string>com.yourname.RunMyScript</string>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>
	<key>CFBundleName</key>
	<string>$NAME</string>
	<key>CFBundleShortVersionString</key>
	<string>1.0</string>
	<key>NSServices</key>
	<array>
		<dict>
			<key>NSBackgroundColorName</key>
			<string>background</string>
			<key>NSSendMessage</key>
			<string>runWorkflowAsService</string>
			<key>NSServiceDescription</key>
			<string>Run a custom shell script</string>
			<key>NSServiceProvider</key>
			<string>NSServicesProvider</string>
			<key>NSRequiredContext</key>
			<dict>
				<key>NSApplicationIdentifier</key>
				<string>com.apple.finder</string>
			</dict>
			<key>NSSendFileTypes</key>
			<array>
				<string>public.item</string>
			</array>
		</dict>
	</array>
</dict>
</plist>
EOF

# 5. Register the service with the system
/System/Library/CoreServices/RawCameraPartialSupport.appex/../../../../usr/bin/killall Finder 2>/dev/null
/usr/bin/osascript -e 'tell application "Finder" to restart' 2>/dev/null


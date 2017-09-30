package main

var ignoreTmpl = `
### Code ###
# Visual Studio Code - https://code.visualstudio.com/
.settings/
.vscode/


### Go ###
# Compiled Object files, Static and Dynamic libs (Shared Objects)
*.o
*.a
*.so

# Folders
_obj
_test

# Architecture specific extensions/prefixes
*.[568vq]
[568vq].out

*.cgo1.go
*.cgo2.c
_cgo_defun.c
_cgo_gotypes.go
_cgo_export.*

_testmain.go

*.exe
*.test
*.prof


### OSX ###
.DS_Store
.AppleDouble
.LSOverride

# Icon must end with two \r
Icon


# Thumbnails
._*

# Files that might appear in the root of a volume
.DocumentRevisions-V100
.fseventsd
.Spotlight-V100
.TemporaryItems
.Trashes
.VolumeIcon.icns

# Directories potentially created on remote AFP share
.AppleDB
.AppleDesktop
Network Trash Folder
Temporary Items
.apdisk


### Windows ###
# Windows image file caches
Thumbs.db
ehthumbs.db

# Folder config file
Desktop.ini

# Recycle Bin used on file shares
$RECYCLE.BIN/

# Windows Installer files
*.cab
*.msi
*.msm
*.msp

# Windows shortcuts
*.lnk


### SublimeText ###
# cache files for sublime text
*.tmlanguage.cache
*.tmPreferences.cache
*.stTheme.cache

# workspace files are user-specific
*.sublime-workspace

# project files should be checked into the repository, unless a significant
# proportion of contributors will probably not be using SublimeText
# *.sublime-project

# sftp configuration file
sftp-config.json
`

var circleTmpl = `test:
override:
  - go version
  - go get -u github.com/mattn/goveralls
  - GOMAXPROCS=1 go test -timeout 30s 
  - GOMAXPROCS=4 go test -timeout 30s -race
  - goveralls -service=circleci -repotoken $COVERALLS_TOKEN`

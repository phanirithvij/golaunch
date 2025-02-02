#SHELL=C:/Windows/System32/cmd.exe
#SHELL=D:/dev/msys64/usr/bin/bash.exe
SHELL=bash.exe
VERSION=$(shell node -e "console.log(require('./package.json').version);")

watch:
	go run watch.go

dev:
	npm start

buildplugins:
	@# flatc -g query_result.fbs request.fbs response.fbs
	@# flatc -s -o ./app/js/flatapi query_result.fbs request.fbs response.fbs
	go build -o ./plugins/golaunch-programs/golaunch-programs.exe ./plugins/golaunch-programs
	go build -o ./plugins/golaunch-process-killer/golaunch-process-killer.exe ./plugins/golaunch-process-killer
	go build -o ./plugins/golaunch-helloworld/golaunch-helloworld.exe ./plugins/golaunch-helloworld

buildjs:
	gobble build -f build

package:buildplugins buildjs
	./node_modules/.bin/electron-packager . GoLaunch \
	  --overwrite --prune --platform=win32 --arch=x64 --version=0.35.1 --out=dist \
		--ignore=node_modules/\.bin \
		--ignore="^\\." \
		--ignore="media" \
		--ignore="src" \
		--icon=src/icon.ico \
		--app-version=$(VERSION)

# rebrand:
# 	rcedit electron\electron.exe --set-icon app\icon.ico --set-version-string "CompanyName" "Outroot" --set-version-string "FileDescription" "GoLaunch application launcher" --set-version-string "LegalCopyright" "Copyright (C) 2015 Kevin Darlington. All rights reserved." --set-version-string "ProductName" "GoLaunch" --set-file-version "0.0.1" --set-product-version "0.0.1"

# bundle:
# 	electron-packager . Playback --platform=win32 --arch=ia32 --version=0.27.2 --icon=icon.ico

.PHONY: watch dev build package rebrand bundle

default:
	@just --list

_build platform exe:
	wails build -trimpath -platform {{platform}}/amd64 -o {{platform}}/{{exe}}

build-linux: (_build "linux" "srcvox")

build-windows: (_build "windows" "srcvox.exe")

build: build-linux build-windows


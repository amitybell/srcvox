default:
	@just --list

_build target:
	wails build -trimpath -platform {{target}}/amd64

build: (_build "linux") (_build "windows")


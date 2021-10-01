.PHONY:windows
windows:
	set GOOS=windows&&build -o cache_managerr -ldflags "-s -w"

.PHONY:linux
linux:
	set GOOS=linux&&go build -o cache_managerr -ldflags "-s -w"

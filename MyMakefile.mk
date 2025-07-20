umake: uclean ubuild

uclean:
	cd E:/Projects/PKr-Service/PKr-Installer/mywixfile/ && rm -rf PKr-Service.exe PKr-Cli.exe PKr-Base.exe
ubuild:
	GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o E:/Projects/PKr-Service/PKr-Installer/mywixfile/PKr-Service.exe
# 	cd E:/Projects/PKR-Re/PKr-Cli/ && GOOS=windows GOARCH=amd64 go build -o E:/Projects/PKr-Service/PKr-Installer/PKr-Cli.exe
# 	cd E:/Projects/PKR-Re/PKr-Base/ && GOOS=windows GOARCH=amd64 go build -o E:/Projects/PKr-Service/PKr-Installer/PKr-Base.exe

.PHONY: umake uclean ubuild
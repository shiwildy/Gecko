cd cmd/gecko
rsrc -ico ../../etc/logo/logo.ico -o ../../cmd/gecko/rsrc.syso
go build -ldflags="-s -w" -trimpath -o gecko.exe .
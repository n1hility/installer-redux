cd podman-msihooks
go build -buildmode=c-shared  -o ../artifacts/podman-msihooks.dll ./cmd/podman-msihooks || exit /b 1
go build -ldflags -H=windowsgui -o ../artifacts/podman-kerninst.exe ./cmd/podman-kerninst || exit /b 1
cd ..

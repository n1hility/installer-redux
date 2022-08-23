cd podman-msihooks
go build -buildmode=c-shared  -o ../artifacts/podman-msihooks.dll ./cmd/podman-msihooks || exit /b 1
go build -ldflags -H=windowsgui -o ../artifacts/podman-kerninst.exe ./cmd/podman-kerninst || exit /b 1
cd ..
heat dir docs -var var.ManSource -cg ManFiles -dr INSTALLDIR -gg -g1 -srd -out pages.wxs || exit /b 1
candle -ext WixUIExtension -ext WixUtilExtension -ext .\PanelSwWixExtension.dll -arch x64 -dManSource="docs" -dVERSION="4.2.0" podman.wxs pages.wxs podman-ui.wxs welcome-install-dlg.wxs || exit /b 1
light -ext WixUIExtension -ext WixUtilExtension -ext .\PanelSwWixExtension.dll .\podman.wixobj .\pages.wixobj .\podman-ui.wixobj .\welcome-install-dlg.wixobj -out podman.msi || exit /b 1
candle -ext WixUIExtension -ext WixUtilExtension -ext WixBalExtension -arch x64 -dManSource="docs" -dVERSION="4.2.0" burn.wxs || exit /b 1
light -ext WixUIExtension -ext WixUtilExtension -ext WixBalExtension .\burn.wixobj || exit /b 1

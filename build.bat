cd podman-msihooks
go build -buildmode=c-shared  -o ../artifacts/podman-msihooks.dll ./cmd/podman-msihooks
go build -ldflags -H=windowsgui -o ../artifacts/podman-kerninst.exe ./cmd/podman-kerninst
cd ..
heat dir docs -var var.ManSource -cg ManFiles -dr INSTALLDIR -gg -g1 -srd -out pages.wxs
candle -ext WixUIExtension -ext WixUtilExtension -ext .\PanelSwWixExtension.dll -arch x64 -dManSource="docs" -dVERSION="4.2.0" podman.wxs pages.wxs podman-ui.wxs welcome-install-dlg.wxs
light -ext WixUIExtension -ext WixUtilExtension -ext .\PanelSwWixExtension.dll .\podman.wixobj .\pages.wixobj .\podman-ui.wixobj .\welcome-install-dlg.wixobj -out podman.msi
candle -ext WixUIExtension -ext WixUtilExtension -ext WixBalExtension -arch x64 -dManSource="docs" -dVERSION="4.2.0" burn.wxs
light -ext WixUIExtension -ext WixUtilExtension -ext WixBalExtension .\burn.wixobj

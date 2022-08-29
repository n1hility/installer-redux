function Copy-Artifact {
    param(
        [Parameter(Mandatory)]
        [string]$fileName
    )
    $file = Get-ChildItem -Recurse -Path . -Name $fileName
    if (!$file) {
        throw "Could not find $filename"
    }
    Write-Host "file:" $file
    Copy-Item -Path $file -Destination "..\artifacts\$filename" -ErrorAction Stop
}

function DownloadOrSkip {
    param(
        [Parameter(Mandatory)]
        [string]$url,
        [Parameter(Mandatory)]
        [string]$file
    )
    $ProgressPreference = 'SilentlyContinue';
    try {
        Invoke-WebRequest -UseBasicParsing -ErrorAction Stop -Uri $url -OutFile $file
    } Catch {
        if ($_.Exception.Response.StatusCode -eq 404) {
            Write-Host "URL not availahble, signaling skip:"
            Write-Host "URL: $url"
            Exit 2
        }

        throw $_.Exception
    }
}

if ($args.Count -lt 1) {
    Write-Host "Usage: " $MyInvocation.MyCommand.Name "<version>"
    Exit 1
}

$base_url = "$ENV:FETCH_BASE_URL"
if ($base_url.Length -le 0) {
    $base_url = "https://github.com/containers/podman"
}

$version = $args[0]
if ($version -notmatch '^v?([0-9]+\.[0-9]+\.[0-9]+)(-.*)?$') {
    Write-Host "Invalid version"
    Exit 1
}

# WiX burn requires a QWORD version only, numeric only
$Env:INSTVER=$Matches[1]

if ($version[0] -ne 'v') {
    $version = 'v' + $version
}

$restore = 0
$exitCode = 0

try {
    Write-Host "Cleaning up old artifacts"
    Remove-Item -Force -Recurse -Path .\docs -ErrorAction SilentlyContinue | Out-Null
    Remove-Item -Force -Recurse -Path .\artifacts -ErrorAction SilentlyContinue | Out-Null
    Remove-Item -Force -Recurse -Path .\fetch -ErrorAction SilentlyContinue | Out-Null

    New-Item fetch -ItemType Directory | Out-Null
    New-Item artifacts -ItemType Directory | Out-Null

    Write-Host "Fetching zip release"

    Push-Location fetch -ErrorAction Stop
    $restore = 1
    $ProgressPreference = 'SilentlyContinue';
    DownloadOrSkip "$base_url/releases/download/$version/podman-remote-release-windows_amd64.zip"  "release.zip"
    Expand-Archive -Path release.zip
    $loc = Get-ChildItem -Recurse -Path . -Name win-sshproxy.exe
    if (!$loc) {
        Write-Host "Old release, zip does not include win-sshproxy.exe, fetching via msi"
        DownloadOrSkip "$base_url/releases/download/$version/podman-$version.msi" "podman.msi"
        dark -x expand ./podman.msi
        if (!$?) {
            throw "Dark command failed"
        }
        $loc = Get-ChildItem -Recurse -Path expand -Name 4A2AD125-34E7-4BD8-BE28-B2A9A5EDBEB5
        if (!$loc) {
            throw "Could not obtain win-sshproxy.exe"
        }
        Copy-Item -Path "expand\$loc" -Destination "win-sshproxy.exe" -ErrorAction Stop
        Remove-Item -Recurse -Force -Path expand
    }

    Write-Host "Copying artifacts"
    Foreach ($fileName in "win-sshproxy.exe", "podman.exe") {
        Copy-Artifact($fileName)
    }

    $loc = Get-ChildItem -Recurse -Path . -Name podman-for-windows.html
    if (!$loc) {
        Write-Host "Old release did not include welcome page, using podman-machine instead"
        $loc = Get-ChildItem -Recurse -Path . -Name podman-machine.html ..\docs\podman-for-windows.html
    }

    Write-Host "Copying docs"
    $loc = Get-ChildItem -Path . -Name docs -Recurse

    Copy-Item -Recurse -Path $loc -Destination ..\docs -ErrorAction Stop
    Write-Host "Done!"

    if (!$loc) {
        throw "Could not find docs"
    }
}
catch {
    Write-Host $_

    $exitCode = 1
}
finally {
    if ($restore) {
        Pop-Location
    }
}

exit $exitCode
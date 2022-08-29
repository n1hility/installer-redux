function SkipExists {
    param(
        [Parameter(Mandatory)]
        [string]$url
    )
    try {
        Invoke-WebRequest -Method HEAD -UseBasicParsing -ErrorAction Stop -Uri $url
        Write-Host "Installer already uploaded, skipping"
        Exit 2
    } Catch {
        if ($_.Exception.Response.StatusCode -eq 404) {
            Write-Host "Installer does not exist,  continuing..."
            Return
        }

        throw $_.Exception
    }
}

iif ($args.Count -lt 1 -or $args[0].Length -lt 2) {
    Write-Host "Usage: " $MyInvocation.MyCommand.Name "<version>"
    Exit 1
}

$release = $args[0]
$version = $release
if ($release[0] -eq "v") {
    $version = $release.Substring(1)
}

$ENV:UPLOAD_ASSET_NAME = "$base_url/releases/download/$release/podman-$version-setup.exe"
SkipExists "$base_url/releases/download/$release/podman-$version-setup.exe"
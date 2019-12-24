#Set-ExecutionPolicy -ExecutionPolicy Unrestricted -Scope CurrentUser

function checkGo {
    $software = "Go Programming Language amd64 go1.13.5";
    $installed = (Get-ItemProperty HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\* | Where { $_.DisplayName -eq $software }) -ne $null

    If(-Not $installed) {
	    Write-Host "'$software' NOT is installed.";
        $answer = Read-Host -Prompt "Install? (y/n)"
        if ($answer -eq "y") {
            installGo
        }else{
            Write-Host "Aborted by user."
            exit 0
        }
    } else {
	    Write-Host "'$software' is already installed."
    }
}

function installGo {
    $URL = "https://dl.google.com/go/go1.13.5.windows-amd64.msi"
    $filename = $URL.Substring($URL.LastIndexOf("/") + 1)
    (New-Object System.Net.WebClient).DownloadFile("https://dl.google.com/go/go1.13.5.windows-amd64.msi","$PSScriptRoot\$filename")

    $hashFromFile = get-filehash -Path "$PSScriptRoot\$filename" -Algorithm SHA256
    # checksum from https://golang.org/dl/
    $verifychecksum = "eabcf66d6d8f44de21a96a18d957990c62f21d7fadcb308a25c6b58c79ac2b96"

    if ($hashFromFile.Hash -eq $verifychecksum.toUpper()) {
        Write-Host "$PSScriptRoot\$filename" verified -ForegroundColor Green
        Start-Process -FilePath "msiexec.exe" -ArgumentList "/i","$PSScriptRoot\$filename","INSTALLDIR=$env:APPDATA\go","/qb" -Wait
    } else {
        Write-Host "$PSScriptRoot\$filename" is corrupt -ForegroundColor Red
        exit 1
    }
}

function buildOptimizelyAgent {
    Write-Host "Building Optimizely Agent..." -ForegroundColor Green
    $env:GO111MODULE = "on"
    go build -ldflags "-s -w -X main.Version=0.8.1" -o bin\optimizely.exe cmd\sidedoor\main.go
    dir bin
}

function refreshPath {
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") +
                ";" +
                [System.Environment]::GetEnvironmentVariable("Path","User")
}

function runOptimizelyAgentTests {
    Write-Host "Running tests..." -ForegroundColor Green
    go test ./...
    if ( $LASTEXITCODE -ne 0 ){
        Write-Host "Tests failed, refusing to build." -ForegroundColor Red
        exit 1
    } else {
        Write-Host "Tests passed!" -ForegroundColor Green
    }
}

function main {

# check if go is installed, if not, install it.
checkGo

# refresh $PATH
refreshPath

# run tests
runOptimizelyAgentTests

# build optimizely agent
buildOptimizelyAgent

}

main

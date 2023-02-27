#Set-ExecutionPolicy -ExecutionPolicy Unrestricted -Scope CurrentUser

function checkPrereq($software, $URL, $SHA, $mode) {

    $installed = (Get-ItemProperty HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\* | Where { $_.DisplayName -eq $software }) -ne $null

    If(-Not $installed) {
	    Write-Host "'$software' is NOT installed. (You may (s)kip if you already have it installed and it is in your PATH)";
        if ($mode -eq "noninteractive") {
            $answer = "y"
        } else {
            $answer = Read-Host -Prompt "Install? (y)es, (n)o or (s)kip"
        }
        if ($answer -eq "y") {
            installPrereq $URL $SHA
        } elseif ($answer -eq "s") {
            Write-Host "Skipped by user."
        } else {
            Write-Host "Aborted by user."
            exit 0
        }
    } else {
	    Write-Host "'$software' is already installed."
    }
}

function installPrereq($URL, $SHA) {
    $RANDOM = Get-Random
    $filename = $URL.Substring($URL.LastIndexOf("/") + 1)
    # https://github.com/lukesampson/scoop/pull/2065#issuecomment-369669048
    if (([System.Net.ServicePointManager]::SecurityProtocol).HasFlag([Net.SecurityProtocolType]::Tls12) -eq $false) {
        [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
    }
    (New-Object System.Net.WebClient).DownloadFile("$URL","$PSScriptRoot\$filename")
    Write-Host "$PSScriptRoot\$filename" downloaded. -ForegroundColor Green
    $hashFromFile = get-filehash -Path "$PSScriptRoot\$filename" -Algorithm SHA256

    if ($hashFromFile.Hash -eq $SHA.toUpper()) {
        Write-Host "$PSScriptRoot\$filename" verified. -ForegroundColor Green
        $extension = $filename.Substring($filename.LastIndexOf("."))
        if ($extension -eq ".msi") {
            Start-Process -FilePath "msiexec.exe" -ArgumentList "/i","$PSScriptRoot\$filename","INSTALLDIR=$env:APPDATA\$RANDOM","/qb" -Wait
        } elseif ($extension -eq ".exe") {
            Start-Process -FilePath "$PSScriptRoot\$filename" -ArgumentList "/DIR=$env:APPDATA\$RANDOM","/SILENT"
            Write-Host "Installing $filename..."
            Start-Sleep -s 60
        } else {
            Write-Host "Unrecognized extension: $extension" -ForegroundColor Red
            exit 1
        }
        Write-Host "$PSScriptRoot\$filename" installed. -ForegroundColor Green
    } else {
        Write-Host "$PSScriptRoot\$filename" is corrupt. -ForegroundColor Red
        Write-Host "expected: "$SHA.toUpper()
        Write-Host "got:       "$hashFromFile.Hash
        exit 1
    }
}

function buildOptimizelyAgent {
    Write-Host "Building Optimizely Agent..." -ForegroundColor Green
    git status
    if ( $LASTEXITCODE -ne 0 ) {
        Write-Host "Optimizely Agent needs to build from its own git repository in order to determine its version." -ForegroundColor Red
        exit 1
    }
    $VERSION = (git describe --tags)
    go build -ldflags "-s -w -X main.Version=$VERSION" -o bin\optimizely.exe cmd\optimizely\main.go
    if (!$?) {
        exit 1
    }
    Get-ChildItem -Path "bin\optimizely.exe"
}

function refreshPath {
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") +
                ";" +
                [System.Environment]::GetEnvironmentVariable("Path","User")
}

function installDeps {
    Write-Host "Installing build dependencies for Optimizely Agent..." -ForegroundColor Green
    $env:GO111MODULE = "off"
    go get -u github.com/rakyll/statik
    statik -src=api/openapi-spec
    $env:GO111MODULE = "on"
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

function main($mode) {

    # noninteractive mode: ./build.ps1 noninteractive (default: interactive)

    # check if go is installed, if not, install it.
    checkPrereq 'Go Programming Language amd64 go1.20.1' https://dl.google.com/go/go1.20.1.windows-amd64.msi f06fdfa56f3aba62cbf80dacddbcc1150f4990cc117a9477047d3a3529ee3e80 $mode
    # same but with git
    checkPrereq 'Git version 2.39.2' https://github.com/git-for-windows/git/releases/download/v2.39.2.windows.1/Git-2.39.2-64-bit.exe d7608fbd854b3689102ff48b03c8cc77b35138f9f7350d134306da0ba5751464 $mode

    # refresh $PATH
    refreshPath

    # Install additional build deps
    installDeps

    # run tests
    runOptimizelyAgentTests

    # build optimizely agent
    buildOptimizelyAgent

}


$mode = $args[0]
if ($mode -ne "noninteractive") {
    $mode = "interactive"
}
main $mode

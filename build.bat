:: This batch file builds Optimizely Agent for Windows

@echo off
:: The name of the executable (default is current directory name)
set TARGET=sidedoor
set VERSION=0.8.1

:: Go parameters
::GO111MODULE //set in environment variables
set GOCMD=C:\Go\bin\go.exe
set GOPATH=C:\go-work
set GOBUILD=%GOCMD% build
set GOTEST=%GOCMD% test

:: Use linker flags to strip debugging info from the binary.
:: -s Omit the symbol table and debug information.
:: -w Omit the DWARF symbol table.
set LDFLAGS=-ldflags "-s -w -X main.Version=%VERSION%"

echo "Running tests..."
%GOTEST% ./...
if errorlevel 1 goto fail

echo "Running build..."
%GOBUILD% %LDFLAGS% -o %GOPATH%/%TARGET%.exe cmd/%TARGET%/main.go
goto end

:fail
echo ERROR: tests failed

:end

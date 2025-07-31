@echo off
echo Building k8sGO for multiple platforms...
echo.

cd cmd\k8sgo

echo Building for Windows (amd64)...
set GOOS=windows
set GOARCH=amd64
go build -buildvcs=false -o ..\..\builds\windows\k8sgo-windows-amd64.exe
if %ERRORLEVEL% NEQ 0 (
    echo Failed to build for Windows amd64
    exit /b 1
)

echo Building for Windows (arm64)...
set GOOS=windows
set GOARCH=arm64
go build -buildvcs=false -o ..\..\builds\windows\k8sgo-windows-arm64.exe
if %ERRORLEVEL% NEQ 0 (
    echo Failed to build for Windows arm64
    exit /b 1
)

echo Building for Linux (amd64) - Ubuntu/Red Hat/CentOS...
set GOOS=linux
set GOARCH=amd64
go build -buildvcs=false -o ..\..\builds\linux\k8sgo-linux-amd64
if %ERRORLEVEL% NEQ 0 (
    echo Failed to build for Linux amd64
    exit /b 1
)

echo Building for Linux (arm64)...
set GOOS=linux
set GOARCH=arm64
go build -buildvcs=false -o ..\..\builds\linux\k8sgo-linux-arm64
if %ERRORLEVEL% NEQ 0 (
    echo Failed to build for Linux arm64
    exit /b 1
)

echo Building for macOS (amd64)...
set GOOS=darwin
set GOARCH=amd64
go build -buildvcs=false -o ..\..\builds\linux\k8sgo-darwin-amd64
if %ERRORLEVEL% NEQ 0 (
    echo Failed to build for macOS amd64
    exit /b 1
)

echo Building for macOS (arm64) - Apple Silicon...
set GOOS=darwin
set GOARCH=arm64
go build -buildvcs=false -o ..\..\builds\linux\k8sgo-darwin-arm64
if %ERRORLEVEL% NEQ 0 (
    echo Failed to build for macOS arm64
    exit /b 1
)

cd ..\..

echo.
echo ===============================================
echo Build Summary:
echo ===============================================
echo Windows amd64: builds\windows\k8sgo-windows-amd64.exe
echo Windows arm64: builds\windows\k8sgo-windows-arm64.exe
echo Linux amd64:   builds\linux\k8sgo-linux-amd64
echo Linux arm64:   builds\linux\k8sgo-linux-arm64
echo macOS amd64:   builds\linux\k8sgo-darwin-amd64
echo macOS arm64:   builds\linux\k8sgo-darwin-arm64
echo ===============================================
echo.
echo All builds completed successfully!
echo.
echo Usage Instructions:
echo - Windows: Run k8sgo-windows-amd64.exe (or arm64 version)
echo - Linux/Red Hat/CentOS/Ubuntu: chmod +x k8sgo-linux-amd64 && ./k8sgo-linux-amd64
echo - macOS: chmod +x k8sgo-darwin-amd64 && ./k8sgo-darwin-amd64
echo.
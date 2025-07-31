@echo off
REM K8sgo Build Script for Windows

echo Building K8sgo...

REM Create build directory
if not exist build mkdir build

REM Build for Windows
echo Building for Windows...
go build -buildvcs=false -ldflags "-X main.Version=1.0.0" -o build\k8sgo.exe .\cmd\k8sgo

if %errorlevel% equ 0 (
    echo Build successful! Binary created at build\k8sgo.exe
    echo.
    echo To run: build\k8sgo.exe
) else (
    echo Build failed!
    exit /b 1
)

pause
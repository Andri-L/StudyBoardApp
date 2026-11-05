param(
    [switch]$SkipInstall
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Push-Location $root
try {
    $wails = Get-Command wails -ErrorAction SilentlyContinue
    if (-not $wails) {
        $candidate = Join-Path $env:USERPROFILE "go\bin\wails.exe"
        if (-not (Test-Path $candidate)) {
            throw "No se encontró Wails. Instala con: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
        }
        $wails = $candidate
    } else {
        $wails = $wails.Source
    }

    if (-not $SkipInstall) {
        Push-Location (Join-Path $root "frontend")
        try {
            npm install
        } finally {
            Pop-Location
        }
    }

    & $wails build -clean -nocolour
    if ($LASTEXITCODE -ne 0) {
        throw "La compilación de Wails falló."
    }

    $output = Join-Path $root "build\bin"
    $engine = Join-Path $output "bin"
    if (Test-Path $engine) {
        Remove-Item $engine -Recurse -Force
    }
    $localEngine = Join-Path $root "bin"
    if (Test-Path $localEngine) {
        Copy-Item $localEngine $engine -Recurse -Force
    } else {
        Write-Warning "No hay carpeta bin\. Ejecuta primero scripts\setup-engine.ps1 para instalar Whisper/FFmpeg/modelos."
    }

    # Segunda compilación (subsistema de consola) para el servidor MCP.
    # OpenCode necesita stdio limpio; este binario no abre la ventana.
    & go build -o (Join-Path $output "StudyBoardMCP.exe") .
    if ($LASTEXITCODE -ne 0) {
        throw "La compilación del servidor MCP falló."
    }

    Write-Host ""
    Write-Host "Listo: $output\StudyBoardApp.exe" -ForegroundColor Green
    Write-Host "MCP:  $output\StudyBoardMCP.exe" -ForegroundColor Green
    Write-Host "Motor copiado a: $engine" -ForegroundColor DarkGray
} finally {
    Pop-Location
}

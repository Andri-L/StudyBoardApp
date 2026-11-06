param(
    [ValidateSet("turbo", "small-q5", "small", "base-q5", "none")]
    [string]$Model = "turbo",
    [switch]$Force
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$bin = Join-Path $root "bin"
$whisperDir = Join-Path $bin "whisper"
$ffmpegDir = Join-Path $bin "ffmpeg"
$modelDir = Join-Path $env:LOCALAPPDATA "StudyBoard App\models"
$temp = Join-Path $env:TEMP ("studyboard-setup-" + [guid]::NewGuid().ToString("N"))

$whisperURL = "https://github.com/jerryshell/whisper.cpp-windows-vulkan-bin/releases/download/v1.0.0/whisper.cpp-windows-vulkan.zip"
$ffmpegURL = "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip"
$vadURL = "https://huggingface.co/ggml-org/whisper-vad/resolve/main/ggml-silero-v5.1.2.bin"

$models = @{
    "turbo"    = @{ File = "ggml-large-v3-turbo-q5_0.bin"; URL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3-turbo-q5_0.bin"; SHA256 = "394221709cd5ad1f40c46e6031ca61bce88931e6e088c188294c6d5a55ffa7e2" }
    "small-q5" = @{ File = "ggml-small-q5_1.bin"; URL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small-q5_1.bin"; SHA256 = "ae85e4a935d7a567bd102fe55afc16bb595bdb618e11b2fc7591bc08120411bb" }
    "small"    = @{ File = "ggml-small.bin"; URL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin"; SHA256 = "1be3a9b2063867b937e64e2ec7483364a79917e157fa98c5d94b5c1fffea987b" }
    "base-q5"  = @{ File = "ggml-base-q5_1.bin"; URL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base-q5_1.bin"; SHA256 = "422f1ae452ade6f30a004d7e5c6a43195e4433bc370bf23fac9cc591f01a8898" }
}
$vadSHA256 = "29940d98d42b91fbd05ce489f3ecf7c72f0a42f027e4875919a28fb4c04ea2cf"

New-Item -ItemType Directory -Force -Path $bin, $whisperDir, $ffmpegDir, $modelDir, $temp | Out-Null

function Download-File([string]$url, [string]$destination) {
    Write-Host "Descargando $url" -ForegroundColor Cyan
    Invoke-WebRequest -Uri $url -OutFile $destination -UseBasicParsing
}

function Assert-Hash([string]$path, [string]$expected) {
    if (-not $expected) { return }
    $actual = (Get-FileHash -Path $path -Algorithm SHA256).Hash
    if ($actual -ne $expected.ToUpperInvariant()) {
        throw "Hash incorrecto para $path. Esperado $expected, obtenido $actual."
    }
}

try {
    # 1. Whisper.cpp con backend Vulkan
    if ($Force -or -not (Test-Path (Join-Path $whisperDir "whisper-cli.exe"))) {
        $zip = Join-Path $temp "whisper.zip"
        Download-File $whisperURL $zip
        $extract = Join-Path $temp "whisper"
        Expand-Archive -Path $zip -DestinationPath $extract -Force
        Copy-Item (Join-Path $extract "*") $whisperDir -Recurse -Force
        Write-Host "Whisper Vulkan instalado en $whisperDir" -ForegroundColor Green
    } else {
        Write-Host "Whisper ya está instalado." -ForegroundColor DarkGray
    }

    # 2. FFmpeg + ffprobe
    if ($Force -or -not (Test-Path (Join-Path $ffmpegDir "ffmpeg.exe"))) {
        $zip = Join-Path $temp "ffmpeg.zip"
        Download-File $ffmpegURL $zip
        $extract = Join-Path $temp "ffmpeg"
        Expand-Archive -Path $zip -DestinationPath $extract -Force
        $ffmpegRoot = Get-ChildItem $extract -Directory | Select-Object -First 1
        Copy-Item (Join-Path $ffmpegRoot.FullName "bin\ffmpeg.exe") $ffmpegDir -Force
        Copy-Item (Join-Path $ffmpegRoot.FullName "bin\ffprobe.exe") $ffmpegDir -Force
        Write-Host "FFmpeg instalado en $ffmpegDir" -ForegroundColor Green
    } else {
        Write-Host "FFmpeg ya está instalado." -ForegroundColor DarkGray
    }

    # 3. VAD de Silero
    $vadPath = Join-Path $modelDir "ggml-silero-v5.1.2.bin"
    if ($Force -or -not (Test-Path $vadPath)) {
        Download-File $vadURL $vadPath
        try { Assert-Hash $vadPath $vadSHA256 } catch { Remove-Item $vadPath -Force -ErrorAction SilentlyContinue; throw }
        Write-Host "VAD instalado en $vadPath" -ForegroundColor Green
    } else {
        Write-Host "VAD ya está instalado." -ForegroundColor DarkGray
    }

    # 4. Modelo de Whisper opcional
    if ($Model -ne "none") {
        $spec = $models[$Model]
        $modelPath = Join-Path $modelDir $spec.File
        if ($Force -or -not (Test-Path $modelPath)) {
            $part = $modelPath + ".part"
            Download-File $spec.URL $part
            try { Assert-Hash $part $spec.SHA256 } catch { Remove-Item $part -Force -ErrorAction SilentlyContinue; throw }
            Move-Item $part $modelPath -Force
            Write-Host "Modelo instalado en $modelPath" -ForegroundColor Green
        } else {
            Write-Host "El modelo $($spec.File) ya está instalado." -ForegroundColor DarkGray
        }
    }

    Write-Host ""
    Write-Host "Motor listo." -ForegroundColor Green
    Write-Host "Siguiente paso: powershell -ExecutionPolicy Bypass -File .\scripts\build.ps1" -ForegroundColor Cyan
} finally {
    Remove-Item $temp -Recurse -Force -ErrorAction SilentlyContinue
}

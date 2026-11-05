# StudyBoard App

**Espacio de trabajo visual y transcripción local para clases y tutoriales.**

StudyBoard App organiza carpetas y videos en tableros, transcribe las grabaciones
con Whisper en tu propia máquina y permite consultar el contenido desde
**OpenCode TUI** mediante un servidor MCP.

La app **nunca mueve, renombra ni borra tus archivos**: las tarjetas son
referencias a rutas reales.

## Características

- Tablero visual con tarjetas de carpetas y videos (React Flow).
- Clic derecho en el tablero para crear **carpetas reales**, **notas adhesivas** y **flechas** de anotación.
- Las notas y flechas guardan texto, color y coordenadas en `state.json`.
- Varios tableros por materia y navegador de carpetas integrado.
- Transcripción local con Whisper.cpp y GPU AMD mediante Vulkan.
- Timestamps por segmento y por palabra en el JSON interno.
- Visor de video sincronizado: clic en un timestamp salta al momento exacto.
- Búsqueda dentro de la transcripción, modo “Seguir” y captura de frames.
- Cola de transcripciones con progreso, cancelación y reintento.
- Exportación a Markdown con timestamps y SRT.
- Descarga y selección de modelos desde la interfaz, sin salir de la app.
- Selector de idioma: Español, English o detección automática.
- Servidor MCP para consultar las clases desde OpenCode sin abrir la app.
- La ventana no se queda siempre encima.

## Stack

- **Go + Wails v2** para la aplicación de escritorio (WebView2).
- **React + TypeScript + Vite** para la interfaz.
- **React Flow** para el tablero.
- **Whisper.cpp + FFmpeg** como binarios locales.
- Servidor HTTP local con token para servir video con soporte de `Range`.
- Servidor MCP JSON-RPC por stdio para OpenCode.

## Requisitos

- Windows 10/11.
- Go 1.22+ (probado con 1.26).
- Node.js 20+.
- Wails CLI:
  ```powershell
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- WebView2 Runtime (incluido en Windows 10/11).
- Driver AMD/Intel/NVIDIA actualizado si se usa GPU.

## Instalación

```powershell
git clone https://github.com/Andri-L/StudyBoardApp.git
cd StudyBoardApp

# Descarga Whisper Vulkan, FFmpeg, VAD y el modelo turbo (~574 MB).
powershell -ExecutionPolicy Bypass -File .\scripts\setup-engine.ps1

# Compila la app y el servidor MCP.
powershell -ExecutionPolicy Bypass -File .\scripts\build.ps1
```

El ejecutable queda en:

```text
build\bin\StudyBoardApp.exe
```

El servidor MCP queda en:

```text
build\bin\StudyBoardMCP.exe
```

### Otros modelos en la instalación

El script acepta un modelo inicial distinto:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\setup-engine.ps1 -Model small-q5
```

Valores: `turbo`, `small-q5`, `small`, `base-q5`, `none`. También puedes elegir
`none` y descargar modelos luego desde el menú del modelo en la app.

## Ejecutar en desarrollo

```powershell
$env:WAILS_DEV_ALLOWED_HOSTS = "wails.localhost"
wails dev
```

Si `wails` no está en el PATH:

```powershell
& "$env:USERPROFILE\go\bin\wails.exe" dev
```

## Integración con OpenCode TUI (MCP)

El servidor MCP lee el workspace y las transcripciones guardadas. Funciona aunque
la interfaz de StudyBoard esté cerrada. Regístralo una vez:

```powershell
opencode mcp add studyboard --global -- "C:\ruta\StudyBoardApp\build\bin\StudyBoardMCP.exe"
opencode mcp list
```

Deberías ver:

```text
✓ studyboard  connected
```

Configuración manual en `~/.config/opencode/opencode.jsonc`:

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "servers": {
      "studyboard": {
        "type": "local",
        "command": [
          "C:\\ruta\\StudyBoardApp\\build\\bin\\StudyBoardMCP.exe"
        ]
      }
    }
  }
}
```

`StudyBoardMCP.exe` detecta automáticamente el modo MCP. El ejecutable de la
interfaz (`StudyBoardApp.exe`) también acepta `--mcp`.

### Herramientas expuestas

OpenCode las prefija con el nombre del servidor:

- `studyboard_overview`: workspace, tableros, videos, transcripciones y modelo.
- `studyboard_list_videos`: videos con clase, duración y estado.
- `studyboard_search_transcripts`: búsqueda global sin distinguir tildes.
- `studyboard_get_transcript`: transcripción completa o por rango de segundos.
- `studyboard_capture_frame`: imagen de un segundo exacto del video.

También expone recursos (`studyboard://overview`, `videos`, `transcripts`) y los
prompts `/studyboard:resumir-clase` y `/studyboard:buscar-clases`.

Ejemplos:

```text
Busca en mis clases dónde se explicó regresión logística y dame timestamps.
Resume la clase 22Jul26 y muestra el frame de la matriz de confusión.
Compara lo que se dijo sobre validación cruzada en dos clases.
```

## Modelos

La app incluye un menú rápido en la barra superior con el modelo exacto activo,
por ejemplo `large-v3-turbo-q5_0 · Vulkan`. Desde ahí puedes:

- Ver modelos instalados y su tamaño.
- Descargar `small`, `small-q5`, `base-q5`, `medium-q5`, `tiny` o
  `distil-large-v3`.
- Cambiar el modelo activo para las próximas transcripciones.
- Elegir Español, English o detección automática.
- Abrir la carpeta de modelos.

Las descargas se guardan en:

```text
%LOCALAPPDATA%\StudyBoard App\models\
```

No se suben modelos al repositorio. Los binarios de Whisper/FFmpeg también se
instalan localmente con `scripts/setup-engine.ps1`.

## Datos de la app

- Estado de tableros: `%APPDATA%\StudyBoard App\state.json` (incluye tarjetas, notas y flechas).
- Transcripciones normalizadas: `%LOCALAPPDATA%\StudyBoard App\transcripts\`
- Miniaturas: `%LOCALAPPDATA%\StudyBoard App\thumbs\`
- Modelos descargados: `%LOCALAPPDATA%\StudyBoard App\models\`
- Temporales de trabajos: `%LOCALAPPDATA%\StudyBoard App\jobs\`

Las transcripciones se invalidan automáticamente si el video cambia de tamaño o
fecha de modificación. No se guardan copias de los videos.

## Pipeline de transcripción

```text
video
  └─ ffmpeg -vn -ac 1 -ar 16000 -c:a pcm_s16le → audio.wav temporal
       └─ whisper-cli -l <es|en|auto> -ojf -osrt --vad -pp
            └─ JSON normalizado (segmentos + palabras + confianza)
                 ├─ visor y timestamps
                 ├─ export .md / .srt
                 └─ frames bajo demanda con ffmpeg
```

El modelo recomendado es `ggml-large-v3-turbo-q5_0.bin` con VAD Silero. La cola
procesa una tarea a la vez para no competir por la GPU. El progreso se lee del
callback de Whisper y de `-progress` de FFmpeg.

## Estructura del proyecto

```text
StudyBoardApp/
├── app.go              # bindings, estado, exportaciones
├── jobs.go             # cola, FFmpeg, Whisper y transcripción normalizada
├── media.go            # servidor local de video/miniaturas y frames
├── engine.go           # detección de binarios y modelo
├── models.go           # catálogo y descarga de modelos
├── mcp.go              # servidor MCP JSON-RPC por stdio
├── mcp_data.go         # herramientas, recursos y prompts del MCP
├── store.go            # persistencia JSON, fechas y helpers
├── types.go            # tipos compartidos
├── scripts/
│   ├── setup-engine.ps1
│   └── build.ps1
├── frontend/           # React + React Flow
└── build/              # icono y salida de compilación (build/bin ignorado)
```

## Rendimiento medido

Medición de referencia en Windows 11 con una **Radeon RX 6650 XT (8 GB)** y CS2
abierto:

- Clase de 42:20 en inglés: **11:05** de transcripción (~3,8× tiempo real).
- `large-v3-turbo-q5_0` + VAD, 1.617 segmentos.
- VRAM total del adaptador: 7,4–7,9 GB de 8 GB.
- Con la GPU tan cargada, `small-q5` puede fallar al inicializar Vulkan por falta
  de VRAM. Para transcripciones largas se recomienda no competir con un juego.
- En una máquina sin carga, el rendimiento es considerablemente mejor.

Los valores dependen del video, la GPU, el modelo y las aplicaciones abiertas.

## Limitaciones actuales

- Los timestamps que se muestran en la UI son por segmento; el JSON interno ya
  guarda timestamps por palabra, pendiente de explotar visualmente.
- No hay diarización de hablantes.
- No hay resumen automático con LLM dentro de la app; se exporta el texto.
- MKV/AVI pueden no reproducirse dentro de WebView2; se ofrece abrirlo con el
  reproductor de Windows.
- La búsqueda global está disponible vía MCP; la UI todavía no tiene un buscador
  entre todas las clases.
- El servidor MCP es de solo lectura: no inicia transcripciones.

## Roadmap

1. Alineación forzada a nivel de palabra (WhisperX/wav2vec2 u ONNX).
2. Frames automáticos y resumen con LLM enlazado a timestamps.
3. Diarización profesor/alumnos.
4. Búsqueda global dentro de la UI.
5. Procesamiento por lotes de carpetas completas.

Referencia técnica para alineación y diarización:
[jpgomezv/video-transcriber](https://github.com/jpgomezv/video-transcriber).

## Privacidad

El audio y las transcripciones se procesan en la máquina. La app no sube
contenido a internet. Solo se conecta para descargar binarios/modelos cuando tú
lo solicitas.

## Licencia

GNU General Public License v3.0. Consulta [LICENSE](LICENSE).

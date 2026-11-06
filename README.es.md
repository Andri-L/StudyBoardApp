# StudyBoard App

**Espacio de trabajo visual y transcripción local para clases y tutoriales.**

[English](README.md) · [Español](README.es.md)

StudyBoard App organiza carpetas y videos en tableros, transcribe las grabaciones
con Whisper en tu propia máquina y permite consultar el contenido desde
**OpenCode TUI** mediante un servidor MCP.

La app **nunca mueve, renombra ni borra tus archivos**: las tarjetas son
referencias a rutas reales.

## Características

- Tablero visual con tarjetas de carpetas y videos (React Flow).
- Clic derecho en el tablero para crear **carpetas reales**, **notas adhesivas** y **flechas** de anotación.
- Renombrar archivos y carpetas reales desde el tablero o el navegador.
- Nueve colores pastel de fondo para tarjetas, notas e informes.
- Flechas rectas y en curva S con selección, movimiento, redimensionado de extremos y color.
- Temas claro y oscuro persistentes.
- Las notas y flechas guardan texto, color y coordenadas en `state.json`.
- Varios tableros por materia y navegador de carpetas integrado.
- Transcripción local con Whisper.cpp y GPU AMD mediante Vulkan.
- Timestamps por segmento y por palabra en el JSON interno.
- Visor de video sincronizado: clic en un timestamp salta al momento exacto.
- Conversión a MP4 bajo demanda para formatos que WebView2 no reproduce.
- Búsqueda dentro de la transcripción, modo “Seguir” y captura de frames.
- Cola de transcripciones con progreso, cancelación y reintento.
- Exportación a Markdown con timestamps y SRT.
- Descarga y selección de modelos desde la interfaz, sin salir de la app.
- Verificación SHA256 de los modelos descargados.
- Selector de idioma: Español, English o detección automática.
- **Workflows automáticos de OpenCode** para extraer tareas, preparar parciales, crear planes de estudio y extraer fechas.
- Traducción de informes entre español e inglés.
- Los informes aparecen como notas adhesivas junto al video y se guardan como Markdown junto a la grabación.
- Selector para ejecutar varios workflows sobre varias clases a la vez.
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
- Go 1.26.10+ (incluye correcciones de seguridad de la librería estándar).
- Node.js 20+.
- Wails CLI:
  ```powershell
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- WebView2 Runtime (incluido en Windows 10/11).
- Driver AMD/Intel/NVIDIA actualizado si se usa GPU.
- CLI de OpenCode instalado y autenticado (solo para workflows):
  ```powershell
  npm install -g @opencode/cli
  opencode
  ```
  StudyBoard reutiliza la autenticación que ya tienes en OpenCode. No pide
  credenciales ni guarda tokens del proveedor.

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

En desarrollo, indica explícitamente el motor del repositorio:

```powershell
$env:STUDYBOARD_ENGINE_DIR = (Get-Location).Path
```

Las pruebas E2E del frontend usan un backend Wails simulado:

```powershell
cd frontend
npm run build
npm run test:e2e
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
- `studyboard_list_workflow_results`: informes de tareas o parciales generados.
- `studyboard_get_workflow_result`: informe completo por id, video o workflow.

También expone recursos (`studyboard://overview`, `videos`, `transcripts`,
`workflow-results`) y los prompts `/studyboard:resumir-clase`,
`/studyboard:buscar-clases`, `/studyboard:extraer-tareas`,
`/studyboard:puntos-parciales`, `/studyboard:plan-de-estudio` y
`/studyboard:fechas-calendario`.

Ejemplos:

```text
Busca en mis clases dónde se explicó regresión logística y dame timestamps.
Resume la clase 22Jul26 y muestra el frame de la matriz de confusión.
Compara lo que se dijo sobre validación cruzada en dos clases.
Muéstrame las tareas extraídas de la clase 22Jul26.
```

## Workflows con OpenCode

StudyBoard puede ejecutar OpenCode automáticamente sobre una clase transcrita.
La app lanza `opencode run` con una sesión nueva y reutiliza la autenticación que
ya configuraste en OpenCode; no hay un login aparte dentro de StudyBoard.

Incluye cuatro workflows:

- **Extracción de tareas** (`task-extraction`): actividades explícitas,
  requisitos, preferencias del profesor, plazos, canal de entrega, modo
  individual/grupal, pistas de rúbrica, evidencia y dudas abiertas. Si no hay
  tareas, devuelve una respuesta corta.
- **Puntos para parciales** (`exam-focus`): solo información explícita del
  parcial. Si el profesor no mencionó temas, formato, fecha ni criterios,
  devuelve una respuesta corta de “sin información explícita” en vez de un plan
  largo.
- **Plan de estudio** (`study-plan`): plan breve y priorizado basado en el
  contenido real de la clase, con práctica y evidencia.
- **Fechas para calendario** (`calendar-dates`): exámenes, entregas y eventos
  con fecha, además de un archivo `.ics` para importar al calendario.

Para ejecutar uno:

1. Abre una clase transcrita y pulsa **Ejecutar workflow** en la barra del
   tablero, o **Workflows** en el visor de transcripción.
2. Selecciona uno o varios workflows y uno o varios videos transcritos.
3. Elige el idioma de salida (Español o English).
4. Pulsa **Ejecutar**.

Cada resultado se coloca en el tablero que contiene el video de origen como una
tarjeta y se guarda como Markdown junto al video (`<video> - tareas.md`,
`<video> - parcial.md`, `<video> - plan.md` o `<video> - fechas.md`). Los
workflows de calendario también escriben un archivo `.ics`. Puedes arrastrar la
tarjeta, copiar su contenido, traducirlo a español o inglés, mostrar el archivo
en Explorer o quitarla del tablero sin borrar el Markdown.

Las plantillas de prompt están en inglés para mayor fiabilidad del modelo. El
informe final se escribe en el idioma elegido por el usuario. Los workflows
requieren el CLI de OpenCode instalado y autenticado.

Las ejecuciones están aisladas: OpenCode arranca con una configuración generada
que deniega shell, edición de archivos, lecturas directas, acceso web,
subagentes, skills y LSP. Solo se permiten las herramientas MCP de solo lectura
de StudyBoard. El proceso también se ejecuta en un directorio dedicado, no en el
workspace de clases, para que la configuración del proyecto y los archivos
locales no puedan influirlo. Consulta `SECURITY.md` para más detalles.

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

No se suben modelos al repositorio. Las descargas se verifican contra el hash
SHA256 del catálogo antes de poder seleccionarlas. Los binarios de
Whisper/FFmpeg también se instalan localmente con `scripts/setup-engine.ps1`.

## Datos de la app

- Estado de tableros: `%APPDATA%\StudyBoard App\state.json` (incluye tarjetas, notas, flechas y resultados de workflows).
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
├── security.go         # contención de rutas, escritura atómica, validación de URL
├── playback.go         # conversión a MP4 para contenedores no soportados
├── workflows.go        # definiciones, prompts y resultados de workflows
├── workflow_jobs.go    # cola y ejecución de OpenCode
├── workflows/          # workflows declarativos y plantillas de prompt
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
- Los workflows requieren el CLI de OpenCode instalado y autenticado.
- Convertir un video no soportado a MP4 puede tardar varios minutos.
- La interfaz está en español; la internacionalización está planificada.
- Las fechas del calendario las extrae el modelo. Revisa el `.ics` antes de
  importarlo.
- MKV/AVI pueden no reproducirse dentro de WebView2; se ofrece reproductor
  externo y conversión a MP4.
- La búsqueda global está disponible vía MCP; la UI todavía no tiene un buscador
  entre todas las clases.
- El servidor MCP es de solo lectura: no inicia transcripciones ni workflows.

## Roadmap

1. Alineación forzada a nivel de palabra (WhisperX/wav2vec2 u ONNX).
2. Diarización profesor/alumnos.
3. Búsqueda global dentro de la UI.
4. Procesamiento por lotes de carpetas completas.
5. Internacionalización de la UI y workflows definidos por el usuario.

Referencia técnica para alineación y diarización:
[jpgomezv/video-transcriber](https://github.com/jpgomezv/video-transcriber).

## Privacidad

El audio y las transcripciones se procesan en la máquina. StudyBoard no sube
contenido por su cuenta. Solo se conecta para descargar binarios/modelos cuando
tú lo solicitas.

Los workflows son distintos: invocan OpenCode con tu proveedor configurado, así
que el contenido de la transcripción consultado por el modelo se envía a ese
proveedor, igual que en cualquier sesión de OpenCode. Si no quieres eso, no
ejecutes workflows.

## Seguridad

La app trata el material de clase, las transcripciones y la salida del modelo
como datos no confiables:

- El servidor local de medios escucha solo en loopback, exige un token aleatorio
  y sirve únicamente videos dentro del workspace.
- Las rutas se resuelven antes de acceder para bloquear escapes por enlaces
  simbólicos o junctions.
- Los informes omiten HTML crudo, bloquean imágenes remotas y solo abren enlaces
  en el navegador del sistema tras validar `http`/`https` y rechazar hosts
  locales.
- Las descargas de modelos exigen HTTPS, verifican el SHA256 del catálogo y el
  formato GGML.
- El estado y los informes se escriben de forma atómica con permisos de usuario.
- `npm audit` y `govulncheck` no reportan vulnerabilidades conocidas.

Consulta [SECURITY.md](SECURITY.md) para el modelo de amenazas, los controles y
los riesgos residuales. La app no puede proteger una máquina ya comprometida por
malware que se ejecute como el mismo usuario de Windows.

## Licencia

GNU General Public License v3.0. Consulta [LICENSE](LICENSE).

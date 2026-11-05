# Changelog

## v1.0.0 — 2026-11-05

Primera versión pública de StudyBoard App.

### Añadido

- Tablero visual de carpetas y videos con React Flow.
- Creación de carpetas reales desde el menú contextual del tablero.
- Notas adhesivas editables con colores y flechas de anotación libre.
- Persistencia de tarjetas, notas, flechas y coordenadas en JSON.
- Varios tableros por materia, navegador de archivos y drag & drop.
- Transcripción local con Whisper.cpp y aceleración Vulkan en GPU AMD.
- Timestamps por segmento y por palabra en el JSON interno.
- Visor de video sincronizado con la transcripción.
- Búsqueda dentro de la transcripción y modo “Seguir”.
- Captura de frames para revisar lo que se mostraba en pantalla.
- Cola de transcripciones con progreso, cancelación y reintento.
- Exportación a Markdown con timestamps y SRT.
- Copia de la transcripción al portapapeles.
- Menú rápido de modelos con descarga integrada (`tiny`, `base`, `small`,
  `small-q5`, `medium-q5`, `large-v3-turbo`, `distil-large-v3`).
- Selector de idioma Español/English/Auto.
- Servidor MCP por stdio para consultar clases desde OpenCode TUI.
- Herramientas MCP: overview, list_videos, search_transcripts, get_transcript y
  capture_frame.
- Script de instalación del motor y de compilación.
- La ventana ya no se queda siempre encima.

### Notas

- Los modelos y binarios no se incluyen en el repositorio; se instalan con
  `scripts/setup-engine.ps1`.
- El servidor MCP es de solo lectura en esta versión.
- La diarización y el resumen automático con LLM quedan para versiones futuras.

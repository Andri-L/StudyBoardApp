# Changelog

[English](CHANGELOG.md) · [Español](CHANGELOG.es.md)

## v1.2.2 — 2026-11-05

Corrección del parpadeo al soltar flechas.

### Corregido

- Al soltar una flecha ya no aparece brevemente su geometría anterior. La flecha
  se actualiza de forma optimista y la posición final usa el último punto real
  del puntero, incluido el ajuste a bordes.

## v1.2.1 — 2026-11-05

Correcciones de flechas y selectores de color.

### Añadido

- Tipo de flecha **cuadrada/escalonada**, además de recta y curva S.
- Los extremos de las flechas se ajustan al centro del borde más cercano de una
  tarjeta al arrastrarlos cerca, con un indicador visual.

### Cambiado

- Todas las paletas de color son ahora paneles desplegables compactos en vez de
  filas de muestras: color de flecha, fondo de tarjeta/informe y color de nota.

### Corregido

- El nombre del tablero activo se ve correctamente en modo oscuro.
- Los desplegables de color se renderizan fuera de las tarjetas, así que no los
  recortan las notas, menús o informes.

## v1.2.0 — 2026-11-05

Interacción del workspace, edición de flechas, modo oscuro y nuevos workflows.

### Añadido

- Renombrar archivos y carpetas reales desde el tablero o el navegador de
  carpetas (clic derecho o lápiz). Se migran rutas, tarjetas, transcripciones,
  informes y archivos Markdown.
- Colores pastel de fondo para tarjetas, notas e informes (paleta de 9 colores).
- Tipos de flecha: recta y curva S. Las flechas se pueden seleccionar, mover
  completas, redimensionar arrastrando cualquiera de sus extremos, cambiar de
  color y eliminar desde una barra flotante.
- Modo oscuro con selector persistente claro/oscuro.
- Workflow **Plan de estudio**: convierte el contenido real de la clase en un
  plan breve y priorizado.
- Workflow **Fechas para calendario**: extrae exámenes, entregas y eventos y
  genera un archivo `.ics` junto al video.
- Traducción de informes: traduce cualquier informe a español o inglés
  conservando timestamps y evidencia.
- Conversión automática a MP4 para formatos que WebView2 no reproduce, con la
  acción “Convertir a MP4” en el visor.
- Verificación SHA256 de descargas de modelos, incluido el script de instalación.
- CI con GitHub Actions: tests de Go, vet, `govulncheck`, build del frontend,
  `npm audit` y pruebas E2E con Playwright.
- Suite E2E de Playwright con backend Wails simulado.

### Cambiado

- **Extracción de tareas** y **Puntos para parciales** ahora son estrictos: si
  no hay tareas o información explícita de parcial, devuelven una respuesta
  corta en vez de un informe largo. El material de estudio general pasa al nuevo
  workflow de plan de estudio.
- Los resultados se colocan automáticamente en el tablero que contiene el video.
- Los workflows auditan las herramientas completadas y abortan si se usó algo
  distinto de StudyBoard.
- Arrastrar tarjetas y flechas es más fluido: los nodos están memoizados y React
  Flow solo renderiza elementos visibles.

### Corregido

- Los informes ya no incluyen la narración intermedia del modelo.
- Renombrar un video renombra sus informes Markdown y migra su transcripción.
- Los resultados traducidos conservan idioma e información de origen.

## v1.1.1 — 2026-11-05

Versión de endurecimiento de seguridad.

### Seguridad

- Los workflows ejecutan OpenCode en una sesión standalone aislada con una
  configuración generada en línea. Se deniegan shell, edición de archivos,
  lectura directa, acceso web, subagentes, skills y LSP; solo se permiten las
  herramientas MCP de solo lectura de StudyBoard, incluso con `--auto`.
- Los workflows se ejecutan en un directorio vacío dedicado, no en el workspace
  de clases, evitando que la configuración del proyecto, `AGENTS.md`, plugins
  locales o DLLs plantadas influyan en el agente.
- Las transcripciones y nombres de archivo se marcan como datos no confiables;
  los metadatos se citan para que un nombre no pueda inyectar líneas al prompt.
- Las rutas absolutas del MCP quedan restringidas al workspace y los enlaces
  simbólicos/junctions se resuelven antes de acceder.
- El servidor local de medios falla cerrado si no puede generar el token
  aleatorio, sirve solo videos del workspace, rechaza rutas que resuelven fuera
  y ya no envía CORS comodín.
- Los informes bloquean imágenes remotas y HTML crudo. Los enlaces se limitan a
  `http`/`https`, rechazan userinfo y hosts locales/loopback, y se abren en el
  navegador del sistema en lugar de navegar dentro de la app.
- La salida de workflows tiene límite de tamaño y se eliminan caracteres de
  control, secuencias ANSI y overrides bidi Unicode antes de guardarla y
  exponerla por MCP.
- Las descargas de modelos exigen HTTPS, limitan redirecciones, verifican el
  tamaño exacto del catálogo y la cabecera GGML, y usan archivos temporales
  seguros.
- Estado, transcripciones e informes se escriben de forma atómica con permisos
  de usuario.
- Los binarios de motor y OpenCode ya no se buscan en el directorio actual.
- La compilación requiere Go 1.26.10, que incluye correcciones de
  vulnerabilidades conocidas de la librería estándar.

### Notas

- `npm audit` y `govulncheck` no reportan vulnerabilidades conocidas.
- Consulta `SECURITY.md` para el modelo de amenazas y los riesgos residuales.

## v1.1.0 — 2026-11-05

Workflows automáticos de OpenCode en el workspace.

### Añadido

- Acción **Ejecutar workflow** en la barra del tablero.
- Selector para ejecutar uno o varios workflows sobre una o varias clases a la
  vez, con salida en español o inglés.
- Workflow **Extracción de tareas**: actividades explícitas y probables,
  requisitos y preferencias del profesor, plazos, canal de entrega, modo
  individual/grupal, pistas de rúbrica, evidencia con timestamps y ambigüedades.
- Workflow **Puntos para parciales**: información explícita del parcial separada
  del énfasis del profesor, temas excluidos, formatos de pregunta, plan de
  estudio priorizado, preguntas de práctica y advertencias. Nunca inventa datos
  del parcial.
- Resultados como tarjetas tipo nota adhesiva junto al video de origen.
- Cada resultado se guarda además como Markdown junto al video
  (`<video> - tareas.md`, `<video> - parcial.md`).
- Panel de ejecuciones con etapa en vivo, cancelación y detalle de errores.
- Botón **Workflows** por video en el visor de transcripción.
- Herramientas MCP `list_workflow_results` y `get_workflow_result`.
- Recurso MCP `studyboard://workflow-results`.
- Prompts MCP `extraer-tareas` y `puntos-parciales`.
- Workflows declarativos en JSON con plantillas de prompt en inglés.

### Cambiado

- `overview` ahora informa resultados de workflows, notas y flechas por tablero.

### Notas

- Los workflows usan el CLI de OpenCode ya instalado y autenticado en la máquina.
  Cada ejecución crea una sesión nueva; StudyBoard no almacena credenciales del
  proveedor.
- Los prompts están en inglés para mayor fiabilidad del modelo. El idioma del
  informe lo elige el usuario.
- La transcripción se lee a través del servidor MCP de StudyBoard. Al ejecutar un
  workflow, el contenido consultado por el modelo se envía al proveedor
  configurado en OpenCode, igual que en cualquier sesión de OpenCode.

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

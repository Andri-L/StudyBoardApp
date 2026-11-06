You are executing the StudyBoard **Study Plan** workflow.

## Context

- Video: {{VIDEO_NAME}}
- Video path: {{VIDEO_PATH}}
- Class folder: {{CLASS_NAME}}
- Class date context: {{CLASS_DATE}}
- Required output language: {{OUTPUT_LANGUAGE}}

The video has already been transcribed. Use the StudyBoard MCP tools as the
only source of evidence:

- `studyboard_get_transcript` to read the complete transcript.
- `studyboard_search_transcripts` to locate exercises, formulas, or procedures.
- `studyboard_capture_frame` when the teacher shows a slide, formula, or board.

## Goal

Turn the actual content of this class into a short, practical study plan. This
is general study material, not an exam prediction.

## Rules

1. Select only the 3 to 7 most important elements of the class: concepts,
   formulas, procedures, comparisons, or exercises.
2. For each element, explain why it matters and give one concrete study action.
3. Order the plan by dependency: foundations first, then practice.
4. Include exercises or examples the teacher used. Do not invent new ones.
5. Cite the timestamp and a short verbatim quote for every element.
6. If the class has no substantive study content, stop after a short answer.
7. Never claim that something will be on an exam unless the teacher explicitly
   said so.

## Output format

Return only the final report in {{OUTPUT_LANGUAGE}}.

# Plan de estudio — <class>

## Resultado

If there is no substantive content, write one short paragraph:

**Sin elementos de estudio suficientes.** La clase no desarrolló contenido que
justifique un plan. Evidence: [hh:mm:ss] "quote" (or "sin menciones").

Otherwise:

## Elementos clave

### 1. <element>
- **Por qué importa:** ...
- **Acción de estudio:** ...
- **Evidencia:** [hh:mm:ss] "verbatim quote"

## Orden sugerido

1. ...
2. ...
3. ...

## Práctica

- Exercises, examples, or procedures from the class, with timestamps.

Do not ask follow-up questions. Do not include your reasoning. Return only the
report.

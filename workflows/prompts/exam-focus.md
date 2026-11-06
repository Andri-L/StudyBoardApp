You are executing the StudyBoard **Exam Focus** workflow.

## Context

- Video: {{VIDEO_NAME}}
- Video path: {{VIDEO_PATH}}
- Class folder: {{CLASS_NAME}}
- Class date context: {{CLASS_DATE}}
- Required output language: {{OUTPUT_LANGUAGE}}

The video has already been transcribed. Use the StudyBoard MCP tools as the
only source of evidence:

- `studyboard_get_transcript` to read the complete transcript.
- `studyboard_search_transcripts` to verify exam mentions.
- `studyboard_capture_frame` when a slide or board shows exam information.

## Goal

Report only what the teacher explicitly says about the exam. This is not a
general study plan and not a class summary.

## Strict rules

1. Develop a point only when the teacher explicitly connects it to an exam or
   evaluation: "this is on the exam", "this will be evaluated", "this is not
   on the exam", exam date, weight, format, allowed materials, question types.
2. Do not infer exam topics from emphasis, repetition, or long explanations.
   If the only signal is emphasis, do not list it as exam content.
3. Never invent exam details.
4. If the class contains no explicit exam information, stop after a short
   answer. Do not add a study plan, practice questions, formulas, warnings, or
   inferred topics.
5. Keep the report short. Prefer bullets over paragraphs.

## Output format

Return only the final report in {{OUTPUT_LANGUAGE}}.

# Parcial — <class>

## Resultado

If there is no explicit exam information, write exactly one short paragraph:

**Sin información explícita de parcial.** El profesor no mencionó temas,
fecha, formato ni criterios de evaluación. Evidence: [hh:mm:ss] "quote" (or
"sin menciones").

If explicit exam information exists, list only that information:

- **Temas incluidos:** ...
- **Temas excluidos:** ...
- **Formato:** ...
- **Fecha / duración / peso:** ...
- **Materiales permitidos:** ...
- **Evidencia:** [hh:mm:ss] "verbatim quote"

## Énfasis explícito del profesor

Only include this section if the teacher explicitly said a topic is important
for the exam. Do not include general emphasis.

- **Tema:** ... — **Evidencia:** [hh:mm:ss] "quote"

Do not ask follow-up questions. Do not include your reasoning. Return only the
report.

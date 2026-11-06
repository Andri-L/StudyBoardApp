You are executing the StudyBoard **Calendar Dates** workflow.

## Context

- Video: {{VIDEO_NAME}}
- Video path: {{VIDEO_PATH}}
- Class folder: {{CLASS_NAME}}
- Class date context: {{CLASS_DATE}}
- Required output language: {{OUTPUT_LANGUAGE}}

The video has already been transcribed. Use the StudyBoard MCP tools as the
only source of evidence:

- `studyboard_get_transcript` to read the complete transcript.
- `studyboard_search_transcripts` to verify dates, weeks, and deadlines.
- `studyboard_capture_frame` when a date is shown on a slide.

## Goal

Extract dates and events that a student could add to a calendar: exams,
deliveries, presentations, holidays, special classes, and deadlines.

## Strict rules

1. Only extract dates the teacher explicitly mentions or that are unambiguous
   from the class date ("next Friday", "July 30").
2. For relative dates, include the literal wording and label any calendar date
   as an inference.
3. Do not invent dates. If no date is mentioned, say so and do not produce an
   ICS block.
4. Include the timestamp and a short verbatim quote for every event.
5. Use ISO dates (YYYY-MM-DD) in the table when the date is known.
6. The ICS block must be valid: `BEGIN:VCALENDAR`, at least one `VEVENT` with
   `UID`, `DTSTAMP`, `DTSTART`, `SUMMARY`, and `END:VCALENDAR`. Use all-day
   events (`DTSTART;VALUE=DATE:YYYYMMDD`) when no time is stated.

## Output format

Return only the final report in {{OUTPUT_LANGUAGE}}.

# Fechas — <class>

## Resultado

If there are no dates, write one short paragraph:

**Sin fechas.** No se mencionaron fechas ni eventos en esta clase. Evidence:
[hh:mm:ss] "quote" (or "sin menciones").

Otherwise:

| Fecha | Evento | Hora | Evidencia | Nota |
|---|---|---|---|---|
| YYYY-MM-DD | ... | ... | [hh:mm:ss] "quote" | explícita / inferida |

## Archivo ICS

Never copy the example event; it only shows the required schema. Replace all
values with the extracted events. If no date exists, omit this entire section.

```ics
BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//StudyBoard//Fechas//ES
BEGIN:VEVENT
UID:studyboard-1@local
DTSTAMP:20260101T000000Z
DTSTART;VALUE=DATE:20260101
SUMMARY:Example event
DESCRIPTION:Replace this example with the extracted events.
END:VEVENT
END:VCALENDAR
```

Do not ask follow-up questions. Do not include your reasoning. Return only the
report.

You are executing the StudyBoard **Task Extraction** workflow.

## Context

- Video: {{VIDEO_NAME}}
- Video path: {{VIDEO_PATH}}
- Class folder: {{CLASS_NAME}}
- Class date context: {{CLASS_DATE}}
- Required output language: {{OUTPUT_LANGUAGE}}

The video has already been transcribed by StudyBoard. Use the StudyBoard MCP
tools as the source of truth:

- `studyboard_get_transcript` to read the complete transcript (or ranges).
- `studyboard_search_transcripts` to verify mentions across the class.
- `studyboard_capture_frame` when an instruction is shown on screen rather than
  only spoken.

Do not rely only on keyword search results. Read the full transcript first, then
use search to verify specific candidates.

## Goal

Find every pending activity, assignment, deliverable, or preparation the teacher
proposes in this class. The student may not have watched the class, so the output
must be actionable without watching it again.

## Extraction rules

1. Distinguish carefully:
   - **Explicit task**: the teacher clearly assigns work ("for next class",
     "you must submit", "in groups of...", "upload to...", "do not forget").
   - **Probable task**: strongly suggested but not fully assigned.
   - **Hypothetical/optional**: "you could", "if you want", "for example".
   - **Administrative note**: dates, channels, or materials without a task.
2. Capture the teacher's exact preferences and constraints, not just the topic:
   project theme restrictions, tools, format, length, language, dataset, team
   size, citations, presentation style, things the teacher explicitly does not
   want, and evaluation hints.
3. For each task record:
   - short title;
   - what must be done;
   - concrete requirements (or "not specified");
   - teacher preferences and restrictions;
   - deadline: quote the literal expression and mark any inferred date;
   - submission channel and format;
   - individual/group mode and team size if stated;
   - grading or rubric hints;
   - evidence timestamp and a short verbatim quote;
   - confidence: explicit, probable, or ambiguous.
4. Relative deadlines ("next class", "Friday") must include the literal phrase.
   Only infer a calendar date when the class date makes it unambiguous; label it
   as an inference.
5. If an instruction is visual (slide, screen, board), call
   `studyboard_capture_frame` and mention the exact timestamp. Do not describe a
   frame you did not inspect.
6. Never invent requirements, dates, tools, or submission channels. Write
   "not specified" when the transcript does not say it.
7. If there are no real tasks, say so clearly and list only administrative
   announcements separately. Do not turn examples into homework.

## Output format

Return only the final report in {{OUTPUT_LANGUAGE}}, using Markdown:

# Task extraction — <class>

## Summary
A short paragraph with the number of tasks and the most urgent one.

## Tasks

### 1. <title>
- **Status:** explicit / probable / ambiguous / optional
- **What to do:** ...
- **Requirements:** ...
- **Teacher preferences:** ...
- **Deadline:** ...
- **Submission:** ...
- **Team:** ...
- **Evidence:** [hh:mm:ss] "verbatim quote"
- **Confidence:** high / medium / low

(repeat for each task)

## Visual material to review
- [hh:mm:ss] what should be checked in the frame

## Ambiguities and missing information
- Questions the student should confirm with the teacher or the campus.

## Administrative notes
- Dates, channels, or announcements that are not tasks.

Do not ask follow-up questions. Do not include your reasoning. Return only the
report.

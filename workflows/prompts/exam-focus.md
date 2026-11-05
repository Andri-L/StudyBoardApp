You are executing the StudyBoard **Exam Focus** workflow.

## Context

- Video: {{VIDEO_NAME}}
- Video path: {{VIDEO_PATH}}
- Class folder: {{CLASS_NAME}}
- Class date context: {{CLASS_DATE}}
- Required output language: {{OUTPUT_LANGUAGE}}

The video has already been transcribed by StudyBoard. Use the StudyBoard MCP
tools as the source of truth:

- `studyboard_get_transcript` to read the complete transcript (or ranges).
- `studyboard_search_transcripts` to verify exam-related mentions.
- `studyboard_capture_frame` when a formula, slide, diagram, or instruction is
  shown on screen.

Read the full transcript first. Do not rely only on keyword search results.

## Goal

Extract everything useful for preparing an exam, while clearly separating what
the teacher actually said from what is only inferred from emphasis. The student
needs an effective study plan, not a generic summary.

## Extraction rules

1. Separate three evidence levels:
   - **Explicit exam information**: "this is on the exam", "this is not on the
     exam", exam date, weight, format, allowed materials, question types.
   - **Teacher emphasis**: repeated topics, "this is very important", long
     worked examples, warnings. Label these as inferences, not facts.
   - **High-yield class material**: concepts, formulas, procedures, comparisons
     and exercises that carry significant weight in the class.
2. Identify:
   - topics that will be evaluated;
   - topics explicitly excluded;
   - expected question formats (multiple choice, coding, derivation, essay,
     case study, oral, etc.);
   - date, duration, weight, and allowed materials if stated;
   - exercises, examples, or problem types the teacher used as practice;
   - common mistakes the teacher warned about;
   - formulas, definitions, and procedures that must be memorized or practiced.
3. For every claim, include a timestamp and a short verbatim quote.
4. If information comes from a slide or screen, call
   `studyboard_capture_frame` and reference the timestamp. Do not describe a
   frame you did not inspect.
5. Never invent exam details. If there is no explicit exam information, say so
   prominently and offer a high-yield study plan based only on class emphasis,
   clearly labeled as an inference.
6. If this does not look like a university class, do not pretend an exam exists.
   Adapt the output to the material and state the limitation.
7. Prioritize study actions by evidence and impact, not by topic order in the
   video.

## Output format

Return only the final report in {{OUTPUT_LANGUAGE}}, using Markdown:

# Exam focus — <class>

## Bottom line
What is explicitly known and what is only inferred.

## Explicit exam information
- **Topics included:** ...
- **Topics excluded:** ...
- **Format:** ...
- **Date / duration / weight:** ...
- **Allowed materials:** ...
- **Evidence:** [hh:mm:ss] "verbatim quote"

Use "not specified" when absent.

## Topics with emphasis (inference)
| Topic | Why it matters | Evidence | Confidence |
|---|---|---|---|

## High-yield study plan
1. ...
2. ...
3. ...

## Practice questions
- ...
- ...

## Formulas, definitions and procedures to review
- ...

## Warnings and common mistakes
- ...

## Gaps to confirm
- What the student should verify with the teacher or campus.

Do not ask follow-up questions. Do not include your reasoning. Return only the
report.

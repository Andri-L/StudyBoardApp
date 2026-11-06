export const PASTEL_COLORS = [
  '#fff4b8',
  '#ffe2e2',
  '#ffe8cc',
  '#e8f5e9',
  '#e0f2fe',
  '#ede9fe',
  '#fce7f3',
  '#e5e7eb',
  '#fef9c3',
] as const;

export const ARROW_COLORS = [
  '#0067c0',
  '#c42b1c',
  '#0f7b0f',
  '#b8860b',
  '#7c3aed',
  '#0f766e',
  '#db2777',
  '#475569',
  '#111827',
] as const;

export type PastelColor = (typeof PASTEL_COLORS)[number];
export type ArrowColor = (typeof ARROW_COLORS)[number];

import { test, expect, type Page } from '@playwright/test';

interface MockFixture {
  state: any;
}

function fixture(): MockFixture {
  const now = 1_700_000_000_000;
  return {
    state: {
      version: 1,
      workspace: 'C:\\Clases',
      activeBoardId: 'board-1',
      language: 'es',
      engine: {
        ready: true,
        whisperPath: 'C:\\engine\\whisper.exe',
        ffmpegPath: 'C:\\engine\\ffmpeg.exe',
        ffprobePath: 'C:\\engine\\ffprobe.exe',
        modelPath: 'C:\\engine\\model.bin',
        modelName: 'tiny',
        vadPath: '',
        gpu: 'CPU',
        missing: [],
        models: [],
      },
      mediaBase: '',
      boards: [
        {
          id: 'board-1',
          name: 'IA',
          viewport: { x: 0, y: 0, zoom: 1 },
          arrows: [
            {
              id: 'arrow-1',
              x1: 520,
              y1: 560,
              x2: 680,
              y2: 620,
              type: 'straight',
              color: '#0067c0',
              createdAt: now,
            },
            {
              id: 'arrow-2',
              x1: 232,
              y1: 90,
              x2: 545,
              y2: 500,
              type: 'square',
              color: '#c42b1c',
              createdAt: now,
            },
          ],
          notes: [
            {
              id: 'note-1',
              text: 'Nota de prueba',
              color: '#fff4b8',
              x: 430,
              y: 500,
              width: 230,
              height: 170,
              createdAt: now,
            },
          ],
          cards: [
            {
              id: 'card-1',
              kind: 'video',
              path: 'C:\\Clases\\IA\\22Jul26\\clase-ia-22jul26.mp4',
              label: 'clase-ia-22jul26.mp4',
              dateLabel: '22 jul 2026',
              x: 80,
              y: 90,
              width: 304,
              height: 260,
              addedAt: now,
            },
            {
              id: 'card-2',
              kind: 'folder',
              path: 'C:\\Clases\\IA\\22Jul26',
              label: '22Jul26',
              dateLabel: '22 jul 2026',
              x: 80,
              y: 400,
              width: 286,
              height: 124,
              addedAt: now,
            },
          ],
          results: [
            {
              id: 'result-1',
              workflowId: 'task-extraction',
              workflowName: 'Extracción de tareas',
              videoPath: 'C:\\Clases\\IA\\22Jul26\\clase-ia-22jul26.mp4',
              videoName: 'clase-ia-22jul26.mp4',
              className: '22Jul26',
              content:
                '# Tareas — 22Jul26\n\n## Resultado\n\n**Sin tareas.** [Enlace peligroso](javascript:alert(1)) y [Web](https://example.com).\n\n![tracking](https://tracker.test/pixel.png)',
              filePath: 'C:\\Clases\\IA\\22Jul26\\clase-ia-22jul26 - tareas.md',
              language: 'es',
              sessionId: 'ses_test',
              x: 430,
              y: 90,
              width: 430,
              height: 380,
              createdAt: now,
            },
          ],
        },
      ],
    },
  };
}

async function mockWails(page: Page) {
  const data = fixture();
  await page.addInitScript((initial) => {
    const clone = (value: any) => JSON.parse(JSON.stringify(value));
    const state = clone(initial.state);
    const calls: Array<{ name: string; args: any[] }> = [];
    (window as any).__calls = calls;
    const record = (name: string, args: any[]) => {
      calls.push({ name, args });
    };
    const app: Record<string, (...args: any[]) => Promise<any>> = {
      GetState: async () => clone(state),
      GetJobs: async () => [],
      GetWorkflows: async () => [
        { id: 'task-extraction', name: 'Extracción de tareas', shortName: 'Tareas', description: 'Tareas', icon: 'clipboard-list', outputFile: 'tareas' },
        { id: 'exam-focus', name: 'Puntos para parciales', shortName: 'Parciales', description: 'Parcial', icon: 'target', outputFile: 'parcial' },
        { id: 'study-plan', name: 'Plan de estudio', shortName: 'Plan', description: 'Plan', icon: 'book-open', outputFile: 'plan' },
        { id: 'calendar-dates', name: 'Fechas para calendario', shortName: 'Fechas', description: 'Fechas', icon: 'calendar', outputFile: 'fechas' },
      ],
      GetWorkflowVideos: async () => [
        { path: state.boards[0].cards[0].path, name: 'clase-ia-22jul26.mp4', class: '22Jul26', duration: 48, hasTranscript: true, segments: 3 },
      ],
      GetWorkflowJobs: async () => [],
      StartWorkflow: async (boardId: string, workflowId: string, videoPath: string, outputLanguage: string) => {
        record('StartWorkflow', [boardId, workflowId, videoPath, outputLanguage]);
        return {
          id: 'job-1', boardId, workflowId, workflowName: workflowId, videoPath,
          videoName: videoPath.split('\\').pop(), outputLanguage, status: 'queued',
          stage: 'En cola', error: '', sessionId: '', createdAt: Date.now(), startedAt: 0, finishedAt: 0,
        };
      },
      StartTranslation: async (boardId: string, resultId: string, targetLanguage: string) => {
        record('StartTranslation', [boardId, resultId, targetLanguage]);
        return {
          id: 'job-t', boardId, workflowId: 'task-extraction', workflowName: 'Traducción',
          videoPath: state.boards[0].results[0].videoPath, videoName: 'clase-ia-22jul26.mp4',
          outputLanguage: targetLanguage, kind: 'translation', sourceResultId: resultId,
          targetLanguage, status: 'queued', stage: 'En cola', error: '', sessionId: '',
          createdAt: Date.now(), startedAt: 0, finishedAt: 0,
        };
      },
      CancelWorkflow: async () => undefined,
      GetMediaInfo: async (path: string) => ({
        path, name: 'clase-ia-22jul26.mp4', size: 1000, modTime: 1,
        duration: 48, kind: 'video',
        transcript: { status: 'done', progress: 100, stage: 'Listo', segments: 3, language: 'es', createdAt: 1, error: '' },
      }),
      GetTranscript: async () => ({
        version: 1, mediaPath: state.boards[0].cards[0].path, mediaSize: 1000,
        mediaModTime: 1, duration: 48, language: 'es', model: 'tiny', createdAt: 1,
        segments: [{ start: 0, end: 4, text: 'Hola' }],
      }),
      ListDir: async () => [],
      SaveViewport: async () => undefined,
      SetActiveBoard: async () => clone(state),
      RenamePath: async (path: string, newName: string) => {
        record('RenamePath', [path, newName]);
        const card = state.boards[0].cards.find((item: any) => item.path === path);
        if (card) {
          const oldExt = path.match(/\.[^.]+$/)?.[0] || '';
          const finalName = /\.[^.]+$/.test(newName) ? newName : newName + oldExt;
          const newPath = path.replace(/[^\\/]+$/, finalName);
          card.path = newPath;
          card.label = finalName;
          const result = state.boards[0].results[0];
          result.videoPath = newPath;
          result.videoName = finalName;
          result.filePath = newPath.replace(/\.[^.]+$/, ' - tareas.md');
        }
        return clone(state);
      },
      SetItemColor: async (_boardId: string, itemId: string, color: string) => {
        record('SetItemColor', [itemId, color]);
        const card = state.boards[0].cards.find((item: any) => item.id === itemId);
        if (card) card.color = color;
        const result = state.boards[0].results.find((item: any) => item.id === itemId);
        if (result) result.color = color;
        return clone(state);
      },
      AddArrow: async (...args: any[]) => {
        record('AddArrow', args);
        return clone(state);
      },
      UpdateArrow: async (...args: any[]) => {
        record('UpdateArrow', args);
        const arrow = state.boards[0].arrows.find((item: any) => item.id === args[1]);
        if (arrow) Object.assign(arrow, { x1: args[2], y1: args[3], x2: args[4], y2: args[5], type: args[6], color: args[7] });
        return clone(state);
      },
      RemoveArrow: async () => clone(state),
      AddNote: async () => clone(state),
      UpdateNote: async () => clone(state),
      MoveNote: async () => undefined,
      RemoveNote: async () => clone(state),
      AddCard: async () => clone(state),
      MoveCard: async () => undefined,
      RemoveCard: async () => clone(state),
      CreateFolderCard: async () => clone(state),
      CopyText: async (text: string) => { record('CopyText', [text]); },
      Reveal: async (path: string) => { record('Reveal', [path]); },
      OpenExternal: async () => undefined,
      OpenURL: async (url: string) => { record('OpenURL', [url]); },
      PreparePlayback: async (path: string) => path,
      CaptureFrame: async () => '',
      ProbeMedia: async () => ({ path: '', duration: 0 }),
      SelectModel: async () => clone(state),
      SetLanguage: async () => clone(state),
      DownloadModel: async () => ({ file: '', status: 'done', progress: 100, bytes: 0, total: 0, error: '' }),
      CancelModelDownload: async () => undefined,
      OpenModelsFolder: async () => undefined,
      ExportTranscript: async () => '',
      CreateBoard: async () => clone(state),
      RenameBoard: async () => clone(state),
      DeleteBoard: async () => clone(state),
      ChooseWorkspace: async () => clone(state),
    };
    (window as any).go = { main: { App: app } };
    (window as any).runtime = {
      EventsOn: () => undefined,
      EventsOff: () => undefined,
      BrowserOpenURL: () => undefined,
    };
  }, data);
}

test.beforeEach(async ({ page }) => {
  await mockWails(page);
  await page.goto('/');
  await expect(page.locator('.board-toolbar')).toBeVisible();
});

test('loads the board and workspace cards', async ({ page }) => {
  await expect(page.getByRole('heading', { name: 'IA' })).toBeVisible();
  await expect(page.locator('.video-node')).toBeVisible();
  await expect(page.locator('.folder-node')).toBeVisible();
});

test('opens the workflow picker and queues a workflow', async ({ page }) => {
  await page.getByRole('button', { name: 'Ejecutar workflow' }).first().click();
  await expect(page.locator('.workflow-dialog')).toBeVisible();
  await page.getByRole('button', { name: /Plan de estudio/ }).click();
  await page.getByRole('button', { name: /clase-ia-22jul26\.mp4/ }).click();
  await page.locator('.workflow-dialog-foot').getByRole('button', { name: /Ejecutar workflow/ }).click();
  await expect.poll(async () =>
    page.evaluate(() => (window as any).__calls.some((call: any) => call.name === 'StartWorkflow')),
  ).toBe(true);
  const call = await page.evaluate(() => (window as any).__calls.find((item: any) => item.name === 'StartWorkflow'));
  expect(call.args[1]).toBe('study-plan');
});

test('switches to dark mode and persists the preference', async ({ page }) => {
  await page.getByRole('button', { name: 'Cambiar a modo oscuro' }).click();
  await expect.poll(async () => page.evaluate(() => document.documentElement.dataset.theme)).toBe('dark');
  await expect(page.getByRole('button', { name: 'Cambiar a modo claro' })).toBeVisible();
  expect(await page.evaluate(() => localStorage.getItem('studyboard.theme'))).toBe('dark');
  expect(await page.locator('.board-toolbar').evaluate((el) => getComputedStyle(el).backgroundColor)).toBe('rgb(32, 32, 32)');
  expect(await page.locator('.video-node').evaluate((el) => getComputedStyle(el).backgroundColor)).toBe('rgb(32, 32, 32)');
  const activeName = page.locator('.board-row.active .board-name');
  await expect(activeName).toBeVisible();
  expect(await activeName.evaluate((el) => getComputedStyle(el).color)).not.toBe('rgb(0, 62, 115)');
});

test('sanitizes model-generated links and images', async ({ page }) => {
  const body = page.locator('.workflow-result-body');
  await expect(body).toBeVisible();
  await expect(body.locator('img')).toHaveCount(0);
  await expect(body.locator('.md-image-blocked')).toHaveCount(1);
  await expect(body.locator('.md-link-blocked')).toHaveCount(1);
  await body.locator('.md-link').click();
  await expect.poll(async () =>
    page.evaluate(() => (window as any).__calls.some((call: any) => call.name === 'OpenURL')),
  ).toBe(true);
});

test('renames a real item from the right-click menu', async ({ page }) => {
  await page.locator('.video-node').click({ button: 'right' });
  await page.getByRole('button', { name: 'Renombrar en el disco' }).click();
  const input = page.locator('#rename-input');
  await input.fill('clase-nueva');
  await page.getByRole('button', { name: 'Renombrar', exact: true }).click();
  await expect.poll(async () =>
    page.evaluate(() => (window as any).__calls.some((call: any) => call.name === 'RenamePath')),
  ).toBe(true);
  await expect(page.locator('.video-node')).toContainText('clase-nueva.mp4');
});

test('changes a card background from the right-click palette', async ({ page }) => {
  await page.locator('.video-node').click({ button: 'right' });
  await page.locator('.node-menu-palette .color-picker-trigger').click();
  await page.locator('.color-picker-panel .color-swatch').first().click();
  await expect.poll(async () =>
    page.evaluate(() => (window as any).__calls.some((call: any) => call.name === 'SetItemColor')),
  ).toBe(true);
  await expect(page.locator('.video-node')).toHaveCSS('background-color', 'rgb(255, 244, 184)');
});

test('selects an arrow and changes its type', async ({ page }) => {
  await page.locator('.board-arrow').first().locator('path').first().click({ force: true });
  await expect(page.locator('.selection-toolbar')).toContainText('Flecha');
  await page.getByRole('button', { name: /Curva S/ }).click();
  await expect.poll(async () =>
    page.evaluate(() => (window as any).__calls.some((call: any) => call.name === 'UpdateArrow' && call.args[6] === 's')),
  ).toBe(true);
  await page.getByRole('button', { name: /Cuadrada/ }).click();
  await expect.poll(async () =>
    page.evaluate(() => (window as any).__calls.some((call: any) => call.name === 'UpdateArrow' && call.args[6] === 'square')),
  ).toBe(true);
});

test('drags an arrow endpoint and persists the geometry', async ({ page }) => {
  await page.locator('.board-arrow').first().locator('path').first().click({ force: true });
  const handle = page.locator('.arrow-handle').first();
  const box = await handle.boundingBox();
  expect(box).not.toBeNull();
  if (!box) return;
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
  await page.mouse.down();
  await page.mouse.move(box.x + box.width / 2 + 35, box.y + box.height / 2 + 25, { steps: 6 });
  await page.mouse.up();
  await expect.poll(async () =>
    page.evaluate(() => {
      const call = (window as any).__calls.find((item: any) => item.name === 'UpdateArrow');
      return Boolean(call && (call.args[2] !== 520 || call.args[3] !== 560));
    }),
  ).toBe(true);
});

test('snaps an arrow endpoint to the center of a card edge', async ({ page }) => {
  await page.locator('.board-arrow').first().locator('path').first().click({ force: true });
  const card = await page.locator('.video-node').boundingBox();
  const handle = await page.locator('.arrow-handle').first().boundingBox();
  expect(card).not.toBeNull();
  expect(handle).not.toBeNull();
  if (!card || !handle) return;
  await page.mouse.move(handle.x + handle.width / 2, handle.y + handle.height / 2);
  await page.mouse.down();
  await page.mouse.move(card.x + card.width + 9, card.y + card.height / 2 + 12, { steps: 4 });
  await page.mouse.up();
  await expect.poll(async () =>
    page.evaluate(() => {
      const call = (window as any).__calls.filter((item: any) => item.name === 'UpdateArrow').at(-1);
      return call ? { x: Math.round(call.args[2]), y: Math.round(call.args[3]) } : null;
    }),
  ).toEqual({ x: 384, y: 220 });
});

test('routes a square arrow around snapped cards', async ({ page }) => {
  const path = await page.locator('.board-arrow').nth(1).locator('path').nth(1).getAttribute('d');
  expect(path).toContain('545 66');
  expect(path).toContain('545 476');
});

test('queues a report translation from the result card', async ({ page }) => {
  await page.getByRole('button', { name: 'Traducir informe' }).click();
  await page.getByRole('button', { name: 'English', exact: true }).click();
  await expect.poll(async () =>
    page.evaluate(() => (window as any).__calls.some((call: any) => call.name === 'StartTranslation')),
  ).toBe(true);
});


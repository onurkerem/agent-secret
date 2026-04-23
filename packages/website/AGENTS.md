# packages/website — agent-secret marketing site

Astro-based marketing website for agent-secret.

## Stack
- Astro 6, Tailwind 4, TypeScript

## Commands
- `npm run dev` — start dev server
- `npm run build` — production build
- `npm run preview` — preview production build
- `npm run check` — type check

## Structure
- `src/pages/` — Astro pages (single-page: index.astro)
- `src/layouts/` — layout components
- `src/styles/` — global CSS with Tailwind v4 theme tokens

## Rules
- All content must reflect actual CLI behavior — read the README and command source before changing copy.
- The terminal components use a dark inverse-surface style with macOS traffic lights. Maintain this pattern.
- No component extraction unless the site grows beyond one page.
- Accent color is muted green (#006d4a). The full tonal palette is in `src/styles/global.css`.

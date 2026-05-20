# FRONTEND

React 19 SPA admin dashboard. Vite 8 + Tailwind v4 + shadcn/ui.

## STRUCTURE

```
src/
├── main.tsx              # Entry → App.tsx
├── App.tsx               # Router + auth guard + layout shell
├── pages/                # Route pages (all lazy-loaded)
│   ├── SetupPage, LoginPage, OverviewPage
│   ├── KeysPage, ProvidersPage, ModelsPage
│   └── UsagePage          # Charts: Recharts stacked bars, custom dropdowns
├── components/
│   ├── Sidebar, GenerateKeyDialog, ToastContainer
│   └── ui/               # 13 shadcn/ui primitives
├── lib/
│   ├── api.ts            # Single fetch wrapper, all endpoints
│   ├── auth.tsx           # AuthContext provider
│   ├── utils.ts           # cn() helper
│   └── useDebounce.ts
└── hooks/
    ├── use-theme.tsx
    └── use-toast.tsx
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add a page | `src/pages/` + `App.tsx` + `Sidebar.tsx` | Lazy via `React.lazy()` |
| Add API call | `src/lib/api.ts` | Object `api` with named methods |
| Add UI component | `src/components/` or `ui/` | shadcn primitives in `ui/` |
| Change design tokens | `src/index.css` `@theme` block | CSS-based, no JS config |
| Change auth flow | `src/lib/auth.tsx` | AuthContext + useAuth hook |
| Add a hook | `src/hooks/` | Co-locate with component if single-use |
| Add chart/widget | `src/pages/UsagePage.tsx` | Recharts, custom dropdowns, CSS variable theming |

## CONVENTIONS

- **TypeScript**: strict, `verbatimModuleSyntax` (MUST `import type`), `noUnusedLocals`, `noUnusedParameters`, `erasableSyntaxOnly` (no runtime `enum`)
- **Tailwind v4**: Config in `src/index.css` `@theme` block only. No `tailwind.config.js`.
- **Design tokens**: `canvas`/`ink`/`body`/`hairline` (surfaces+text), `accent-sunset`/`dusk`/`twilight`/`breeze`/`midnight` (accents)
- **Radius**: single `--radius: 8px` everywhere
- **Animations**: `pulse-dot`, `fade-in-up`, `shimmer`, `skeleton`, `fade-in-stagger` (cascading children)
- **Fonts**: Inter (variable), Geist Mono (variable)
- **API client**: `api` object with methods: `checkSetup`, `setup`, `login`, `getMe`, `getStats`, `getUsageLogs`, `listProviders`, `createProvider`, `deleteProvider`, `listKeys`, `createKey`, `toggleKey`, `deleteKey`, `listModels`, `toggleModel`, `updateModel`
- **Key deps**: React Router v7, `@base-ui/react` v1, recharts v3.8.1, lucide-react
- **Vite proxy**: `/api` + `/v1` → `localhost:8080`
- **ESLint**: flat config (ESLint 10), typescript-eslint recommended, react-hooks, react-refresh
- **Charts**: CSS variables for ALL colors (`--color-chart-*`, `--color-card`, etc.), custom tooltips/legends

## ANTI-PATTERNS

- No `tailwind.config.js` exists; editing it is wrong
- Runtime `enum` banned by `erasableSyntaxOnly`; use string unions
- `import type` mandatory for type-only imports; `verbatimModuleSyntax` enforces this

## TESTS

- Vitest configured (`vitest.config.ts`): jsdom environment, globals enabled
- Setup: `src/test/setup.ts` imports `@testing-library/jest-dom`
- Test files: `KeysPage.test.tsx`, `GenerateKeyDialog.test.tsx`, `setup.test.ts`
- Run: `npm run test` (vitest run)

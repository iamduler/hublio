# Hublio — Design Rules

> Brand tokens ported from the Figma design system · Integration + Orchestration platform  
> Keep CSS variables — never hardcode hex in components when a token exists.

---

## Brand identity

| | |
|---|---|
| **Name** | Hublio |
| **Positioning** | Business Integration Platform + Business Orchestration Platform |
| **Tone** | Precise · connected · calm — trustworthy ops tooling |

---

## CSS tokens — required

```css
:root {
  --primary:      #2563EB;
  --primary-ink:  #1D4ED8;
  --primary-soft: #EFF6FF;
  --accent:       #3B82F6;

  --slate-900: #0F172A;
  --slate-800: #1E293B;

  --ink:     #0F172A;
  --ink-2:   #334155;
  --muted:   #64748B;
  --line:    #E2E8F0;
  --bg:      #FAFAFA;
  --surface: #FFFFFF;

  --success: #16A34A;
  --amber:   #D97706;
  --danger:  #DC2626;

  --radius: 0.75rem;
  --r-sm: 8px;
  --r:    12px;
  --r-lg: 18px;
}
```

Dark mode remaps **neutrals only** (`--bg`, `--white`, `--ink*`, `--line*`). Keep `--primary` (blue) and slate chrome.

---

## Typography

| Font | Use |
|------|-----|
| **Inter** (400–800) | Display / headings and UI body, labels, buttons |
| **JetBrains Mono** (400–600) | Code, payloads, IDs, monospaced values |

Match the Figma design system: Inter for all UI text, JetBrains Mono for technical/monospaced content.

---

## UI rules

- Prefer CSS variables over raw hex.
- One job per section; avoid dashboard clutter on marketing surfaces.
- Loading / empty / error / success states for async views.
- shadcn primitives live in `components/ui`.

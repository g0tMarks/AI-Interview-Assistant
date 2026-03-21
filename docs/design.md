# design.md — Frontend Design System
**Style:** Warm Minimalism
**Context:** Education / EdTech tool
**Last updated:** 2026-03-21

---

## 1. Design Philosophy

This frontend uses **Warm Minimalism** — a visual language that is calm, considered, and materially grounded. It prioritises clarity over decoration, breathing room over density, and warmth over clinical neutrality.

The interface should feel like a well-designed notebook or a thoughtfully printed textbook — not a SaaS dashboard or a marketing page. Every design decision should serve the learner's focus, not compete with it.

**Core principles:**
- Hierarchy through weight and spacing, never through decoration
- Quiet by default — elements earn attention, they don't demand it
- Warmth is non-negotiable; nothing should feel cold, corporate, or sterile
- Responsive first — the experience must be equally intentional on mobile and desktop

---

## 2. Color Palette

All colors are drawn from a warm, natural material palette. Never use pure black or pure white.

### CSS Variables

```css
:root {
  /* Surfaces */
  --color-bg:           #FAF8F5;   /* Warm off-white — primary background */
  --color-bg-subtle:    #F2EDE8;   /* Warm cream — cards, sidebars, inputs */
  --color-bg-muted:     #EDE8E3;   /* Deeper warm — hover states, dividers */
  --color-border:       #E5DDD5;   /* Warm border — inputs, cards, dividers */
  --color-border-strong:#C8BDB5;   /* Stronger border — focused states, separators */

  /* Text */
  --color-text-primary: #1C1917;   /* Warm near-black — headings, primary body */
  --color-text-secondary:#44403C;  /* Warm dark-grey — supporting text, labels */
  --color-text-muted:   #78716C;   /* Warm mid-grey — captions, placeholders */
  --color-text-inverse: #FAF8F5;   /* For use on dark/accent backgrounds */

  /* Accent — Terracotta / Clay */
  --color-accent:       #C4714A;   /* Primary accent — CTAs, highlights, links */
  --color-accent-hover: #A85C38;   /* Darker accent — hover/active states */
  --color-accent-subtle:#F5E8E0;   /* Tinted accent — badges, tags, highlights */
  --color-accent-border:#E0B89E;   /* Accent-family border */

  /* Semantic */
  --color-success:      #7D9B76;   /* Warm sage green */
  --color-warning:      #B8973E;   /* Aged gold */
  --color-error:        #C0533A;   /* Deep warm red */
  --color-info:         #6B8FA3;   /* Muted slate blue — only cool tone permitted */
}
```

### Rules
- **Never** use `#000000` or `#FFFFFF` — always use the warm equivalents above
- **Never** use cool greys, blue-greys, or any tone without warm undertones
- The terracotta accent is the **only** saturated color; use it sparingly and intentionally
- Semantic colors (success, warning, error) are permitted for functional feedback only

---

## 3. Typography

### Typefaces

Three font families are used, all loaded locally. Files live in `/fonts/` relative to your CSS entry point.

| Variable | Family | Role |
|---|---|---|
| `--font-sans` | Styrene A | All UI chrome — nav, labels, buttons, metadata |
| `--font-serif` | Tiempos Text | Body prose, long-form reading content |
| `--font-display` | Tiempos Headline | Large headings (h1, hero text) |
| `--font-mono` | system monospace | Code blocks |

> **Tiempos Fine** is available in your library but not used in the system — it is optimised for very large display sizes (60px+) and is unlikely to be needed in a typical edtech UI. Add it if you introduce a true hero display moment.

---

### @font-face Declarations

Paste this block at the top of your global CSS file (e.g. `globals.css` or `index.css`). Assumes fonts are stored in a `fonts/` directory at the same level.

```css
/* ─── Styrene A (UI / Sans) ─────────────────────────────────────────── */

@font-face {
  font-family: 'Styrene A';
  src: url('../fonts/StyreneA-Thin.otf') format('opentype');
  font-weight: 100;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Styrene A';
  src: url('../fonts/StyreneA-ThinItalic.otf') format('opentype');
  font-weight: 100;
  font-style: italic;
  font-display: swap;
}
@font-face {
  font-family: 'Styrene A';
  src: url('../fonts/StyreneA-Light.otf') format('opentype');
  font-weight: 300;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Styrene A';
  src: url('../fonts/StyreneA-LightItalic.otf') format('opentype');
  font-weight: 300;
  font-style: italic;
  font-display: swap;
}
@font-face {
  font-family: 'Styrene A';
  src: url('../fonts/StyreneA-Regular.otf') format('opentype');
  font-weight: 400;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Styrene A';
  src: url('../fonts/StyreneA-RegularItalic.otf') format('opentype');
  font-weight: 400;
  font-style: italic;
  font-display: swap;
}
@font-face {
  font-family: 'Styrene A';
  src: url('../fonts/StyreneA-Medium.otf') format('opentype');
  font-weight: 500;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Styrene A';
  src: url('../fonts/StyreneA-MediumItalic.otf') format('opentype');
  font-weight: 500;
  font-style: italic;
  font-display: swap;
}
@font-face {
  font-family: 'Styrene A';
  src: url('../fonts/StyreneA-Bold.otf') format('opentype');
  font-weight: 700;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Styrene A';
  src: url('../fonts/StyreneA-BoldItalic.otf') format('opentype');
  font-weight: 700;
  font-style: italic;
  font-display: swap;
}
@font-face {
  font-family: 'Styrene A';
  src: url('../fonts/StyreneA-Black.otf') format('opentype');
  font-weight: 900;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Styrene A';
  src: url('../fonts/StyreneA-BlackItalic.otf') format('opentype');
  font-weight: 900;
  font-style: italic;
  font-display: swap;
}

/* ─── Tiempos Text (Body / Prose) ───────────────────────────────────── */

@font-face {
  font-family: 'Tiempos Text';
  src: url('../fonts/TiemposText-Regular.otf') format('opentype');
  font-weight: 400;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Text';
  src: url('../fonts/TiemposText-RegularItalic.otf') format('opentype');
  font-weight: 400;
  font-style: italic;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Text';
  src: url('../fonts/TiemposText-Medium.otf') format('opentype');
  font-weight: 500;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Text';
  src: url('../fonts/TiemposText-MediumItalic.otf') format('opentype');
  font-weight: 500;
  font-style: italic;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Text';
  src: url('../fonts/TiemposText-Semibold.otf') format('opentype');
  font-weight: 600;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Text';
  src: url('../fonts/TiemposText-SemiboldItalic.otf') format('opentype');
  font-weight: 600;
  font-style: italic;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Text';
  src: url('../fonts/TiemposText-Bold.otf') format('opentype');
  font-weight: 700;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Text';
  src: url('../fonts/TiemposText-BoldItalic.otf') format('opentype');
  font-weight: 700;
  font-style: italic;
  font-display: swap;
}

/* ─── Tiempos Headline (Display / Large Headings) ────────────────────── */

@font-face {
  font-family: 'Tiempos Headline';
  src: url('../fonts/TiemposHeadline-Light.otf') format('opentype');
  font-weight: 300;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Headline';
  src: url('../fonts/TiemposHeadline-LightItalic.otf') format('opentype');
  font-weight: 300;
  font-style: italic;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Headline';
  src: url('../fonts/TiemposHeadline-Regular.otf') format('opentype');
  font-weight: 400;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Headline';
  src: url('../fonts/TiemposHeadline-RegularItalic.otf') format('opentype');
  font-weight: 400;
  font-style: italic;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Headline';
  src: url('../fonts/TiemposHeadline-Medium.otf') format('opentype');
  font-weight: 500;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Headline';
  src: url('../fonts/TiemposHeadline-MediumItalic.otf') format('opentype');
  font-weight: 500;
  font-style: italic;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Headline';
  src: url('../fonts/TiemposHeadline-Semibold.otf') format('opentype');
  font-weight: 600;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Headline';
  src: url('../fonts/TiemposHeadline-SemiboldItalic.otf') format('opentype');
  font-weight: 600;
  font-style: italic;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Headline';
  src: url('../fonts/TiemposHeadline-Bold.otf') format('opentype');
  font-weight: 700;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Headline';
  src: url('../fonts/TiemposHeadline-BoldItalic.otf') format('opentype');
  font-weight: 700;
  font-style: italic;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Headline';
  src: url('../fonts/TiemposHeadline-Black.otf') format('opentype');
  font-weight: 900;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'Tiempos Headline';
  src: url('../fonts/TiemposHeadline-BlackItalic.otf') format('opentype');
  font-weight: 900;
  font-style: italic;
  font-display: swap;
}
```

> **Note on font paths:** The `../fonts/` path assumes your CSS lives one level below the project root (e.g. `src/globals.css`). Adjust to `./fonts/` if your CSS is at the root, or use an absolute path if your build tool requires it (e.g. `/fonts/` in a Vite/Next.js project with a `public/fonts/` directory).

---

### Font Stack Variables

```css
:root {
  --font-sans:    'Styrene A', system-ui, sans-serif;
  --font-serif:   'Tiempos Text', Georgia, serif;
  --font-display: 'Tiempos Headline', Georgia, serif;
  --font-mono:    'SF Mono', 'Fira Code', 'Cascadia Code', monospace;
}
```

---

### Phosphor Icons

Icons are loaded via the Phosphor CDN. Use the `regular` weight by default; `bold` for emphasis or touch targets; `light` sparingly for decorative contexts.

```html
<!-- In your <head> -->
<script src="https://unpkg.com/@phosphor-icons/web"></script>
```

**Usage:**
```html
<!-- Regular (default) -->
<i class="ph ph-book-open"></i>

<!-- Bold weight -->
<i class="ph-bold ph-check-circle"></i>

<!-- Light weight -->
<i class="ph-light ph-arrow-right"></i>
```

**Sizing with CSS:**
```css
.icon-sm { font-size: 16px; }   /* Inline, labels */
.icon-md { font-size: 20px; }   /* Buttons, nav items */
.icon-lg { font-size: 24px; }   /* Section headers, empty states */
```

**Icon rules:**
- Always inherit or explicitly set `color` — never hardcode icon colors separately from their context
- Use `ph-bold` weight for icons inside buttons and touch targets (improves legibility at small sizes)
- Pair icons with visible text labels wherever space allows — never icon-only on primary actions
- Do not mix Phosphor weights within the same component

**Recommended icons for edtech UI:**

| Context | Icon |
|---|---|
| Lesson / reading | `ph-book-open` |
| Progress / complete | `ph-check-circle` |
| Quiz / assessment | `ph-pencil-simple` |
| Video content | `ph-play-circle` |
| Discussion | `ph-chat-circle` |
| Settings | `ph-sliders` |
| Navigation back | `ph-arrow-left` |
| External link | `ph-arrow-square-out` |
| Search | `ph-magnifying-glass` |
| User / profile | `ph-user-circle` |

---

### Type Scale

```css
:root {
  --text-xs:   0.75rem;    /* 12px — captions, metadata */
  --text-sm:   0.875rem;   /* 14px — labels, secondary UI */
  --text-base: 1rem;       /* 16px — body text */
  --text-lg:   1.125rem;   /* 18px — lead text, intro paragraphs */
  --text-xl:   1.25rem;    /* 20px — card headings */
  --text-2xl:  1.5rem;     /* 24px — section headings */
  --text-3xl:  1.875rem;   /* 30px — page headings */
  --text-4xl:  2.25rem;    /* 36px — hero headings */
  --text-5xl:  3rem;       /* 48px — display headings (use sparingly) */
}
```

### Typography Rules

```css
/* Large display headings — Tiempos Headline */
h1 {
  font-family: var(--font-display);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: -0.02em;
  line-height: 1.15;
}

/* Section headings — Tiempos Headline, lighter weight */
h2, h3 {
  font-family: var(--font-display);
  font-weight: 500;
  color: var(--color-text-primary);
  letter-spacing: -0.015em;
  line-height: 1.25;
}

/* Component/card headings — Styrene A */
h4, h5, h6 {
  font-family: var(--font-sans);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: -0.01em;
  line-height: 1.3;
}

/* Body prose — Tiempos Text */
body, p, li, blockquote {
  font-family: var(--font-serif);
  font-weight: 400;
  color: var(--color-text-secondary);
  line-height: 1.75;
}

/* UI labels, nav, buttons, metadata — Styrene A */
button, label, nav, .label, .caption, .meta {
  font-family: var(--font-sans);
}

.label, .caption {
  font-weight: 500;
  font-size: var(--text-sm);
  letter-spacing: 0.01em;
  color: var(--color-text-muted);
}

/* Code */
code, pre {
  font-family: var(--font-mono);
  font-size: 0.9em;
  background: var(--color-bg-muted);
  color: var(--color-text-primary);
}
```

**Typography rules:**
- `--font-display` (Tiempos Headline) for h1–h3; `--font-sans` (Styrene A) for h4–h6 and all UI chrome
- `--font-serif` (Tiempos Text) for all reading content — lesson text, descriptions, explanations
- Left-align all body text and most headings; centre only for short, deliberate moments (e.g. empty states, hero sections)
- No all-caps except for very short metadata labels (max 3–4 words) set in Styrene A
- Never exceed `font-weight: 700` in Tiempos; never exceed `font-weight: 900` in Styrene A (Black reserved for rare display use only)
- Generous `line-height` (1.7–1.75) for all Tiempos Text — readability is paramount

---

## 4. Spacing System

Based on a 4px base unit. All spacing must use this scale.

```css
:root {
  --space-1:  0.25rem;   /*  4px */
  --space-2:  0.5rem;    /*  8px */
  --space-3:  0.75rem;   /* 12px */
  --space-4:  1rem;      /* 16px */
  --space-5:  1.25rem;   /* 20px */
  --space-6:  1.5rem;    /* 24px */
  --space-8:  2rem;      /* 32px */
  --space-10: 2.5rem;    /* 40px */
  --space-12: 3rem;      /* 48px */
  --space-16: 4rem;      /* 64px */
  --space-20: 5rem;      /* 80px */
  --space-24: 6rem;      /* 96px */
}
```

**Spacing principles:**
- Be generous — breathing room communicates calm and clarity
- Section padding: `--space-16` to `--space-24` on desktop; `--space-10` to `--space-16` on mobile
- Component internal padding: `--space-6` to `--space-8`
- Stack gaps between related elements: `--space-4` to `--space-6`
- Stack gaps between sections: `--space-12` to `--space-20`

---

## 5. Shape & Borders

```css
:root {
  --radius-sm:   4px;    /* Inputs, tags, badges */
  --radius-md:   6px;    /* Cards, buttons, panels */
  --radius-lg:   10px;   /* Modals, large containers */
  --radius-full: 9999px; /* Pills — use sparingly */

  --border-width: 1px;
  --border-color: var(--color-border);
  --border: var(--border-width) solid var(--color-border);
  --border-strong: var(--border-width) solid var(--color-border-strong);
}
```

**Shape rules:**
- Use `--radius-md` (6px) as the default for most interactive elements
- No sharp 0px corners — this is Warm Minimalism, not Brutalism
- No large pill shapes on primary UI — reserve `--radius-full` for small tags or avatar indicators only
- All borders use warm tones from the palette; never `#ccc`, `#ddd`, or any cool-grey

---

## 6. Shadows & Elevation

Shadows must be subtle, warm-tinted, and used sparingly to denote elevation — not decoration.

```css
:root {
  --shadow-sm:  0 1px 2px rgba(101, 70, 50, 0.06);
  --shadow-md:  0 2px 8px rgba(101, 70, 50, 0.08), 0 1px 2px rgba(101, 70, 50, 0.04);
  --shadow-lg:  0 8px 24px rgba(101, 70, 50, 0.10), 0 2px 6px rgba(101, 70, 50, 0.05);
  --shadow-focus: 0 0 0 3px rgba(196, 113, 74, 0.20); /* Accent-tinted focus ring */
}
```

**Shadow rules:**
- No harsh, dark, or large drop shadows
- `--shadow-sm` for cards and inputs at rest
- `--shadow-md` for cards on hover, popovers
- `--shadow-lg` for modals and overlays only
- All focus states use `--shadow-focus` — never the browser default blue ring

---

## 7. Components

### Buttons

```css
/* Primary */
.btn-primary {
  background: var(--color-accent);
  color: var(--color-text-inverse);
  border: 1px solid var(--color-accent);
  border-radius: var(--radius-md);
  padding: var(--space-3) var(--space-6);
  font-weight: 500;
  font-size: var(--text-base);
  transition: background 150ms ease, box-shadow 150ms ease;
}
.btn-primary:hover {
  background: var(--color-accent-hover);
  box-shadow: var(--shadow-sm);
}

/* Secondary */
.btn-secondary {
  background: transparent;
  color: var(--color-text-primary);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-md);
  padding: var(--space-3) var(--space-6);
  font-weight: 500;
}
.btn-secondary:hover {
  background: var(--color-bg-muted);
}

/* Ghost */
.btn-ghost {
  background: transparent;
  color: var(--color-accent);
  border: none;
  padding: var(--space-3) var(--space-4);
  font-weight: 500;
}
.btn-ghost:hover {
  background: var(--color-accent-subtle);
}
```

### Cards

```css
.card {
  background: var(--color-bg-subtle);
  border: var(--border);
  border-radius: var(--radius-md);
  padding: var(--space-6) var(--space-8);
  box-shadow: var(--shadow-sm);
  transition: box-shadow 200ms ease;
}
.card:hover {
  box-shadow: var(--shadow-md);
}
```

### Inputs & Form Fields

```css
.input {
  background: var(--color-bg-subtle);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: var(--space-3) var(--space-4);
  font-family: var(--font-sans);
  font-size: var(--text-base);
  color: var(--color-text-primary);
  transition: border-color 150ms ease, box-shadow 150ms ease;
  width: 100%;
}
.input:focus {
  outline: none;
  border-color: var(--color-accent);
  box-shadow: var(--shadow-focus);
}
.input::placeholder {
  color: var(--color-text-muted);
}
```

### Tags & Badges

```css
.tag {
  display: inline-flex;
  align-items: center;
  background: var(--color-accent-subtle);
  color: var(--color-accent-hover);
  border: 1px solid var(--color-accent-border);
  border-radius: var(--radius-full);
  padding: var(--space-1) var(--space-3);
  font-size: var(--text-xs);
  font-weight: 500;
}
```

---

## 8. Responsive Design

This design system is **mobile-first**. All base styles target small screens; breakpoints layer in progressively.

### Breakpoints

```css
:root {
  --bp-sm:  480px;   /* Large phones */
  --bp-md:  768px;   /* Tablets */
  --bp-lg:  1024px;  /* Small desktops / landscape tablets */
  --bp-xl:  1280px;  /* Standard desktop */
  --bp-2xl: 1536px;  /* Wide desktop */
}

/* Usage */
@media (min-width: 768px)  { /* tablet+  */ }
@media (min-width: 1024px) { /* desktop+ */ }
@media (min-width: 1280px) { /* wide+    */ }
```

### Layout Grid

```css
.container {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 var(--space-4);        /* 16px on mobile */
}

@media (min-width: 768px) {
  .container { padding: 0 var(--space-8); }  /* 32px on tablet */
}

@media (min-width: 1024px) {
  .container { padding: 0 var(--space-12); } /* 48px on desktop */
}
```

### Responsive Typography

```css
/* Scale headings down on mobile */
h1 { font-size: var(--text-3xl); }
h2 { font-size: var(--text-2xl); }
h3 { font-size: var(--text-xl); }

@media (min-width: 1024px) {
  h1 { font-size: var(--text-4xl); }
  h2 { font-size: var(--text-3xl); }
  h3 { font-size: var(--text-2xl); }
}
```

### Touch Targets
- All interactive elements must be at minimum **44px tall** on mobile
- Increase button padding on mobile: `padding: var(--space-4) var(--space-6)`
- Ensure tap targets have at least `--space-2` gap between them

---

## 9. Motion & Interaction

Transitions should feel unhurried but responsive. Nothing should snap or jolt.

```css
:root {
  --transition-fast:   100ms ease;
  --transition-base:   150ms ease;
  --transition-slow:   250ms ease;
  --transition-layout: 300ms ease;
}

/* Default interactive transition */
button, a, input, .card {
  transition: all var(--transition-base);
}
```

**Motion rules:**
- No bouncy or elastic animations on functional UI
- Use `opacity` and `transform` for performance — avoid animating `height`, `width`, or `margin`
- Page/section entrances: subtle `fade-in` + `translateY(8px)` over `--transition-layout`
- Hover states are the primary motion moment — keep them gentle

---

## 10. Accessibility

- All text must meet **WCAG AA** contrast ratio (4.5:1 for body, 3:1 for large text)
- Focus states must always be visible — use `--shadow-focus` on all interactive elements
- Never remove `outline` without providing an alternative focus indicator
- All images need meaningful `alt` text; decorative images use `alt=""`
- Use semantic HTML (`<nav>`, `<main>`, `<section>`, `<article>`, `<button>`) throughout
- Minimum touch target: 44×44px on mobile

---

## 11. Forbidden Rules (Strict)

Do **NOT** violate these in any component or screen:

- ❌ Pure `#000000` or `#FFFFFF` — always use the warm palette equivalents
- ❌ Cool greys, blue-greys, or any tone without warm undertones
- ❌ Saturated or neon colors outside of semantic feedback states
- ❌ Heavy drop shadows — max `--shadow-md` on hover; `--shadow-lg` for modals only
- ❌ Gradients (except extremely subtle warm surface gradients, e.g. `#FAF8F5` → `#F2EDE8`)
- ❌ Glassmorphism, Neumorphism, Brutalism, or Claymorphism elements
- ❌ Sharp 0px corners — always use the defined radius scale
- ❌ Decorative icons, illustrations, or textures unless part of the component spec
- ❌ Anything that feels clinical, cold, or corporate
- ❌ Font weights above 700
- ❌ Centered body text in multi-line paragraphs

---

## 12. Quick Reference

| Token | Value | Use |
|---|---|---|
| `--color-bg` | `#FAF8F5` | Page background |
| `--color-bg-subtle` | `#F2EDE8` | Cards, inputs |
| `--color-accent` | `#C4714A` | CTAs, links, highlights |
| `--color-text-primary` | `#1C1917` | Headings, primary body |
| `--color-text-secondary` | `#44403C` | Supporting text |
| `--font-sans` | Styrene A | UI chrome — nav, buttons, labels |
| `--font-serif` | Tiempos Text | Body prose, reading content |
| `--font-display` | Tiempos Headline | h1–h3, hero headings |
| `--radius-md` | `6px` | Default corner radius |
| `--shadow-sm` | warm tint | Card resting state |
| `--shadow-focus` | accent tint | Focus ring |
| `--transition-base` | `150ms ease` | Hover states |

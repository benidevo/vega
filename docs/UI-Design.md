# Vega AI UI Design System

Quick reference for consistent UI across the application.

## Design Foundations

**Typography:**

- **Body**: DM Sans
- **Headings**: Space Grotesk
- **Scale**: text-xs to text-6xl

**Colors:**

- **Primary**: #0D9488 (teal)
- **Primary-Dark**: #0B7A70 (hover states)
- **Secondary**: #F59E0B (amber)
- **Background**: slate-900/800/700 (dark theme)

**Tech Stack:**

- Go templates + HTMX + Hyperscript + Tailwind CSS
- No JavaScript frameworks

## Key Components

### Buttons

Use `partials/button.html` component:

```html
{{template "partials/button.html" (dict "variant" "primary" "text" "Save")}}
{{template "partials/button.html" (dict "variant" "danger" "text" "Delete")}}
{{template "partials/button.html" (dict "text" "Cancel")}} <!-- secondary -->
```

### Form Fields

Use `partials/form-input.html` component:

```html
{{template "partials/form-input.html" (dict "type" "text" "label" "Name" "required" true)}}
{{template "partials/form-input.html" (dict "type" "textarea" "label" "Description")}}
{{template "partials/form-input.html" (dict "type" "select" "options" $options)}}
```

### Cards & Containers

```html
<!-- Standard card -->
<div class="bg-slate-800 rounded-lg p-4 sm:p-6 border border-slate-700">
  <!-- Content -->
</div>

<!-- Stats card -->
<div class="bg-slate-800 rounded-lg p-3 sm:p-4 lg:p-6 border border-slate-700">
  <div class="flex items-center justify-between">
    <div class="min-w-0 flex-1">
      <p class="text-gray-400 text-xs sm:text-sm mb-1 truncate">Label</p>
      <p class="text-lg sm:text-xl lg:text-2xl font-bold text-white">Value</p>
    </div>
    <svg class="h-5 w-5 text-primary opacity-50"><!-- Icon --></svg>
  </div>
</div>
```

### Status Badges

```html
<span class="px-2 py-1 bg-blue-500/20 text-blue-400 text-xs font-medium rounded-full">Interested</span>
<span class="px-2 py-1 bg-primary/20 text-primary text-xs font-medium rounded-full">Applied</span>
<span class="px-2 py-1 bg-green-500/20 text-green-400 text-xs font-medium rounded-full">Interviewing</span>
```

## Style Guidelines

**Spacing:** Use Tailwind's spacing scale consistently (p-3, p-4, p-6, etc.)

**Responsive:** Always include sm:, md:, lg: breakpoints for key elements

**Colors:**

- Interactive elements: `text-primary`, `hover:text-primary-dark`
- Secondary text: `text-gray-400`, `text-gray-300`
- Backgrounds: `bg-slate-800`, `bg-slate-700`

**Focus States:** All interactive elements have `focus:ring-2 focus:ring-primary`

**Transitions:** Use `transition-colors duration-150` for smooth interactions

## Navigation

**Desktop:** Fixed sidebar (`templates/partials/sidebar.html`)
**Mobile:** Hamburger menu that slides in from left

```html
<!-- Sidebar nav item -->
<a href="/path" class="{{if eq .activeNav "key"}}bg-slate-700 text-white{{else}}text-gray-300 hover:bg-slate-700{{end}} group flex items-center px-3 py-3 text-sm font-medium rounded-md">
  <svg class="h-5 w-5 mr-3 {{if eq .activeNav "key"}}text-primary{{else}}text-gray-400 group-hover:text-primary{{end}}">
    <!-- Icon -->
  </svg>
  Label
</a>
```

## HTMX Patterns

**Form submissions:** `hx-post="/endpoint" hx-target="#result"`
**Dynamic content:** `hx-get="/data" hx-trigger="change" hx-target="#container"`
**Loading states:** `hx-indicator="#spinner"` with `.htmx-indicator` class

## Template Structure

```
layouts/base.html          # Main layout
templates/[page]/           # Page-specific templates
partials/                   # Reusable components
  ├── button.html
  ├── form-input.html
  ├── sidebar.html
  └── stats-card.html
```

Use `{{template "partials/component.html" (dict "key" "value")}}` for reusable components.

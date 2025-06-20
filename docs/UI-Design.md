# Vega UI Design System

This document outlines the UI design system for Vega, providing guidelines for consistent styling and components across the application.

## 1. Design Foundations

### 1.1 Typography

* **Primary Font (Body)**: Inter - A clean, modern sans-serif with excellent readability
* **Heading Font**: Poppins - A geometric sans-serif with distinctive character
* **Font Weights**: 300/400/500/600/700 for varied emphasis and hierarchy
* **Scale**: Text-xs to text-5xl with responsive adjustments

### 1.2 Color Theme

* **Primary**: Teal (#0D9488)
* **Primary-Dark**: Darker Teal (#0B7A70) - Used for hover states
* **Secondary**: Amber (#F59E0B)
* **Additional Accent Colors**: Light Blue (#0EA5E9), Purple (#8B5CF6)
* **Background**: Slate gradient (slate-900 via slate-800 to slate-900)
* **Text**: White/Gray for readability

### 1.3 UI Layout

* Responsive design with mobile-first approach
* Glass-morphism effects (backdrop blur) for containers
* Clean, minimalist interface with subtle animations
* Multi-color particle backgrounds with interactive hover effects

### 1.4 UI Framework

* Go templates with template inheritance
* HTMX for interactive UI elements and form submissions
* Tailwind CSS for responsive styling
* Particles.js for background animations
* Minimal JavaScript approach

## 2. UI Components

### 2.1 Container Components

```html
<!-- Main container with glow effect -->
<div class="w-full max-w-5xl z-10 relative">
  <!-- Glow effect -->
  <div class="absolute -inset-1 rounded-xl bg-gradient-to-r from-primary via-secondary to-primary opacity-30 blur-xl"></div>

  <!-- Content container -->
  <div class="relative p-8 md:p-12 bg-slate-900 bg-opacity-70 backdrop-blur-xl rounded-xl shadow-2xl border border-white border-opacity-10">
    <!-- Content goes here -->
  </div>
</div>
```

### 2.2 Animated Logo

```html
<div class="animate-float">
  <div class="p-6 bg-slate-800 bg-opacity-30 rounded-full border border-primary border-opacity-30 shadow-lg shadow-primary/20">
    <svg class="h-20 w-20 text-primary" viewBox="0 0 24 24"><!-- SVG content --></svg>
  </div>
  <div class="absolute -inset-1 bg-primary opacity-20 blur-xl rounded-full animate-pulse-slow"></div>
</div>
```

### 2.3 Gradient Text Headings

```html
<h1 class="text-5xl font-bold text-center mb-4 bg-clip-text text-transparent bg-gradient-to-r from-primary to-secondary">
  Heading Text
</h1>
```

### 2.4 Feature Cards

```html
<div class="group p-6 bg-slate-800 bg-opacity-50 rounded-xl border border-slate-700 hover:border-primary transition-all duration-300 transform hover:-translate-y-1 hover:shadow-lg hover:shadow-primary/20">
  <div class="flex items-center justify-center mb-4">
    <div class="p-3 bg-slate-700 bg-opacity-50 rounded-lg group-hover:bg-primary group-hover:bg-opacity-20 transition-all duration-300">
      <svg class="h-10 w-10 text-secondary group-hover:text-primary transition-colors duration-300">
        <!-- SVG content -->
      </svg>
    </div>
  </div>
  <h3 class="text-lg font-semibold text-center text-white mb-3">Feature Title</h3>
  <p class="text-gray-300 text-center">Feature description text</p>
</div>
```

### 2.5 Buttons

```html
<!-- Primary Button (Large) -->
<button class="px-8 py-4 bg-primary hover:bg-primary-dark text-white font-semibold text-lg rounded-lg transition-colors duration-300">
  Button Text
</button>

<!-- Primary Button (Standard) -->
<button class="px-6 py-2.5 bg-primary hover:bg-primary-dark text-white font-semibold rounded-lg transition-colors duration-300">
  Button Text
</button>

<!-- Primary Button with Icon -->
<button class="px-6 py-2.5 bg-primary hover:bg-primary-dark text-white font-semibold rounded-lg transition-colors duration-300 flex items-center justify-center">
  <svg class="h-5 w-5 mr-2"><!-- SVG content --></svg>
  Button Text
</button>

<!-- Secondary Button -->
<button class="px-5 py-2.5 bg-slate-700 hover:bg-slate-600 text-white rounded-md text-sm transition-colors text-center">
  Cancel
</button>

<!-- Danger Button -->
<button class="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-md text-sm font-medium transition-colors">
  Delete
</button>

<!-- Loading Button with Spinner -->
<button class="px-6 py-2.5 bg-primary hover:bg-primary-dark text-white font-semibold rounded-lg transition-colors duration-300 flex items-center justify-center">
  <span class="htmx-indicator mr-2">
    <svg class="animate-spin h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
    </svg>
  </span>
  Loading...
</button>
```

### 2.6 Form Fields with Icons

```html
<div class="relative">
  <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
    <svg class="h-5 w-5 text-gray-400"><!-- SVG content --></svg>
  </div>
  <input type="text" class="w-full pl-10 pr-4 py-3 rounded-lg bg-slate-800 bg-opacity-50 border border-slate-700 text-white focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary transition-colors">
</div>
```

### 2.7 Custom Animations

```css
@keyframes float {
  0% { transform: translateY(0px); }
  50% { transform: translateY(-10px); }
  100% { transform: translateY(0px); }
}

.animate-float {
  animation: float 6s ease-in-out infinite;
}

@keyframes pulse-slow {
  0% { opacity: 0.8; }
  50% { opacity: 0.3; }
  100% { opacity: 0.8; }
}

.animate-pulse-slow {
  animation: pulse-slow 5s ease-in-out infinite;
}
```

## 3. UI Consistency Guidelines

To ensure visual consistency across the application, follow these guidelines for all new pages and components:

### 3.1 Button Styles

* Use solid background colors instead of gradients for buttons
* Primary action buttons should use the primary color (#0D9488) with hover state of primary-dark (#0B7A70)
* Secondary actions should use slate-700 with hover state of slate-600
* Danger actions should use red-600 with hover state of red-700
* Button sizes should be consistent across similar contexts (e.g., form submissions)
* Button corners should use rounded-lg (large radius) for primary actions and rounded-md (medium radius) for secondary actions
* Loading/spinner buttons should position the spinner icon to the left of the text with mr-2 spacing

### 3.2 Form Elements

* Input fields should use bg-slate-700 bg-opacity-70 for the background
* Border should be border-slate-600 with focus:ring-2 focus:ring-primary
* Text inputs should have consistent padding: px-3 py-2
* Labels should be positioned above inputs with text-sm font-medium text-gray-300 mb-1

### 3.3 Card & Container Components

* Use consistent container styles with glass-morphism effects as defined in the template
* Card elements should have hover effects that are subtle and consistent
* Maintain consistent spacing within and between elements

### 3.4 Color Usage

* Use the primary color for main interactive elements and key UI components
* Use the secondary color for accents and to create visual interest
* Use slate colors for backgrounds and non-interactive elements
* For any gradient effects, use the established from-primary via-secondary to-primary pattern

### 3.5 Responsive Design

* All new pages must be fully responsive with a mobile-first approach
* Use the established breakpoint patterns (sm, md, lg, xl) consistently
* Stack elements vertically on mobile and use grid/flex layouts on larger screens

## 4. Page Templates

### 4.1 Dashboard Layout

The dashboard layout uses a sidebar for navigation with a main content area.

### 4.2 Form Pages

Form pages should follow a consistent structure with clear section headings and grouped fields.

### 4.3 Detail Pages

Detail pages should present information in clearly defined sections with consistent spacing.

## 5. UI States

### 5.1 Loading States

* Use spinner indicators for buttons during loading
* Apply skeleton loaders for content areas during data fetching

### 5.2 Error States

* Display validation errors inline with form fields
* Use consistent error messages with appropriate icons
* Provide clear next steps for resolution

### 5.3 Empty States

* All empty states should have helpful illustrations or icons
* Include clear messages explaining why content is missing
* Offer actionable next steps when appropriate

## 6. Accessibility Considerations

* Maintain color contrast ratios that meet WCAG 2.1 AA standards
* Ensure all interactive elements have appropriate focus states
* Use semantic HTML elements appropriately
* Include aria attributes where needed for complex components

## 7. Implementation Notes

### 7.1 Tailwind Configuration

The Tailwind configuration extends the default theme with custom colors and fonts.

```javascript
tailwind.config = {
  theme: {
    fontFamily: {
      'sans': ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
      'heading': ['Poppins', 'ui-sans-serif', 'system-ui', 'sans-serif']
    },
    extend: {
      colors: {
        primary: '#0D9488',
        'primary-dark': '#0B7A70',
        secondary: '#F59E0B',
      }
    }
  }
}
```

### 7.2 HTMX Integration

HTMX is used for interactive elements with consistent patterns for indicators and swapping.

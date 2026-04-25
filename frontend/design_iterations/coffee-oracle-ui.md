# Coffee Oracle UI – Iteration 01

## Stage 1 · Research & Goals

- Align visual language with espresso-inspired palette introduced at scaffold stage
- Emphasize safe upload workflow (name + creativity slider + drag/drop area)
- Provide streaming placeholder state mirroring Go backend SSE behaviour
- Accessibility priorities: color contrast, keyboard access, descriptive labels

## Stage 2 · Layout System

- Two-column hero for large screens (Form left, Results right) collapsing to stacked layout on mobile
- Sticky guidance column with process steps + best practices
- Form components grouped inside translucent panel with focus outlines and semantic hints

## Stage 3 · Visual + Theme Decisions

- Use Tailwind tokens from `tailwind.config.ts` (coffee-night, crema, foam)
- Flowbite buttons for consistent micro-interactions + progress states
- Rounded glassmorphism containers (border-white/10, background blur) to mimic steamed glass

## Stage 4 · Implementation + Handoff Notes

- `UploadForm` component encapsulates inputs, validation copy, disabled states
- `ResultsPanel` component handles streaming placeholder + eventual SSE hook
- Breakpoints: `lg` for split layout, `sm` adjustments for stack
- Next iteration will wire actual streaming hook & error toasts (Subtask 07)

# TimelineViewer Agent Guidelines

## Build Commands
- **Frontend Dev**: `cd client && bun run dev`
- **Run Dev Build**: `./stop_dev.sh && ./start_dev.sh`
- **Build**: `./build.sh` (builds frontend and backend)
- **Type Check**: `cd client && bun run check`
- **Run Tests**: No test commands defined yet

## Code Style

### Coding workflow
- After each step that is performed make a git commit with the changed files.
- After each set of steps is completed if any Go code was modified then Run Dev Build.

### TypeScript/Svelte
- Use TypeScript with strict type checking
- Define interfaces in `src/lib/types.ts`
- Use Svelte component structure: `<script lang="ts">`, markup, `<style>`
- Prefer explicit typing over `any`
- Use ES modules (`import`/`export`)

### Go
- Follow standard Go formatting conventions
- Use error handling with proper returns
- Organize code with clear function responsibilities
- Use structs with JSON tags for data models

### Naming Conventions
- TypeScript: camelCase for variables/functions, PascalCase for components/interfaces
- Go: CamelCase for exported functions/variables, camelCase for private
- Files: PascalCase for Svelte components, lowercase for Go files

### Error Handling
- TypeScript: Use try/catch with specific error messages
- Go: Return errors with context using fmt.Errorf
- Log errors with appropriate detail

### Project Structure
- Frontend: Svelte components in `client/src/components/`
- Backend: Go server in `server/`
- Data storage in `server/data/` directory

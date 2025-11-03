# Frontend - Smart Company Discovery Assistant

Modern React + TypeScript frontend application with a beautiful UI for managing Q&A knowledge base and chatting with an AI assistant. Built with Vite, Tailwind CSS, and Redux Toolkit.

## 📋 Table of Contents

- [Overview](#overview)
- [Technology Stack](#technology-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Setup Instructions](#setup-instructions)
- [Environment Configuration](#environment-configuration)
- [Features](#features)
- [Component Architecture](#component-architecture)
- [State Management](#state-management)
- [API Integration](#api-integration)
- [Styling Approach](#styling-approach)
- [Development Workflows](#development-workflows)
- [Build and Deployment](#build-and-deployment)
- [Troubleshooting](#troubleshooting)

## Overview

The frontend provides an intuitive, modern web interface for the Smart Company Discovery Assistant. Users can manage the Q&A knowledge base and interact with an AI-powered chat assistant through a responsive, polished UI.

### Key Features

✅ **Q&A Management Interface** - Full CRUD operations with search and pagination  
✅ **AI Chat Interface** - Conversational AI with streaming responses  
✅ **Real-time Updates** - Optimistic updates and automatic cache invalidation  
✅ **Modern UI Components** - ShadCN UI with Radix primitives  
✅ **Type Safety** - Full TypeScript coverage  
✅ **Responsive Design** - Works on desktop, tablet, and mobile  
✅ **Toast Notifications** - User-friendly feedback for all actions  
✅ **Loading States** - Skeleton loaders and spinners

## Technology Stack

### Core Framework
- **React 19.1.1** - Latest React with concurrent features
- **TypeScript 5.9.3** - Type-safe JavaScript
- **Vite 7.1.7** - Lightning-fast build tool and dev server

### UI & Styling
- **Tailwind CSS 4.1.16** - Utility-first CSS framework
- **ShadCN UI** - High-quality, accessible components built on Radix UI
- **Radix UI** - Unstyled, accessible component primitives
- **Lucide React** - Beautiful icon library
- **Sonner** - Elegant toast notifications

### State Management
- **Redux Toolkit 2.9.2** - Modern Redux with less boilerplate
- **RTK Query** - Powerful data fetching and caching
- **React Redux 9.2.0** - Official React bindings

### Routing & Navigation
- **React Router 7.9.5** - Declarative routing for React

### Development Tools
- **ESLint 9.36.0** - Code linting
- **TypeScript ESLint** - TypeScript-specific linting rules
- **Vite Plugin React** - Fast refresh and JSX transform

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         UI Layer                            │
│  • Pages           • Components        • Layouts            │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────────┐
│                     Feature Modules                          │
│  ┌─────────────────┐              ┌─────────────────┐      │
│  │  Q&A Feature    │              │  Chat Feature   │      │
│  │  • QATable      │              │  • ChatMessage  │      │
│  │  • QAForm       │              │  • ChatInput    │      │
│  │  • QADialogs    │              │  • Streaming    │      │
│  └─────────────────┘              └─────────────────┘      │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────────┐
│                    State Management                          │
│  • Redux Store    • RTK Query APIs    • Cache Management    │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────────┐
│                      API Layer                               │
│  • qaApi (RTK Query)                                        │
│  • chatApi (RTK Query + Streaming)                          │
└───────────────────────────┬─────────────────────────────────┘
                            │
                  ┌─────────┴─────────┐
                  │                   │
          ┌───────▼────────┐  ┌──────▼──────────┐
          │  Backend API   │  │  Python Agent   │
          │  :8080         │  │  :8000          │
          └────────────────┘  └─────────────────┘
```

### Design Patterns

1. **Feature-Based Structure**: Code organized by feature, not technical role
2. **Atomic Components**: Reusable UI components with single responsibility
3. **Container/Presenter Pattern**: Smart components handle logic, presenters handle UI
4. **RTK Query Integration**: Automatic caching, invalidation, and refetching
5. **Type-First Development**: TypeScript interfaces for all data structures

## Project Structure

```
frontend/
├── src/
│   ├── app/
│   │   ├── store.ts              # Redux store configuration
│   │   └── hooks.ts              # Typed Redux hooks
│   │
│   ├── features/
│   │   ├── qa/                   # Q&A Management Feature
│   │   │   ├── api/
│   │   │   │   └── qaApi.ts      # RTK Query API slice
│   │   │   ├── components/
│   │   │   │   ├── QATable.tsx   # Q&A list table
│   │   │   │   ├── QAForm.tsx    # Create/edit form
│   │   │   │   ├── CreateQADialog.tsx
│   │   │   │   ├── EditQADialog.tsx
│   │   │   │   └── DeleteQADialog.tsx
│   │   │   ├── pages/
│   │   │   │   └── QAManagementPage.tsx
│   │   │   └── types.ts          # Feature-specific types
│   │   │
│   │   └── chat/                 # AI Chat Feature
│   │       ├── api/
│   │       │   └── chatApi.ts    # RTK Query + Streaming
│   │       ├── components/
│   │       │   ├── ChatMessage.tsx
│   │       │   ├── ChatInput.tsx
│   │       │   └── StreamingMessage.tsx
│   │       ├── pages/
│   │       │   └── ChatPage.tsx
│   │       └── types.ts
│   │
│   ├── components/               # Shared Components
│   │   ├── Layout.tsx            # App layout wrapper
│   │   ├── Navigation.tsx        # Top navigation bar
│   │   ├── ai/                   # AI-specific components
│   │   │   ├── message-content.tsx
│   │   │   └── tool.tsx
│   │   └── ui/                   # ShadCN UI Components
│   │       ├── button.tsx
│   │       ├── card.tsx
│   │       ├── dialog.tsx
│   │       ├── input.tsx
│   │       ├── table.tsx
│   │       └── ... (other UI primitives)
│   │
│   ├── types/                    # Global Types
│   │   ├── api.ts               # API request/response types
│   │   └── models.ts            # Domain models
│   │
│   ├── lib/
│   │   └── utils.ts             # Utility functions (cn, etc.)
│   │
│   ├── App.tsx                  # Root component
│   ├── App.css                  # App-level styles
│   ├── main.tsx                 # Entry point
│   └── index.css                # Global styles + Tailwind
│
├── public/
│   └── vite.svg                 # Public assets
│
├── index.html                   # HTML template
├── package.json                 # Dependencies
├── tsconfig.json                # TypeScript config
├── vite.config.ts              # Vite configuration
├── tailwind.config.js          # Tailwind configuration
├── postcss.config.js           # PostCSS configuration
└── eslint.config.js            # ESLint configuration
```

### Directory Explanations

- **app/**: Global app configuration (Redux store, hooks)
- **features/**: Feature modules (Q&A, Chat) with co-located components, API, and types
- **components/**: Shared components used across features
- **types/**: Global TypeScript type definitions
- **lib/**: Utility functions and helpers

## Setup Instructions

### Prerequisites

- **Node.js 18+** (LTS recommended)
- **npm** or **yarn** or **pnpm**
- **Backend running** at `http://localhost:8080`

### Installation

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

The app will be available at **http://localhost:5173**

### First Time Setup

1. **Ensure backend is running**:
   ```bash
   curl http://localhost:8080/health
   ```

2. **Start frontend**:
   ```bash
   npm run dev
   ```

3. **Open browser**:
   - Navigate to http://localhost:5173
   - You should see the Q&A Management page

### Development Mode

In development mode, Vite provides:
- ⚡ Lightning-fast hot module replacement (HMR)
- 🔥 Instant feedback on code changes
- 🔍 Detailed error messages
- 🌐 Automatic proxy to backend API

## Environment Configuration

### Environment Variables

Create a `.env` file in the `frontend/` directory:

```bash
# Backend API URL
VITE_API_BASE_URL=http://localhost:8080

# Python Agent URL (for chat)
VITE_AGENT_URL=http://localhost:8000
```

**Note**: Vite requires `VITE_` prefix for environment variables to be exposed to the client.

### Vite Proxy Configuration

The Vite dev server is configured to proxy API requests:

```typescript
// vite.config.ts
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
    },
  },
}
```

This means:
- Frontend can call `/api/qa-pairs` and it proxies to `http://localhost:8080/api/qa-pairs`
- Avoids CORS issues in development
- Simulates production environment

## Features

### Q&A Management

**Location**: `/` (Home page)

**Capabilities**:
- ✅ **View all Q&A pairs** in a responsive table
- ✅ **Search** with debounced full-text search
- ✅ **Pagination** with cursor-based navigation
- ✅ **Create** new Q&A pairs via dialog modal
- ✅ **Edit** existing Q&A pairs with pre-filled form
- ✅ **Delete** with confirmation dialog
- ✅ **Toast notifications** for success/error feedback
- ✅ **Loading states** with skeleton loaders
- ✅ **Error handling** with user-friendly messages

**Features in Detail**:

1. **Search**:
   - Real-time debounced search (300ms delay)
   - Searches both questions and answers
   - Clears pagination cursor on new search

2. **Pagination**:
   - Cursor-based (efficient for large datasets)
   - Next/Previous navigation
   - Shows current page status
   - Preserves search query across pages

3. **Create Q&A**:
   - Modal dialog with form
   - Question and answer fields (required)
   - Validation feedback
   - Success toast on creation
   - Automatic table refresh

4. **Edit Q&A**:
   - Pre-populated form with existing data
   - Same validation as create
   - Optimistic updates
   - Success toast on update

5. **Delete Q&A**:
   - Confirmation dialog
   - Prevents accidental deletion
   - Success toast on deletion
   - Automatic table refresh

### AI Chat Interface

**Location**: `/chat`

**Capabilities**:
- ✅ **Ask questions** in natural language
- ✅ **View AI responses** with streaming
- ✅ **Message history** displayed chronologically
- ✅ **Clear conversation** to start fresh
- ✅ **Tool call visualization** (see what tools AI uses)
- ✅ **Auto-scroll** to latest messages
- ✅ **Keyboard shortcuts** (Enter to send, Shift+Enter for newline)
- ✅ **Loading indicators** during AI processing
- ✅ **Error handling** with retry capability

**Chat Flow**:

1. User types question
2. Message sent to Python AI agent
3. Agent streams response in real-time
4. Messages saved to backend
5. UI updates with streaming content
6. Tool calls displayed (optional)

**Streaming Implementation**:
- Uses Server-Sent Events (SSE) pattern
- Newline-delimited JSON (NDJSON) format
- Real-time token-by-token display
- Handles disconnections gracefully

## Component Architecture

### Atomic Design Principles

1. **Atoms** (Basic UI elements):
   - `Button`, `Input`, `Label`, `Card`
   - From ShadCN UI library
   - Highly reusable, no business logic

2. **Molecules** (Composite components):
   - `QAForm`, `ChatInput`, `ChatMessage`
   - Combine atoms with minimal logic
   - Feature-specific

3. **Organisms** (Complex components):
   - `QATable`, `CreateQADialog`, `StreamingMessage`
   - Contain business logic
   - Feature-complete sections

4. **Pages** (Route-level components):
   - `QAManagementPage`, `ChatPage`
   - Compose organisms
   - Handle routing and layout

### Key Components

#### QATable (`features/qa/components/QATable.tsx`)

Displays Q&A pairs with actions:

```typescript
interface QATableProps {
  data: QAPair[]
  loading: boolean
  onEdit: (qa: QAPair) => void
  onDelete: (qa: QAPair) => void
}
```

**Features**:
- Responsive table layout
- Action buttons (edit, delete)
- Loading skeleton
- Empty state

#### QAForm (`features/qa/components/QAForm.tsx`)

Reusable form for create/edit:

```typescript
interface QAFormProps {
  initialData?: QAPair
  onSubmit: (data: QAFormData) => void
  loading: boolean
}
```

**Features**:
- Controlled inputs
- Validation
- Loading states
- Error display

#### ChatMessage (`features/chat/components/ChatMessage.tsx`)

Displays a single chat message:

```typescript
interface ChatMessageProps {
  message: ChatMessage
  streaming?: boolean
}
```

**Features**:
- Role-based styling (user vs assistant)
- Markdown rendering
- Tool call visualization
- Timestamp display

#### StreamingMessage (`features/chat/components/StreamingMessage.tsx`)

Handles real-time streaming:

```typescript
interface StreamingMessageProps {
  conversationId: string
  onComplete: () => void
}
```

**Features**:
- SSE connection management
- Token accumulation
- Tool call parsing
- Error handling

## State Management

### Redux Store Configuration

```typescript
// app/store.ts
export const store = configureStore({
  reducer: {
    [qaApi.reducerPath]: qaApi.reducer,
    [chatApi.reducerPath]: chatApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware()
      .concat(qaApi.middleware, chatApi.middleware),
})
```

**State Structure**:
- RTK Query manages all server state
- No manual Redux slices needed
- Automatic cache management
- Optimistic updates

### RTK Query APIs

#### QA API (`features/qa/api/qaApi.ts`)

```typescript
export const qaApi = createApi({
  reducerPath: 'qaApi',
  baseQuery: fetchBaseQuery({ baseUrl: `${API_BASE_URL}/api` }),
  tagTypes: ['QAPair'],
  endpoints: (builder) => ({
    listQAPairs: builder.query<ListQAResponse, CursorParams>(...),
    getQAPair: builder.query<QAPair, string>(...),
    createQAPair: builder.mutation<CreateQAResponse, CreateQARequest>(...),
    updateQAPair: builder.mutation<UpdateQAResponse, {...}>(...),
    deleteQAPair: builder.mutation<{ success: boolean }, string>(...),
  }),
})
```

**Auto-generated Hooks**:
- `useListQAPairsQuery` - List with caching
- `useGetQAPairQuery` - Single item
- `useCreateQAPairMutation` - Create with invalidation
- `useUpdateQAPairMutation` - Update with invalidation
- `useDeleteQAPairMutation` - Delete with invalidation

**Cache Tags**:
- `{ type: 'QAPair', id: 'LIST' }` - All Q&A lists
- `{ type: 'QAPair', id: <uuid> }` - Individual Q&A pair

**Invalidation Strategy**:
- Create → Invalidates `LIST`
- Update → Invalidates `LIST` + specific `id`
- Delete → Invalidates `LIST` + specific `id`

#### Chat API (`features/chat/api/chatApi.ts`)

```typescript
export const chatApi = createApi({
  reducerPath: 'chatApi',
  baseQuery: fetchBaseQuery({ baseUrl: `${AGENT_URL}/chat` }),
  tagTypes: ['Conversation', 'Message'],
  endpoints: (builder) => ({
    createConversation: builder.mutation(...),
    getMessages: builder.query(...),
  }),
})
```

**Special Handling**:
- Streaming not handled by RTK Query
- Uses custom fetch with SSE
- Messages cached after streaming completes

## API Integration

### Backend API (Go)

**Base URL**: `http://localhost:8080/api`

**Endpoints Used**:
- `GET /qa-pairs` - List Q&A (with search, pagination)
- `GET /qa-pairs/:id` - Get single Q&A
- `POST /qa-pairs` - Create Q&A
- `PUT /qa-pairs/:id` - Update Q&A
- `DELETE /qa-pairs/:id` - Delete Q&A

**Request Examples**:

```typescript
// List Q&A with search
const { data } = useListQAPairsQuery({ 
  search: 'docker', 
  limit: 20 
})

// Create Q&A
const [createQA] = useCreateQAPairMutation()
await createQA({ 
  question: '...', 
  answer: '...' 
})
```

### Python Agent API

**Base URL**: `http://localhost:8000/chat`

**Endpoints Used**:
- `POST /conversations` - Create conversation
- `POST /conversations/:id/messages` - Send message (streaming)
- `GET /conversations/:id/messages` - Get message history

**Streaming Implementation**:

```typescript
const response = await fetch(`${url}/messages`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ message: userMessage }),
})

const reader = response.body?.getReader()
const decoder = new TextDecoder()

while (true) {
  const { done, value } = await reader.read()
  if (done) break
  
  const text = decoder.decode(value)
  const lines = text.split('\n')
  
  for (const line of lines) {
    if (line) {
      const event = JSON.parse(line)
      // Handle event...
    }
  }
}
```

## Styling Approach

### Tailwind CSS

**Configuration**: `tailwind.config.js`

```javascript
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      // Custom theme extensions
    },
  },
}
```

**Usage**:
```tsx
<button className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
  Click me
</button>
```

### ShadCN UI Components

**Installation**: Components are copied into `src/components/ui/`

**Benefits**:
- Full control over component code
- Customizable styling
- Consistent design system
- Accessible by default (Radix UI)

**Usage**:
```tsx
import { Button } from '@/components/ui/button'

<Button variant="default" size="lg">
  Create Q&A
</Button>
```

### Custom Utilities

**lib/utils.ts**:
```typescript
import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

Combines `clsx` for conditional classes with `twMerge` for Tailwind class conflicts.

### Global Styles

**index.css**:
- Tailwind directives (`@tailwind base`, etc.)
- CSS custom properties for theming
- Global resets and base styles

## Development Workflows

### Adding a New Feature

1. **Create feature directory**:
   ```
   src/features/my-feature/
   ├── api/
   ├── components/
   ├── pages/
   └── types.ts
   ```

2. **Define types** (`types.ts`):
   ```typescript
   export interface MyData {
     id: string
     name: string
   }
   ```

3. **Create RTK Query API** (`api/myApi.ts`):
   ```typescript
   export const myApi = createApi({
     reducerPath: 'myApi',
     baseQuery: fetchBaseQuery({ baseUrl: API_BASE_URL }),
     endpoints: (builder) => ({
       // Define endpoints
     }),
   })
   ```

4. **Add to store** (`app/store.ts`):
   ```typescript
   reducer: {
     [myApi.reducerPath]: myApi.reducer,
   },
   middleware: (gDM) => gDM().concat(myApi.middleware),
   ```

5. **Create components** and **pages**

6. **Add route** (`App.tsx`)

### Adding a UI Component

Using ShadCN CLI:

```bash
npx shadcn-ui@latest add [component-name]
```

Example:
```bash
npx shadcn-ui@latest add alert
npx shadcn-ui@latest add badge
```

This copies the component to `src/components/ui/`

### Code Style

**TypeScript**:
- Use interfaces for object types
- Use type for unions/primitives
- Explicit return types on functions
- Prefer `const` over `let`

**React**:
- Functional components only
- Named exports for components
- Props interface above component
- Destructure props in parameters

**Imports**:
```typescript
// External libraries first
import { useState } from 'react'

// Internal imports
import { Button } from '@/components/ui/button'
import { useListQAPairsQuery } from '@/features/qa/api/qaApi'
import type { QAPair } from '@/types/api'
```

### Git Workflow

```bash
# Create feature branch
git checkout -b feature/my-feature

# Make changes
# ...

# Commit with descriptive message
git add .
git commit -m "feat: add my feature"

# Push and create PR
git push origin feature/my-feature
```

## Build and Deployment

### Build for Production

```bash
# Build
npm run build

# Output in dist/ directory
ls -la dist/
```

**Build output**:
- Optimized JavaScript bundles
- CSS extracted and minified
- Assets hashed for cache busting
- HTML with preloaded resources

### Preview Production Build

```bash
npm run preview
```

Serves the production build locally at http://localhost:4173

### Environment Variables for Production

Create `.env.production`:

```bash
VITE_API_BASE_URL=https://api.yourcompany.com
VITE_AGENT_URL=https://agent.yourcompany.com
```

### Deployment Options

#### Static Hosting (Recommended)

Since this is a SPA, deploy to:
- **Vercel**: `vercel deploy`
- **Netlify**: `netlify deploy`
- **AWS S3 + CloudFront**
- **GitHub Pages**

#### Docker

```dockerfile
FROM node:18-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

#### Nginx Configuration

```nginx
server {
  listen 80;
  root /usr/share/nginx/html;
  index index.html;

  location / {
    try_files $uri $uri/ /index.html;
  }

  location /api {
    proxy_pass http://backend:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
  }
}
```

## Troubleshooting

### Port 5173 Already in Use

```bash
# Find process
lsof -i :5173

# Kill process
kill -9 <PID>

# Or use different port
npm run dev -- --port 3000
```

### Cannot Connect to Backend

1. **Check backend is running**:
   ```bash
   curl http://localhost:8080/health
   ```

2. **Check CORS**: Backend must allow `http://localhost:5173`

3. **Check proxy** in `vite.config.ts`

4. **Check environment variable**:
   ```bash
   echo $VITE_API_BASE_URL
   ```

### Module Resolution Errors

```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install

# Clear Vite cache
rm -rf node_modules/.vite
```

### TypeScript Errors

```bash
# Check types
npm run type-check

# Restart TypeScript server in VSCode
Cmd/Ctrl + Shift + P → "TypeScript: Restart TS Server"
```

### Build Errors

```bash
# Clear dist folder
rm -rf dist

# Rebuild
npm run build
```

### Hot Reload Not Working

1. Save files to trigger HMR
2. Check console for errors
3. Restart dev server
4. Check `.gitignore` doesn't ignore source files

## Performance Optimization

### Current Optimizations

1. **Code Splitting**: Automatic route-based splitting
2. **Tree Shaking**: Vite removes unused code
3. **Asset Optimization**: Images and CSS optimized
4. **Lazy Loading**: Components loaded on demand
5. **Cache Headers**: Immutable assets with hashing

### Future Enhancements

1. **React.lazy**: Lazy load heavy components
2. **Virtual Scrolling**: For large Q&A lists
3. **Service Worker**: Offline support
4. **Image Optimization**: WebP format, lazy loading
5. **Bundle Analysis**: Identify large dependencies

## Testing (Future)

### Recommended Testing Stack

```bash
npm install -D vitest @testing-library/react @testing-library/user-event jsdom
```

### Test Structure

```
features/
├── qa/
│   ├── __tests__/
│   │   ├── QATable.test.tsx
│   │   ├── QAForm.test.tsx
│   │   └── qaApi.test.ts
```

### Example Test

```typescript
import { render, screen } from '@testing-library/react'
import { QATable } from './QATable'

test('renders Q&A table', () => {
  render(<QATable data={[]} loading={false} />)
  expect(screen.getByText('No Q&A pairs')).toBeInTheDocument()
})
```

## Related Documentation

- [Main README](../README.md) - Full application setup
- [Backend README](../backend/README.md) - Go backend API
- [Python Agent README](../python-agent/README.md) - AI agent details
- [ShadCN UI Docs](https://ui.shadcn.com/) - Component library
- [Tailwind CSS Docs](https://tailwindcss.com/) - Styling framework
- [Redux Toolkit Docs](https://redux-toolkit.js.org/) - State management
- [Vite Docs](https://vitejs.dev/) - Build tool

---

**Built with modern React best practices** ⚡

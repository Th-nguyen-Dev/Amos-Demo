# Smart Company Discovery - Frontend

A modern React + TypeScript frontend application for the Smart Company Discovery Assistant.

## 🚀 Tech Stack

- **Framework**: React 18 + Vite
- **Language**: TypeScript
- **UI Library**: ShadCN UI (Radix UI primitives + Tailwind CSS)
- **State Management**: Redux Toolkit + RTK Query
- **Routing**: React Router v6
- **Styling**: Tailwind CSS
- **Notifications**: Sonner (toast notifications)
- **Icons**: Lucide React

## 📋 Prerequisites

- Node.js 18+ 
- npm or yarn
- Backend Go API running on `http://localhost:8080`

## 🛠️ Installation

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

The app will be available at **http://localhost:5173**

## 🌐 Environment Configuration

The app uses environment variables for configuration. The default values work with the local backend:

```env
VITE_API_BASE_URL=http://localhost:8080
```

The Vite dev server includes a proxy configuration that forwards `/api/*` requests to the backend.

## ✨ Features

### Q&A Knowledge Base Management
- ✅ **List Q&A pairs** with cursor-based pagination
- ✅ **Search Q&A pairs** with debounced full-text search
- ✅ **Create new Q&A pairs** via modal dialog
- ✅ **Edit existing Q&A pairs** with pre-populated forms
- ✅ **Delete Q&A pairs** with confirmation dialog
- ✅ **Toast notifications** for all CRUD operations
- ✅ **Loading states** during API operations
- ✅ **Error handling** with user-friendly messages

### AI Chat Interface
- ✅ **Ask questions** to the AI assistant
- ✅ **View AI responses** in a chat-like interface
- ✅ **Message history** displayed chronologically
- ✅ **Clear conversation** functionality
- ✅ **Auto-scroll** to latest messages
- ✅ **Keyboard shortcuts** (Enter to send, Shift+Enter for new line)
- ✅ **Loading indicators** during AI processing
- ✅ **Error handling** with retry capability

## 📁 Project Structure

```
src/
├── app/                        # Redux store configuration
│   ├── store.ts               # Redux store setup
│   └── hooks.ts               # Typed Redux hooks
├── features/                   # Feature-based modules
│   ├── qa/                    # Q&A Management feature
│   │   ├── api/               # RTK Query API slice
│   │   ├── components/        # Feature components
│   │   ├── pages/             # Feature pages
│   │   └── types.ts           # Feature types
│   └── chat/                  # Chat feature
│       ├── api/               # RTK Query API slice
│       ├── components/        # Feature components
│       ├── pages/             # Feature pages
│       └── types.ts           # Feature types
├── components/                 # Shared components
│   ├── ui/                    # ShadCN UI components
│   ├── Layout.tsx             # Main layout wrapper
│   └── Navigation.tsx         # Navigation bar
├── lib/                       # Utility functions
│   └── utils.ts              # cn() for className merging
├── types/                     # Shared TypeScript types
│   ├── models.ts             # Domain models
│   └── api.ts                # API request/response types
├── App.tsx                    # Root component with routing
├── main.tsx                   # Application entry point
└── index.css                  # Global styles + Tailwind
```

## 🎨 Component Architecture

### Q&A Management Components
- **QAManagementPage**: Main page with CRUD operations
- **QATable**: Displays Q&A pairs in a table
- **QAForm**: Reusable form for create/edit
- **CreateQADialog**: Modal for creating new Q&A
- **EditQADialog**: Modal for editing existing Q&A
- **DeleteQADialog**: Confirmation dialog for deletion

### Chat Components
- **ChatPage**: Main chat interface
- **ChatMessage**: Individual message display
- **ChatInput**: Input field with submit functionality

### Shared UI Components (ShadCN)
- Button, Input, Textarea, Label
- Card, Table, Dialog, Alert
- Toast notifications (Sonner)

## 🔄 State Management

### RTK Query API Slices

**QA API** (`features/qa/api/qaApi.ts`):
- `listQAPairs` - GET with pagination and search
- `getQAPair` - GET single Q&A by ID
- `createQAPair` - POST new Q&A
- `updateQAPair` - PUT update Q&A
- `deleteQAPair` - DELETE Q&A

**Chat API** (`features/chat/api/chatApi.ts`):
- `askQuestion` - POST question to AI assistant

### Automatic Cache Management
RTK Query handles:
- Automatic caching and invalidation
- Optimistic updates
- Loading and error states
- Request deduplication

## 🚦 Development Commands

```bash
# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Type checking
npx tsc --noEmit

# Lint (if configured)
npm run lint
```

## 🌐 API Integration

### Backend Endpoints

The frontend integrates with these Go backend endpoints:

**Q&A Management**:
- `GET /api/qa-pairs` - List Q&A pairs
- `GET /api/qa-pairs/:id` - Get single Q&A
- `POST /api/qa-pairs` - Create Q&A
- `PUT /api/qa-pairs/:id` - Update Q&A
- `DELETE /api/qa-pairs/:id` - Delete Q&A

**Chat**:
- `POST /api/ask` - Ask question (proxies to Python agent)

### Query Parameters

**Pagination**:
- `limit` (default: 10) - Number of items per page
- `cursor` - Cursor for pagination
- `direction` - `next` or `prev`
- `search` - Search query string

## 🎯 TypeScript Types

All types match the backend Go models:

```typescript
interface QAPair {
  id: string              // UUID
  question: string
  answer: string
  created_at: string
  updated_at: string
}

interface CursorPagination {
  next_cursor?: string
  prev_cursor?: string
  has_next: boolean
  has_prev: boolean
}
```

## 🎨 Styling

### Tailwind CSS
- Custom color scheme matching ShadCN defaults
- Responsive design utilities
- CSS variables for theming
- Dark mode support (configured but not activated)

### Customization
Edit `tailwind.config.js` and `src/index.css` to customize:
- Colors and themes
- Border radius
- Spacing
- Typography

## 🔧 Configuration Files

- `vite.config.ts` - Vite configuration with path aliases
- `tsconfig.json` - TypeScript configuration
- `tailwind.config.js` - Tailwind CSS configuration
- `postcss.config.js` - PostCSS configuration

## 🐛 Troubleshooting

### Build Errors
```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

### Type Errors
```bash
# Run type checking
npx tsc --noEmit
```

### API Connection Issues
- Ensure backend is running on `http://localhost:8080`
- Check browser console for CORS errors
- Verify Vite proxy configuration in `vite.config.ts`

### Port Already in Use
```bash
# Vite will automatically try the next available port
# Or specify a different port:
npm run dev -- --port 5174
```

## 📝 Code Style

- **Component naming**: PascalCase for components
- **File naming**: PascalCase for component files, camelCase for utilities
- **Imports**: Use absolute imports with `@/` alias
- **State management**: RTK Query for server state, useState for local UI state
- **Error handling**: Try-catch with toast notifications
- **TypeScript**: Strict mode enabled, explicit types preferred

## 🚀 Deployment

### Production Build

```bash
# Build for production
npm run build

# Output will be in dist/ directory
```

### Environment Variables for Production

Set these in your hosting platform:

```env
VITE_API_BASE_URL=https://api.yourdomain.com
```

### Hosting Options

- **Vercel** - Automatic deployment from Git
- **Netlify** - Static site hosting
- **AWS S3 + CloudFront** - Scalable hosting
- **Docker** - Container deployment (can be added)

## 📄 License

MIT

## 🤝 Contributing

This is part of the Smart Company Discovery Assistant project. See the main README for overall project information.

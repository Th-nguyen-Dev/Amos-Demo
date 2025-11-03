# Smart Company Discovery - Frontend

A React + TypeScript frontend application for the Smart Company Discovery Assistant.

## Tech Stack

- **Framework**: React 18 + Vite
- **Language**: TypeScript
- **UI Library**: ShadCN UI (Radix UI + Tailwind CSS)
- **State Management**: Redux Toolkit + RTK Query
- **Routing**: React Router v6
- **Styling**: Tailwind CSS

## Getting Started

### Prerequisites

- Node.js 18+ 
- npm or yarn
- Backend API running on http://localhost:8080

### Installation

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

The app will be available at http://localhost:5173

### Environment Variables

Create a `.env` file based on `.env.example`:

```
VITE_API_BASE_URL=http://localhost:8080
```

## Features

### Q&A Management
- List Q&A pairs with pagination
- Search Q&A pairs
- Create new Q&A pairs
- Edit existing Q&A pairs
- Delete Q&A pairs

### Chat Interface
- Ask questions to the AI assistant
- View AI-generated responses
- Clear conversation history

## Project Structure

```
src/
├── app/                    # Redux store setup
├── features/               # Feature modules
│   ├── qa/                # Q&A Management
│   └── chat/              # Chat Interface
├── components/            # Shared components
│   └── ui/               # ShadCN components
├── lib/                   # Utilities
└── types/                 # TypeScript types
```

## Development

```bash
# Run dev server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Type check
npx tsc --noEmit
```

## API Integration

The frontend connects to the Go backend API:

- **Base URL**: `http://localhost:8080`
- **Q&A Endpoints**: `/api/qa-pairs/*`
- **Chat Endpoint**: `/api/ask`

API requests are proxied through Vite dev server.

## License

MIT

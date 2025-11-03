# Frontend Setup Guide

This guide will help you get the React frontend up and running.

## Quick Start

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

The frontend will be available at **http://localhost:5173**

## Prerequisites

1. **Node.js 18+** installed
2. **Backend Go API** running on `http://localhost:8080`
3. **Python AI Agent** running (backend will proxy requests)

## Detailed Setup

### 1. Install Dependencies

```bash
cd frontend
npm install
```

This installs all required packages:
- React 18 + TypeScript
- Redux Toolkit + RTK Query
- React Router v6
- ShadCN UI components
- Tailwind CSS
- Sonner (toast notifications)

### 2. Environment Configuration

The app uses Vite environment variables. Default configuration works for local development:

```env
VITE_API_BASE_URL=http://localhost:8080
```

No `.env` file is needed for local development with default settings.

### 3. Start Development Server

```bash
npm run dev
```

Output:
```
  VITE v5.x.x  ready in 500 ms

  ➜  Local:   http://localhost:5173/
  ➜  Network: use --host to expose
  ➜  press h + enter to show help
```

### 4. Access the Application

Open your browser and navigate to:
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080 (should already be running)

## Features Checklist

After starting the app, you should be able to:

### Q&A Management Page (`/qa`)
- [ ] View list of Q&A pairs
- [ ] Search Q&A pairs (type in search box)
- [ ] Create new Q&A pair (click "Create New" button)
- [ ] Edit existing Q&A pair (click "Edit" button)
- [ ] Delete Q&A pair (click "Delete" button with confirmation)
- [ ] Navigate pages (click "Previous" / "Next" buttons)
- [ ] See toast notifications for all actions

### Chat Page (`/chat`)
- [ ] Type a question in the input box
- [ ] Submit question (click "Send" or press Enter)
- [ ] See your question appear in chat
- [ ] See AI assistant's response
- [ ] Clear chat history (click "Clear Chat" button)
- [ ] See loading indicator while waiting for response

## Common Issues

### Port Already in Use

If port 5173 is occupied, Vite will automatically use the next available port (5174, 5175, etc.)

### Cannot Connect to Backend

**Symptom**: API requests fail with network errors

**Solutions**:
1. Ensure backend is running: `cd backend && go run cmd/server/main.go`
2. Check backend is on port 8080: `http://localhost:8080/health`
3. Check browser console for CORS errors
4. Verify proxy config in `vite.config.ts`

### TypeScript Errors

**Symptom**: Red squiggly lines in VS Code or build errors

**Solutions**:
```bash
# Type check without building
npx tsc --noEmit

# If errors persist, try reinstalling
rm -rf node_modules
npm install
```

### Styling Not Working

**Symptom**: Components appear unstyled

**Solutions**:
1. Check `index.css` is imported in `main.tsx`
2. Verify Tailwind CSS is configured:
   ```bash
   # Should exist
   ls tailwind.config.js postcss.config.js
   ```
3. Restart dev server

### 404 on Page Refresh

**Symptom**: Refreshing a page like `/chat` gives 404

**Solution**: This is expected in development. Vite dev server handles client-side routing automatically, but on production you need proper server configuration.

## Development Workflow

### Making Changes

1. **Components**: Edit files in `src/components/` or `src/features/*/components/`
2. **Styles**: Use Tailwind classes or edit `src/index.css`
3. **API**: Modify API slices in `src/features/*/api/`
4. **Types**: Update types in `src/types/` to match backend

### Hot Module Replacement (HMR)

Vite provides instant updates when you save files:
- Component changes → instant update
- CSS changes → instant update
- Type changes → requires manual refresh

### Debugging

**Browser DevTools**:
1. Open DevTools (F12)
2. Check Console for errors
3. Check Network tab for API calls
4. Use Redux DevTools extension for state inspection

**VS Code**:
1. Install recommended extensions
2. Use breakpoints in browser debugger
3. Hover over variables to see types

## Building for Production

```bash
# Create optimized build
npm run build

# Preview production build locally
npm run preview
```

The build output will be in `dist/` directory.

## Project Structure Quick Reference

```
src/
├── app/                    # Redux store
├── features/
│   ├── qa/                # Q&A CRUD feature
│   └── chat/              # Chat feature
├── components/
│   ├── ui/                # ShadCN components
│   ├── Layout.tsx         # Main layout
│   └── Navigation.tsx     # Nav bar
├── types/                 # TypeScript types
├── lib/                   # Utilities
├── App.tsx                # Root with routing
└── main.tsx               # Entry point
```

## Next Steps

1. ✅ Start the frontend
2. ✅ Test Q&A CRUD operations
3. ✅ Test chat functionality
4. 📖 Read `README.md` for detailed documentation
5. 🔧 Customize styling in `tailwind.config.js`
6. 🚀 Deploy to production

## Getting Help

- Check browser console for errors
- Check terminal for Vite errors
- Review backend logs for API issues
- See main project README for overall architecture

## Summary

The frontend is a modern React application that:
- Uses TypeScript for type safety
- Uses RTK Query for automatic API state management
- Uses ShadCN UI for beautiful, accessible components
- Follows best practices for React development
- Provides a great developer experience with Vite

Enjoy building! 🚀



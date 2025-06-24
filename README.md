# Timeline Media Viewer

A timeline-based media viewer application for visualizing and organizing media files (images, audio, video) on a timeline.

## Project Overview

This application consists of:

- **Backend**: Go server that handles file uploads, serves media files, and manages metadata
- **Frontend**: Svelte + TypeScript application with vis-timeline for timeline visualization

## Features

- Upload media files (images, audio, video)
- Visualize media on an interactive timeline
- View media files with basic playback controls
- Store media files and metadata locally

## Project Structure

```
/timelineviewer
├── /client               # Svelte frontend app (with bun)
│   ├── /src              # Frontend source code
│   ├── bun.lockb         # Bun lock file
│   ├── bunfig.toml       # Bun configuration
│   └── tsconfig.json     # TypeScript configuration
├── /server               # Go backend
│   └── main.go           # Main Go server code
├── /data                 # Local media + metadata store
│   ├── /media            # Uploaded media files
│   ├── /metadata         # JSON metadata for media files
│   └── timeline.json     # Timeline data
├── dev.sh                # Script: starts bun + Go server in dev mode
├── build.sh              # Script: builds Svelte, then runs Go server
└── README.md             # This file
```

## Prerequisites

- [Go](https://golang.org/) (1.16 or later)
- [Bun](https://bun.sh/) (latest version)

## Getting Started

### Development Mode

To run the application in development mode:

```bash
./dev.sh
```

This will:
1. Start the Svelte development server on port 5173
2. Start the Go backend server on port 8080
3. Set up proxy for API requests from frontend to backend

### Production Mode

To build and run the application in production mode:

```bash
./build.sh
```

This will:
1. Build the Svelte frontend to static files
2. Build the Go backend
3. Start the Go server which serves both the API and the static frontend files

## API Endpoints

- `GET /api/timeline` - Get timeline data
- `POST /api/upload` - Upload a media file
- `GET /api/metadata/:filename` - Get metadata for a specific file
- `GET /media/:filename` - Serve a media file

## Future Enhancements

- WhisperX integration for audio/video transcription
- ffmpeg integration for media processing
- Advanced timeline filtering and search
- User authentication and multi-user support
- Cloud storage options

## License

MIT
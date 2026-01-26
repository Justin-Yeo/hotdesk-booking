# Hotdesk Booking System

A modern hot-desking management platform for flexible office spaces, enabling employees to reserve desks, track availability, and manage workspace resources efficiently.

## Overview

This application provides a comprehensive solution for managing hot-desking in modern workplaces. Built with a focus on user experience and real-time updates, it helps organizations optimize their workspace utilization while providing employees with an intuitive booking experience.

## Tech Stack

### Frontend
- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS + shadcn/ui
- **State Management**: Zustand
- **API Communication**: React Query

### Backend
- **Language**: Go
- **Framework**: Fiber (Express-like framework)
- **Database**: PostgreSQL
- **ORM**: GORM
- **Authentication**: JWT

## Project Structure

```
hotdesk-booking/
├── frontend/          # Next.js frontend application
├── backend/           # Go backend API
├── docs/              # Documentation (coming soon)
└── README.md
```

## Getting Started

### Prerequisites
- Node.js 18+ (for frontend)
- Go 1.21+ (for backend)
- PostgreSQL 15+

### Frontend Setup
```bash
cd frontend
npm install
npm run dev
```

### Backend Setup
```bash
cd backend
go mod download
go run cmd/api/main.go
```

## Features (Planned)

- 🪑 Real-time desk availability tracking
- 📅 Advanced booking system with recurring reservations
- 👥 User authentication and role-based access control
- 📊 Analytics dashboard for workspace utilization
- 🔔 Notifications and reminders
- 📱 Responsive design for mobile and desktop

## Development Status

🚧 This project is currently in the initial setup phase. See [TASK_BREAKDOWN.md](TASK_BREAKDOWN.md) for detailed development progress.

## Documentation

- [Implementation Plan](IMPLEMENTATION_PLAN.md) - Detailed technical architecture and design decisions
- [Task Breakdown](TASK_BREAKDOWN.md) - Phase-by-phase development tasks
- [Project Idea](idea.md) - Original concept and requirements
- [Tech Stack Details](tech-stack.md) - Technology selection rationale

## License

This project is currently private and not licensed for public use.

## Authors

Built with ❤️ for modern workplaces

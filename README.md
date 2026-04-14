# StartUp

## Background
For a patient with Alzheimer's or similar type of dementia who feels unsettled by their surroundings, who wants to know where they are and not feel abandoned when they wake up, StartUp is a dementia assistance tool acting as a surrogate for a loved one, reducing the physiological anxiety of the person with dementia and improving the customer’s quality of life. Unlike face-to-face interactions in real-time, our project is more scalable and available, allowing customers to feel connected despite the caregiver's busy schedules.

StartUp enables caregivers to create, manage, and deliver personalized content for carepartners. The system supports multiple module types including media (photo, video, audio), prompts (messages), and widgets (weather, news, sports), all displayed through a simple and accessible interface.

---

## Key Features
- Upload and manage media (photos, videos, audio)
- Create and edit prompts (messages with optional images)
- Add dynamic widgets (weather, news, sports)
- Assign content to specific carepartners
- Interactive library for browsing and selecting content
- Agenda integration for structured content delivery
- Secure authentication using Google OAuth

---

## System Architecture

### Frontend
- React
- TypeScript
- Tailwind CSS

### Backend
- Go (Golang)
- RESTful API using Go standard `net/http`
- Handles:
  - Authentication (Google OAuth)
  - Module management
  - Media upload/storage metadata
  - Widget integrations
  - Configuration scheduling logic

### Storage Layer
- PostgreSQL: Stores all structured metadata (modules, users, configurations, prompts, widgets)
- MinIO: Stores raw uploaded files (media) before deployment migration to AWS S3

### External Integrations
- Google OAuth authentication
- Weather API
- News API
- Sports API

## Data Flow

1. Caregiver logs in and pairs with carepartner via Google OAuth
2. Caregiver uploads media → BLOB stored in MinIO
3. Metadata (URI, filename, type) created → stored in PostgreSQL
4. Caregiver creates widgets and/or prompts
5. Modules are created linking metadata to carepartners
6. Frontend fetches modules and resolves metadata via REST endpoints
7. Carepartner views scheduled content based on Configuration rules
8. External APIs enrich widget-based modules (weather/news/sports)

---

## Data Model

The system uses a modular architecture centered around the `Module` table.

### Module 
- id 
- module_type (photo, video, audio, weather, news, sports, messages) 
- caregiver_id 
- carepartner_id 
- metadata_id 

### Each module links to a metadata record: 

### Media (photo/video/audio) 
- id 
- uri 
- displayed_name 
- filename 
- extension 

### Prompts (locations, events)
- id 
- message 
- image_uri 

### Widgets (sports, news, weather)
- id
- zipcode

---

## Configuration System

The Configuration system controls **when and how modules are displayed** to carepartners.

### Purpose
- Defines ordering of modules in a daily schedule
- Supports timed or interval-based content delivery
- Enables flexible reordering without modifying modules themselves
- Supports exceptions (special cases like skipping or replacing content)

### How It Works

Each configuration row links:
- a carepartner
- a module
- a position in the schedule
- timing rules (optional)

### Key Features
- Ordered scheduling via `position`
- Time-based or interval-based display (`timed`, `time`, `duration`)
- Batch updates via PostgreSQL stored procedure
- Exception handling for overrides

### Stored Procedures

#### insert_configurations
- Inserts multiple configuration rows atomically
- Maintains ordering and shifts existing positions
- Ensures no gaps in schedule ordering

#### delete_configurations
- Deletes configurations by ID list
- Re-indexes remaining positions to maintain order integrity

---

## Setup Instructions

### Prerequisites
- npm
- Node.js
- Go 1.20+
- Docker
- PostgreSQL
- MinIO

---

### Run locally
#### For more options, run *npm run*

#### Backend
```bash
cd server
go run cmd/main.go
```

#### Frontend Caregiver
```bash
cd caregiver
npm run dev
```

#### Frontend Carepartner
```bash
cd carepartner
npm start
```

# Kiosk Server

Kiosk Server is a Go-based backend for a campus print kiosk and notes distribution platform. It powers the full workflow from user authentication and file upload to print pricing, payment, kiosk job retrieval, and notes catalog management for department admins.

## What this project does

This server supports four main user journeys:

- Student / user flow
  - Sign in with Google
  - Upload documents for printing
  - Choose print settings such as page ranges, copies, duplex, and color mode
  - Pay securely through Razorpay
  - Retrieve the print job later using a 6-digit token from the kiosk

- Kiosk flow
  - Look up a job by token
  - Mark the session as completed after printing

- Department admin flow
  - Log in as a department admin or super admin
  - Manage subjects and notes for a specific branch
  - Upload notes and make them available to students

- Admin reporting flow
  - Review print history and revenue statistics
  - Track totals for sheets printed, color vs black-and-white output, and payment activity

---

## Core architecture

The project follows a clean layered architecture:

- Entry point: main.go
  - Starts the HTTP server and wires the application
- App bootstrap: internal/app/app.go
  - Initializes the database, S3 client, mail client, repositories, services, and handlers
- Routing: internal/routes/router.go
  - Defines API routes and middleware
- Handlers: internal/handler
  - Accepts HTTP requests and validates input
- Services: internal/service
  - Contains the business logic for authentication, uploads, payments, notes, and admin reporting
- Repositories: internal/repository
  - Handles persistence with PostgreSQL
- Shared packages: pkg/
  - JWT, S3, mail, password hashing, utilities, and pricing helpers

---

## End-to-end workflow

### 1. Authentication
A user logs in through Google OAuth at /auth/google. The server verifies the Google ID token, creates or updates the user record, and returns a signed JWT for subsequent requests.

### 2. File upload session
When a user starts uploading files:
1. The server creates an upload session and generates a 6-digit token.
2. It returns presigned S3 upload URLs for each file.
3. Files are initially stored in a staging area in S3.

### 3. Confirm upload and pricing
After the files are uploaded:
1. The server verifies the files exist in storage.
2. It calculates the print price based on selected page ranges, copies, layout, color mode, and duplex settings.
3. The session is marked as priced and becomes ready for payment.

### 4. Payment
The frontend calls the payment endpoint to create a Razorpay order. Once payment succeeds, the server receives a webhook and promotes the staged files to their final storage location.

### 5. Kiosk retrieval
The kiosk uses the token provided to the user to fetch the job details. After printing is complete, the session is marked as completed.

### 6. Notes catalog
Dept admins manage academic notes by branch, semester, subject, and module. Notes are uploaded to S3 and served via dedicated notes endpoints.

---

## Tech stack

- Go 1.25
- Chi router
- PostgreSQL + pgx
- AWS S3-compatible object storage
- Razorpay for payments
- Resend for email delivery
- JWT-based authentication
- Swagger/OpenAPI documentation

---

## Project structure

```text
.
├── main.go
├── Makefile
├── internal/
│   ├── app/
│   ├── env/
│   ├── handler/
│   ├── middlewares/
│   ├── models/
│   ├── repository/
│   ├── routes/
│   ├── service/
│   └── validator/
├── pkg/
│   ├── apperror/
│   ├── db/
│   ├── jwt/
│   ├── mail/
│   ├── password/
│   ├── s3/
│   └── utils/
├── migrations/
└── docs/
```

---

## Prerequisites

Before running the server locally, make sure you have:

- Go 1.25 or newer
- PostgreSQL running and accessible
- An S3-compatible bucket and credentials
- Razorpay API credentials
- Resend API credentials for mail sending

---

## Environment variables

Create a .env file in the project root with the following variables:

```env
PORT=17069

DATABASE_URL=postgres://user:password@host:5432/database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_NAME=kiosk_db
DATABASE_PASSWORD=your_password

REGION=ap-south-1
ACCESS_KEY=your_access_key
SECRETE_KEY=your_secret_key
BUCKET=your_bucket_name

JWT_SECRET=your_jwt_secret
GOOGLE_CLIENT_ID=your_google_client_id

RZP_KEY=your_razorpay_key
RZP_SECRET=your_razorpay_secret
RZP_WEBHOOK_SECRET=your_webhook_secret

SUPER_ADMIN_EMAIL=super@example.com
SUPER_ADMIN_PASSWORD=super_password
```

> The application uses these values during bootstrap and for service integrations such as S3, Razorpay, JWT, and mail.

---

## Running the server locally

### 1. Install dependencies

```bash
go mod download
```

### 2. Run database migrations

```bash
make up
```

### 3. Start the server

```bash
go run .
```

The server will run on:

```text
http://localhost:17069
```

### 4. Explore the API docs

Swagger UI is available at:

```text
http://localhost:17069/swagger/index.html
```

---

## Useful Make targets

```bash
make db-status
make up
make down
make reset
make create
```

These targets help manage PostgreSQL migrations using Goose.

---

## API highlights

### Authentication
- POST /auth/google

### File and print jobs
- POST /files/upload/init
- POST /files/upload/confirm
- GET /files/jobs/recent
- GET /files/jobs/active
- POST /print/jobs/token
- POST /print/jobs/error
- POST /print/jobs/expire
- GET /files/job/session/{session_id}

### Payments
- POST /payments/create
- GET /payments/status/{session_id}
- POST /webhooks/razorpay

### Notes
- GET /notes/branches
- GET /notes/branches/{branch_id}/semesters
- GET /notes/semesters/{semester_id}/subjects
- GET /notes/subjects/{subject_id}/modules
- GET /notes/modules/{module_id}/notes
- POST /notes/print/init

### Dept admin
- POST /deptadmin/auth/super/login
- POST /deptadmin/auth/login
- POST /deptadmin/dept-admins
- GET /deptadmin/dept-admins
- POST /deptadmin/notes/upload/init
- POST /deptadmin/notes/upload/confirm
- PUT /deptadmin/notes/{note_id}
- DELETE /deptadmin/notes/{note_id}
- POST /deptadmin/subjects

### Admin reports
- GET /admin/print/history
- GET /admin/print/revenue
- GET /admin/print/totalsheetsprinted
- GET /admin/print/colorsheets
- GET /admin/print/blackandwhite
- GET /admin/print/revenue/{date}
- GET /admin/print/sheets/{date}
- GET /admin/print/history/{date}

---

## Notes for contributors

- The server uses structured JSON logging for easier debugging.
- Errors are returned in a consistent envelope format.
- Most write operations are protected by authentication and role-based middleware.
- The project is designed to be extended with additional services or admin features without changing the core flow.

---

## License

This project is currently maintained as an internal service. Please confirm the license terms with the repository owner before redistribution or reuse.

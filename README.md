# Go Patient Portal API: A Role-Based Patient Management System

Welcome to the Go Patient Portal API! This project provides a robust backend for managing patient records, complete with secure **user authentication**, **role-based access control**, and efficient **CRUD operations**.

## Key Features & Enhancements

This API goes beyond basic patient management by incorporating:

*   **Role-Based Access Control:**
    
    *   **Receptionists** have permissions to register new users, **add new patients** (without diagnosis), **update patient details** (excluding diagnosis), **retrieve patient lists**, and **delete patients**.
        
    *   **Doctors** can **retrieve patient lists**, **update patient diagnosis only**, and do not have permissions to add or delete patients.
        
    *   The system strictly enforces these roles; for example, a receptionist cannot set a patient's diagnosis, and a doctor cannot update a patient's name or delete their record.
        
*   **Patient Search:** Easily find patients by their **name** using partial or incomplete queries. The search is case-insensitive, making it user-friendly.
    
*   **Pagination:** Efficiently retrieve large datasets of patients by specifying `page` and `limit` parameters in GET requests. This improves performance by reducing the data transferred and enhances the user experience for browsing patient records.
    
*   **CSV Export:** Download a single patient's complete details as a **CSV file**, providing a convenient and standard way to extract data for reporting or analysis.
    
*   **Secure Authentication:** Utilizes **JWT (JSON Web Tokens)** for stateless user authentication, ensuring secure communication between the client and API.
    
*   **Robust Error Handling:** The API includes comprehensive validation and explicit error handling for all endpoints, ensuring data integrity and providing clear feedback for invalid requests or internal issues. Edge cases are robustly handled by the API's logic.
    
*   **Database Persistence:** Leverages **PostgreSQL** as a reliable and scalable relational database for data storage.
    

## Getting Started

Follow these simple steps to get the API running on your local machine.

### Prerequisites

Before you begin, ensure you have the following installed on your system:

1.  **Git**: For cloning the repository.
    
2.  **Go (Golang)**: Version 1.22 or higher.
    
    *   [Download Go](https://golang.org/doc/install "null")
        
3.  **PostgreSQL**: A running PostgreSQL instance that your application can connect to.
    
    *   You can install it directly on your machine, or use a tool like Docker to run a temporary instance (e.g., `docker run --name some-postgres -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d postgres`).
        

### Project Setup

1.  **Clone the Repository:** Open your terminal or command prompt and clone the project:
    
        git clone https://github.com/Faizan2005/Golang_Coding_Assessment_Makerble.git # Replace with your actual repo name if different
        cd Golang_Coding_Assessment_Makerble
        
    
2.  **Create Your Environment File (`.env`):** This file stores your sensitive configuration (like database credentials and JWT secret) and is **not committed to Git** for security reasons.
    
    *   Create a new file named `.env` in the **root of your project directory**:
        
            touch .env
            
        
    *   Open `.env` in a text editor and paste the following content. **Crucially, replace placeholder values with your actual PostgreSQL connection details and a strong `JWT_SECRET`.**
        
            # .env - Environment variables for the application
            DB_USER=your_postgres_user      # e.g., postgres
            DB_PASSWORD=your_postgres_password # <--- CHANGE THIS!
            DB_NAME=hospital
            DB_HOST=localhost               # Or your DB host IP/hostname
            DB_PORT=5432
            JWT_SECRET=your_super_strong_random_jwt_secret_key_atleast_32_chars # <--- CHANGE THIS!
            LISTEN_ADDR=:3000               # Port your Go app will listen on
            
        
        *   **Important:** Ensure your `JWT_SECRET` is a long, random string for security.
            
3.  **Prepare Database Schema Initialization:** The application is designed to create its necessary database tables (`users` and `patients`) automatically on startup if they don't already exist.
    
    *   **ACTION REQUIRED IN YOUR GO CODE:** You need to implement a function in your Go application (e.g., in your `config` package or a new `migrations` package) that reads and executes the SQL content from `migrations/init.sql` against your connected PostgreSQL database. This function should be called from `main.go` right after the database connection is established.
        
    *   Create a `migrations` directory and `init.sql` file:
        
            mkdir -p migrations
            touch migrations/init.sql
            
        
    *   Open `migrations/init.sql` and paste the exact table schemas (as provided in the "Database Schema" section below):
        
            -- migrations/init.sql
            
            -- Create the users table if it doesn't exist
            CREATE TABLE IF NOT EXISTS users (
                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                name VARCHAR(255) NOT NULL,
                email VARCHAR(255) UNIQUE NOT NULL,
                password VARCHAR(255) NOT NULL,
                role VARCHAR(255) NOT NULL
            );
            
            -- Add the CHECK constraint for the role column if it doesn't exist
            DO $$ BEGIN
                IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_role_check') THEN
                    ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role::text = ANY (ARRAY['receptionist'::character varying, 'doctor'::character varying]::text[]));
                END IF;
            END $$;
            
            
            -- Create the patients table if it doesn't exist
            CREATE TABLE IF NOT EXISTS patients (
                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                name VARCHAR(255) NOT NULL,
                age INTEGER NOT NULL,
                gender VARCHAR(255) NOT NULL,
                diagnosis TEXT, -- This column is NULLABLE
                created_by UUID NOT NULL
            );
            
            -- Add the Foreign Key constraint if it doesn't exist
            DO $$ BEGIN
                IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'patients_created_by_fkey') THEN
                    ALTER TABLE patients ADD CONSTRAINT patients_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT;
                END IF;
            END $$;
            
            
            -- Optional: Create indexes for improved query performance
            CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
            CREATE INDEX IF NOT EXISTS idx_patients_name ON patients (name);
            
        
4.  **Install Go Dependencies:**
    
        go mod tidy
        
    
5.  **Run the Go Application:**
    
        go run main.go
        
    
    *   Your application will attempt to connect to the PostgreSQL database using the credentials from your `.env` file and then run the schema initialization.
        
    *   Look for the message: `Application starting on :3000` (or whatever `LISTEN_ADDR` you set).
        
    *   Once this message appears, your API is up and running on `http://localhost:3000`.
        

## How to Use the API

The API is accessible at `http://localhost:3000`. You'll use `curl` (command line) or tools like Postman/Insomnia to send requests.

### Authentication Flow

1.  **Register:** Create a new user account, specifying their **`role`** (either `receptionist` or `doctor`).
    
2.  **Login:** Authenticate with your email and password to receive a **JSON Web Token (JWT)**.
    
3.  **Use JWT:** For all authenticated endpoints, you **must** include the JWT in the `Authorization` header as a `Bearer` token (e.g., `Authorization: Bearer YOUR_AUTH_TOKEN`).
    

### Example API Requests (Test Steps)

These `curl` commands demonstrate the core functionality and role-based access for your API. Run them sequentially.

#### First, set up a variable for your base URL:

    BASE_URL="http://localhost:3000"
    
    # Placeholders for JWT Tokens (You'll update these after successful login)
    RECEPTIONIST_TOKEN=""
    DOCTOR_TOKEN=""
    
    # Placeholder for a Patient ID (You'll update this after adding a patient)
    PATIENT_ID=""
    

#### 1\. `POST /register` – Register New Users

Let's create accounts for a receptionist and a doctor.

**Register a Receptionist:**

    curl -X POST \
      "$BASE_URL/register" \
      -H 'Content-Type: application/json' \
      -d '{
        "name": "Alice Receptionist",
        "email": "alice.r@example.com",
        "password": "alicepass",
        "role": "receptionist"
      }'
    

**Register a Doctor:**

    curl -X POST \
      "$BASE_URL/register" \
      -H 'Content-Type: application/json' \
      -d '{
        "name": "Dr. Bob Physician",
        "email": "bob.d@example.com",
        "password": "bobpass",
        "role": "doctor"
      }'
    

#### 2\. `POST /login` – Authenticate and Get Tokens

Log in as each user to obtain their **JWT token**.

**Login as Receptionist (and get token):** **ACTION:** From the JSON response, copy the **entire `Bearer <TOKEN>` string** (including "Bearer ").

    curl -X POST \
      "$BASE_URL/login" \
      -H 'Content-Type: application/json' \
      -d '{
        "email": "alice.r@example.com",
        "password": "alicepass"
      }'
    

**Then, in your terminal, paste it like this:** `RECEPTIONIST_TOKEN="<PASTE_YOUR_RECEPTIONIST_BEARER_TOKEN_HERE>"`

**Login as Doctor (and get token):** **ACTION:** From the JSON response, copy the **entire `Bearer <TOKEN>` string**.

    curl -X POST \
      "$BASE_URL/login" \
      -H 'Content-Type: application/json' \
      -d '{
        "email": "bob.d@example.com",
        "password": "bobpass"
      }'
    

**Then, in your terminal, paste it like this:** `DOCTOR_TOKEN="<PASTE_YOUR_DOCTOR_BEARER_TOKEN_HERE>"`

#### Now, use `$RECEPTIONIST_TOKEN` or `$DOCTOR_TOKEN` for subsequent authenticated requests.

### Receptionist Portal Endpoints (`/api/v1/receptionist/patients`)

Receptionists manage core patient information.

#### 3\. `POST /api/v1/receptionist/patients` – Add a New Patient

**Receptionists add patients without specifying diagnosis.** The API explicitly sets diagnosis to `NULL`. **Attempts to include a `diagnosis` field in the request by a receptionist will be rejected.**

    curl -X POST \
      "$BASE_URL/api/v1/receptionist/patients" \
      -H 'Content-Type: application/json' \
      -H "Authorization: $RECEPTIONIST_TOKEN" \
      -d '{
        "name": "Patient X",
        "age": 25,
        "gender": "Female"
      }'
    

**ACTION:** From the successful JSON response, copy the **`id`** value (e.g., `"id": "some-uuid"`) and update the `PATIENT_ID` variable: `PATIENT_ID="<COPIED_PATIENT_UUID_HERE>"`

#### 4\. `GET /api/v1/receptionist/patients` – Get Patients (Search & Pagination)

Receptionists can retrieve patient lists with powerful filtering and pagination.

*   **Get all patients (default pagination: page 1, limit 10, no search query):**
    
        curl -X GET \
          "$BASE_URL/api/v1/receptionist/patients" \
          -H "Authorization: $RECEPTIONIST_TOKEN"
        
    
*   **Get patients filtered by name (partial & case-insensitive search):**
    
        curl -X GET \
          "$BASE_URL/api/v1/receptionist/patients?name=x" \
          -H "Authorization: $RECEPTIONIST_TOKEN"
        
    
*   **Get patients with specific pagination (e.g., page 1, limit 1 result):**
    
        curl -X GET \
          "$BASE_URL/api/v1/receptionist/patients?page=1&limit=1" \
          -H "Authorization: $RECEPTIONIST_TOKEN"
        
    

#### 5\. `PUT /api/v1/receptionist/patients/:id` – Update Patient Details

**Receptionists can only update `name`, `age`, and `gender`.** If a `diagnosis` field is included in the request body, the API will specifically reject the request with a `400 Bad Request` error.

    curl -X PUT \
      "$BASE_URL/api/v1/receptionist/patients/$PATIENT_ID" \
      -H 'Content-Type: application/json' \
      -H "Authorization: $RECEPTIONIST_TOKEN" \
      -d '{
        "name": "Patient X (Updated by Receptionist)",
        "age": 26,
        "gender": "Female"
      }'
    

### Doctor Portal Endpoints (`/api/v1/doctor/patients`)

Doctors focus primarily on patient diagnosis and viewing records.

#### 6\. `PUT /api/v1/doctor/patients/:id` – Update Patient Diagnosis

**Doctors can ONLY update the `diagnosis` field.** Any other fields (like `name`, `age`, `gender`) provided in the request body will be **ignored** by the API, ensuring strict adherence to doctor's specific responsibilities.

    curl -X PUT \
      "$BASE_URL/api/v1/doctor/patients/$PATIENT_ID" \
      -H 'Content-Type: application/json' \
      -H "Authorization: $DOCTOR_TOKEN" \
      -d '{
        "diagnosis": "Common cold, advised rest and fluids."
        # Example: "name": "Dr. trying to change name" will be ignored by the API
      }'
    

#### 7\. `GET /api/v1/doctor/patients` – Get Patients (Search & Pagination)

Doctors have the same search and pagination capabilities as receptionists for retrieving patient lists.

    curl -X GET \
      "$BASE_URL/api/v1/doctor/patients?name=patient&page=1&limit=10" \
      -H "Authorization: $DOCTOR_TOKEN"
    

### Special Functionality: CSV Export

Both roles can export patient data.

#### 8\. `GET /api/v1/receptionist/patients/:id/export/csv` – Export Patient to CSV (Receptionist)

    curl -X GET \
      "$BASE_URL/api/v1/receptionist/patients/$PATIENT_ID/export/csv" \
      -H "Authorization: $RECEPTIONIST_TOKEN" \
      -o "receptionist_patient_demo.csv" # Saves the CSV output to a local file
    

#### 9\. `GET /api/v1/doctor/patients/:id/export/csv` – Export Patient to CSV (Doctor)

    curl -X GET \
      "$BASE_URL/api/v1/doctor/patients/$PATIENT_ID/export/csv" \
      -H "Authorization: $DOCTOR_TOKEN" \
      -o "doctor_patient_demo.csv"
    

### Role-Based Access Control in Action (Forbidden Actions)

These examples explicitly demonstrate the API's strict role enforcement.

#### 10\. Doctor Tries to Add a Patient (`POST /api/v1/receptionist/patients`)

    curl -X POST \
      "$BASE_URL/api/v1/receptionist/patients" \
      -H 'Content-Type: application/json' \
      -H "Authorization: $DOCTOR_TOKEN" \
      -d '{
        "name": "Forbidden Patient",
        "age": 30,
        "gender": "Male"
      }'
    

**Expected:** `403 Forbidden` status with an "Access denied: Insufficient permissions" error.

#### 11\. Receptionist Tries to Update Patient Diagnosis (`PUT /api/v1/receptionist/patients/:id`)

    curl -X PUT \
      "$BASE_URL/api/v1/receptionist/patients/$PATIENT_ID" \
      -H 'Content-Type: application/json' \
      -H "Authorization: $RECEPTIONIST_TOKEN" \
      -d '{
        "diagnosis": "Attempted to diagnose by receptionist."
      }'
    

**Expected:** `400 Bad Request` status (as the API specifically rejects diagnosis updates from receptionists on this endpoint).

## Design Decisions & Proven Concepts

*   **Role-Based Access Control:** Implemented using **JWT (JSON Web Tokens)** for stateless authentication. JWT claims include user role, which is then validated by Fiber middleware (`auth.RoleMiddleware`) to enforce permissions on specific API routes. This ensures that a receptionist cannot perform doctor actions, and vice-versa, providing a secure and granular access model.
    
*   **Repository Design Pattern:** The application correctly separates data persistence logic from business logic.
    
    *   **Interfaces (`models.Storage`, `models.Account`)** define contracts for data operations, abstracting the underlying database.
        
    *   A **concrete implementation (`models.PostgresStore`)** interacts with the PostgreSQL database.
        
    *   API Handlers (`routes.APIServer`) depend only on these interfaces, making the application **highly testable** and **maintainable** (easy to swap out database implementations without altering handlers).
        
*   **Unit Testing for API Routes:** Comprehensive unit tests are provided for the API handlers in the `routes` package. These tests leverage Go's built-in `testing` package and **manual mock implementations** of the `Storage` and `Account` interfaces. This approach ensures that handler logic, request parsing, response formatting, and error handling for various scenarios (including role-based access) are verified in isolation, without requiring a live database connection.
    
*   **Standard Go Practices:**
    
    *   **Explicit Error Handling:** Errors are returned as values and checked meticulously. Error **wrapping** (`fmt.Errorf("...: %w", err)`) is used to preserve the original error context for better debugging.
        
    *   **Clarity and Simplicity:** Code prioritizes readability and straightforward logic, adhering to Go's idiomatic conventions.
        
    *   **Environment Variables:** Sensitive configurations are loaded from environment variables, promoting security and portability.
        

## Database Schema

The API's database schema is defined as follows:

    -- Table "public.users"
    CREATE TABLE users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        name VARCHAR(255) NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        password VARCHAR(255) NOT NULL,
        role VARCHAR(255) NOT NULL
    );
    
    -- Constraint to ensure only 'receptionist' or 'doctor' roles
    ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role::text = ANY (ARRAY['receptionist'::character varying, 'doctor'::character varying]::text[]));
    
    
    -- Table "public.patients"
    CREATE TABLE patients (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        name VARCHAR(255) NOT NULL,
        age INTEGER NOT NULL,
        gender VARCHAR(255) NOT NULL,
        diagnosis TEXT, -- This column is NULLABLE
        created_by UUID NOT NULL,
        CONSTRAINT fk_user
            FOREIGN KEY(created_by)
            REFERENCES users(id)
            ON DELETE RESTRICT
    );
    
    -- Indexes for performance
    CREATE INDEX idx_users_email ON users (email);
    CREATE INDEX idx_patients_name ON patients (name);

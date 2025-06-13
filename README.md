# Go Patient Portal API: A Role-Based Patient Management System

This project provides a robust backend for managing patient records, complete with secure **user authentication**, **role-based access control**, and efficient **CRUD operations**.

## Key Features & Enhancements

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


## Demo Video Link
[![Watch the video](preview.png)](https://drive.google.com/file/d/13L5KZUUhUw1fndn6n9ap8tRca4Ox7S8R/view)

## Getting Started

Follow these simple steps to get the API running on your local machine.

# Project Setup Guide

### Prerequisites

Before you begin, ensure you have the following installed on your system:

- **Git**: For cloning the repository.  
- **Go (Golang)**: Version 1.22 or higher.  
  [Download Go](https://go.dev/dl/)  
- **Docker**: Required to run PostgreSQL using a container.  
  [Download Docker Desktop](https://www.docker.com/products/docker-desktop)

> You do **not** need to install PostgreSQL or psql locally.

---

### Step 1: Clone the Repository

```bash
git clone https://github.com/Faizan2005/Golang_Coding_Assessment_Makerble.git
cd Golang_Coding_Assessment_Makerble
```

---

### Step 2: Set Up PostgreSQL with Docker

Start a PostgreSQL instance using Docker:

```bash
docker run --name some-postgres -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d postgres
```

This starts a container with:
- Username: `postgres` (default)
- Password: `mysecretpassword`
- Port: `5432`

---

### Step 3: Create the Database Inside Docker

Access the PostgreSQL container:

```bash
docker exec -it some-postgres psql -U postgres
```

Inside the PostgreSQL shell, run:

```sql
CREATE DATABASE hospital;
\q
```

This creates the `hospital` database.

---

### Step 4: Apply the Schema Using Docker

Schema file is located at `./migrations/init.sql`, copy it into the container:

```bash
docker cp ./migrations/init.sql some-postgres:/init.sql
```

Then apply the schema:

```bash
docker exec -it some-postgres psql -U postgres -d hospital -f /init.sql
```

---

### Step 5: Create the `.env` File

Create a `.env` file in the root of your project with the following content:

```env
# .env - Environment variables for the application
DB_USER=postgres
DB_PASSWORD=mysecretpassword
DB_NAME=hospital
DB_HOST=localhost
DB_PORT=5432
JWT_SECRET=cjnvjerfg48unvbjirnv9854hg8945tu895hgf8tu34
LISTEN_ADDR=3000
```

Make sure the values here match your Docker configuration.

---

### Step 6: Install Go Dependencies

```bash
go mod tidy
```

---

### Step 7: Run the Application

```bash
go run main.go
```

Once the app starts, you should see a message like:

```
Application starting on :3000
Connected to database!
```

The API is now running at:  
**http://localhost:3000**

---

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

### Receptionist Portal Endpoints (`/api/receptionist/patients`)

Receptionists manage core patient information.

#### 3\. `POST /api/receptionist/patients` – Add a New Patient

**Receptionists add patients without specifying diagnosis.** The API explicitly sets diagnosis to `NULL`. **Attempts to include a `diagnosis` field in the request by a receptionist will be rejected.**

    curl -X POST \
      "$BASE_URL/api/receptionist/patients" \
      -H 'Content-Type: application/json' \
      -H "Authorization: $RECEPTIONIST_TOKEN" \
      -d '{
        "name": "Patient X",
        "age": 25,
        "gender": "Female"
      }'
    

**ACTION:** From the successful JSON response, copy the **`id`** value (e.g., `"id": "some-uuid"`) and update the `PATIENT_ID` variable: `PATIENT_ID="<COPIED_PATIENT_UUID_HERE>"`

#### 4\. `GET /api/receptionist/patients` – Get Patients (Search & Pagination)

Receptionists can retrieve patient lists with powerful filtering and pagination.

*   **Get all patients (default pagination: page 1, limit 10, no search query):**
    
        curl -X GET \
          "$BASE_URL/api/receptionist/patients" \
          -H "Authorization: $RECEPTIONIST_TOKEN"
        
    
*   **Get patients filtered by name (partial & case-insensitive search):**
    
        curl -X GET \
          "$BASE_URL/api/receptionist/patients?name=x" \
          -H "Authorization: $RECEPTIONIST_TOKEN"
        
    
*   **Get patients with specific pagination (e.g., page 1, limit 1 result):**
    
        curl -X GET \
          "$BASE_URL/api/receptionist/patients?page=1&limit=1" \
          -H "Authorization: $RECEPTIONIST_TOKEN"
        
    

#### 5\. `PUT /api/receptionist/patients/:id` – Update Patient Details

**Receptionists can only update `name`, `age`, and `gender`.** If a `diagnosis` field is included in the request body, the API will specifically reject the request with a `400 Bad Request` error.

    curl -X PUT \
      "$BASE_URL/api/receptionist/patients/$PATIENT_ID" \
      -H 'Content-Type: application/json' \
      -H "Authorization: $RECEPTIONIST_TOKEN" \
      -d '{
        "name": "Patient X (Updated by Receptionist)",
        "age": 26,
        "gender": "Female"
      }'
    

### Doctor Portal Endpoints (`/api/doctor/patients`)

Doctors focus primarily on patient diagnosis and viewing records.

#### 6\. `PUT /api/doctor/patients/:id` – Update Patient Diagnosis

**Doctors can ONLY update the `diagnosis` field.** Any other fields (like `name`, `age`, `gender`) provided in the request body will be **ignored** by the API, ensuring strict adherence to doctor's specific responsibilities.

    curl -X PUT \
      "$BASE_URL/api/doctor/patients/$PATIENT_ID" \
      -H 'Content-Type: application/json' \
      -H "Authorization: $DOCTOR_TOKEN" \
      -d '{
        "diagnosis": "Common cold, advised rest and fluids."
        # Example: "name": "Dr. trying to change name" will be ignored by the API
      }'
    

#### 7\. `GET /api/doctor/patients` – Get Patients (Search & Pagination)

Doctors have the same search and pagination capabilities as receptionists for retrieving patient lists.

    curl -X GET \
      "$BASE_URL/api/doctor/patients?name=patient&page=1&limit=10" \
      -H "Authorization: $DOCTOR_TOKEN"
    

### Special Functionality: CSV Export

Both roles can export patient data.

#### 8\. `GET /api/receptionist/patients/:id/export/csv` – Export Patient to CSV (Receptionist)

    curl -X GET \
      "$BASE_URL/api/receptionist/patients/$PATIENT_ID/export/csv" \
      -H "Authorization: $RECEPTIONIST_TOKEN" \
      -o "receptionist_patient_demo.csv" # Saves the CSV output to a local file
    

#### 9\. `GET /api/doctor/patients/:id/export/csv` – Export Patient to CSV (Doctor)

    curl -X GET \
      "$BASE_URL/api/doctor/patients/$PATIENT_ID/export/csv" \
      -H "Authorization: $DOCTOR_TOKEN" \
      -o "doctor_patient_demo.csv"
    

### Role-Based Access Control in Action (Forbidden Actions)

These examples explicitly demonstrate the API's strict role enforcement.

#### 10\. Doctor Tries to Add a Patient (`POST /api/receptionist/patients`)

    curl -X POST \
      "$BASE_URL/api/receptionist/patients" \
      -H 'Content-Type: application/json' \
      -H "Authorization: $DOCTOR_TOKEN" \
      -d '{
        "name": "Forbidden Patient",
        "age": 30,
        "gender": "Male"
      }'
    

**Expected:** `403 Forbidden` status with an "Access denied: Insufficient permissions" error.

#### 11\. Receptionist Tries to Update Patient Diagnosis (`PUT /api/receptionist/patients/:id`)

    curl -X PUT \
      "$BASE_URL/api/receptionist/patients/$PATIENT_ID" \
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

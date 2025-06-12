Go Patient Portal API: A Role-Based Patient Management System
=============================================================

Welcome to the Go Patient Portal API! This project provides a robust backend for managing patient records, complete with secure **user authentication**, **role-based access control**, and efficient **CRUD operations**.

Key Features & Enhancements
---------------------------

This API goes beyond basic patient management by incorporating:

*   **Role-Based Access Control:**
    
    *   **Receptionists** have permissions to register new users, **add new patients** (without diagnosis), **update patient details** (excluding diagnosis), **retrieve patient lists**, and **delete patients**.
        
    *   **Doctors** can **retrieve patient lists**, **update patient diagnosis only**, and do not have permissions to add or delete patients.
        
    *   The system strictly enforces these roles; for example, a receptionist cannot set a patient's diagnosis, and a doctor cannot update a patient's name or delete their record.
        
*   **Patient Search:** Easily find patients by their **name** using partial or incomplete queries. The search is case-insensitive, making it user-friendly.
    
*   **Pagination:** Efficiently retrieve large datasets of patients by specifying page and limit parameters in GET requests. This improves performance by reducing the data transferred and enhances the user experience for browsing patient records.
    
*   **CSV Export:** Download a single patient's complete details as a **CSV file**, providing a convenient and standard way to extract data for reporting or analysis.
    
*   **Secure Authentication:** Utilizes **JWT (JSON Web Tokens)** for stateless user authentication, ensuring secure communication between the client and API.
    
*   **Robust Error Handling:** The API includes comprehensive validation and explicit error handling for all endpoints, ensuring data integrity and providing clear feedback for invalid requests or internal issues. Edge cases are robustly handled by the API's logic.
    
*   **Database Persistence:** Leverages **PostgreSQL** as a reliable and scalable relational database for data storage.
    

Getting Started
---------------

Follow these simple steps to get the API running on your local machine.

### Prerequisites

Before you begin, ensure you have the following installed on your system:

1.  **Git**: For cloning the repository.
    
2.  **Go (Golang)**: Version 1.22 or higher.
    
    *   [Download Go](https://golang.org/doc/install)
        
3.  **PostgreSQL**: A running PostgreSQL instance that your application can connect to.
    
    *   You can install it directly on your machine or use a service like docker run --name some-postgres -e POSTGRES\_PASSWORD=mysecretpassword -p 5432:5432 -d postgres for a quick temporary database.
        

### Project Setup

1.  git clone https://github.com/Faizan2005/Golang\_Coding\_Assessment\_Makerble.git # Replace with your actual repo name if differentcd Golang\_Coding\_Assessment\_Makerble
    
2.  **Create Your Environment File (.env):**This file stores your sensitive configuration (like database credentials and JWT secret) and is **not committed to Git** for security reasons.
    
    *   touch .env
        
    *   \# .env - Environment variables for the applicationDB\_USER=your\_postgres\_user # e.g., postgresDB\_PASSWORD=your\_postgres\_password # <--- CHANGE THIS!DB\_NAME=hospitalDB\_HOST=localhost # Or your DB host IP/hostnameDB\_PORT=5432JWT\_SECRET=your\_super\_strong\_random\_jwt\_secret\_key\_atleast\_32\_chars # <--- CHANGE THIS!LISTEN\_ADDR=:3000 # Port your Go app will listen on
        
        *   **Important:** Ensure your JWT\_SECRET is a long, random string for security.
            
3.  **Prepare Database Schema Initialization:**For ease of use, the application is designed to create its necessary database tables (users and patients) automatically on startup if they don't already exist.
    
    *   **ACTION REQUIRED IN YOUR GO CODE:** You need to implement a function in your Go application (e.g., in config/config.go or a new migrations/migrations.go file) that reads and executes the init.sql content against your connected PostgreSQL database. This function should be called from main.go after the database connection is established.
        
    *   mkdir -p migrationstouch migrations/init.sql
        
    *   \-- migrations/init.sql-- Create the users table if it doesn't existCREATE TABLE IF NOT EXISTS users ( id UUID PRIMARY KEY DEFAULT gen\_random\_uuid(), name VARCHAR(255) NOT NULL, email VARCHAR(255) UNIQUE NOT NULL, password VARCHAR(255) NOT NULL, role VARCHAR(255) NOT NULL);-- Add the CHECK constraint for the role column if it doesn't existDO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg\_constraint WHERE conname = 'users\_role\_check') THEN ALTER TABLE users ADD CONSTRAINT users\_role\_check CHECK (role::text = ANY (ARRAY\['receptionist'::character varying, 'doctor'::character varying\]::text\[\])); END IF;END $$;-- Create the patients table if it doesn't existCREATE TABLE IF NOT EXISTS patients ( id UUID PRIMARY KEY DEFAULT gen\_random\_uuid(), name VARCHAR(255) NOT NULL, age INTEGER NOT NULL, gender VARCHAR(255) NOT NULL, diagnosis TEXT, -- This column is NULLABLE created\_by UUID NOT NULL);-- Add the Foreign Key constraint if it doesn't existDO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg\_constraint WHERE conname = 'patients\_created\_by\_fkey') THEN ALTER TABLE patients ADD CONSTRAINT patients\_created\_by\_fkey FOREIGN KEY (created\_by) REFERENCES users(id) ON DELETE RESTRICT; END IF;END $$;-- Optional: Create indexes for improved query performanceCREATE INDEX IF NOT EXISTS idx\_users\_email ON users (email);CREATE INDEX IF NOT EXISTS idx\_patients\_name ON patients (name);
        
4.  go mod tidy
    
5.  go run main.go
    
    *   Your application will attempt to connect to the PostgreSQL database using the credentials from your .env file and then run the init.sql migrations.
        
    *   Look for the message: Application starting on :3000 (or whatever LISTEN\_ADDR you set).
        
    *   Once this message appears, your API is up and running on http://localhost:3000.
        

How to Use the API
------------------

The API is accessible at http://localhost:3000. You'll use curl (command line) or tools like Postman/Insomnia to send requests.

### Authentication Flow

1.  **Register:** Create a new user account, specifying their role (either receptionist or doctor).
    
2.  **Login:** Authenticate with your email and password to receive a **JSON Web Token (JWT)**.
    
3.  **Use JWT:** For all authenticated endpoints, you **must** include the JWT in the Authorization header as a Bearer token (e.g., Authorization: Bearer YOUR\_AUTH\_TOKEN).
    

### Example API Requests (Test Steps)

These curl commands demonstrate the core functionality and role-based access for your API. Remember to replace placeholder values with actual tokens and IDs obtained from previous steps.

#### First, set up a variable for your base URL:

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   BASE_URL="http://localhost:3000"   `

#### 1\. POST /register – Register a New User

Register both a receptionist and a doctor account to test role-based access.

**Register Receptionist:**

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   curl -X POST \    "$BASE_URL/register" \    -H "Content-Type: application/json" \    -d '{      "name": "Alice Receptionist",      "email": "alice.r@example.com",      "password": "alicepass",      "role": "receptionist"    }'   `

**Register Doctor:**

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   curl -X POST \    "$BASE_URL/register" \    -H "Content-Type: application/json" \    -d '{      "name": "Dr. Bob Physician",      "email": "bob.d@example.com",      "password": "bobpass",      "role": "doctor"    }'   `

#### 2\. POST /login – Authenticate and Get a Token

**Login as Receptionist (Copy Bearer to RECEPTIONIST\_TOKEN):**

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   curl -X POST \    "$BASE_URL/login" \    -H "Content-Type: application/json" \    -d '{      "email": "alice.r@example.com",      "password": "alicepass"    }'   `

**ACTION:** From the JSON response, copy the entire Bearer string. For your terminal session, you can set a variable:RECEPTIONIST\_TOKEN="Bearer eyJhbGciOiJIUzI1Ni..."

**Login as Doctor (Copy Bearer to DOCTOR\_TOKEN):**

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   curl -X POST \    "$BASE_URL/login" \    -H "Content-Type: application/json" \    -d '{      "email": "bob.d@example.com",      "password": "bobpass"    }'   `

**ACTION:** Copy the entire Bearer string.DOCTOR\_TOKEN="Bearer eyJhbGciOiJIUzI1Ni..."

#### Now, use RECEPTIONIST\_TOKEN or DOCTOR\_TOKEN for subsequent authenticated requests.

### Receptionist Portal Endpoints (/api/v1/receptionist/patients)

Receptionists manage core patient information.

#### 3\. POST /api/v1/receptionist/patients – Add a New Patient

**Receptionists add patients without specifying diagnosis.** The API explicitly sets diagnosis to NULL for newly added patients. Attempts to include a diagnosis field in the request by a receptionist will result in a 400 Bad Request error.

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   curl -X POST \    "$BASE_URL/api/v1/receptionist/patients" \    -H "Content-Type: application/json" \    -H "Authorization: $RECEPTIONIST_TOKEN" \    -d '{      "name": "Emma Watson",      "age": 33,      "gender": "Female"    }'   `

**ACTION:** From the successful JSON response, copy the id value (e.g., "id": "some-uuid") and save it for future requests:PATIENT\_ID=""

#### 4\. GET /api/v1/receptionist/patients – Get Patients (Search & Pagination)

Receptionists can retrieve patient lists with powerful filtering and pagination.

**Get all patients (default pagination: page 1, limit 10, no search query):**

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   curl -X GET \    "$BASE_URL/api/v1/receptionist/patients" \    -H "Authorization: $RECEPTIONIST_TOKEN"   `

**Get patients filtered by name (partial & case-insensitive search):**

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   curl -X GET \    "$BASE_URL/api/v1/receptionist/patients?name=emma" \    -H "Authorization: $RECEPTIONIST_TOKEN"   `

**Get patients with specific pagination (e.g., page 1, limit 5 results):**

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   curl -X GET \    "$BASE_URL/api/v1/receptionist/patients?page=1&limit=5" \    -H "Authorization: $RECEPTIONIST_TOKEN"   `

_(The API handles edge cases like invalid page/limit values gracefully.)_

#### 5\. PUT /api/v1/receptionist/patients/:id – Update Patient Details

**Receptionists can only update name, age, and gender.** If a diagnosis field is included in the request body, the API will specifically reject the request with a 400 Bad Request error.

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   curl -X PUT \    "$BASE_URL/api/v1/receptionist/patients/$PATIENT_ID" \    -H "Content-Type: application/json" \    -H "Authorization: $RECEPTIONIST_TOKEN" \    -d '{      "name": "Emma Watson Updated",      "age": 34,      "gender": "Female"    }'   `

#### 6\. GET /api/v1/receptionist/patients/:id/export/csv – Export Single Patient to CSV

This endpoint allows receptionists to download a CSV file containing all details for a specific patient by their ID.

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   curl -X GET \    "$BASE_URL/api/v1/receptionist/patients/$PATIENT_ID/export/csv" \    -H "Authorization: $RECEPTIONIST_TOKEN" \    -o "patient_data_$PATIENT_ID.csv" # Saves the CSV output to a local file   `

_(Verify the .csv file content after download.)_

#### 7\. DELETE /api/v1/receptionist/patients/:id – Delete a Patient

Receptionists have the authority to delete patient records.

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   curl -X DELETE \    "$BASE_URL/api/v1/receptionist/patients/$PATIENT_ID" \    -H "Authorization: $RECEPTIONIST_TOKEN"   `

_(Expected: 204 No Content on successful deletion.)_

### Doctor Portal Endpoints (/api/v1/doctor/patients)

Doctors focus primarily on patient diagnosis and viewing records.

#### 8\. PUT /api/v1/doctor/patients/:id – Update Patient Diagnosis

**Doctors can only update the diagnosis field.** Any other fields (like name, age, gender) provided in the request body will be **ignored** by the API, ensuring strict adherence to doctor's specific responsibilities.

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   curl -X PUT \    "$BASE_URL/api/v1/doctor/patients/$PATIENT_ID" \    -H "Content-Type: application/json" \    -H "Authorization: $DOCTOR_TOKEN" \    -d '{      "diagnosis": "Acute Bronchitis, prescribed antibiotics and rest."      # Example: "name": "Dr. trying to change name" will be ignored by the API    }'   `

#### 9\. GET /api/v1/doctor/patients – Get Patients (Search & Pagination)

Doctors have the same search and pagination capabilities as receptionists for retrieving patient lists.

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   curl -X GET \    "$BASE_URL/api/v1/doctor/patients?name=emma&page=1&limit=10" \    -H "Authorization: $DOCTOR_TOKEN"   `

#### 10\. GET /api/v1/doctor/patients/:id/export/csv – Export Single Patient to CSV

Doctors can also export a single patient's details to a CSV file.

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   curl -X GET \    "$BASE_URL/api/v1/doctor/patients/$PATIENT_ID/export/csv" \    -H "Authorization: $DOCTOR_TOKEN" \    -o "doctor_patient_data_$PATIENT_ID.csv"   `

_(Verify the .csv file content after download.)_

Design Decisions & Proven Concepts
----------------------------------

*   **Role-Based Access Control:** Implemented using **JWT (JSON Web Tokens)** for stateless authentication. JWT claims include user role, which is then validated by Fiber middleware (auth.RoleMiddleware) to enforce permissions on specific API routes. This ensures that a receptionist cannot perform doctor actions, and vice-versa, providing a secure and granular access model.
    
*   **Repository Design Pattern:** The application correctly separates data persistence logic from business logic.
    
    *   **Interfaces (models.Storage, models.Account)** define contracts for data operations, abstracting the underlying database.
        
    *   A **concrete implementation (models.PostgresStore)** interacts with the PostgreSQL database.
        
    *   API Handlers (routes.APIServer) depend only on these interfaces, making the application **highly testable** and **maintainable** (easy to swap out database implementations without altering handlers).
        
*   **Unit Testing for API Routes:** Comprehensive unit tests are provided for the API handlers in the routes package. These tests leverage Go's built-in testing package and **manual mock implementations** of the Storage and Account interfaces. This approach ensures that handler logic, request parsing, response formatting, and error handling for various scenarios (including role-based access) are verified in isolation, without requiring a live database connection.
    
*   **Standard Go Practices:**
    
    *   **Explicit Error Handling:** Errors are returned as values and checked meticulously. Error **wrapping** (fmt.Errorf("...: %w", err)) is used to preserve the original error context for better debugging.
        
    *   **Clarity and Simplicity:** Code prioritizes readability and straightforward logic, adhering to Go's idiomatic conventions.
        
    *   **Environment Variables:** Sensitive configurations are loaded from environment variables, promoting security and portability.
        

Database Schema
---------------

The API's database schema is defined as follows:

Plain textANTLR4BashCC#CSSCoffeeScriptCMakeDartDjangoDockerEJSErlangGitGoGraphQLGroovyHTMLJavaJavaScriptJSONJSXKotlinLaTeXLessLuaMakefileMarkdownMATLABMarkupObjective-CPerlPHPPowerShell.propertiesProtocol BuffersPythonRRubySass (Sass)Sass (Scss)SchemeSQLShellSwiftSVGTSXTypeScriptWebAssemblyYAMLXML`   -- Table "public.users"  CREATE TABLE users (      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),      name VARCHAR(255) NOT NULL,      email VARCHAR(255) UNIQUE NOT NULL,      password VARCHAR(255) NOT NULL,      role VARCHAR(255) NOT NULL  );  -- Constraint to ensure only 'receptionist' or 'doctor' roles  ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role::text = ANY (ARRAY['receptionist'::character varying, 'doctor'::character varying]::text[]));  -- Table "public.patients"  CREATE TABLE patients (      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),      name VARCHAR(255) NOT NULL,      age INTEGER NOT NULL,      gender VARCHAR(255) NOT NULL,      diagnosis TEXT, -- This column is NULLABLE      created_by UUID NOT NULL,      CONSTRAINT fk_user          FOREIGN KEY(created_by)          REFERENCES users(id)          ON DELETE RESTRICT  );  -- Indexes for performance  CREATE INDEX idx_users_email ON users (email);  CREATE INDEX idx_patients_name ON patients (name);   `
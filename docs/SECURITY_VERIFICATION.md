# Security Verification: User Registration

## ✅ Verified: Users CANNOT Register as Admin

### Code Review Summary

#### 1. Registration Request Handler
**File:** `internal/handlers/auth_handler.go` (Lines 20-24)

```go
type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    // ✅ NO role field - users cannot specify role
}
```

**✅ SECURE:** The registration request struct does NOT include a `role` field, so users cannot pass it.

#### 2. User Creation Logic
**File:** `internal/services/auth_service.go` (Lines 60-70)

```go
user := &models.User{
    Username:           username,
    Email:              email,
    PasswordHash:       string(hashedPassword),
    Role:               "user",  // ✅ HARDCODED to "user"
    EmailVerified:      false,
    VerificationToken:  token,
    VerificationExpiry: s.emailService.GetVerificationExpiry(),
    CreatedAt:          time.Now(),
    UpdatedAt:          time.Now(),
}
```

**✅ SECURE:** The role is hardcoded to `"user"` in the service layer. There is no parameter or variable that could be manipulated.

#### 3. No Admin Registration Endpoint
**File:** `internal/routes/routes.go`

```go
// Public Routes
r.POST("/auth/register", authHandler.Register)  // Uses Register handler above
r.POST("/auth/login", authHandler.Login)
r.GET("/scoreboard", scoreboardHandler.GetScoreboard)
r.GET("/auth/verify-email", authHandler.VerifyEmail)

// ✅ NO endpoint to register as admin
```

**✅ SECURE:** There is no special endpoint like `/auth/register-admin` or any way to specify role during registration.

#### 4. User Model Definition
**File:** `internal/models/user.go` (Lines 9-23)

```go
type User struct {
    ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Username            string             `bson:"username" json:"username"`
    Email               string             `bson:"email" json:"email"`
    PasswordHash        string             `bson:"password_hash" json:"-"`
    Role                string             `bson:"role" json:"role"` // "admin" or "user"
    // ... other fields
}
```

**ℹ️ INFO:** The model supports both "admin" and "user" roles, but only "user" is set during registration.

#### 5. Admin Middleware Protection
**File:** `internal/middleware/auth_middleware.go` (Line 60)

```go
if !exists || role != "admin" {
    c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
    c.Abort()
    return
}
```

**✅ SECURE:** Admin-only endpoints check the JWT token's role claim, which is set during login based on the database value.

## 🔐 Security Flow

### Normal Registration Flow:
```
1. User sends: POST /auth/register
   Body: { "username": "john", "email": "john@example.com", "password": "pass123" }
   
2. RegisterRequest struct validates input
   ✅ No role field accepted
   
3. AuthService.Register() is called
   ✅ Role is hardcoded to "user"
   
4. User is created in database with role="user"
   
5. User receives verification email
   
6. User verifies email and logs in
   
7. JWT token is generated with role="user"
   ✅ Cannot access admin endpoints
```

### Attempting to Inject Admin Role:
```
1. Malicious user tries: POST /auth/register
   Body: { 
     "username": "hacker", 
     "email": "hacker@example.com", 
     "password": "pass123",
     "role": "admin"  ← Attempting to inject
   }
   
2. RegisterRequest struct IGNORES the role field
   ✅ Go's struct binding only maps defined fields
   
3. AuthService.Register() executes
   ✅ Role is still hardcoded to "user"
   
4. User is created with role="user"
   ❌ Attack fails - user is NOT admin
```

## 🛡️ Security Guarantees

| Attack Vector | Protection | Status |
|---------------|-----------|---------|
| Direct role field in registration | No role field in RegisterRequest | ✅ Protected |
| JSON injection with role | Struct binding ignores unmapped fields | ✅ Protected |
| API parameter manipulation | Role hardcoded in service layer | ✅ Protected |
| Database direct injection | Outside scope of this verification | ⚠️ Depends on DB security |
| JWT token manipulation | JWT signed with secret key | ✅ Protected |
| Admin endpoint bypass | Middleware checks role from verified JWT | ✅ Protected |

## 🎯 How to Create Admin Users

Since registration always creates regular users, admins must be created through:

### Method 1: Admin CLI Tool (Recommended)
```bash
cd backend
go build -o admin-tool cmd/admin/main.go
./admin-tool
# Choose option 1 to create admin user
```

### Method 2: Promote Existing User
```bash
./admin-tool
# Choose option 2 to promote user to admin
```

### Method 3: Direct Database Update (Advanced)
```bash
# Connect to MongoDB
mongo root_access

# Promote user
db.users.updateOne(
  { "username": "johndoe" },
  { $set: { "role": "admin" } }
)
```

## ✅ Conclusion

**Registration is SECURE by design:**
- ✅ Users are created with role="user" by default
- ✅ No way to register as admin through the API
- ✅ Role cannot be manipulated during registration
- ✅ Admin privileges require manual assignment
- ✅ JWT tokens reflect actual database role

**The system follows security best practices:**
- Principle of Least Privilege (users start with minimal access)
- Defense in Depth (multiple layers prevent admin escalation)
- Explicit Authorization (admin must be explicitly granted)

## 📅 Last Verified
- Date: 2026-01-13
- Verified By: Code Review
- Files Checked: 5 core security files
- Status: ✅ SECURE

# Quick Start: Creating Your First Admin

## ✅ Verified: Registration is Secure

**By default, ALL registered users are regular users (NOT admins).**

The role is hardcoded to `"user"` in the backend code and cannot be changed through the API.

## 🚀 Create Your First Admin (3 Steps)

### Step 1: Build the Admin Tool
```bash
cd backend
go build -o admin-tool cmd/admin/main.go
```

### Step 2: Run the Tool
```bash
./admin-tool
```

### Step 3: Create Admin User
```
Choose option: 1

Username: admin
Email: admin@yourdomain.com
Password: YourSecurePassword123!

✅ Done! Admin created.
```

## 📋 Quick Commands

### Create Admin User
```bash
cd backend && go run cmd/admin/main.go
# Then choose option 1
```

### Promote Existing User to Admin
```bash
cd backend && go run cmd/admin/main.go
# Then choose option 2 and enter username/email
```

### List All Users
```bash
cd backend && go run cmd/admin/main.go
# Then choose option 4
```

## 🔒 Security Verification

✅ **Checked:** `internal/handlers/auth_handler.go`
- RegisterRequest has NO role field

✅ **Checked:** `internal/services/auth_service.go`
- Role is HARDCODED to `"user"` on line 64

✅ **Checked:** `internal/routes/routes.go`
- NO admin registration endpoint exists

✅ **Checked:** `internal/middleware/auth_middleware.go`
- Admin routes protected by role check

## 📖 More Information

- **Full Documentation:** See `ADMIN_TOOL.md`
- **Security Details:** See `SECURITY_VERIFICATION.md`
- **Auth Setup:** See `../AUTH_SETUP.md`

## ⚡ Pro Tip

Add this to your `.gitignore`:
```
backend/admin-tool
backend/admin-tool.exe
```

This prevents accidentally committing the compiled binary.

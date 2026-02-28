# OAuth-Based Authentication with Email Verification - Setup Guide

## 🔐 Features Implemented

### Backend (Go)
- ✅ Email-based registration with validation
- ✅ Email domain verification (checks if domain exists)
- ✅ Email verification tokens with expiry
- ✅ Password reset functionality
- ✅ JWT token-based authentication
- ✅ Password hashing with bcrypt
- ✅ Beautiful HTML email templates
- ✅ SMTP email sending
- ✅ Change password for logged-in users

### Frontend (Angular)
- ✅ Registration with username, email, and password
- ✅ Login with username or email
- ✅ Email verification page
- ✅ Success/error message handling
- ✅ Responsive UI with dark/light theme support
- ✅ Form validation

## 📧 Email Configuration

### 1. Using Gmail (Recommended for Testing)

1. **Enable 2-Factor Authentication** on your Gmail account
2. **Generate an App Password**:
   - Go to: https://myaccount.google.com/apppasswords
   - Select "Mail" and "Other (Custom name)"
   - Enter "RootAccess CTF"
   - Copy the generated 16-character password

3. **Update `.env` file** in `backend/`:
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-16-char-app-password-here
SMTP_FROM=RootAccess CTF <noreply@rootaccess.ctf>
```

### 2. Using Other Email Providers

#### Outlook/Hotmail
```bash
SMTP_HOST=smtp.office365.com
SMTP_PORT=587
SMTP_USER=your-email@outlook.com
SMTP_PASS=your-password
```

#### Yahoo Mail
```bash
SMTP_HOST=smtp.mail.yahoo.com
SMTP_PORT=587
SMTP_USER=your-email@yahoo.com
SMTP_PASS=your-app-password
```

#### SendGrid
```bash
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASS=your-sendgrid-api-key
```

## 🚀 Getting Started

### Backend Setup

1. **Install dependencies**:
```bash
cd backend
go mod tidy
```

2. **Configure environment**:
Create a `.env` file in the `backend/` directory:
```bash
MONGO_URI=mongodb://localhost:27017
DB_NAME=go_ctf
JWT_SECRET=your-super-secret-jwt-key-change-in-production
PORT=8080
FRONTEND_URL=http://localhost:4200
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-gmail-app-password
SMTP_FROM=RootAccess CTF <noreply@rootaccess.ctf>
```

3. **Start MongoDB**:
```bash
# Using Docker
docker run -d -p 27017:27017 --name mongodb mongo:latest

# Or using Docker Compose
docker-compose up -d mongodb
```

4. **Run the backend**:
```bash
cd backend
go run cmd/api/main.go
```

Server will start on `http://localhost:8080`

### Frontend Setup

1. **Install dependencies**:
```bash
cd frontend
npm install
```

2. **Start the frontend**:
```bash
npm start
```

App will run on `http://localhost:4200`

## 📋 API Endpoints

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Register new user |
| POST | `/auth/login` | Login user |
| GET/POST | `/auth/verify-email?token=xxx` | Verify email |
| POST | `/auth/resend-verification` | Resend verification email |
| POST | `/auth/forgot-password` | Request password reset |
| POST | `/auth/reset-password` | Reset password with token |
| GET | `/scoreboard` | View scoreboard |

### Protected Endpoints (Requires Authentication)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/change-password` | Change password |
| GET | `/challenges` | List all challenges |
| GET | `/challenges/:id` | Get challenge details |
| POST | `/challenges/:id/submit` | Submit flag |

### Admin Endpoints (Requires Admin Role)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/challenges` | Create new challenge |

## 🔄 Authentication Flow

### 1. Registration Flow

```
User submits registration form
    ↓
Backend validates email format and domain
    ↓
Backend creates user (email_verified: false)
    ↓
Backend generates verification token
    ↓
Backend sends verification email
    ↓
User clicks link in email
    ↓
Frontend navigates to /verify-email?token=xxx
    ↓
Backend verifies token and marks email as verified
    ↓
User can now log in
```

### 2. Login Flow

```
User enters username/email and password
    ↓
Backend checks if email is verified
    ↓
Backend verifies password
    ↓
Backend generates JWT token
    ↓
Frontend stores token in localStorage
    ↓
User is authenticated
```

### 3. Email Verification

```
User receives email with verification link
    ↓
Link format: http://localhost:4200/verify-email?token=xxx
    ↓
Frontend calls backend API with token
    ↓
Backend validates token and expiry (24 hours)
    ↓
Backend marks email as verified
    ↓
Success! User can now log in
```

## 🧪 Testing

### Test Registration

1. Navigate to `http://localhost:4200/register`
2. Fill in:
   - Username: testuser
   - Email: your-real-email@gmail.com
   - Password: TestPassword123
3. Click "Create Account"
4. Check your email for verification link
5. Click the link to verify
6. Log in with your credentials

### Test Email Verification

The system validates:
- ✅ Email format (RFC 5322 compliant)
- ✅ Domain MX records (mail server exists)
- ✅ Token expiry (24 hours)
- ✅ Duplicate email prevention

## 🎨 Email Templates

The system sends beautiful, branded HTML emails with:
- Dark theme matching RootAccess design
- Red gradient accents
- Responsive layout
- Clear call-to-action buttons
- Security information
- Professional branding

## 🔒 Security Features

1. **Password Hashing**: bcrypt with default cost
2. **JWT Tokens**: 7-day expiry, HS256 signing
3. **Email Validation**: Format + Domain verification
4. **Token Expiry**: 
   - Email verification: 24 hours
   - Password reset: 1 hour
5. **No Password Storage**: Only bcrypt hashes stored
6. **CORS Protection**: Configured for frontend origin
7. **Rate Limiting**: Implement in production

## 📝 Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| MONGO_URI | MongoDB connection string | mongodb://localhost:27017 |
| DB_NAME | Database name | go_ctf |
| JWT_SECRET | Secret for JWT signing | random-secure-string |
| PORT | Backend server port | 8080 |
| FRONTEND_URL | Frontend URL for emails | http://localhost:4200 |
| SMTP_HOST | SMTP server hostname | smtp.gmail.com |
| SMTP_PORT | SMTP server port | 587 |
| SMTP_USER | SMTP username/email | your-email@gmail.com |
| SMTP_PASS | SMTP password/app password | your-app-password |
| SMTP_FROM | From email address | RootAccess CTF <noreply@rootaccess.ctf> |

## 🐛 Troubleshooting

### Email Not Sending

1. **Check SMTP credentials**:
   - Verify username and password
   - For Gmail, ensure you're using App Password, not regular password

2. **Check firewall**:
   - Ensure port 587 is not blocked
   - Some networks block SMTP

3. **Check logs**:
   - Backend will log email sending errors
   - Look for connection errors

### Verification Link Not Working

1. **Check token expiry**: Tokens expire after 24 hours
2. **Check frontend URL**: Must match FRONTEND_URL in backend `.env`
3. **Resend verification**: Use "Resend Verification" feature

### Login Issues

1. **Email not verified**: Check inbox for verification email
2. **Wrong credentials**: Username/email and password must match
3. **Token expired**: Log in again to get new token

## 📚 Additional Resources

- [Go Mail Documentation](https://github.com/go-mail/mail)
- [JWT Documentation](https://jwt.io/)
- [Angular HTTP Client](https://angular.io/guide/http)
- [MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current/)

## 🎯 Next Steps

1. **Set up production SMTP**: Use SendGrid, AWS SES, or similar
2. **Add rate limiting**: Prevent spam registrations
3. **Add CAPTCHA**: Prevent bot registrations
4. **Add email change**: Allow users to update email
5. **Add social OAuth**: Google, GitHub, etc.
6. **Add 2FA**: Two-factor authentication
7. **Add email templates**: More email types
8. **Add logging**: Comprehensive logging system

## 💡 Tips

- **Testing**: Use a real email for testing, temp emails may not work
- **Security**: Change JWT_SECRET in production
- **SMTP**: Gmail has daily sending limits (500/day)
- **Production**: Use dedicated email service (SendGrid, AWS SES)
- **Monitoring**: Monitor email delivery rates

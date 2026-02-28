# Google OAuth Implementation Summary

Google OAuth sign-in/sign-up has been successfully implemented for the RootAccess CTF Platform.

## What Was Implemented

### Backend Changes

1. **Dependencies Added**
   - `golang.org/x/oauth2` - OAuth2 client library
   - `golang.org/x/oauth2/google` - Google-specific OAuth2 configuration

2. **Configuration Updates** (`backend/internal/config/config.go`)
   - Added `GoogleClientID`, `GoogleClientSecret`, and `GoogleRedirectURL` fields
   - Environment variables: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`

3. **OAuth Service** (`backend/internal/services/oauth_service.go`)
   - `GetGoogleAuthURL()` - Generates OAuth consent URL with CSRF state
   - `ExchangeGoogleCode()` - Exchanges authorization code for access token
   - `GetGoogleUserInfo()` - Fetches user profile from Google API
   - `HandleGoogleCallback()` - Full OAuth callback flow:
     - Finds existing users by email or creates new ones
     - Auto-links Google accounts to existing email/password accounts
     - Marks OAuth users as email-verified automatically
     - Generates JWT token for authentication

4. **OAuth Handler** (`backend/internal/handlers/oauth_handler.go`)
   - `GoogleLogin()` - Initiates OAuth flow with CSRF protection (Redis-based state tokens)
   - `GoogleCallback()` - Handles OAuth callback, validates state, sets auth cookie

5. **Routes Added** (`backend/internal/routes/routes.go`)
   - `GET /auth/google` - Initiates Google OAuth flow
   - `GET /auth/google/callback` - Handles OAuth callback from Google

6. **Repository Updates** (`backend/internal/repositories/user_repository.go`)
   - Added `FindByProviderID()` method for finding users by OAuth provider

7. **Security Features**
   - CSRF protection using Redis-stored state tokens (10-minute expiry)
   - HTTP-only cookies for JWT tokens
   - Auto-linking by email (existing users can sign in with Google)
   - Email verification bypass for OAuth users (Google already verified)

### Frontend Changes

1. **Environment Configuration**
   - Added `googleAuthUrl` to both development and production environments
   - Development: `http://localhost:8080/auth/google`
   - Production: `http://148.100.79.152:8080/auth/google`

2. **Auth Service Updates** (`frontend/src/app/services/auth.ts`)
   - Added `loginWithGoogle()` method that redirects to backend OAuth endpoint

3. **Login Component** (`frontend/src/app/components/login/`)
   - Added "Continue with Google" button with Google branding
   - Includes Google logo SVG
   - OR divider between email/password and Google sign-in

4. **Register Component** (`frontend/src/app/components/register/`)
   - Added "Sign up with Google" button
   - Same styling and branding as login component

5. **OAuth Callback Component** (`frontend/src/app/components/oauth-callback/`)
   - Shows loading state during authentication
   - Displays success message with username
   - Shows error message if authentication fails
   - Auto-redirects to challenges page on success
   - Auto-redirects to login page on error
   - Route: `/auth/callback`

### Documentation

1. **GOOGLE_OAUTH_SETUP.md** - Comprehensive guide for:
   - Creating Google Cloud project
   - Configuring OAuth consent screen
   - Creating OAuth 2.0 credentials
   - Setting up redirect URIs
   - Troubleshooting common errors
   - Security best practices

## How It Works

### OAuth Flow

```
1. User clicks "Continue with Google"
   ↓
2. Frontend redirects to backend /auth/google
   ↓
3. Backend generates CSRF state token → stores in Redis → redirects to Google
   ↓
4. User approves on Google consent screen
   ↓
5. Google redirects to backend /auth/google/callback with code
   ↓
6. Backend validates state → exchanges code for token → gets user info
   ↓
7. Backend finds/creates user → generates JWT → sets HTTP-only cookie
   ↓
8. Backend redirects to frontend /auth/callback?success=true
   ↓
9. Frontend checks auth status → redirects to /challenges
```

### Account Linking

- **Existing Users**: If a user with the same email exists, the Google account is linked automatically
- **New Users**: A new account is created with:
  - Email from Google
  - Username derived from Google name or email
  - `EmailVerified: true` (no verification email needed)
  - `PasswordHash: ""` (OAuth users don't have passwords)
  - OAuth credentials stored in `User.OAuth` field

### Security Features

1. **CSRF Protection**: State tokens stored in Redis with 10-minute expiry
2. **HTTP-only Cookies**: JWT tokens cannot be accessed by JavaScript
3. **Email Verification**: OAuth users automatically marked as verified
4. **Auto-linking**: Prevents duplicate accounts for same email

## Setup Instructions

### 1. Google Cloud Console Setup

Follow the instructions in `GOOGLE_OAUTH_SETUP.md` to:
- Create a Google Cloud project
- Configure OAuth consent screen
- Get Client ID and Client Secret
- Add redirect URIs

### 2. Update Backend Environment

Edit `backend/.env`:

```env
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback
```

### 3. Restart Backend

```bash
cd backend
go run cmd/api/main.go
```

### 4. Test the Flow

1. Navigate to `http://localhost:4200/login`
2. Click "Continue with Google"
3. Select your Google account
4. Approve the permissions
5. You should be redirected back and logged in

## Files Created

**Backend:**
- `backend/internal/services/oauth_service.go`
- `backend/internal/handlers/oauth_handler.go`

**Frontend:**
- `frontend/src/app/components/oauth-callback/oauth-callback.ts`
- `frontend/src/app/components/oauth-callback/oauth-callback.html`
- `frontend/src/app/components/oauth-callback/oauth-callback.scss`

**Documentation:**
- `GOOGLE_OAUTH_SETUP.md`
- `OAUTH_IMPLEMENTATION_SUMMARY.md`

## Files Modified

**Backend:**
- `backend/go.mod` - Added OAuth dependencies
- `backend/internal/config/config.go` - Added OAuth config fields
- `backend/.env` - Added OAuth credentials
- `backend/internal/routes/routes.go` - Added OAuth routes and wiring
- `backend/internal/repositories/user_repository.go` - Added `FindByProviderID()`

**Frontend:**
- `frontend/src/environments/environment.ts` - Added `googleAuthUrl`
- `frontend/src/environments/environment.prod.ts` - Added `googleAuthUrl`
- `frontend/src/app/services/auth.ts` - Added `loginWithGoogle()`
- `frontend/src/app/components/login/login.ts` - Added Google login handler
- `frontend/src/app/components/login/login.html` - Added Google button
- `frontend/src/app/components/register/register.ts` - Added Google signup handler
- `frontend/src/app/components/register/register.html` - Added Google button
- `frontend/src/app/app.routes.ts` - Added OAuth callback route

## Extensibility

The implementation is designed to be easily extended for additional OAuth providers:

1. The OAuth service uses standard OAuth2 patterns
2. The handler can be extended with similar methods for other providers
3. Routes can be added following the pattern: `/auth/{provider}` and `/auth/{provider}/callback`
4. Frontend can add buttons for other providers using the same pattern

Example for adding GitHub OAuth:
- Backend: Add `GitHubLogin()` and `GitHubCallback()` handlers
- Backend: Add `/auth/github` and `/auth/github/callback` routes
- Frontend: Add `loginWithGitHub()` to auth service
- Frontend: Add "Sign in with GitHub" button to login/register components

## Testing Checklist

- [ ] New user can sign up with Google
- [ ] Existing user can sign in with Google (account linking)
- [ ] OAuth users are marked as email-verified
- [ ] JWT cookie is set correctly
- [ ] User is redirected to challenges page
- [ ] CSRF protection works (state token validation)
- [ ] Error handling works (declined consent, invalid code)
- [ ] Works on both localhost and production domains

## Known Limitations

1. **Email Required**: Users must have a verified email on their Google account
2. **Username Conflicts**: If username derived from Google name already exists, a numeric suffix is added
3. **Redis Required**: State tokens are stored in Redis; if Redis is down, OAuth won't work
4. **No Token Refresh**: Google refresh tokens are stored but not currently used for long-lived sessions

## Future Enhancements

1. Add support for additional OAuth providers (GitHub, Discord, etc.)
2. Implement OAuth token refresh for long-lived sessions
3. Allow users to link/unlink multiple OAuth accounts
4. Add OAuth account management UI in settings
5. Support for OAuth-only accounts (users who never set a password)

## Support

If you encounter issues:
1. Check backend logs for detailed error messages
2. Verify Redis is running and accessible
3. Confirm Google Cloud Console configuration matches `.env` settings
4. Review `GOOGLE_OAUTH_SETUP.md` for troubleshooting steps

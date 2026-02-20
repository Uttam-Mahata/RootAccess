# Cookie-Based Authentication Migration

## ✅ Completed: Migration from localStorage to HTTP-only Cookies

Successfully migrated the authentication system from **localStorage-based JWT tokens** to **secure HTTP-only cookie-based authentication**.

---

## 🔐 Security Improvements

| Feature | Before (localStorage) | After (Cookies) |
|---------|----------------------|-----------------|
| **XSS Protection** | ❌ Vulnerable | ✅ Protected (httpOnly) |
| **CSRF Protection** | ✅ Not needed | ⚠️ Consider adding tokens |
| **Token Storage** | JavaScript accessible | Browser-only (httpOnly) |
| **Automatic Sending** | ❌ Manual headers | ✅ Automatic with requests |
| **Security Level** | Medium | High |

---

## 🔧 Backend Changes

### 1. **Auth Handler** (`backend/internal/handlers/auth_handler.go`)

#### ✅ Updated Login Handler
- Sets JWT token in **HTTP-only cookie** instead of JSON response
- Cookie name: `auth_token`
- Expiry: 7 days
- Path: `/`
- HttpOnly: `true` (prevents JavaScript access)
- Secure: `false` (set to `true` in production with HTTPS)

```go
// Set HTTP-only cookie with JWT token
c.SetCookie(
    "auth_token",    // name
    token,           // value
    7*24*60*60,      // maxAge (7 days)
    "/",             // path
    "",              // domain
    false,           // secure (set true for HTTPS)
    true,            // httpOnly
)
```

#### ✅ Added Logout Handler
- Clears the authentication cookie
- Sets maxAge to -1 to delete cookie

```go
func (h *AuthHandler) Logout(c *gin.Context) {
    c.SetCookie("auth_token", "", -1, "/", "", false, true)
    c.JSON(200, gin.H{"message": "Logged out successfully!"})
}
```

---

### 2. **Auth Service** (`backend/internal/services/auth_service.go`)

#### ✅ Updated Login Method
- Returns both **JWT token** AND **user info**
- User info sent in response body (not sensitive)
- Token stored in cookie (secure)

```go
type UserInfo struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role"`
}

func (s *AuthService) Login(...) (string, *UserInfo, error)
```

---

### 3. **Auth Middleware** (`backend/internal/middleware/auth_middleware.go`)

#### ✅ Updated to Read from Cookie
- **Primary**: Reads token from `auth_token` cookie
- **Fallback**: Still supports `Authorization` header (backward compatibility)

```go
// Try to get token from cookie first
tokenString, err := c.Cookie("auth_token")

// Fallback to Authorization header
if err != nil || tokenString == "" {
    authHeader := c.GetHeader("Authorization")
    // ...
}
```

---

### 4. **Routes** (`backend/internal/routes/routes.go`)

#### ✅ Added New Endpoints
- `POST /auth/logout` - Logout and clear cookie
- `GET /auth/me` - Get current user info (checks cookie)

#### ✅ Updated CORS Configuration
- Added `DELETE` to allowed methods
- Already had `Access-Control-Allow-Credentials: true`

```go
c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
```

---

## 🎨 Frontend Changes

### 1. **Auth Service** (`frontend/src/app/services/auth.ts`)

#### ✅ Removed localStorage
- Deleted: `getToken()`, `setToken()`, `hasToken()`, `tokenKey`
- Replaced with: Server-side cookie management

#### ✅ Added User State Management
```typescript
private currentUserSubject = new BehaviorSubject<UserInfo | null>(null);
currentUser$ = this.currentUserSubject.asObservable();
```

#### ✅ Updated All Methods
- All HTTP requests now use `{ withCredentials: true }`
- Cookies sent automatically with every request

```typescript
login(data: LoginData): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/login`, data, { 
        withCredentials: true 
    }).pipe(
        tap(response => {
            if (response.user) {
                this.currentUserSubject.next(response.user);
                this.isAuthenticatedSubject.next(true);
            }
        })
    );
}
```

#### ✅ Added Authentication Check
```typescript
checkAuthStatus(): void {
    this.http.get<AuthResponse>(`${this.apiUrl}/me`, { 
        withCredentials: true 
    }).subscribe({
        next: (response) => {
            if (response.authenticated && response.user) {
                this.currentUserSubject.next(response.user);
                this.isAuthenticatedSubject.next(true);
            }
        }
    });
}
```

---

### 2. **HTTP Interceptor** (`frontend/src/app/interceptors/credentials.interceptor.ts`)

#### ✅ Created Credentials Interceptor
- Automatically adds `withCredentials: true` to **all HTTP requests**
- No need to manually add to each request

```typescript
export const credentialsInterceptor: HttpInterceptorFn = (req, next) => {
    const clonedRequest = req.clone({
        withCredentials: true
    });
    return next(clonedRequest);
};
```

---

### 3. **App Config** (`frontend/src/app/app.config.ts`)

#### ✅ Registered Interceptor
```typescript
provideHttpClient(
    withFetch(),
    withInterceptors([credentialsInterceptor])
)
```

---

### 4. **App Component** (`frontend/src/app/app.component.ts` & `.html`)

#### ✅ Updated Logout Method
- Now calls server to clear cookie
- Redirects to `/login` after logout

```typescript
logout(): void {
    this.authService.logout().subscribe(() => {
        window.location.href = '/login';
    });
}
```

#### ✅ Enhanced Header UI
- **Signed Out**: Shows "Welcome, Guest!" + Login/Register buttons
- **Signed In**: Shows "Welcome, [Username]!" + Logout button
- Username displayed dynamically from currentUser$

---

## 🧪 Testing

### Test Authentication Flow

1. **Register New User**:
   ```bash
   POST http://localhost:8080/auth/register
   {
       "username": "testuser",
       "email": "test@example.com",
       "password": "Password123!"
   }
   ```

2. **Verify Email**:
   - Check email for verification link
   - Click link or visit: `/verify-email?token=xxx`

3. **Login**:
   ```bash
   POST http://localhost:8080/auth/login
   {
       "username": "testuser",
       "password": "Password123!"
   }
   ```
   - Cookie `auth_token` is set automatically
   - User info returned in response body

4. **Check Auth Status**:
   ```bash
   GET http://localhost:8080/auth/me
   ```
   - Cookie sent automatically
   - Returns user info if authenticated

5. **Access Protected Route**:
   ```bash
   GET http://localhost:8080/challenges
   ```
   - Cookie sent automatically
   - No need to add Authorization header

6. **Logout**:
   ```bash
   POST http://localhost:8080/auth/logout
   ```
   - Cookie cleared
   - User redirected to login

---

## 🔍 Browser DevTools Verification

### Check Cookies
1. Open DevTools → Application → Cookies
2. Look for `auth_token` cookie:
   - ✅ Name: `auth_token`
   - ✅ Value: JWT token
   - ✅ Path: `/`
   - ✅ HttpOnly: ✓ (checked)
   - ✅ Secure: (should be checked in production)
   - ✅ SameSite: `Lax` or `Strict` (recommended)

### Check Network Requests
1. Open DevTools → Network
2. Look at any API request
3. Verify:
   - ✅ Request Headers contain: `Cookie: auth_token=...`
   - ✅ No `Authorization` header needed
   - ✅ Response headers include: `Access-Control-Allow-Credentials: true`

---

## 🚀 Deployment Considerations

### Production Settings

1. **Enable HTTPS**:
   ```go
   c.SetCookie(
       "auth_token",
       token,
       7*24*60*60,
       "/",
       "",
       true,  // ← Set to true for HTTPS
       true,
   )
   ```

2. **Add SameSite Attribute**:
   ```go
   c.SetSameSite(http.SameSiteStrictMode)
   c.SetCookie(...)
   ```

3. **Update CORS Origin**:
   ```go
   // In routes.go, replace wildcard with specific domain
   c.Writer.Header().Set("Access-Control-Allow-Origin", "https://yourdomain.com")
   ```

4. **Add CSRF Protection** (Optional but recommended):
   - Implement CSRF tokens for state-changing operations
   - Use libraries like `gorilla/csrf`

---

## ✅ Advantages of Cookie-Based Auth

1. **🔒 XSS Protection**: Cookies with `httpOnly` flag cannot be accessed by JavaScript
2. **🚀 Automatic**: Cookies sent automatically with every request
3. **📱 Better Mobile Support**: Works better with mobile apps and PWAs
4. **🔐 Server Control**: Server can invalidate cookies anytime
5. **⏰ Built-in Expiry**: Browsers handle cookie expiration automatically

---

## ⚠️ Important Notes

- **Backward Compatibility**: Backend still supports `Authorization` header
- **CSRF Tokens**: Consider adding for production (especially for sensitive operations)
- **Cookie Domains**: Configure properly for subdomains if needed
- **HTTPS Only**: Always use `Secure` flag in production
- **SameSite**: Consider using `SameSite=Strict` for better security

---

## 📚 Resources

- [OWASP JWT Security](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
- [MDN: HTTP Cookies](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies)
- [Angular HTTP Client with Credentials](https://angular.io/guide/http-send-data-to-server)

---

## 🎯 Summary

✅ **Secure**: HTTP-only cookies prevent XSS attacks  
✅ **Simple**: Automatic cookie handling, no manual token management  
✅ **Scalable**: Works with any frontend framework or mobile app  
✅ **Modern**: Industry best practice for web authentication  
✅ **User-Friendly**: Shows username in header when signed in  

**Migration Complete!** 🎉

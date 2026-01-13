import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, BehaviorSubject } from 'rxjs';
import { tap } from 'rxjs/operators';
import { environment } from '../../environments/environment';

export interface RegisterData {
  username: string;
  email: string;
  password: string;
}

export interface LoginData {
  username: string;
  password: string;
}

export interface UserInfo {
  id: string;
  username: string;
  email: string;
  role: string;
}

export interface AuthResponse {
  message: string;
  user?: UserInfo;
  authenticated?: boolean;
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private apiUrl = `${environment.apiUrl}/auth`;
  private isAuthenticatedSubject = new BehaviorSubject<boolean>(false);
  private currentUserSubject = new BehaviorSubject<UserInfo | null>(null);
  private authCheckCompleteSubject = new BehaviorSubject<boolean>(false);

  isAuthenticated$ = this.isAuthenticatedSubject.asObservable();
  currentUser$ = this.currentUserSubject.asObservable();
  authCheckComplete$ = this.authCheckCompleteSubject.asObservable();

  constructor(private http: HttpClient) {
    // Check authentication status on init
    this.checkAuthStatus();
  }

  register(data: RegisterData): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/register`, data, { withCredentials: true });
  }

  login(data: LoginData): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/login`, data, { withCredentials: true }).pipe(
      tap(response => {
        if (response.user) {
          this.currentUserSubject.next(response.user);
          this.isAuthenticatedSubject.next(true);
        }
      })
    );
  }

  verifyEmail(token: string): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/verify-email`, { token }, { withCredentials: true });
  }

  resendVerification(email: string): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/resend-verification`, { email }, { withCredentials: true });
  }

  forgotPassword(email: string): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/forgot-password`, { email }, { withCredentials: true });
  }

  resetPassword(token: string, newPassword: string): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/reset-password`, {
      token,
      new_password: newPassword
    }, { withCredentials: true });
  }

  changePassword(oldPassword: string, newPassword: string): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/change-password`, {
      old_password: oldPassword,
      new_password: newPassword
    }, { withCredentials: true });
  }

  logout(): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/logout`, {}, { withCredentials: true }).pipe(
      tap(() => {
        this.currentUserSubject.next(null);
        this.isAuthenticatedSubject.next(false);
      })
    );
  }

  isLoggedIn(): boolean {
    return this.isAuthenticatedSubject.value;
  }

  getCurrentUser(): UserInfo | null {
    return this.currentUserSubject.value;
  }

  getUserInfo(): UserInfo | null {
    return this.currentUserSubject.value;
  }

  isAdmin(): boolean {
    const user = this.getCurrentUser();
    return user !== null && user.role === 'admin';
  }

  getUserRole(): string | null {
    const user = this.getCurrentUser();
    return user ? user.role : null;
  }

  checkAuthStatus(): void {
    this.http.get<AuthResponse>(`${this.apiUrl}/me`, { withCredentials: true }).subscribe({
      next: (response) => {
        if (response.authenticated && response.user) {
          this.currentUserSubject.next(response.user);
          this.isAuthenticatedSubject.next(true);
        } else {
          this.currentUserSubject.next(null);
          this.isAuthenticatedSubject.next(false);
        }
        this.authCheckCompleteSubject.next(true);
      },
      error: () => {
        this.currentUserSubject.next(null);
        this.isAuthenticatedSubject.next(false);
        this.authCheckCompleteSubject.next(true);
      }
    });
  }

  isAuthCheckComplete(): boolean {
    return this.authCheckCompleteSubject.value;
  }
}

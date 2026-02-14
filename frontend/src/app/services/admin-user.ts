import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface IPRecord {
  ip: string;
  timestamp: string;
  action?: string;
}

export interface AdminUser {
  id: string;
  username: string;
  email: string;
  role: string;
  status: string;
  email_verified: boolean;
  last_ip?: string;
  ip_history?: IPRecord[];
  last_login_at?: string;
  team_id?: string;
  team_name?: string;
  ban_reason?: string;
  oauth_provider?: string;
  created_at: string;
  updated_at: string;
}

@Injectable({
  providedIn: 'root'
})
export class AdminUserService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) { }

  listUsers(): Observable<AdminUser[]> {
    return this.http.get<AdminUser[]>(`${this.apiUrl}/admin/users`);
  }

  getUser(id: string): Observable<AdminUser> {
    return this.http.get<AdminUser>(`${this.apiUrl}/admin/users/${id}`);
  }

  updateUserStatus(id: string, status: string, banReason?: string): Observable<any> {
    return this.http.put<any>(`${this.apiUrl}/admin/users/${id}/status`, { status, ban_reason: banReason || '' });
  }

  updateUserRole(id: string, role: string): Observable<any> {
    return this.http.put<any>(`${this.apiUrl}/admin/users/${id}/role`, { role });
  }

  deleteUser(id: string): Observable<any> {
    return this.http.delete<any>(`${this.apiUrl}/admin/users/${id}`);
  }
}

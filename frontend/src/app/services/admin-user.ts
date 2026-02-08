import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface AdminUser {
  id: string;
  username: string;
  email: string;
  role: string;
  status: string;
  email_verified: boolean;
  created_at: string;
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

  updateUserStatus(id: string, status: string, banReason?: string): Observable<any> {
    return this.http.put<any>(`${this.apiUrl}/admin/users/${id}/status`, { status, ban_reason: banReason || '' });
  }

  updateUserRole(id: string, role: string): Observable<any> {
    return this.http.put<any>(`${this.apiUrl}/admin/users/${id}/role`, { role });
  }
}

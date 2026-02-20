import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface TeamMemberInfo {
  id: string;
  username: string;
  email: string;
  is_leader: boolean;
}

export interface AdminTeam {
  id: string;
  name: string;
  description: string;
  avatar?: string;
  leader_id: string;
  leader_name: string;
  member_count: number;
  members: TeamMemberInfo[];
  invite_code: string;
  score: number;
  created_at: string;
  updated_at: string;
}

@Injectable({
  providedIn: 'root'
})
export class AdminTeamService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) { }

  listTeams(): Observable<AdminTeam[]> {
    return this.http.get<AdminTeam[]>(`${this.apiUrl}/admin/teams`);
  }

  getTeam(id: string): Observable<AdminTeam> {
    return this.http.get<AdminTeam>(`${this.apiUrl}/admin/teams/${id}`);
  }

  updateTeam(id: string, name: string, description: string): Observable<any> {
    return this.http.put<any>(`${this.apiUrl}/admin/teams/${id}`, { name, description });
  }

  updateTeamLeader(id: string, newLeaderId: string): Observable<any> {
    return this.http.put<any>(`${this.apiUrl}/admin/teams/${id}/leader`, { new_leader_id: newLeaderId });
  }

  removeMember(teamId: string, memberId: string): Observable<any> {
    return this.http.delete<any>(`${this.apiUrl}/admin/teams/${teamId}/members/${memberId}`);
  }

  deleteTeam(id: string): Observable<any> {
    return this.http.delete<any>(`${this.apiUrl}/admin/teams/${id}`);
  }

  adjustScore(id: string, delta: number, reason: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/admin/teams/${id}/score-adjust`, {
      delta,
      reason
    });
  }
}

import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface Contest {
  id: string;
  name: string;
  description: string;
  start_time: string;
  end_time: string;
  is_active: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface ContestConfig {
  contest_id?: string;
  title: string;
  start_time: string;
  end_time: string;
  freeze_time?: string;
  is_active: boolean;
  is_paused?: boolean;
  scoreboard_visibility?: string;
}

@Injectable({
  providedIn: 'root'
})
export class ContestService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) {}

  getUpcomingContests(): Observable<Contest[]> {
    return this.http.get<Contest[]>(`${this.apiUrl}/contests/upcoming`);
  }

  registerTeamForContest(contestId: string): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(`${this.apiUrl}/contests/${contestId}/register`, {});
  }

  unregisterTeamFromContest(contestId: string): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(`${this.apiUrl}/contests/${contestId}/unregister`, {});
  }

  getRegistrationStatus(contestId: string): Observable<{ registered: boolean }> {
    return this.http.get<{ registered: boolean }>(`${this.apiUrl}/contests/${contestId}/registration-status`);
  }

  getRegisteredTeamsCount(contestId: string): Observable<{ count: number }> {
    return this.http.get<{ count: number }>(`${this.apiUrl}/contests/${contestId}/registered-count`);
  }

  getContestConfig(): Observable<{ config: ContestConfig }> {
    return this.http.get<{ config: ContestConfig }>(`${this.apiUrl}/admin/contest`);
  }

  updateContestConfig(title: string, startTime: string, endTime: string, isActive: boolean): Observable<{ message: string }> {
    return this.http.put<{ message: string }>(`${this.apiUrl}/admin/contest`, {
      title,
      start_time: startTime,
      end_time: endTime,
      is_active: isActive
    });
  }
}

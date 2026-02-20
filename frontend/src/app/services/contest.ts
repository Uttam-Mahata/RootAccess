import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { tap } from 'rxjs/operators';
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

export interface ScoreboardContest {
  id: string;
  name: string;
  description: string;
  start_time: string;
  end_time: string;
  status: 'running' | 'ended';
  scoreboard_visibility: string;
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

const STORAGE_KEY = 'ra_contest_reg';

@Injectable({
  providedIn: 'root'
})
export class ContestService {
  private apiUrl = environment.apiUrl;

  // Persists across navigations within a session AND across hard refreshes via localStorage
  readonly registrationStatusCache = new Map<string, boolean>();

  constructor(private http: HttpClient) {
    this.loadCache();
  }

  private loadCache(): void {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (raw) {
        const stored = JSON.parse(raw) as Record<string, boolean>;
        Object.entries(stored).forEach(([k, v]) => this.registrationStatusCache.set(k, v));
      }
    } catch { /* ignore parse errors */ }
  }

  private saveCache(): void {
    try {
      const obj: Record<string, boolean> = {};
      this.registrationStatusCache.forEach((v, k) => (obj[k] = v));
      localStorage.setItem(STORAGE_KEY, JSON.stringify(obj));
    } catch { /* ignore storage errors */ }
  }

  clearRegistrationCache(): void {
    this.registrationStatusCache.clear();
    try { localStorage.removeItem(STORAGE_KEY); } catch { /* ignore */ }
  }

  getUpcomingContests(): Observable<Contest[]> {
    return this.http.get<Contest[]>(`${this.apiUrl}/contests/upcoming`);
  }

  getScoreboardContests(): Observable<{ contests: ScoreboardContest[] }> {
    return this.http.get<{ contests: ScoreboardContest[] }>(`${this.apiUrl}/contests/active`);
  }

  registerTeamForContest(contestId: string): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(`${this.apiUrl}/contests/${contestId}/register`, {}).pipe(
      tap(() => { this.registrationStatusCache.set(contestId, true); this.saveCache(); })
    );
  }

  unregisterTeamFromContest(contestId: string): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(`${this.apiUrl}/contests/${contestId}/unregister`, {}).pipe(
      tap(() => { this.registrationStatusCache.set(contestId, false); this.saveCache(); })
    );
  }

  getRegistrationStatus(contestId: string): Observable<{ registered: boolean }> {
    return this.http.get<{ registered: boolean }>(`${this.apiUrl}/contests/${contestId}/registration-status`).pipe(
      tap(resp => { this.registrationStatusCache.set(contestId, resp.registered); this.saveCache(); })
    );
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

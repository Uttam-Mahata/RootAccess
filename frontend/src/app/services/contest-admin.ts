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

export interface ContestRound {
  id: string;
  contest_id: string;
  name: string;
  description: string;
  order: number;
  visible_from: string;
  start_time: string;
  end_time: string;
  created_at?: string;
  updated_at?: string;
}

@Injectable({
  providedIn: 'root'
})
export class ContestAdminService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) {}

  listContests(): Observable<Contest[]> {
    return this.http.get<Contest[]>(`${this.apiUrl}/admin/contest-entities`);
  }

  getContest(id: string): Observable<Contest> {
    return this.http.get<Contest>(`${this.apiUrl}/admin/contest-entities/${id}`);
  }

  createContest(name: string, description: string, startTime: string, endTime: string): Observable<Contest> {
    return this.http.post<Contest>(`${this.apiUrl}/admin/contest-entities`, {
      name,
      description,
      start_time: startTime,
      end_time: endTime
    });
  }

  updateContest(id: string, name: string, description: string, startTime: string, endTime: string, isActive: boolean): Observable<Contest> {
    return this.http.put<Contest>(`${this.apiUrl}/admin/contest-entities/${id}`, {
      name,
      description,
      start_time: startTime,
      end_time: endTime,
      is_active: isActive
    });
  }

  deleteContest(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/admin/contest-entities/${id}`);
  }

  setActiveContest(contestId: string): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(`${this.apiUrl}/admin/contest-entities/set-active`, { contest_id: contestId });
  }

  listRounds(contestId: string): Observable<ContestRound[]> {
    return this.http.get<ContestRound[]>(`${this.apiUrl}/admin/contest-entities/${contestId}/rounds`);
  }

  createRound(contestId: string, name: string, description: string, order: number, visibleFrom: string, startTime: string, endTime: string): Observable<ContestRound> {
    return this.http.post<ContestRound>(`${this.apiUrl}/admin/contest-entities/${contestId}/rounds`, {
      name,
      description,
      order,
      visible_from: visibleFrom,
      start_time: startTime,
      end_time: endTime
    });
  }

  updateRound(contestId: string, roundId: string, name: string, description: string, order: number, visibleFrom: string, startTime: string, endTime: string): Observable<ContestRound> {
    return this.http.put<ContestRound>(`${this.apiUrl}/admin/contest-entities/${contestId}/rounds/${roundId}`, {
      name,
      description,
      order,
      visible_from: visibleFrom,
      start_time: startTime,
      end_time: endTime
    });
  }

  deleteRound(contestId: string, roundId: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/admin/contest-entities/${contestId}/rounds/${roundId}`);
  }

  getRoundChallenges(contestId: string, roundId: string): Observable<string[]> {
    return this.http.get<string[]>(`${this.apiUrl}/admin/contest-entities/${contestId}/rounds/${roundId}/challenges`);
  }

  attachChallenges(contestId: string, roundId: string, challengeIds: string[]): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/admin/contest-entities/${contestId}/rounds/${roundId}/challenges`, {
      challenge_ids: challengeIds
    });
  }

  detachChallenges(contestId: string, roundId: string, challengeIds: string[]): Observable<void> {
    return this.http.request<void>('delete', `${this.apiUrl}/admin/contest-entities/${contestId}/rounds/${roundId}/challenges`, {
      body: { challenge_ids: challengeIds }
    });
  }
}

import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface ChallengePopularity {
  challenge_id: string;
  title: string;
  category: string;
  solve_count: number;
  attempt_count: number;
  success_rate: number;
}

export interface TimeSeriesEntry {
  date: string;
  count: number;
}

export interface AdminAnalytics {
  total_users: number;
  total_teams: number;
  total_challenges: number;
  total_submissions: number;
  total_correct: number;
  success_rate: number;
  challenge_popularity: ChallengePopularity[];
  category_breakdown: { [key: string]: number };
  difficulty_breakdown: { [key: string]: number };
  solves_over_time: TimeSeriesEntry[];
}

@Injectable({
  providedIn: 'root'
})
export class AnalyticsService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) { }

  getPlatformAnalytics(): Observable<AdminAnalytics> {
    return this.http.get<AdminAnalytics>(`${this.apiUrl}/admin/analytics`);
  }
}

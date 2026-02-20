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

export interface TeamStats {
  team_id: string;
  name: string;
  score: number;
  member_count: number;
  solve_count?: number;
}

export interface UserStats {
  user_id: string;
  username: string;
  score: number;
  solve_count: number;
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
  submissions_over_time?: TimeSeriesEntry[];
  // Enhanced statistics
  active_users: number;
  banned_users: number;
  verified_users: number;
  admin_count: number;
  new_users_today: number;
  new_users_this_week: number;
  new_teams_today: number;
  new_teams_this_week: number;
  submissions_today: number;
  solves_today: number;
  average_team_size: number;
  user_growth: TimeSeriesEntry[];
  team_growth: TimeSeriesEntry[];
  top_teams: TeamStats[];
  top_users: UserStats[];
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

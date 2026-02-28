import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface CategoryStat {
  total: number;
  solved: number;
  points: number;
}

export interface SolveEntry {
  challenge_id: string;
  challenge_title: string;
  category: string;
  points: number;
  solved_at: string;
}

export interface Achievement {
  id: string;
  user_id: string;
  type: string;
  name: string;
  description: string;
  icon: string;
  earned_at: string;
}

export interface UserActivity {
  user_id: string;
  username: string;
  total_solves: number;
  total_points: number;
  category_progress: { [key: string]: CategoryStat };
  recent_solves: SolveEntry[];
  achievements: Achievement[];
  rank: number;
  solve_streak: number;
}

@Injectable({
  providedIn: 'root'
})
export class ActivityService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) { }

  getMyActivity(): Observable<UserActivity> {
    return this.http.get<UserActivity>(`${this.apiUrl}/activity/me`);
  }

  getMyAchievements(): Observable<Achievement[]> {
    return this.http.get<Achievement[]>(`${this.apiUrl}/achievements/me`);
  }
}

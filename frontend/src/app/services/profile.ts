import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface SolvedChallenge {
  id: string;
  title: string;
  category: string;
  difficulty: string;
  points: number;
  solved_at: string;
}

export interface CategoryStats {
  category: string;
  solve_count: number;
  total_points: number;
}

export interface UserProfile {
  username: string;
  joined_at: string;
  team_id?: string;
  team_name?: string;
  total_points: number;
  solve_count: number;
  total_submissions: number;
  solved_challenges: SolvedChallenge[];
  category_stats: CategoryStats[];
}

@Injectable({
  providedIn: 'root'
})
export class ProfileService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) {}

  getUserProfile(username: string): Observable<UserProfile> {
    return this.http.get<UserProfile>(`${this.apiUrl}/users/${username}/profile`);
  }
}

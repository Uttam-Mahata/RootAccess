import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

// Challenge interface for public view
export interface Challenge {
  id: string;
  title: string;
  description: string;
  category: string;
  difficulty: string;
  max_points: number;
  current_points: number;
  solve_count: number;
  files: string[];
}

// Challenge interface for admin view
export interface ChallengeAdmin {
  id: string;
  title: string;
  description: string;
  category: string;
  difficulty: string;
  max_points: number;
  min_points: number;
  decay: number;
  solve_count: number;
  current_points: number;
  files: string[];
}

// Request interface for creating/updating challenges
export interface ChallengeRequest {
  title: string;
  description: string;
  category: string;
  difficulty: string;
  max_points: number;
  min_points: number;
  decay: number;
  flag: string;
  files: string[];
}

// Response interface for flag submission
export interface SubmitFlagResponse {
  correct: boolean;
  already_solved: boolean;
  message: string;
  points?: number;
  solve_count?: number;
  team_name?: string;
}

@Injectable({
  providedIn: 'root'
})
export class ChallengeService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) { }

  // Public methods
  getChallenges(): Observable<Challenge[]> {
    return this.http.get<Challenge[]>(`${this.apiUrl}/challenges`);
  }

  getChallenge(id: string): Observable<Challenge> {
    return this.http.get<Challenge>(`${this.apiUrl}/challenges/${id}`);
  }

  submitFlag(id: string, flag: string): Observable<SubmitFlagResponse> {
    return this.http.post<SubmitFlagResponse>(`${this.apiUrl}/challenges/${id}/submit`, { flag });
  }

  // Admin methods
  getChallengesForAdmin(): Observable<ChallengeAdmin[]> {
    return this.http.get<ChallengeAdmin[]>(`${this.apiUrl}/admin/challenges`);
  }

  createChallenge(challenge: ChallengeRequest): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/admin/challenges`, challenge);
  }

  updateChallenge(id: string, challenge: ChallengeRequest): Observable<any> {
    return this.http.put<any>(`${this.apiUrl}/admin/challenges/${id}`, challenge);
  }

  deleteChallenge(id: string): Observable<any> {
    return this.http.delete<any>(`${this.apiUrl}/admin/challenges/${id}`);
  }
}

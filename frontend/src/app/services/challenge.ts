import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

// Challenge interface for public view
export interface Challenge {
  id: string;
  title: string;
  description: string;
  description_format?: string; // "markdown" or "html"
  category: string;
  difficulty: string;
  max_points: number;
  current_points: number;
  scoring_type: string;
  solve_count: number;
  files: string[];
  tags: string[];
  hint_count: number;
}

// Challenge interface for admin view
export interface ChallengeAdmin {
  id: string;
  title: string;
  description: string;
  description_format?: string; // "markdown" or "html"
  category: string;
  difficulty: string;
  max_points: number;
  min_points: number;
  decay: number;
  scoring_type: string;
  solve_count: number;
  current_points: number;
  files: string[];
  tags: string[];
  hint_count: number;
}

// Hint interfaces
export interface HintResponse {
  id: string;
  cost: number;
  order: number;
  content?: string;
  revealed: boolean;
}

export interface HintRequest {
  content: string;
  cost: number;
  order: number;
}

// Request interface for creating/updating challenges
export interface ChallengeRequest {
  title: string;
  description: string;
  description_format?: string; // "markdown" or "html"
  category: string;
  difficulty: string;
  max_points: number;
  min_points: number;
  decay: number;
  scoring_type: string;
  flag: string;
  files: string[];
  tags: string[];
  hints: HintRequest[];
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

  getChallengeSolves(id: string): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/challenges/${id}/solves`);
  }

  submitFlag(id: string, flag: string): Observable<SubmitFlagResponse> {
    return this.http.post<SubmitFlagResponse>(`${this.apiUrl}/challenges/${id}/submit`, { flag });
  }

  // Admin methods
  getChallengesForAdmin(listOnly = false): Observable<ChallengeAdmin[]> {
    const url = listOnly ? `${this.apiUrl}/admin/challenges?list=1` : `${this.apiUrl}/admin/challenges`;
    return this.http.get<ChallengeAdmin[]>(url);
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

  // Hint methods
  getHints(challengeId: string): Observable<HintResponse[]> {
    return this.http.get<HintResponse[]>(`${this.apiUrl}/challenges/${challengeId}/hints`);
  }

  revealHint(challengeId: string, hintId: string): Observable<HintResponse> {
    return this.http.post<HintResponse>(`${this.apiUrl}/challenges/${challengeId}/hints/${hintId}/reveal`, {});
  }

  // Writeup methods
  submitWriteup(challengeId: string, content: string, contentFormat: string = 'markdown'): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/challenges/${challengeId}/writeups`, { 
      content,
      content_format: contentFormat
    });
  }

  getWriteups(challengeId: string): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/challenges/${challengeId}/writeups`);
  }

  getMyWriteups(): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/writeups/my`);
  }

  updateWriteup(writeupId: string, content: string): Observable<any> {
    return this.http.put<any>(`${this.apiUrl}/writeups/${writeupId}`, { content });
  }

  toggleWriteupUpvote(writeupId: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/writeups/${writeupId}/upvote`, {});
  }
}

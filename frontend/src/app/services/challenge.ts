import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { AuthService } from './auth';

@Injectable({
  providedIn: 'root'
})
export class ChallengeService {
  private apiUrl = 'http://localhost:8080';

  constructor(private http: HttpClient, private authService: AuthService) { }

  private getHeaders(): HttpHeaders {
    const token = this.authService.getToken();
    return new HttpHeaders().set('Authorization', `Bearer ${token}`);
  }

  getChallenges(): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/challenges`, { headers: this.getHeaders() });
  }

  getChallenge(id: string): Observable<any> {
    return this.http.get<any>(`${this.apiUrl}/challenges/${id}`, { headers: this.getHeaders() });
  }

  submitFlag(id: string, flag: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/challenges/${id}/submit`, { flag }, { headers: this.getHeaders() });
  }

  createChallenge(challenge: any): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/challenges`, challenge, { headers: this.getHeaders() });
  }
}

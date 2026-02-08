import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class BulkChallengeService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) { }

  importChallenges(challenges: any[]): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/admin/challenges/import`, challenges);
  }

  exportChallenges(): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/admin/challenges/export`);
  }

  duplicateChallenge(id: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/admin/challenges/${id}/duplicate`, {});
  }
}
